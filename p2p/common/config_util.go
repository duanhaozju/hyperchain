//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

import (
	//"hyperchain/p2p/message"
)

type Config interface {
	LocalID() int
	LocalIP() string
	LocalGRPCPort() int
	LocalJsonRPCPort() int
	IntroIP() string
	IntroID() int
	IntroJSONRPCPort() int
	IntroPort() int
	IsOrigin() bool
	IsVP() bool
	GetID(nodeID int) int
	GetPort(nodeID int) int
	GetRPCPort(nodeID int) int
	GetIP(nodeID int) string
	MaxNum() int
	//AddNodesAndPersist(addrs map[string]message.PeerAddr)
	//DelNodesAndPersist(addrs map[string]message.PeerAddr)
	Peers() []PeerConfigNodes
}

type ConfigWriter interface {
	SaveAddress(addr Address) error
}
