package vm

import "hyperchain/common"

type Log interface {
	// getter
	GetType() string
	GetTxHash() common.Hash
	GetBlockHash() common.Hash
	GetTxIndex() uint
	GetBlockNumber() uint64
	GetIndex() uint
	// setter
	SetTxHash(common.Hash)
	SetBlockHash(common.Hash)
	SetTxIndex(uint)
	SetBlockNumber(uint64)
	SetIndex(uint)
}

type Logs []Log
