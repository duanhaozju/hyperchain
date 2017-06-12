//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package message

import (
	"encoding/hex"
	"hyperchain/crypto"
)

func GetHash(needHashString string) string {
	hasher := crypto.NewKeccak256Hash("keccak256Haner")
	return hex.EncodeToString(hasher.ByteHash([]byte(needHashString)).Bytes())
}
