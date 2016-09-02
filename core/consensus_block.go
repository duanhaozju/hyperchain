package core

import (
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/common"
)

type ConsensusBlock struct {
	ParentHash   []byte
	BlockHash    []byte
	Transactions []*types.Transaction
	Timestamp    int64
	MerkleRoot   []byte
	Number       int64
	Now          int64
	Pre          int64
}



func (self ConsensusBlock)Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
}
