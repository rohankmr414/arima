package fsm

import (
	"io"
	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
)

type ArimaFSM struct {
	Conn *badger.DB
}

type LogStruct struct {
	Op  string
	Key []byte
	Val []byte
}

func NewArimaFSM(path string) (*ArimaFSM, error) {
	var err error
	opts := badger.DefaultOptions(path)
	opts.Logger = nil
	opts.SyncWrites = true

	handle, err := badger.Open(opts)

	if err != nil {
		return nil, err
	}

	return &ArimaFSM{
		Conn: handle,
	}, nil
}

func (fsm *ArimaFSM) Apply(log *raft.Log) interface{} {
	var data LogStruct
	var err error

	err = decodeMsgPack(log.Data, &data)
	if err != nil {
		return err
	}
	if data.Op == "set" {
		err = fsm.Conn.Update(func(txn *badger.Txn) error {
			return txn.Set(data.Key, data.Val)
		})
	} else if data.Op == "delete" {
		err = fsm.Conn.Update(func(txn *badger.Txn) error {
			return txn.Delete(data.Key)
		})
	}
	if err != nil {
		return err
	}
	return nil
}

func (fsm *ArimaFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &ArimaSnapshot{Conn: fsm.Conn}, nil
}

func (fsm *ArimaFSM) Restore(r io.ReadCloser) error {
	err := fsm.Conn.DropAll()
	if err != nil {
		return err
	}
	err = fsm.Conn.Load(r, 100)
	if err != nil {
		return err
	}
	return nil
}