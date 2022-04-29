package fsm

import (
	"io"

	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/utils"
)

type ArimaFSM struct {
	Conn *badger.DB
}

// type LogStruct struct {
// 	Op  string
// 	Key []byte
// 	Val []byte
// }

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

// Apply log is invoked once a log entry is committed.
// It returns a value which will be made available in the
// ApplyFuture returned by Raft.Apply method if that
// method was called on the same Raft node as the FSM.
func (fsm *ArimaFSM) Apply(log *raft.Log) interface{} {
	// var data CommandPayload
	// var err error

	// err = decodeMsgPack(log.Data, &data)
	// if err != nil {
	// 	return err
	// }
	// if data.Operation == "set" {
	// 	err = fsm.Conn.Update(func(txn *badger.Txn) error {
	// 		return txn.Set(data.Key, data.Value)
	// 	})
	// } else if data.Operation == "delete" {
	// 	err = fsm.Conn.Update(func(txn *badger.Txn) error {
	// 		return txn.Delete(data.Key)
	// 	})
	// }
	// if err != nil {
	// 	return err
	// }
	// return nil
	switch log.Type {
	case raft.LogCommand:
		var payload CommandPayload
		if err := utils.DecodeMsgPack(log.Data, &payload); err != nil {
			return err
		}
		if payload.Operation == "set" {
			return &ApplyResponse{
				Error: fsm.Conn.Update(func(txn *badger.Txn) error {
					return txn.Set(payload.Key, payload.Value)
				}),
				Data: payload.Value,
			}
		} else if payload.Operation == "delete" {
			return &ApplyResponse{
				Error: fsm.Conn.Update(func(txn *badger.Txn) error {
					return txn.Delete(payload.Key)
				}),
				Data: nil,
			}
		} else if payload.Operation == "get" {
			data, err := fsm.Get(payload.Key)
			if err != nil {
				return &ApplyResponse{
					Error: err,
					Data:  data,
				}
			}
		}
	}

	return nil
}

// Snapshot is used to support log compaction. This call should
// return an FSMSnapshot which can be used to save a point-in-time snapshot of the FSM.
func (fsm *ArimaFSM) Snapshot() (raft.FSMSnapshot, error) {
	return &ArimaSnapshot{Conn: fsm.Conn}, nil
}

// Restore is used to restore an FSM from a snapshot. It is not called
// concurrently with any other command. The FSM must discard all previous
// state.
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

func (fsm *ArimaFSM) Get(key []byte) ([]byte, error) {
	var val []byte
	err := fsm.Conn.View(func(txn *badger.Txn) error {
		item, err := txn.Get(key)
		if err != nil {
			return err
		}
		val, err = item.ValueCopy(val)
		return err
	})
	if err != nil {
		return nil, err
	}
	return val, nil
}
