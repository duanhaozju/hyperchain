//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package message

import (
	"hyperchain/crypto/sha3"
	"hyperchain/common"
)

func GetHash(needHashString string) string {
	hasher := sha3.NewKeccak256()
	hasher.Write([]byte(needHashString))
	hash := hasher.Sum(nil)
	return common.ToHex(hash)
}
