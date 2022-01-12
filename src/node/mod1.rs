use std::time::{Duration, Instant};
use std::sync::Arc;
use std::collections::HashMap;

use crossbeam_channel::{Receiver, RecvTimeoutError};
use raft::eraftpb::{EntryType, Message as RaftMessage};
use raft::{self, Config, RawNode, SnapshotStatus, Storage as RaftStorage};
use raft::ProgressState::Snapshot;
use rocket::futures::AsyncReadExt;
use rocksdb::{Options, WriteBatch, WriteOptions, DB};
use serde::{Deserialize, Serialize};
use serde_json;
use slog_stdlog::*;
use slog::{Logger, Drain, o};
use crate::node::Msg::Propose;

use crate::storage::{ArimaStorage};
use crate::storage::util::*;
use crate::storage::keys::*;

// op: read 1, write 2, delete 3.
// op: status 128
#[derive(Serialize, Deserialize, Default)]
pub struct Request {
    pub id: u64,
    pub op: u32,
    pub row: Row,
}

#[derive(Serialize, Deserialize, Default)]
pub struct Response {
    pub id: u64,
    pub ok: bool,
    pub op: u32,
    pub value: Option<Vec<u8>>,
}

#[derive(Serialize, Deserialize, Default)]
pub struct Status {
    pub leader_id: u64,
    pub id: u64,
    pub first_index: u64,
    pub last_index: u64,
    pub term: u64,
    pub apply_index: u64,
    pub commit_index: u64,
}

pub type RequestCallback = Box<dyn Fn(Response) + Send>;

pub enum Msg {
    Propose {
        request: Request,
        cb: RequestCallback,
    },
    Raft(RaftMessage),
    ReportUnreachable(u64),
    ReportSnapshot{
        id: u64,
        status: SnapshotStatus,
    },
}

struct ArimaNode {
    id: u64,
    tag: String,
    r: RawNode<ArimaStorage>,
    cbs: HashMap<u64, RequestCallback>,
    db: Arc<DB>,
}

impl ArimaNode {
    pub fn new(id: u64, db: Arc<DB>) -> ArimaNode {
        let storage = ArimaStorage::new(id, db.clone());
        let tag = format!("node-{}", id);
        let logger = slog::Logger::root(slog_stdlog::StdLog.fuse(), o!());
        let cfg = Config {
            id,
            election_tick: 10,
            heartbeat_tick: 3,
            max_size_per_msg: 1024 * 1024 * 1024,
            max_inflight_msgs: 256,
            applied: storage.apply_index,
            ..Default::default()
        };

        cfg.validate().unwrap();

        let r = RawNode::new(&cfg, storage, &logger).unwrap();

        ArimaNode {
            id,
            tag,
            r,
            cbs: HashMap::new(),
            db,
        }
    }

    fn handle_status(&self, req: Request, cb: RequestCallback) {
        let raft_status = self.r.status();
        let s = Status {
            leader_id: self.r.raft.leader_id,
            id: self.id,
            first_index: self.r.store().first_index().unwrap(),
            last_index: self.r.store().last_index().unwrap(),
            term: raft_status.hs.get_term(),
            apply_index: self.r.store().apply_index,
            commit_index: raft_status.hs.get_commit(),
        };

        cb(Response {
            id: req.id,
            ok: false,
            op: req.op,
            value: Some(serde_json::to_vec(&s).unwrap()),
        });
    }

    pub fn on_msg(&mut self, msg: Msg) {
        match msg {
            Msg::Raft(m) => self.r.step(m).unwrap(),
            Msg::Propose { request, cb } => {
                if request.op == 128 {
                    self.handle_status(request, cb);
                    return;
                }
                if self.r.raft.leader_id != self.id || self.cbs.contains_key(&request.id) {
                    cb(Response {
                        id: request.id,
                        ok: false,
                        op: request.op,
                        ..Default::default()
                    });
                    return;
                }

                let data = serde_json::to_vec(&request).unwrap();
                self.r.propose(vec![], data).unwrap();
                self.cbs.insert(request.id, cb);
            }
            Msg::ReportUnreachable(id) => {
                self.r.report_unreachable(id);
            }
            Msg::ReportSnapshot { id, status } => {
                self.r.report_snapshot(id, status);
            }
        }
    }

    pub fn on_tick(&mut self) {
        if !self.r.has_ready() {
            return;
        }
        let is_leader = self.r.raft.leader_id == self.id;
        let mut ready = self.r.ready();

        if is_leader {
            let messages = ready.messages();
            for msg in messages {
                todo!()
            }
        }

        if !raft::is_empty_snap(&ready.snapshot()) {
            dbg!("{} begin to append {} entries", self.tag, ready.entries().len());
            self.r.mut_store().apply_snapshot(&ready.snapshot());
        }

        if !ready.entries().is_empty() {
            dbg!("{} begin to append {} entries", self.tag, ready.entries().len());
            self.r.mut_store().append(&ready.entries());
        }

        // hard state changed persist it
        if let Some(ref hs) = ready.hs() {
            put_msg(&self.db, RAFT_HARD_STATE_KEY, hs);
        }

        {
            let msgs = ready.messages();
            for msg in msgs {
                todo!()
            }
        }

        if let Some(committed_entries) = ready.committed_entries().take() {
            if !committed_entries.is_empty() {
                dbg!("{} begin to apply {} committed entries", self.tag, committed_entries.len());
            }
            let mut last_applying_index = 0;

            for entry in committed_entries {
                last_applying_index = entry.get_index();
                if entry.get_data().is_empty() {
                    continue;
                }

                if entry.get_entry_type() == EntryType::EntryNormal {
                    let req: Request = serde_json::from_slice(entry.get_data()).unwrap();
                    self.on_request(req);
                }
            }

            if last_applying_index > 0 {
                self.r.mut_store().apply_index = last_applying_index;
                put_u64(&self.db, RAFT_APPLY_INDEX_KEY, last_applying_index);
            }
        }

        self.r.advance(ready);
    }

    fn on_request(&mut self, req: Request) {
        let mut resp = Response {
            id: req.id,
            op: req.op,
            ok: true,
            value: None,
        };

        dbg!("{} handle command {}", self.tag, req.op);

        match req.op {
            1 => {
                if let Some(v) = self.db.get(&req.row.key).unwrap() {
                    resp.value = Some(v.to_vec());
                }
            }
            2 => {
                self.db.put(&req.row.key, &req.row.value).unwrap();
            }
            3 => {
                self.db.delete(&req.row.key).unwrap();
            }
            _ => unreachable!(),
        }

        if let Some(cb) = self.cbs.remove(&req.id) {
            cb(resp);
        }
    }
}

pub fn run_node(mut node: ArimaNode, ch: Receiver<Msg>) {
    let mut t = Instant::now();
    let d = Duration::from_millis(100);
    loop {
        for _ in 0..4096 {
            match ch.recv_timeout(d) {
                Ok(msg) => node.on_msg(msg),
                Err(RecvTimeoutError::Timeout) => break,
                Err(RecvTimeoutError::Disconnected) => return,
            }
        }

        if t.elapsed() >= d {
            t = Instant::now();
            node.r.tick();
        }
        node.on_tick();
    }
}
