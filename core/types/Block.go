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
	coinBase node.Node // 打包该Block的地址
	merkleRoot string // merkleRoot 的hash值
}

