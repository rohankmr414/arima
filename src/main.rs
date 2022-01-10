mod db;
mod storage;

#[macro_use]
extern crate slog;

fn main() {
    let db = db::dbinit::open("/tmp/rocksdb");
    db::transactions::put(&db, b"key", b"value");
    let val = db::transactions::get(&db, b"key");
    println!("{:?}", val);
}
