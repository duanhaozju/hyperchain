package manager

import (
	"hyperchain-alpha/event"
	"hyperchain-alpha/p2p"
	"hyperchain-alpha/core"
)

func New(eventMux *event.TypeMux, peerManager p2p.PeerManager) (error) {

	//var wg sync.WaitGroup

	peerManager.Start()

	allAlive := peerManager.JudgeAlivePeers()

	if allAlive {

		fetcher := core.NewFetcher()
		protocolManager := NewProtocolManager(eventMux, peerManager , fetcher)

		protocolManager.Start()

	}

	return nil
}


