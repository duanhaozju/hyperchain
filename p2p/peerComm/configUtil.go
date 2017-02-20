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
	GetLocalID() int
	GetLocalIP() string
	GetLocalGRPCPort() int
	GetLocalJsonRPCPort() int
	GetIntroducerIP() string
	GetIntroducerID() int
	GetIntroducerJSONRPCPort() int
	GetIntroducerPort() int
	IsOrigin() bool
	IsVP() bool
	GetID(nodeID int) int
	GetPort(nodeID int) int
	GetRPCPort(nodeID int) int
	GetIP(nodeID int) string
	GetMaxPeerNumber() int
	AddNodesAndPersist(addrs map[string]peermessage.PeerAddr)
	DelNodesAndPersist(addrs map[string]peermessage.PeerAddr)
}

type ConfigWriter interface {
	SaveAddress(addr Address) error
}
