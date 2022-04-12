package store

import (
	"errors"

	"github.com/dgraph-io/badger/v3"
	"github.com/rohankmr414/arima/utils"
)

var ErrKeyNotFound = errors.New("not found")

type StableStore struct {
	Conn *badger.DB
}

func NewStableStore(path string) (*StableStore, error) {
	var err error
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	opts.SyncWrites = true

	handle, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}

	return &StableStore{
		Conn: handle,
	}, nil
}

// Set is used to set a key/value set
func (store *StableStore) Set(key, val []byte) error {
	return store.Conn.Update(func(txn *badger.Txn) error {
		err := txn.Set(key, val)
		if err != nil {
			return err
		}
		return nil
	})
}

// Get returns the value for key, or an empty byte slice if key was not found.
func (store *StableStore) Get(key []byte) ([]byte, error) {
	var value []byte
	err := store.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			if err.Error() == badger.ErrKeyNotFound.Error() {
				return ErrKeyNotFound
			}
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
func (store *StableStore) SetUint64(key []byte, val uint64) error {
	return store.Set(key, utils.Uint64ToBytes(val))
}

// GetUint64 returns the uint64 value for key, or 0 if key was not found.
func (store *StableStore) GetUint64(key []byte) (uint64, error) {
	val, err := store.Get(key)
	if err != nil {
		return 0, err
	}
	return utils.BytesToUint64(val), nil
}
