//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerComm

import (
	"hyperchain/p2p/peermessage"
	"github.com/op/go-logging"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/peerComm")
}
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
	AddNodesAndPersist(addrs map[string]peermessage.PeerAddr)
	DelNodesAndPersist(addrs map[string]peermessage.PeerAddr)
	Peers() []PeerConfigNodes
}

type ConfigWriter interface {
	SaveAddress(addr Address) error
}
