//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import "github.com/hyperchain/hyperchain/common"

// hash interface
type CommonHash interface {
	Hash(x interface{}) (h common.Hash)
	ByteHash(data ...[]byte) (h common.Hash)
}
