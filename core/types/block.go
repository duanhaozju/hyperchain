package types

import (
	"hyperchain/common"
	"hyperchain/crypto"
)

func (self Block)Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
}
