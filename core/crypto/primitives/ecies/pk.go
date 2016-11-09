//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ecies

import (
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"io"

	"hyperchain/core/crypto/primitives"
)

type publicKeyImpl struct {
	pub    *ecdsa.PublicKey
	rand   io.Reader
	params *Params
}

func (pk *publicKeyImpl) GetRand() io.Reader {
	return pk.rand
}

func (pk *publicKeyImpl) IsPublic() bool {
	return true
}

type publicKeySerializerImpl struct{}

func (pks *publicKeySerializerImpl) ToBytes(key interface{}) ([]byte, error) {
	if key == nil {
		return nil, primitives.ErrInvalidNilKeyParameter
	}

	switch pk := key.(type) {
	case *publicKeyImpl:
		return x509.MarshalPKIXPublicKey(pk.pub)
	default:
		return nil, primitives.ErrInvalidPublicKeyType
	}
}

func (pks *publicKeySerializerImpl) FromBytes(bytes []byte) (interface{}, error) {
	key, err := x509.ParsePKIXPublicKey(bytes)
	if err != nil {
		return nil, err
	}

	// TODO: add params here
	return &publicKeyImpl{key.(*ecdsa.PublicKey), rand.Reader, nil}, nil
}
