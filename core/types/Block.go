package types

import (
	"time"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/encrypt"
)

type Block struct {
	ParentHash string
	BlockHash string
	Transactions []Transaction
	TimeStramp time.Time
	CoinBase node.Node // 打包该Block的地址
	MerkleRoot string // merkleRoot 的hash值
}

// TODO 构造一个新的区块
func NewBlock(trans Transactions,node node.Node,timestamp time.Time) *Block{


	return &Block{
		ParentHash: "",
		BlockHash: ,
		Transactions: trans,
		CoinBase: node,
		TimeStramp: timestamp,
	}
}

func (b *Block) Hash() string{
	return encrypt.GetHash([]byte(b.ParentHash + b.Transactions + b.TimeStramp + b.CoinBase + b.MerkleRoot))
}
