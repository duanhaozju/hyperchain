package types

import (
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/encrypt"
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

func NewBlock(trans Transactions,node node.Node,timestamp int64,parentHash string) *Block{


	var block = &Block{
		ParentHash: parentHash,
		Transactions: trans,
		CoinBase: node,
		TimeStramp: timestamp,
		MerkleRoot: "test",
	}

	bHash := block.Hash()

	block.BlockHash = bHash

	return block
}

func (b *Block) Hash() string{
	return string(encrypt.GetHash([]byte(b.ParentHash + b.Transactions + b.TimeStramp + b.CoinBase + b.MerkleRoot)))
}

