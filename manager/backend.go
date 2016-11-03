// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-29
package manager

import (
	"hyperchain/consensus"
	"hyperchain/core/blockpool"

	"hyperchain/crypto"
	"hyperchain/p2p"

	"hyperchain/accounts"
	"hyperchain/protos"

	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/event"
	"time"
)

// init protocol manager params and start
func New(eventMux *event.TypeMux, blockPool *blockpool.BlockPool, peerManager p2p.PeerManager, consenter consensus.Consenter,
	am *accounts.AccountManager, commonHash crypto.CommonHash, nodeId int, syncReplicaInterval time.Duration, syncReplica bool) *ProtocolManager {

	aliveChan := make(chan int)
	go peerManager.Start(aliveChan, eventMux)

	//wait for all peer are connected
	select {
	case initType <- aliveChan:
		{
			protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, consenter, am, commonHash, syncReplicaInterval, syncReplica, initType)
			protocolManager.Start()
			//start server
			return protocolManager
		}
	}

	//return nil
}
