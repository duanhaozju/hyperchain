//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ecdh

import (
	"crypto"
	"io"
)

// ECDH is the main interface for ECDH key exchange.
type ECDH interface {
	GenerateKey(io.Reader) (crypto.PrivateKey, crypto.PublicKey, error)
	Marshal(crypto.PublicKey) []byte
	Unmarshal([]byte) (crypto.PublicKey, bool)
	GenerateSharedSecret(crypto.PrivateKey, crypto.PublicKey) ([]byte, error)
}
