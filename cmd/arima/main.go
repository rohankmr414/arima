package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/urfave/cli/v2"
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
	// The maxPool controls how many connections we will pool.
	maxPool = 3

	// The timeout is used to apply I/O deadlines. For InstallSnapshot, we multiply
	// the timeout by (SnapshotSize / TimeoutScale).
	tcpTimeout = 10 * time.Second

	// The `retain` parameter controls how many
	// snapshots are retained. Must be at least 1.
	raftSnapShotRetain = 2
)

var (
	svport    string
	raftport  string
	nodeid    string
	volumedir string
)

func main() {
	app := &cli.App{
		Name:                 "arima",
		Usage:                "A simple fault tolerant key-value store",
		Description:          "A distributed fault-tolerant key-value store which uses Raft for consensus.",
		EnableBashCompletion: true,
		Commands: []*cli.Command{
			{
				Name:    "run",
				Usage:   "Run the server",
				Aliases: []string{"r"},
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:        "server-port",
						Value:       "",
						Usage:       "The port to listen on for HTTP requests",
						Required:    true,
						Aliases:     []string{"s"},
						Destination: &svport,
					},
					&cli.StringFlag{
						Name:        "node-id",
						Value:       "",
						Usage:       "The raft node id",
						Required:    true,
						Aliases:     []string{"i"},
						Destination: &nodeid,
					},
					&cli.StringFlag{
						Name:        "raft-port",
						Value:       "",
						Usage:       "The port to listen on for raft requests",
						Required:    true,
						Aliases:     []string{"r"},
						Destination: &raftport,
					},
					&cli.PathFlag{
						Name:        "volume-dir",
						Value:       "",
						Usage:       "The directory to store the data",
						Required:    true,
						Aliases:     []string{"v"},
						Destination: &volumedir,
					},
				},
				Action: func(c *cli.Context) error {
					fmt.Println("Starting arima")
					err := startNode(svport, raftport, nodeid, volumedir)
					if err != nil {
						log.Fatal(err)
						return err
					}
					return nil
				},
			},
		},
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
