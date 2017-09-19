package hts

import (
	"crypto"
	"io"
)

type Crypto interface {
	GeneratePriKey(io.Reader) (priKey, pubKey interface{}, err error)
	NegoSharedKey(remotePubkey, extra []byte) ([]byte, error)
	VeryCert(cert, rootCert []byte) (bool, error)
	SymmetricEnc(secKey, msg []byte) (encMsg []byte, err error)
	SymmetricDec(secKey, encMsg []byte) (decMsg []byte, err error)
}

type CryptoImpl struct {
	symmetricEncAlgo func(secKey, msg []byte) (encMsg []byte, err error)
	symmetricDecAlgo func(secKey, msg []byte) (decMsg []byte, err error)
	signAlgo         func(pri crypto.PrivateKey, msg []byte) (signedMsg []byte, err error)
	veriAlgo         func(pub crypto.PublicKey, signature, msg []byte) (result bool, err error)
	veriCertAlgo     func(cert []byte, want []byte) (result bool, err error)
}

func (ciml *CryptoImpl) SetSymmEnc(foo func(secKey, msg []byte) ([]byte, error)) {
	ciml.symmetricEncAlgo = foo
}

func (ciml *CryptoImpl) SetSymmDec(foo func(secKey, msg []byte) ([]byte, error)) {
	ciml.symmetricEncAlgo = foo
}

func (ciml *CryptoImpl) SetSign(foo func(pri crypto.PrivateKey, msg []byte) ([]byte, error)) {
	ciml.signAlgo = foo
}

func (ciml *CryptoImpl) SetVerify(foo func(pub crypto.PublicKey, signature, msg []byte) (bool, error)) {
	ciml.veriAlgo = foo
}
