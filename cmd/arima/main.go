package main

import (
	"os"
	"log"
	// "github.com/rohankmr414/arima/arimanode"
	"github.com/urfave/cli/v2"
	"github.com/google/uuid"
)

type NodeConf struct {
	Id string
	BindAddr string
	Path string
	IsLeader bool
}

var nconf NodeConf

var path string
// var rn string

func Init() {
	nconf.Id = uuid.New().String()
	var err error
	nconf.Path, err = os.UserHomeDir()
	if err != nil {
		log.Fatal(err)
	}
}



func main() {
	app := cli.App{}
	app.Name = "arima"
	app.Usage = "A fault tolerant key value store using raft"
	app.Commands = []*cli.Command{
		{
			Name: "start",
			Usage: "Start a new arima node",
			Flags: []cli.Flag{
				&cli.StringFlag{
					Name: "path",
					Usage: "Path to store the data",
					Required: true,
					Value: path + "/arimatest",
					Aliases: []string{"p"},
					Destination: &nconf.Path,
				},
				&cli.StringFlag{
					Name: "address",
					Usage: "Address to bind to",
					Required: true,
					Aliases: []string{"a"},
					Destination: &nconf.BindAddr,
				},
				// &cli.StringFlag{
				// 	Name: "raft-address",
				// 	Usage: "Peer address to connect to",
				// 	Required: false,
				// 	Aliases: []string{"r"},
				// 	Destination: &rn,
				// },
			},
			Action: func(c *cli.Context) error {
				// Init()
				nconf.Id = uuid.New().String()
				startnode()
				return nil
			},
		},
		// {
		// 	Name: "get",
		// 	Usage: "Get a value from the store",
		// 	Flags: []cli.Flag{
		// 		&cli.StringFlag{
		// 			Name: "key",
		// 			Usage: "Key to get",
		// 			Required: true,
		// 			Aliases: []string{"k"},
		// 		}
		// 	}

		// },
	}
	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}
