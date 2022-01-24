use rocksdb::{DB, Options, WriteBatch, WriteOptions};

pub struct KV {
    db: DB,
}

impl KV {
    pub fn new(path: &str) -> KV {
        let mut opts = Options::default();
        opts.create_if_missing(true);
        KV {
            db: DB::open(&opts, path).unwrap(),
        }
    }

    pub fn put(&self, key: &[u8], value: &[u8]) {
        self.db.put(key, value).unwrap();
    }

    pub fn get(&self, key: &[u8]) -> Option<Vec<u8>> {
        self.db.get(key).unwrap()
    }

    pub fn delete(&self, key: &[u8]) {
        self.db.delete(key).unwrap();
    }
}