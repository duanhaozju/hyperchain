//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package sha3

// This file provides functions for creating instances of the SHA-3
// and SHAKE hash functions, as well as utility functions for hashing
// bytes.

import (
	"hash"
)

// NewKeccak256 creates a new Keccak-256 hash.
func NewKeccak256() hash.Hash { return &state{rate: 136, outputLen: 32, dsbyte: 0x01} }
