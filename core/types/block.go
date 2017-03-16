package types

import (
	"hyperchain/common"
	"hyperchain/crypto"
)

func (self *Block) Hash() common.Hash {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	return kec256Hash.Hash([]interface{}{
		self.ParentHash,
		self.Number,
		self.Timestamp,
		self.TxRoot,
		self.ReceiptRoot,
		self.MerkleRoot,
	})
}

