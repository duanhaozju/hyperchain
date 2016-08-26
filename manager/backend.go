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

func New(eventMux *event.TypeMux, peerManager p2p.PeerManager, path string, isFirst bool) (error) {

	//var wg sync.WaitGroup

	aliveChan := make(chan bool)
	peerManager.Start(path, isFirst,aliveChan)



	//peerManager.JudgeAlivePeers()
	select {

	case <-aliveChan:

		fetcher := core.NewFetcher()
		protocolManager := NewProtocolManager(eventMux, peerManager, fetcher)

		protocolManager.Start()

	}

	return nil
}


