package manager

import (
	"hyperchain-alpha/event"

	"hyperchain-alpha/p2p"

	"hyperchain-alpha/node"

	"hyperchain-alpha/core"

)

func New(eventMux *event.TypeMux, peerManager p2p.PeerManager, node node.Node) (error) {

	peerManager.Start()
	node.Start()

	allAlive := peerManager.JudgeAlivePeers()

	if allAlive {

		fetcher := core.NewFetcher()
		protocolManager := NewProtocolManager(eventMux, peerManager, node, fetcher)

		protocolManager.Start()
	}

	return nil
}


