//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package message

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/sha3"
)

func GetHash(needHashString string) string {
	hasher := sha3.NewKeccak256()
	hasher.Write([]byte(needHashString))
	hash := hasher.Sum(nil)
	return common.ToHex(hash)
}
