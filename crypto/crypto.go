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

// Keccak256Hash hash data by keccak256 hasher.
func Keccak256Hash(data ...[]byte) (h common.Hash) {
	d := sha3.NewKeccak256()
	for _, b := range data {
		d.Write(b)
	}
	d.Sum(h[:0])
	return h
}

// Creates an hyperchain address given the bytes and the nonce.
func CreateAddress(addr common.Address, nonce uint64) common.Address {
	data, _ := rlp.EncodeToBytes([]interface{}{addr, nonce})
	return common.BytesToAddress(Keccak256(data)[12:])
}

// Sha256 hash data by Sha256 hasher.
func Sha256(data []byte) []byte {
	hash := sha256.Sum256(data)

	return hash[:]
}

// Ripemd160 hash data by Ripemd160 hasher.
func Ripemd160(data []byte) []byte {
	ripemd := ripemd160.New()
	ripemd.Write(data)

	return ripemd.Sum(nil)
}

// Ecrecover recover public key using msg and digest in secp256k1.
func Ecrecover(hash, sig []byte) ([]byte, error) {
	return secp256k1.RecoverPubkey(hash, sig)
}

// ValidateSignatureValues determine whether the signature is valid.
func ValidateSignatureValues(v byte, r, s *big.Int, homestead bool) bool {
	if r.Cmp(common.Big1) < 0 || s.Cmp(common.Big1) < 0 {
		return false
	}
	vint := uint32(v)

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