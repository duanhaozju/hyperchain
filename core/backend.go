package core

import (



	"sync"

	"hyperchain-alpha/event"
	"hyperchain-alpha/common"
	"hyperchain-alpha/manager"


	"hyperchain-alpha/transaction"

	"hyperchain-alpha/block"

)

type Ethereum struct {

			       // Channel for shutting down the ethereum
	shutdownChan chan bool

			       // DB interfaces
	//chainDb ethdb.Database // Block chain database


			       // Handlers
	txPool          *transaction.TxPool
	txMu            sync.Mutex
	//blockchain      *core.BlockChain
	//accountManager  *accounts.Manager
	//pow             *ethash.Ethash
	protocolManager *manager.ProtocolManager



	//httpclient *httpclient.HTTPClient

	eventMux *event.TypeMux
	blockMaker    *block.BlockMaker

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
	eth.blockMaker =block.New(eth,  eth.eventMux)
	return eth, nil
}


