//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ecies

import (
	"hyperchain/core/crypto/primitives"
	"hyperchain/core/crypto/utils"
)

type encryptionSchemeImpl struct {
	isForEncryption bool

	// Parameters
	params primitives.AsymmetricCipherParameters
	pub    *publicKeyImpl
	priv   *secretKeyImpl
}

func (es *encryptionSchemeImpl) Init(params primitives.AsymmetricCipherParameters) error {
	if params == nil {
		return primitives.ErrInvalidNilKeyParameter
	}
	es.isForEncryption = params.IsPublic()
	es.params = params

	if es.isForEncryption {
		switch pk := params.(type) {
		case *publicKeyImpl:
			es.pub = pk
		default:
			return primitives.ErrInvalidPublicKeyType
		}
	} else {
		switch sk := params.(type) {
		case *secretKeyImpl:
			es.priv = sk
		default:
			return primitives.ErrInvalidKeyParameter
		}
	}

	return nil
}

func (es *encryptionSchemeImpl) Process(msg []byte) ([]byte, error) {
	if len(msg) == 0 {
		return nil, utils.ErrNilArgument
	}

	if es.isForEncryption {
		// Encrypt
		return eciesEncrypt(es.params.GetRand(), es.pub.pub, nil, nil, msg)
	}

	// Decrypt
	return eciesDecrypt(es.priv.priv, nil, nil, msg)
}
