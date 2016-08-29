package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
)

func (self Transaction)Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
}
