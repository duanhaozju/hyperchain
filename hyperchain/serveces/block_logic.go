package serveces

import (
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core"
	"hyperchain-alpha/encrypt"
	"hyperchain-alpha/core/types"
	"go/types"
)

func NewBlock(trans types.Transactions,node node.Node,timestamp int64) *types.Block{


	var block = &types.Block{
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

func (b *types.Block) Hash() string{

	return string(encrypt.GetHash([]byte(b.ParentHash + b.Transactions + b.TimeStramp + b.CoinBase + b.MerkleRoot)))
}

func (b *types.Block) UpdateLastestBlockHS() {
	core.UpdateChain(b.BlockHash)
}

func (b *types.Block) SubmitBlock(blockHash string) {
	core.PutBlockToLDB(blockHash,b)
}

