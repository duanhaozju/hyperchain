// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package bucket

import (
	"crypto/sha1"
	"hash"
	"hyperchain/crypto"
	"sync"
)

const (
	// Lowest level in hash tree - bucket type node.
	BucketType byte = 1 << iota
	// lower level in hash tree - merkle type node.
	MerkleLow
	// higer level in hash tree - merkle type node.
	MerkleHigh
)

// Hasher is not concurrent safe here.
// You got to required a lock to keep safe by yourself.
type Hasher struct {
	buffer []byte
	typ    byte
	hash   hash.Hash
	keccak *crypto.KeccakHash
	trim   int
}

// Hash calculate a hash for the given data.
// Different hasher type use different hash function.
// If the hasher is a bucket hasher, it uses the simplest sha1 hashfn,
// Otherwise, it uses the kecca256 hashfn.
func (hasher *Hasher) Hash(data []byte) []byte {
	switch hasher.typ {
	case MerkleLow:
		if data == nil || len(data) == 0 {
			return nil
		}
		hasher.hash.Reset()
		hasher.hash.Write(data)
		return hasher.hash.Sum(nil)[:hasher.trim]
	case MerkleHigh:
		return hasher.keccak.Hash(data).Bytes()
	case BucketType:
		hasher.hash.Reset()
		hasher.hash.Write(data)
		return hasher.hash.Sum(nil)[:hasher.trim]
	default:
		// No defined hasher type
		return nil
	}
}

// hashers live in a global pool.
var hasherPool = sync.Pool{
	New: func() interface{} {
		return &Hasher{buffer: nil, hash: sha1.New(), keccak: crypto.NewKeccak256Hash("keccak256")}
	},
}

func newHasher(trim int, typ byte) *Hasher {
	h := hasherPool.Get().(*Hasher)
	h.trim, h.typ = trim, typ
	return h
}

func returnHasher(h *Hasher) {
	hasherPool.Put(h)
}
