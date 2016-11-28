//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"hyperchain/accounts"
	"hyperchain/consensus"
	"hyperchain/core/blockpool"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/p2p"
	"time"
)

// init protocol manager params and start
func New(eventMux *event.TypeMux, blockPool *blockpool.BlockPool, peerManager p2p.PeerManager, consenter consensus.Consenter,
am *accounts.AccountManager, commonHash crypto.CommonHash, nodeId int, syncReplicaInterval time.Duration, syncReplica bool, port int) *ProtocolManager {

	aliveChan := make(chan int)
	go peerManager.Start(aliveChan, eventMux, int64(port))
	//wait for all peer are connected
	initType := <-aliveChan
	//select {
	//case initType := <-aliveChan:
	//	{
	protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, consenter, am, commonHash, syncReplicaInterval, syncReplica, initType)
	protocolManager.Start()
	//start server
	return protocolManager
	//}
	//}
}
