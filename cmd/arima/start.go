package main

import (
	"log"
	// "github.com/hashicorp/raft"
	"github.com/rohankmr414/arima/arimanode"
	// "github.com/google/uuid"
)

func startnode() {
	var arnode arimanode.ArimaNode
	var err error
	// id := uuid.New().String()
	arnode.Raft, arnode.Fsm, err = arnode.New(nconf.Id, nconf.Path, nconf.BindAddr)

	if err != nil {
		log.Fatalln("Error creating node:", err)
	}

}