package types

import (
	"hyperchain-alpha/core/node"
)

type Block struct {
	ParentHash string
	BlockHash string
	Transactions []Transaction
	TimeStramp int64
	CoinBase node.Node  // 打包该Block的地址
	MerkleRoot string   // merkleRoot 的hash值
}

//todo TimeStramp 写错了 应该该掉
