package types

import (
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/encrypt"
	"hyperchain-alpha/core"
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
// TODO 构造一个新的区块
func NewBlock(trans Transactions,node node.Node,timestamp int64) *Block{

	//TODO 得到最新区块HASH

	var block = &Block{
		ParentHash: core.GetLashestBlockHash(),
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

func (b *Block) UpdateLastestBlockHS() {
	core.UpdateChain(b.BlockHash)
}

func (b *Block) SubmitBlock(blockHash string) {
	core.PutBlockToLDB(blockHash,b)
}
