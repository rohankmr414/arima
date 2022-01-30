package arimanode

import (
	"errors"
	"log"
	"net"
	"os"
	"time"

	"github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/fsm"
	"github.com/rohankmr414/arima/store"
)

type ArimaNode struct {
	Fsm *fsm.ArimaFSM
	Raft *raft.Raft
}

func (n *ArimaNode) New(id string, path string, bindaddr string) (*raft.Raft, *fsm.ArimaFSM, error) {
	var node ArimaNode

	config := raft.DefaultConfig()
	config.LocalID = raft.ServerID(id)

	addr, err := net.ResolveTCPAddr("tcp", bindaddr)
	
	if err != nil {
		log.Fatalln("Error resolving TCP address:", err)
		return nil, nil, err
	}
	log.Println(bindaddr)
	transport, err := raft.NewTCPTransport(bindaddr, addr, 3, 10*time.Second, os.Stderr)

	if err != nil {
		log.Fatalln("Error creating TCP transport:", err)
		return nil, nil, err
	}

	snapshot, err := raft.NewFileSnapshotStore(path+"/snap", 3, os.Stderr)
	if err != nil {
		log.Fatalln("Error creating snapshot store:", err)
		return nil, nil, err
	}


	var logstore raft.LogStore
	var stablestore raft.StableStore
	// var sm raft.FSM

	logstore, err = store.NewLogStore(path+"/log")
	if err != nil {
		log.Fatalln("Error creating log store:", err)
		return nil, nil, err
	}

	stablestore, err = store.NewStableStore(path+"/stable")
	if err != nil {
		log.Fatalln("Error creating stable store:", err)
		return nil, nil, err
	}

	sm, err := fsm.NewArimaFSM(path+"/data")
	if err != nil {
		log.Fatalln("Error creating FSM:", err)
		return nil, nil, err
	}

	node.Fsm = sm
	
	rft, err := raft.NewRaft(config, sm, logstore, stablestore, snapshot, transport)

	if err != nil {
		log.Fatalln("Error creating Raft:", err)
		return nil, nil, err
	}

	node.Raft = rft
	
	return rft, sm, nil

}

func (n *ArimaNode) Get(key []byte) ([]byte, error) {
	if n.Raft.State() != raft.Leader {
		log.Println("Not the leader")
		return nil, raft.ErrNotLeader
	}

	return n.Fsm.Get(key)
}

func (n *ArimaNode) Set(key []byte, val []byte) error {
	if n.Raft.State() != raft.Leader {
		log.Println("Not leader")
		return raft.ErrNotLeader
	}

	var data fsm.LogStruct

	data.Op = "set"
	data.Key = key
	data.Val = val

	databuf, err := encodeMsgPack(data)
	if err != nil {
		log.Println("Error encoding data:", err)
		return err
	}
	fut := n.Raft.Apply(databuf.Bytes(), 5*time.Second)

	return fut.Error()
}

func (n *ArimaNode) Delete(key []byte) error {
	if n.Raft.State() != raft.Leader {
		log.Println("Not leader")
		return nil
	}

	var data fsm.LogStruct

	data.Op = "delete"
	data.Key = key
	data.Val = []byte{}

	databuf, err := encodeMsgPack(data)
	if err != nil {
		log.Println("Error encoding data:", err)
		return err
	}
	fut := n.Raft.Apply(databuf.Bytes(), 5*time.Second)

	return fut.Error()
}

func (n *ArimaNode) Join(id, addr string) error {
	log.Panicln("Received join request from ", id, " at ", addr)
	if n.Raft.State() != raft.Leader {
		log.Println("Not leader")
		return raft.ErrNotLeader
	}

	conf := n.Raft.GetConfiguration()

	if err := conf.Error(); err != nil {
		log.Println("Error getting configuration:", err)
		return err
	}

	for _, srv := range conf.Configuration().Servers {
		if srv.ID == raft.ServerID(id) {
			log.Println("Already a member of cluster")
			return errors.New("already a member of cluster")
		}
	}

	fut := n.Raft.AddVoter(raft.ServerID(id), raft.ServerAddress(addr), 0, 0)

	if err := fut.Error(); err != nil {
		return err
	}

	log.Println("Added ", id, " at ", addr, " to cluster")
	return nil
}

func (n *ArimaNode) Leave(id string) error {
	log.Println("Received leave request from ", id)
	if n.Raft.State() != raft.Leader {
		log.Println("Not leader")
		return raft.ErrNotLeader
	}

	conf := n.Raft.GetConfiguration()

	if err := conf.Error(); err != nil {
		log.Println("Error getting configuration:", err)
		return err
	}

	for _, srv := range conf.Configuration().Servers {
		if srv.ID == raft.ServerID(id) {
			fut := n.Raft.RemoveServer(raft.ServerID(id), 0, 0)

			if err := fut.Error(); err != nil {
				log.Println("Error removing server:", err)
				return err
			}

			log.Println("Removed ", id, " from cluster")
			return nil
		}
	}

	log.Println("Not a member of cluster")
	return errors.New("not a member of cluster")
}
