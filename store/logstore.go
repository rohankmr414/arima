package store

import (
	"log"
	"github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/utils"
	"github.com/dgraph-io/badger/v3"
)

type LogStore struct {
	Conn *badger.DB
}

func NewLogStore(path string) (*LogStore, error) {
	var err error
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	opts.SyncWrites = true

	handle, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}

	return &LogStore{
		Conn: handle,
	}, nil
}

// FirstIndex returns the first index written. 0 for no entries.
func (store *LogStore) FirstIndex() (uint64, error) {
	var key uint64
	err := store.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		it.Rewind()

		if it.Valid() {
			item := it.Item()
			key = utils.BytesToUint64(item.Key()) 
		} else {
			key = 0
		}


		return nil
	})

	if err != nil {
		log.Fatalf("Error getting first index: %v", err)
		return 0, err
	}

	return key, nil
}

// LastIndex returns the last index written. 0 for no entries.
func (store *LogStore) LastIndex() (uint64, error) {
	var key uint64
	err := store.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		opts.Reverse = true
		it := txn.NewIterator(opts)
		defer it.Close()

		it.Rewind()

		if it.Valid() {
			item := it.Item()
			key = utils.BytesToUint64(item.Key()) 
		} else {
			key = 0
		}

		return nil
	})

	if err != nil {
		log.Fatalf("Error getting last index: %v", err)
		return 0, err
	}

	return key, nil
}

// GetLog gets a log entry at a given index.
func (store *LogStore) GetLog(index uint64, log *raft.Log) error {
	key := utils.Uint64ToBytes(index)
	var value []byte
	err := store.Conn.View(func (txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		} 
		value, err = item.ValueCopy(value)
		return err
	})

	if err != nil {
		return err
	}

	return utils.DecodeMsgPack(value, log)
}

// StoreLog stores a log entry.
func (store *LogStore) StoreLog(log *raft.Log) error {
	key := utils.Uint64ToBytes(log.Index)
	val, err := utils.EncodeMsgPack(log)
	if err != nil {
		return err
	}
	storeerr := store.Conn.Update(func (txn *badger.Txn) error {
		err := txn.Set(key, val.Bytes())
		if err != nil {
			return err
		}
		return nil
	})

	return storeerr
}

// StoreLogs stores multiple log entries.
func (store *LogStore) StoreLogs(logs []*raft.Log) error {
	for _, log := range logs {
		if err := store.StoreLog(log); err != nil {
			return err
		}
	}
	return nil
}

// DeleteRange deletes a range of log entries. The range is inclusive.
func (store *LogStore) DeleteRange(min, max uint64) error {
	minkey := utils.Uint64ToBytes(min)

	txn := store.Conn.NewTransaction(true)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	it.Seek(minkey)

	for {
		key := it.Item().Key()

		if utils.BytesToUint64(key) > max {
			break
		}
		err := txn.Delete(key)
		if err != nil {
			return err
		}
		it.Next()
	}
	if err := txn.Commit(); err != nil {
		log.Fatalf("Error deleting range: %v", err)
		return err
	}
	return nil
}
