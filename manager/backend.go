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
	"hyperchain/common"
)

// init protocol manager params and start
func New(
	eventMux *event.TypeMux,
	blockPool *blockpool.BlockPool,
	peerManager p2p.PeerManager,
	consenter consensus.Consenter,
	am *accounts.AccountManager,
	exist chan bool,
	expiredTime time.Time, cm *admittance.CAManager, config *common.Config) *ProtocolManager {

	syncReplicaInterval := config.GetDuration(common.SYNC_REPLICA_INFO_INTERVAL)
	syncReplica := config.GetBool(common.SYNC_REPLICA)

	//init hash object
	commonHash := crypto.NewKeccak256Hash("keccak256")
	protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, consenter, am, commonHash, syncReplicaInterval, syncReplica, exist, expiredTime)
	aliveChan := make(chan int)
	protocolManager.Start(aliveChan, cm)

	return protocolManager
}
