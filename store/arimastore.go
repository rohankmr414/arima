package store

import (
	"log"
	"github.com/hashicorp/raft"
	"github.com/dgraph-io/badger/v3"
)

type ArimaStore struct {
	Conn *badger.DB
}

func NewArimaStore(path string) (*ArimaStore, error) {
	var err error
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	opts.SyncWrites = true

	handle, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}

	return &ArimaStore{
		Conn: handle,
	}, nil
}

// FirstIndex returns the first index written. 0 for no entries.
func (store *ArimaStore) FirstIndex() (uint64, error) {
	var key uint64
	err := store.Conn.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchSize = 10
		it := txn.NewIterator(opts)
		defer it.Close()

		it.Rewind()

		if it.Valid() {
			item := it.Item()
			key = bytesToUint64(item.Key()) 
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
func (store *ArimaStore) LastIndex() (uint64, error) {
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
			key = bytesToUint64(item.Key()) 
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
func (store *ArimaStore) GetLog(index uint64, log *raft.Log) error {
	key := uint64ToBytes(index)
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

	return decodeMsgPack(value, log)
}

// StoreLog stores a log entry.
func (store *ArimaStore) StoreLog(log *raft.Log) error {
	key := uint64ToBytes(log.Index)
	val, err := encodeMsgPack(log)
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
func (store *ArimaStore) StoreLogs(logs []*raft.Log) error {
	for _, log := range logs {
		if err := store.StoreLog(log); err != nil {
			return err
		}
	}
	return nil
}

// DeleteRange deletes a range of log entries. The range is inclusive.
func (store *ArimaStore) DeleteRange(min, max uint64) error {
	minkey := uint64ToBytes(min)

	txn := store.Conn.NewTransaction(true)
	defer txn.Discard()

	it := txn.NewIterator(badger.DefaultIteratorOptions)
	it.Seek(minkey)

	for {
		key := it.Item().Key()

		if bytesToUint64(key) > max {
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

// Set is used to set a key/value set
func (store *ArimaStore) Set(key []byte, val []byte) error {
	return store.Conn.Update(func (txn *badger.Txn) error {
		err := txn.Set(key, val)
		if err != nil {
			return err
		}
		return nil
	})
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (store *ArimaStore) Get(key []byte) ([]byte, error) {
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
		val := []byte{}
		return val, err
	}

	return value, err
}

// SetUint64 is like Set, but handles uint64 values
func (store *ArimaStore) SetUint64(key []byte, val uint64) error {
	return store.Set(key, uint64ToBytes(val))
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (store *ArimaStore) GetUint64(key []byte) (uint64, error) {
	val, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	return bytesToUint64(val), nil
}