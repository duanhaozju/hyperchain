// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package manager

import (

	"hyperchain/p2p"
	"hyperchain/core"
	"hyperchain/consensus"
	"hyperchain/crypto"
)

// init protocol manager params and start
func New(peerManager p2p.PeerManager,consenter consensus.Consenter,fetcher *core.Fetcher,encryption crypto.Encryption ,path string, isFirst bool) (error) {



	aliveChan := make(chan bool)
	peerManager.Start(path, isFirst,aliveChan)


	//peerManager.JudgeAlivePeers()
	//wait for all peer are connected
	select {

	case <-aliveChan:


		protocolManager := NewProtocolManager( peerManager, fetcher,consenter,encryption)

		protocolManager.Start()

	}

	return nil
}


