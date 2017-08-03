//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ecies

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"io"

	"hyperchain/crypto/primitives"
)

type secretKeyImpl struct {
	priv   *ecdsa.PrivateKey
	pub    primitives.PublicKey
	params *Params
	rand   io.Reader
}

func (sk *secretKeyImpl) IsPublic() bool {
	return false
}

func (sk *secretKeyImpl) GetRand() io.Reader {
	return sk.rand
}

func (sk *secretKeyImpl) GetPublicKey() primitives.PublicKey {
	if sk.pub == nil {
		sk.pub = &publicKeyImpl{&sk.priv.PublicKey, sk.rand, sk.params}
	}
	return sk.pub
}

type secretKeySerializerImpl struct{}

func (sks *secretKeySerializerImpl) ToBytes(key interface{}) ([]byte, error) {
	if key == nil {
		return nil, primitives.ErrInvalidNilKeyParameter
	}

	switch sk := key.(type) {
	case *secretKeyImpl:
		return x509.MarshalECPrivateKey(sk.priv)
	default:
		return nil, primitives.ErrInvalidKeyParameter
	}
}

func (sks *secretKeySerializerImpl) FromBytes(bytes []byte) (interface{}, error) {
	key, err := x509.ParseECPrivateKey(bytes)
	if err != nil {
		return nil, err
	}

	// TODO: add params here
	return &secretKeyImpl{key, nil, nil, rand.Reader}, nil
}
