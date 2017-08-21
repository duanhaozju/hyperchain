//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import "hyperchain/core/types"

// TxHashBatch contains transactions that batched by primary.
type TxHashBatch struct {
	BatchHash  string
	TxHashList []string
	TxList     []*types.Transaction
}

type MissingTxHashList struct {
	TxHashList	[]string
	BatchHash	string
}

type ReturnFetchTxs struct {
	BatchHash		string
	ReturnedFetchTxs	[]*types.Transaction
}