//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package primitives

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/sha256"
	"encoding/asn1"
	hcrypto "github.com/hyperchain/hyperchain/crypto"
	"math/big"
)

// ECDSASignature represents an ECDSA signature.
type ECDSASignature struct {
	R, S *big.Int
}

// NewECDSAKey generates a new ECDSA Key.
func NewECDSAKey() (*ecdsa.PrivateKey, error) {
	return ecdsa.GenerateKey(GetDefaultCurve(), rand.Reader)
}

// ECDSASign sign a msg by keccak256 hasher.
func ECDSASign(signKey interface{}, msg []byte) ([]byte, error) {
	temp := signKey.(*ecdsa.PrivateKey)

	hasher := hcrypto.NewKeccak256Hash("keccak256Hasher")
	h := hasher.Hash(msg).Bytes()

	r, s, err := ecdsa.Sign(rand.Reader, temp, h)
	if err != nil {
		return nil, err
	}

	raw, err := asn1.Marshal(ECDSASignature{r, s})
	if err != nil {
		return nil, err
	}

	return raw, nil
}

// ECDSAVerify verifies signature by keccak256 hasher.
func ECDSAVerify(verKey interface{}, msg, signature []byte) (bool, error) {
	ecdsaSignature := new(ECDSASignature)
	_, err := asn1.Unmarshal(signature, ecdsaSignature)
	if err != nil {
		return false, err
	}
	temp := verKey.(ecdsa.PublicKey)
	hasher := hcrypto.NewKeccak256Hash("keccak256Hanse")
	h := hasher.Hash(msg).Bytes()
	//h := Hash(msg)
	return ecdsa.Verify(&temp, h, ecdsaSignature.R, ecdsaSignature.S), nil
}

// ECDSAVerifyTransport verify siganture from sdk by sha256 hasher.
func ECDSAVerifyTransport(verKey *ecdsa.PublicKey, msg, signature []byte) (bool, error) {
	ecdsaSignature := new(ECDSASignature)
	_, err := asn1.Unmarshal(signature, ecdsaSignature)
	if err != nil {
		return false, err
	}

	h := sha256.New()
	digest := make([]byte, 32)
	h.Write(msg)
	h.Sum(digest[:0])
	return ecdsa.Verify(verKey, digest, ecdsaSignature.R, ecdsaSignature.S), nil
}
