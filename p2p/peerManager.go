// peer manager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25 Quan Chen
// change log:	modified the  PeerManager interface definition
//		implements GrpcPeerManager methods
package p2p

import (
	peer "hyperchain/p2p/peer"
	"hyperchain/recovery"
	"hyperchain/event"
	"github.com/op/go-logging"
)

// Init the log setting
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
}

type PeerManager interface {
	// judge all peer are connected and return them
	//JudgeAlivePeers(*chan bool)
	GetAllPeers() []*peer.Peer
	Start(aliveChain chan bool,eventMux *event.TypeMux)
	GetNodeId() int
	BroadcastPeers(payLoad []byte)
	SendMsgToPeers(payLoad []byte,peerList []uint64,MessageType recovery.Message_MsgType)
	GetPeerInfos() peer.PeerInfos
}


