// peer manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25 Quan Chen
// change log:	modified the  PeerManager interface definition
//		implements GrpcPeerManager methods
package p2p

import (
	"github.com/op/go-logging"
	"hyperchain/event"
	"hyperchain/recovery"
)

// Init the log setting
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}

// PeerManager provides the basic functions which supports the peer to peer
// data transfer. Those should be invoked by the higher layer.
type PeerManager interface {
	DeleteNode
	// get the all peer list to broadcast
	GetAllPeers() []*Peer
	GetAllPeersWithTemp() []*Peer
	// initialize the peerManager which is for init the local node
	Start(aliveChain chan int, eventMux *event.TypeMux, isReconnect bool, GRPCProt int64)
	// Get local node id
	GetNodeId() int
	// broadcast information to peers
	BroadcastPeers(payLoad []byte)
	// send a message to specific peer  UNICAST
	SendMsgToPeers(payLoad []byte, peerList []uint64, MessageType recovery.Message_MsgType)
	//get the peer information of all nodes.
	GetPeerInfo() PeerInfos
	// set
	SetPrimary(id uint64) error
	// get local node instance
	GetLocalNode() *Node
	// update routing table when new peer's join request is accepted
	UpdateRoutingTable(payLoad []byte)
	// use by new peer when join the chain dynamically only
	ConnectToOthers()
	// set the new node online
	SetOnline()
	// get local address payload
	GetLocalAddressPayload() []byte
}

// delete node interface
type DeleteNode interface{
	GetLocalNodeHash() string
	GetRouterHashifDelete(hash string) (string,uint64)
	DeleteNode(hash string) error// if self {...} else{...}
}
