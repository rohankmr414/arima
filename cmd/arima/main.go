package main

import (
	"fmt"
	"github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/fsm"
	"github.com/rohankmr414/arima/server"
	"github.com/rohankmr414/arima/store"
	"github.com/spf13/viper"
	"log"
	"net"
	"os"
	"path/filepath"
	"time"
)

// configRaft configuration for raft node
type configRaft struct {
	NodeId    string `mapstructure:"node_id"`
	Port      int    `mapstructure:"port"`
	VolumeDir string `mapstructure:"volume_dir"`
}

// configServer configuration for HTTP server
type configServer struct {
	Port int `mapstructure:"port"`
}

// config configuration
type config struct {
	Server configServer `mapstructure:"server"`
	Raft   configRaft   `mapstructure:"raft"`
}

const (
	serverPort = "SERVER_PORT"

	raftNodeId = "RAFT_NODE_ID"
	raftPort   = "RAFT_PORT"
	raftVolDir = "RAFT_VOL_DIR"
)

var confKeys = []string{
	serverPort,

	raftNodeId,
	raftPort,
	raftVolDir,
}

const (
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2
)


func main() {
	var v = viper.New()
	v.AutomaticEnv()

	if err := v.BindEnv(confKeys...); err != nil {
		log.Fatal(err)
		return
	}

	conf := config{
		Server: configServer{
			Port: v.GetInt(serverPort),
		},
		Raft: configRaft{
			NodeId:    v.GetString(raftNodeId),
			Port:      v.GetInt(raftPort),
			VolumeDir: v.GetString(raftVolDir),
		},
	}

	log.Printf("%+v\n", conf)

	var raftBindAddr = fmt.Sprintf("localhost:%d", conf.Raft.Port)

	raftConf := raft.DefaultConfig()
	raftConf.LocalID = raft.ServerID(conf.Raft.NodeId)
	
	arimaFsm, err := fsm.NewArimaFSM(conf.Raft.VolumeDir)
	if err != nil {
		log.Fatalln(err)
		return
	}

	arimaLogStore, err := store.NewLogStore(filepath.Join(conf.Raft.VolumeDir, "log"))
	if err != nil {
		log.Fatalln(err)
		return
	}

	arimaStableStore, err := store.NewStableStore(filepath.Join(conf.Raft.VolumeDir, "stable"))
	if err != nil {
		log.Fatalln(err)
		return
	}

	arimaSnapshotStore, err := raft.NewFileSnapshotStore(conf.Raft.VolumeDir, raftSnapShotRetain, os.Stdout)
	if err != nil {
		log.Fatalln(err)
		return
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", raftBindAddr)
	if err != nil {
		log.Fatalln("Error resolving TCP address:", err)
		return
	}

	transport, err := raft.NewTCPTransport(raftBindAddr, tcpAddr, maxPool, tcpTimeout, os.Stdout)
	if err != nil {
		log.Fatalln("Error creating TCP transport:", err)
		return
	}

	raftServer, err := raft.NewRaft(raftConf, arimaFsm, arimaLogStore, arimaStableStore, arimaSnapshotStore, transport)
	if err != nil {
		log.Fatal(err)
		return
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
	}
}
