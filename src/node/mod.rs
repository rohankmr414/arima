use raft::{Config, raw_node::RawNode};
use std::sync::Arc;
use rocksdb::{DB, Options, WriteBatch,  WriteOptions};

use crate::storage::*;
use slog::{Logger, Drain, o};

struct ArimaNode {
    config: Config,
    log: Logger,
    raw_node: RawNode<ArimaStorage>,
}

impl ArimaNode {
    pub fn new(id: u64) -> ArimaNode {
        let path = format!("/tmp/raft-{}", id);
        let log = slog::Logger::root(slog::Discard, o!());
        let db = Arc::new(DB::open_default(path).unwrap());
        let storage:ArimaStorage = ArimaStorage::new(id,db.clone());
        let config = Config {
            id,
            ..Default::default()
        };
        let mut raw_node = RawNode::new(&config, storage, log).unwrap();
        ArimaNode {
            config,
            log,
            raw_node,
        }
    }
}
