//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"hyperchain/accounts"
	"hyperchain/consensus"
	"hyperchain/core/blockpool"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/admittance"
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
	expiredTime time.Time, cm *admittance.CAManager) *ProtocolManager {

	//add reconnect param

	protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, consenter, am, commonHash, syncReplicaInterval, syncReplica, exist, expiredTime)
	aliveChan := make(chan int)
	protocolManager.Start(aliveChan, cm)

	//wait for all peer are connected
	//select {
	//case initType := <-aliveChan:
	//	{
	//start server
	return protocolManager
	//}
	//}
}
