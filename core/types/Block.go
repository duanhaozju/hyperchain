package types

import (
	"hyperchain-alpha/core/node"
)

type Block struct {
	ParentHash string
	BlockHash string
	Transactions []Transaction
	TimeStramp int64 //unix时间戳
	CoinBase node.Node // 打包该Block的地址
	MerkleRoot string // merkleRoot 的hash值
}

