package main

import (
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"

	"github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/fsm"
	"github.com/rohankmr414/arima/server"
	"github.com/rohankmr414/arima/store"
)

func startNode(svport, raftport, nodeid, volumedir string) error {
	serverPort, err := strconv.Atoi(svport)
	if err != nil {
		return err
	}

	raftPort, err := strconv.Atoi(raftport)
	if err != nil {
		return err
	}

	conf := config{
		Server: configServer{
			Port: serverPort,
		},
		Raft: configRaft{
			NodeId:    nodeid,
			Port:      raftPort,
			VolumeDir: volumedir,
		},
	}

	log.Printf("%+v\n", conf)

	raftBindAddr := fmt.Sprintf("localhost:%d", conf.Raft.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)

	arimaFsm, err := fsm.NewArimaFSM(conf.Raft.VolumeDir)
	if err != nil {
		return err
	}

	arimaLogStore, err := store.NewLogStore(filepath.Join(conf.Raft.VolumeDir, "log"))
	if err != nil {
		return err
	}

	arimaStableStore, err := store.NewStableStore(filepath.Join(conf.Raft.VolumeDir, "stable"))
	if err != nil {
		return err
	}

	arimaSnapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBindAddr)
	if err != nil {
		return fmt.Errorf("error resolving TCP address: %s", err)
	}

	transport, err := raft.NewTCPTransport(raftBindAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		return fmt.Errorf("error creating TCP transport: %s", err)
	}

	raftServer, err := raft.NewRaft(raftConf, arimaFsm, arimaLogStore, arimaStableStore, arimaSnapshotStore, transport)
	if err != nil {
		log.Fatal(err)
		return err
	}

	// always start single server as a leader
	configuration := raft.Configuration{
		Servers: []raft.Server{
			{
				ID:      raft.ServerID(conf.Raft.NodeId),
				Address: transport.LocalAddr(),
			},
		},
	}

	raftServer.BootstrapCluster(configuration)

	srv := server.New(fmt.Sprintf(":%d", conf.Server.Port), arimaFsm.Conn, raftServer)
	if err = srv.Start(); err != nil {
		return fmt.Errorf("failed to start server: %s", err)
	}

	return nil
}
