// peer manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25 Quan Chen
// change log:	modified the  PeerManager interface definition
//		implements GrpcPeerManager methods
// 2016-12-23: change the interface definition (Chen Quan)
package p2p

import (
	"github.com/op/go-logging"
	"hyperchain/event"
	"hyperchain/admittance"
)

// Init the log setting
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}

// PeerManager provides the basic functions which supports the peer to peer
// data transfer. Those should be invoked by the higher layer.
type PeerManager interface {
	AddNode
	DeleteNode
	MsgSender
	InfoGetter
	// initialize the peerManager which is for init the local node
	Start(aliveChain chan int, eventMux *event.TypeMux, cm *admittance.CAManager)
}

// MsgSender Send msg to others peer
type MsgSender interface {
	// broadcast information to peers
	BroadcastPeers(payLoad []byte)
	// send a message to specific peer  UNICAST
	SendMsgToPeers(payLoad []byte, peerList []uint64)
}

// AddNode
type AddNode interface {
	// update routing table when new peer's join request is accepted
	UpdateRoutingTable(payLoad []byte)
	UpdateAllRoutingTable(routerPayload []byte)
	GetLocalAddressPayload() []byte
	SetOnline()
}

// DeleteNode interface
type DeleteNode interface {
	GetLocalNodeHash() string
	GetRouterHashifDelete(hash string) (string, uint64, uint64)
	DeleteNode(hash string) error // if self {...} else{...}
}

// InfoGetter get the peer info to manager
type InfoGetter interface {
	// get the all peer list to broadcast
	GetAllPeers() []*Peer
	GetAllPeersWithTemp() []*Peer
	GetVPPeers() []*Peer
	// get local node instance
	GetLocalNode() *Node
	// Get local node id
	GetNodeId() int
	//get the peer information of all nodes.
	GetPeerInfo() PeerInfos
	// set
	SetPrimary(id uint64) error
	// use by new peer when join the chain dynamically only
	GetRouters() []byte
}
