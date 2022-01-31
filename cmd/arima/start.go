package main

// import (
// 	"log"
// 	// "net"
// 	// "github.com/hashicorp/raft"
// 	"github.com/rohankmr414/arima/arimanode"
// 	// "github.com/google/uuid"
// )

// func startnode() {
// 	var arnode arimanode.ArimaNode
// 	var err error

// 	// _, port, err := net.SplitHostPort(nconf.BindAddr)
// 	// sock, err := net.Listen("tcp", ":"+port)
// 	// id := uuid.New().String()
// 	arnode.Raft, arnode.Fsm, err = arnode.New(nconf.IsLeader, nconf.Id, nconf.Path, nconf.BindAddr)

// 	if err != nil {
// 		log.Fatalln("Error creating node:", err)
// 	}

// }