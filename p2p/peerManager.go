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
	Server "hyperchain/p2p/node"
	peer "hyperchain/p2p/peer"
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
	// get the all peer list to broadcast
	GetAllPeers() []*peer.Peer
	// initialize the peerManager which is for init the local node
	Start(aliveChain chan bool, eventMux *event.TypeMux)
	// Get local node id
	GetNodeId() int
	// broadcast information to peers
	BroadcastPeers(payLoad []byte)
	// send a message to specific peer  UNICAST
	SendMsgToPeers(payLoad []byte, peerList []uint64, MessageType recovery.Message_MsgType)
	//get the peer information of all nodes.
	GetPeerInfo() peer.PeerInfos
	// set
	SetPrimary(id uint64) error
	// get local node instance
	GetLocalNode() *Server.Node
	// update routing table when new peer's join request is accepted
	UpdateRoutingTable([]byte)
}
