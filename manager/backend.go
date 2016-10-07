// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-29
package manager

import (
	"hyperchain/p2p"
	"hyperchain/core"
	"hyperchain/consensus"
	"hyperchain/crypto"

	"hyperchain/event"
	"hyperchain/accounts"

)

// init protocol manager params and start
func New(eventMux *event.TypeMux, blockPool *core.BlockPool, peerManager p2p.PeerManager, consenter consensus.Consenter, fetcher *core.Fetcher,
//encryption crypto.Encryption, commonHash crypto.CommonHash,path string, nodeId int) (error) {
am *accounts.AccountManager, commonHash crypto.CommonHash,path string, nodeId int) (*ProtocolManager) {

	aliveChan := make(chan bool)
	go peerManager.Start(path,nodeId, aliveChan, false, eventMux)

	//wait for all peer are connected
	select {
	case <-aliveChan:
		{

			//protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, fetcher, consenter, encryption, commonHash)
			protocolManager := NewProtocolManager(blockPool, peerManager, eventMux, fetcher, consenter, am, commonHash)
			protocolManager.Start()
			//start server
			return protocolManager


		}
	}

	//return nil
}


