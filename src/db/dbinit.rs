use std::sync::Arc;

use rocksdb::{Options, DB};

pub struct RocksDB {
    pub db: Arc<DB>,
}

pub fn open(path: &str) -> RocksDB {
    let mut opts = Options::default();
    opts.create_if_missing(true);
    let db = DB::open(&opts, path).unwrap();
    RocksDB {
        db: Arc::new(db),
    }
}

pub fn close(db: &RocksDB) {
    drop(db);
}