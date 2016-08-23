package event

import (

	"math/big"

	"hyperchain-alpha/common"
	"hyperchain-alpha/core/types"

)


type ConsensusEvent struct{ Msg *types.Msg }

// TxPreEvent is posted when a transaction enters the transaction pool.
type TxPreEvent struct{ Tx *types.Transaction }

// TxPostEvent is posted when a transaction has been processed.
type TxPostEvent struct{ Tx *types.Transaction }



// PendingStateEvent is posted pre mining and notifies of pending state changes.
type PendingStateEvent struct{}

// NewBlockEvent is posted when a block has been imported.
type NewBlockEvent struct{ Block *types.Block }

// NewMinedBlockEvent is posted when a block has been imported.
type NewMinedBlockEvent struct{ Block *types.Block }

// RemovedTransactionEvent is posted when a reorg happens
type RemovedTransactionEvent struct{ Txs types.Transactions }

//// RemovedLogEvent is posted when a reorg happens
//type RemovedLogsEvent struct{ Logs vm.Logs }

// ChainSplit is posted when a new head is detected
type ChainSplitEvent struct {
	Block *types.Block
	//Logs  vm.Logs
}

type ChainEvent struct {
	Block *types.Block
	Hash  common.Hash
	//Logs  vm.Logs
}

type ChainSideEvent struct {
	Block *types.Block
	//Logs  vm.Logs
}

type PendingBlockEvent struct {
	Block *types.Block
	//Logs  vm.Logs
}

type ChainUncleEvent struct {
	Block *types.Block
}

type ChainHeadEvent struct{ Block *types.Block }

type GasPriceChanged struct{ Price *big.Int }

// Mining operation events
type StartMining struct{}
type TopMining struct{}
