package trie

import (
	"bytes"
	"hash"
	"hyperchain/crypto/sha3"
	"sync"
)

type hashCalculator struct {
	tmp *bytes.Buffer
	sha hash.Hash
}

// hashers live in a global pool.
var hasherPool = sync.Pool{
	New: func() interface{} {
		return &hashCalculator{tmp: new(bytes.Buffer), sha: sha3.NewKeccak256()}
	},
}

func newHashCalculator() *hashCalculator {
	h := hasherPool.Get().(*hashCalculator)
	return h
}

func returnHasherToPool(h *hashCalculator) {
	hasherPool.Put(h)
}
