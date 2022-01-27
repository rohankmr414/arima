package fsm

import (
	"log"

	"github.com/dgraph-io/badger/v3"
	"github.com/hashicorp/raft"
)

type ArimaSnapshot struct {
	Conn *badger.DB
}

// Persist should dump all necessary state to the WriteCloser 'sink',
// and call sink.Close() when finished or call sink.Cancel() on error.
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

// Release is invoked when we are finished with the snapshot.
func (snap *ArimaSnapshot) Release() {
	log.Println("Releasing snapshot")
}