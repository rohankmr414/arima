package main

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/fsm"
	"github.com/rohankmr414/arima/server"
	"github.com/rohankmr414/arima/store"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func startNode(svport, raftport, nodeid, volumedir string) error {
	serverPort, err := strconv.Atoi(svport)
	if err != nil {
		log.Fatal(err)
		return err
	}

	raftPort, err := strconv.Atoi(raftport)
	if err != nil {
		log.Fatal(err)
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

	var raftBindAddr = fmt.Sprintf("localhost:%d", conf.Raft.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)

	arimaFsm, err := fsm.NewArimaFSM(conf.Raft.VolumeDir)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	arimaLogStore, err := store.NewLogStore(filepath.Join(conf.Raft.VolumeDir, "log"))
	if err != nil {
		log.Fatalln(err)
		return err
	}

	arimaStableStore, err := store.NewStableStore(filepath.Join(conf.Raft.VolumeDir, "stable"))
	if err != nil {
		log.Fatalln(err)
		return err
	}

	arimaSnapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		log.Fatalln(err)
		return err
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBindAddr)
	if err != nil {
		log.Fatalln("Error resolving TCP address:", err)
		return err
	}

	transport, err := raft.NewTCPTransport(raftBindAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		log.Fatalln("Error creating TCP transport:", err)
		return err
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
	if err := srv.Start(); err != nil {
		log.Fatalln("failed to start server:", err)
		return err
	}

	return nil
}