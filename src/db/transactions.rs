use crate::db::dbinit::RocksDB;
use rocksdb::{ReadOptions, WriteBatch, WriteOptions};

// function to get the value of a key
pub fn get(db: &RocksDB, key: &[u8]) -> Option<Vec<u8>> {
    let mut opts = ReadOptions::default();
    let value = db.db.get_opt(key, &mut opts).unwrap();
    match value {
        Some(v) => Some(v.to_vec()),
        None => None,
    }
}

// function to put a key-value pair into the database
pub fn put(db: &RocksDB, key: &[u8], value: &[u8]) {
    let mut opts = WriteOptions::default();
    db.db.put_opt(key, value, &mut opts).unwrap();
}

// function to delete a key-value pair from the database
pub fn delete(db: &RocksDB, key: &[u8]) {
    let mut opts = WriteOptions::default();
    db.db.delete_opt(key, &mut opts).unwrap();
}

// function to write a batch of key-value pairs into the database
pub fn write_batch(db: &RocksDB, batch: WriteBatch) {
    let mut opts = WriteOptions::default();
    db.db.write_opt(batch, &mut opts).unwrap();
}
