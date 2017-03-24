//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package manager

import (
	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/consensus"
	"hyperchain/core/executor"
	"hyperchain/manager/event"
	"hyperchain/p2p"
)

// init protocol manager params and start
func New(namespace string, eventMux *event.TypeMux, executor *executor.Executor, peerManager p2p.PeerManager, consenter consensus.Consenter, am *accounts.AccountManager, cm *admittance.CAManager) *EventHub {
	eventHub := NewEventHub(namespace, executor, peerManager, eventMux, consenter, am)
	aliveChan := make(chan int)
	eventHub.Start(aliveChan, cm)
	return eventHub
}
