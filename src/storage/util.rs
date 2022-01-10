use std::sync::Arc;
use std::collections::HashMap;

use rocksdb::{DBIterator, ReadOptions, DB, Options, IteratorMode, Direction, WriteBatch};
use protobuf::{Message};
use byteorder::{BigEndian, ByteOrder};
use rocksdb::IteratorMode::Start;
use serde::{Serialize, Deserialize};
use serde_json;

pub const RAFT_TRUNCATED_STATE_KEY: &[u8] = &[0x03];

#[derive(Serialize, Deserialize, Default)]
pub struct Row {
    pub key: Vec<u8>,
    pub value: Vec<u8>,
}

#[derive(Serialize, Deserialize, Default)]
pub struct TruncatedState {
    pub index: u64,
    pub term: u64,
}

pub fn put_u64(db: &DB, key: &[u8], id: u64) {
    let mut buf = [0u8; 8];
    BigEndian::write_u64(&mut buf, id);
    db.put(key, &buf).unwrap();
}

pub fn put_u64_to_wb(wb: &mut WriteBatch, key: &[u8], id: u64) {
    let mut buf = [0u8; 8];
    BigEndian::write_u64(&mut buf, id);
    wb.put(key, &buf);
}

pub fn get_u64(db: &DB, key: &[u8]) -> Option<u64> {
    let value = db.get(key).unwrap();

    if value.is_none() {
        return None;
    }

    let value = value.unwrap();
    if value.len() != 8 {
        panic!("need 8 bytes but got {}", value.len());
    }

    let n = BigEndian::read_u64(&value);
    Some(n)
}

pub fn get_msg<T: Message>(db: &DB, key: &[u8]) -> Option<T> {
    let value = db.get(key).unwrap();

    if value.is_none() {
        return None;
    }

    let mut msg = T::new();
    msg.merge_from_bytes(&value.unwrap()).unwrap();
    Some(msg)
}

pub fn put_msg<T: Message>(db: &DB, key: &[u8], msg: &T) {
    let value = msg.write_to_bytes().unwrap();
    db.put(key, &value).unwrap();
}

pub fn put_msg_to_wb<T: Message>(wb: &mut WriteBatch, key: &[u8], msg: &T) {
    let value = msg.write_to_bytes().unwrap();
    wb.put(key, &value);
}

pub fn scan<F>(db: &DB, start: &[u8], end: &[u8], f: &mut F) where F: FnMut(&[u8], &[u8]) -> bool {
    let mut  read_opts = ReadOptions::default();
    read_opts.set_iterate_lower_bound(start);
    read_opts.set_iterate_upper_bound(end);
    let mut iter = db.raw_iterator_opt(read_opts);
    iter.seek(start);
    while iter.valid() {
        let r = f(iter.key().unwrap(), iter.value().unwrap());

        if !r {
            break;
        }
    }
}

pub fn seek(db: &DB, start_key: &[u8]) -> Option<(Vec<u8>, Vec<u8>)> {
    let mut opts = ReadOptions::default();
    opts.set_iterate_lower_bound(start_key);
    let mut iter = db.raw_iterator_opt(opts);
    iter.seek(start_key);
    if iter.valid() {
        Some((iter.key().unwrap().to_vec(), iter.value().unwrap().to_vec()))
    } else {
        None
    }
}

pub fn seek_for_prev(db: &DB, end_key: &[u8]) -> Option<(Vec<u8>, Vec<u8>)> {
    let mut opts = ReadOptions::default();
    opts.set_iterate_upper_bound(end_key);
    let mut iter = db.raw_iterator_opt(opts);
    iter.seek_for_prev(end_key);
    if iter.valid() {
        Some((iter.key().unwrap().to_vec(), iter.value().unwrap().to_vec()))
    } else {
        None
    }
}

pub fn put_json<T: Serialize>(db: &DB, key: &[u8], msg: &T) {
    let json = serde_json::to_string(msg).unwrap();
    db.put(key, json.as_bytes()).unwrap();
}

pub fn get_truncated_state(db: &DB) -> TruncatedState {
    let value = db.get(RAFT_TRUNCATED_STATE_KEY).unwrap();
    if value.is_none() {
        return TruncatedState::default();
    }

    let value = value.unwrap();
    let state = serde_json::from_slice(&value).unwrap();
    state
}

pub fn put_truncated_state(db: &DB, state: &TruncatedState) {
    put_json(db, RAFT_TRUNCATED_STATE_KEY, state);
}