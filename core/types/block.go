package types

import (
	"hyperchain/common"
	"hyperchain/crypto"
)

func (self Block)Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash([]interface{}{
		self.ParentHash,
		self.Number,
		self.Timestamp,
		self.TxRoot,
		self.ReceiptRoot,
		self.MerkleRoot,
	})
}

func (self *Block)HashBlock(ch crypto.CommonHash) common.Hash {
	return ch.Hash([]interface{}{
		self.ParentHash,
		self.Number,
		self.Timestamp,
		self.TxRoot,
		self.ReceiptRoot,
		self.MerkleRoot,
	})
}
