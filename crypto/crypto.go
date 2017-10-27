//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"crypto/sha256"
	"github.com/ethereum/go-ethereum/rlp"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/secp256k1"
	"github.com/hyperchain/hyperchain/crypto/sha3"
	"golang.org/x/crypto/ripemd160"
	"math/big"
)

func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Creates an hyperchain address given the bytes and the nonce
func CreateAddress(addr common.Address, nonce uint64) common.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{addr, nonce})
	return common.BytesToAddress(Keccak256(data)[12:])
}

func Sha256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}

func Ripemd160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)

	return ripemd.Sum(nil)
}

func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	vint := uint32(v)
	// reject upper range of s values (ECDSA malleability)
	// see discussion in secp256k1/libsecp256k1/include/secp256k1.h
	if homestead && s.Cmp(secp256k1.HalfN) > 0 {
		return false
	}
	// Frontier: allow s to be in full N range
	if s.Cmp(secp256k1.N) >= 0 {
		return false
	}
	if r.Cmp(secp256k1.N) < 0 && (vint == 27 || vint == 28) {
		return true
	} else {
		return false
	}
}
