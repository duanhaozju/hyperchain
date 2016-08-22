package manager

import (
	"github.com/ethereum/go-ethereum/core/types"

	"time"

	"hyperchain-alpha/block"
)
const (
	forceSyncCycle      = 10 * time.Second // Time interval to force syncs, even if few peers are available


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
	// Wait for different events to fire synchronisation operations
	forceSync := time.Tick(forceSyncCycle)
	for {
		select {
		case <-pm.newPeerCh:
		// Make sure we have peers to select from, then sync

			go pm.BlockMaker.Start()

		case <-forceSync:
		// Force a sync even if not enough peers are present
			go pm.BlockMaker.SyncWithPeer()

		case <-pm.noMorePeers:
			return
		}
	}

}




//跟节点握手后,同步tx
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

