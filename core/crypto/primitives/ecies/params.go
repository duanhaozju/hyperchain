//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ecies

import (
	"crypto"
	"crypto/cipher"
	"hash"
)

// Params ECIES parameters
type Params struct {
	Hash      func() hash.Hash
	hashAlgo  crypto.Hash
	Cipher    func([]byte) (cipher.Block, error)
	BlockSize int
	KeyLen    int
}
