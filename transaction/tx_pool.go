package transaction

import (
	"math/big"


	"hyperchain-alpha/common"
	"sync"

	"hyperchain-alpha/event"

	"hyperchain-alpha/core/types"
)

type txPool interface {
	// AddTransactions should add the given transactions to the pool.
	AddTransactions([]*types.Transaction)

	// GetTransactions should return pending transactions.
	// The slice should be modifiable by the caller.
	GetTransactions() types.Transactions
}

type TxPool struct {



	gasLimit     func() *big.Int // The current gas limit function callback
	minGasPrice  *big.Int
	eventMux     *event.TypeMux
	events       event.Subscription

	mu           sync.RWMutex
	pending      map[common.Hash]*types.Transaction // processable transactions
	queue        map[common.Address]map[common.Hash]*types.Transaction

	wg sync.WaitGroup // for shutdown sync

	homestead bool
}

//new 一个tx pool,然后循环监听中的事件
func NewTxPool( eventMux *event.TypeMux) *TxPool {
	pool := &TxPool{

		pending:      make(map[common.Hash]*types.Transaction),
		queue:        make(map[common.Address]map[common.Hash]*types.Transaction),
		eventMux:     eventMux,


		minGasPrice:  new(big.Int),

		events:       eventMux.Subscribe(event.ChainHeadEvent{},  event.RemovedTransactionEvent{}),
	}

	pool.wg.Add(1)
	go pool.eventLoop()

	return pool
}

func (pool *TxPool) eventLoop() {
	defer pool.wg.Done()

	// Track chain events. When a chain events occurs (new chain canon block)
	// we need to know the new state. The new state will help us determine
	// the nonces in the managed state
	for ev := range pool.events.Chan() {
		switch ev := ev.Data.(type) {
		case event.ChainHeadEvent:
			pool.mu.Lock()
			//TODO
			//pool.resetState()
			pool.mu.Unlock()

		case event.RemovedTransactionEvent:
			pool.AddTransactions(ev.Txs)
		}
	}
}


func (self *TxPool) AddTransactions(txs []*types.Transaction) {
	self.mu.Lock()
	defer self.mu.Unlock()

	for _, tx := range txs {

		//self.pending=tx
		//TODO添加tx到pool中
		go self.eventMux.Post(event.TxPreEvent{tx})

	}

}



func (self *TxPool) GetTransactions() (txs types.Transactions) {
	self.mu.Lock()
	defer self.mu.Unlock()

	// check queue first
	//self.checkQueue()
	//// invalidate any txs
	//self.validatePool()

	txs = make(types.Transactions, len(self.pending))
	i := 0
	for _, tx := range self.pending {
		txs[i] = tx
		i++
	}
	return txs
}
