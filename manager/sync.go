package manager

import (
	"github.com/ethereum/go-ethereum/core/types"

)

type txsync struct {
	p   *peer
	txs []*types.Transaction
}

//接受从handle中传来的同步tx消息
func (pm *ProtocolManager) txsyncLoop() {


	for {
		select {
		case s := <-pm.txsyncCh:
			pm.txpool.AddTransactions(s.txs)
		case <-pm.quitSync:
			return
		}
	}

}
//接受从handle中传来的同步区块消息
func (pm *ProtocolManager) syncer() {

	// Start and ensure cleanup of sync mechanisms
	pm.fetcher.Start()

}



func (pm *ProtocolManager) syncTransactions(p *peer) {
	txs := pm.txpool.GetTransactions()
	if len(txs) == 0 {
		return
	}
	select {
	case pm.txsyncCh <- &txsync{p, txs}:
	case <-pm.quitSync:
	}
}