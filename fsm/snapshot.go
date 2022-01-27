package fsm

import (
	"log"

	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
)

type ArimaSnapshot struct {
	Conn *badger.DB
}

func (snap *ArimaSnapshot) Persist(sink raft.SnapshotSink) error {
	log.Println("Persisting snapshot")
	_, err := snap.Conn.Backup(sink, 0)

	if err != nil {
		log.Fatalln("Error persisting snapshot:", err)
		return err
	}

	err = sink.Close()
	if err != nil {
		log.Fatalln("Error closing snapshot:", err)
		return err
	}
	return nil
}

func (snap *ArimaSnapshot) Release() {
	log.Println("Releasing snapshot")
}