//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"hyperchain/accounts"
	"hyperchain/consensus"
	"hyperchain/core/executor"
	"hyperchain/event"
	"hyperchain/admittance"
	"hyperchain/p2p"
	"time"
)

// init protocol manager params and start
func New(
	namespace string,
	eventMux *event.TypeMux,
	executor *executor.Executor,
	peerManager p2p.PeerManager,
	consenter consensus.Consenter,
	am *accounts.AccountManager,
	exist chan bool,
	expiredTime time.Time, cm *admittance.CAManager) *EventHub {
	protocolManager := NewEventHub(namespace, executor, peerManager, eventMux, consenter, am, exist, expiredTime)
	aliveChan := make(chan int)
	protocolManager.Start(aliveChan, cm)

	return protocolManager
}
