//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package ecies

import (
	"crypto/elliptic"
	"fmt"
	"io"

	"hyperchain/core/crypto/primitives"
)

type keyGeneratorParameterImpl struct {
	rand   io.Reader
	curve  elliptic.Curve
	params *Params
}

type keyGeneratorImpl struct {
	isForEncryption bool
	params          *keyGeneratorParameterImpl
}

func (kgp keyGeneratorParameterImpl) GetRand() io.Reader {
	return kgp.rand
}

func (kg *keyGeneratorImpl) Init(params primitives.KeyGeneratorParameters) error {
	if params == nil {
		return primitives.ErrInvalidKeyGeneratorParameter
	}
	switch kgparams := params.(type) {
	case *keyGeneratorParameterImpl:
		kg.params = kgparams
	default:
		return primitives.ErrInvalidKeyGeneratorParameter
	}

	return nil
}

func (kg *keyGeneratorImpl) GenerateKey() (primitives.PrivateKey, error) {
	if kg.params == nil {
		return nil, fmt.Errorf("Key Generator not initliazed")
	}

	privKey, err := eciesGenerateKey(
		kg.params.rand,
		kg.params.curve,
		kg.params.params,
	)
	if err != nil {
		return nil, err
	}

	return &secretKeyImpl{privKey, nil, kg.params.params, kg.params.rand}, nil
}
