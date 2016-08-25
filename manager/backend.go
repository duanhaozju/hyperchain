package manager

import (
	"hyperchain-alpha/event"

	"hyperchain-alpha/p2p"

	"hyperchain-alpha/node"

	"hyperchain-alpha/core"

	"hyperchain-alpha/crypto"
)

func New(eventMux *event.TypeMux, peerManager p2p.PeerManager, node node.Node, enSign crypto.Encryption) (error) {

	//var wg sync.WaitGroup


	peerManager.Start()
	node.Start()
	enSign.Sign(tx)

	allAlive := peerManager.JudgeAlivePeers()

	if allAlive {

		fetcher := core.NewFetcher()
		protocolManager := NewProtocolManager(eventMux, peerManager, node, fetcher)

		protocolManager.Start()

	}


	return nil
}


