package core

import (
	"github.com/ethereum/go-ethereum/miner"
	"github.com/ethereum/go-ethereum/common/httpclient"
	"github.com/ethereum/ethash"
	"github.com/ethereum/go-ethereum/accounts"

	"sync"
	"github.com/ethereum/go-ethereum/ethdb"
	"hyperchain-alpha/event"
	"hyperchain-alpha/common"
	"hyperchain-alpha/manager"

	"github.com/ethereum/go-ethereum/core"
	"hyperchain-alpha/transaction"
	"hyperchain-alpha/config"
)

type Ethereum struct {

			       // Channel for shutting down the ethereum
	shutdownChan chan bool

			       // DB interfaces
	chainDb ethdb.Database // Block chain database


			       // Handlers
	txPool          *core.TxPool
	txMu            sync.Mutex
	blockchain      *core.BlockChain
	accountManager  *accounts.Manager
	pow             *ethash.Ethash
	protocolManager *manager.ProtocolManager



	httpclient *httpclient.HTTPClient

	eventMux *event.TypeMux
	miner    *miner.Miner

	Mining        bool
	MinerThreads  int
	autodagquit   chan bool
	etherbase     common.Address
	netVersionId  int

}


func New(eventMux *event.TypeMux) (*Ethereum, error) {
	eth := &Ethereum{
		shutdownChan:            make(chan bool),

	}

	newPool := transaction.NewTxPool(eventMux)
	eth.txPool = newPool
	if eth.protocolManager, err = manager.NewProtocolManager(eth.eventMux, eth.txPool); err != nil {
		return nil, err
	}
	eth.protocolManager.Start()
	eth.miner = miner.New(eth,  eth.eventMux)
	return eth, nil
}