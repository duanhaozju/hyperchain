//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"encoding/json"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/sha3"
)

//Keccak256Hash is a kind of hash method which implements CommomHash interface
type KeccakHash struct {
	name string
}

func NewKeccak256Hash(name string) *KeccakHash {
	s256 := &KeccakHash{name: name}
	return s256
}

//Hash transfers object x into common.Hash with length 32
func (k256 *KeccakHash) Hash(x interface{}) (h common.Hash) {
	serialize_data, err := json.Marshal(x)

	if err != nil {
		panic(err)
	}
	hw := sha3.NewKeccak256()
	hw.Write(serialize_data)
	hw.Sum(h[:0])

	return h
}

//ByteHash deals with params which has already been []byte
func (k256 *KeccakHash) ByteHash(data ...[]byte) (h common.Hash) {

	hw := sha3.NewKeccak256()
	for _, d := range data {
		hw.Write(d)
	}
	hw.Sum(h[:0])

	return h
}
