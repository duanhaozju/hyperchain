package types

import (
	"time"
	"hyperchain-alpha/core/node"
)

type Block struct {
	ParentHash string
	BlockHash string
	Transactions []Transaction
	TimeStramp time.Time
	CoinBase node.Node // 打包该Block的地址
	MerkleRoot string // merkleRoot 的hash值
}

