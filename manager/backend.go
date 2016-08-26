// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-24
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
		protocolManager := NewProtocolManager(eventMux, peerManager, fetcher)

		protocolManager.Start()

	}

	return nil
}


