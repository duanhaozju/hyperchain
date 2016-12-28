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

/*
	pm := manager.New(	eventMux,
				blockPool,
				grpcPeerMgr,
				cs,
				am,
				kec256Hash,
				config.getNodeID(),
				syncReplicaInterval,
				syncReplicaEnable,
				exist,
				expiredTime,
				argv.IsReconnect,
				config.getGRPCPort())
 */

// init protocol manager params and start
func New(
	eventMux *event.TypeMux,
	blockPool *blockpool.BlockPool,
	peerManager p2p.PeerManager,
	consenter consensus.Consenter,
	am *accounts.AccountManager,
	commonHash crypto.CommonHash,
	syncReplicaInterval time.Duration,
	syncReplica bool,
	exist chan bool,
	expiredTime time.Time) *ProtocolManager {

	aliveChan := make(chan int)
	//add reconnect param

	go peerManager.Start(aliveChan, eventMux)
	//wait for all peer are connected
	initType := <-aliveChan
	//select {
	//case initType := <-aliveChan:
	//	{
	protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, consenter, am, commonHash, syncReplicaInterval, syncReplica, initType, exist, expiredTime)
	protocolManager.Start()
	//start server
	return protocolManager
	//}
	//}
}
