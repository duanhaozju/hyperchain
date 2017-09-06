package hts

import (
	"crypto/x509"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto"
	"math/big"
	"encoding/asn1"
	"github.com/pkg/errors"
	"hyperchain/crypto/primitives"
)

type CertGroup struct {
	enableEnroll bool
	sign 	     bool

	//ECert group
	eCA          []byte
	eCA_S        *x509.Certificate
	eCERT        []byte
	eCERT_S      *x509.Certificate
	eCERTPriv    []byte
	eCERTPriv_S  *ecdsa.PrivateKey

	//RCert group
	rCA          []byte
	rCA_S        *x509.Certificate
	rCERT        []byte
	rCERT_S      *x509.Certificate
	rCERTPriv    []byte
	rCERTPriv_S  *ecdsa.PrivateKey
}

func (cg *CertGroup)GetECert() []byte {
	return cg.eCERT
}

func (cg *CertGroup)GetRCert() []byte {
	return cg.rCERT
}

func (cg *CertGroup)ESign(data []byte) ([]byte, error) {
	if !cg.sign {
		return nil,nil
	}
	return cg.eCERTPriv_S.Sign(rand.Reader, data, crypto.SHA3_256)
}

func (cg *CertGroup)EVerify(rawcert, sign []byte, hash []byte) (bool, error) {
	if !cg.sign {
		return true,nil
	}
	x509cert, err := primitives.ParseCertificate(rawcert)
	if err != nil {
		return false, errors.New("parse the certificate failed (E verify)")
	}
	pubkey, ok := x509cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("type assert for publick key failed (E verify)")
	}
	ecdsasign := struct {
		R, S *big.Int
	}{}
	_, err = asn1.Unmarshal(sign, &ecdsasign)
	if err != nil {
		return false, errors.New("unmarshal the signature failed. (E verify)")
	}

	b := ecdsa.Verify(pubkey, hash, ecdsasign.R, ecdsasign.S)

	//验证父证书
	err = x509cert.CheckSignatureFrom(cg.eCA_S)
	if err != nil {
		return false, errors.New("verified the parent cert failed!")
	}

	if b {
		return true, nil
	}
	return false, errors.New("verify failed, signature verify not passed.")
}

func (cg *CertGroup)RSign(data []byte) ([]byte, error) {
	if !cg.sign  || !cg.enableEnroll {
		return nil,nil
	}
	hash := cg.Hash(data)
	return cg.rCERTPriv_S.Sign(rand.Reader, hash, crypto.SHA3_256)
}

func (cg *CertGroup)RVerify(rawcert, sign []byte, data []byte) (bool, error) {
	if !cg.sign  || !cg.enableEnroll {
		return true,nil
	}
	x509cert, err := primitives.ParseCertificate(rawcert)
	if err != nil {
		return false, errors.New("parse the certificate failed (R verify)")
	}
	pubkey, ok := x509cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New("type assert for publick key failed (R verify)")
	}
	ecdsasign := struct {
		R, S *big.Int
	}{}
	_, err = asn1.Unmarshal(sign, &ecdsasign)
	if err != nil {
		return false, errors.New("unmarshal the signature failed. (R verify)")
	}
	hash := cg.Hash(data)
	b := ecdsa.Verify(pubkey, hash, ecdsasign.R, ecdsasign.S)

	//验证父证书
	err = x509cert.CheckSignatureFrom(cg.rCA_S)
	if err != nil {
		return false, errors.New("verified the parent cert failed!")
	}

	if b {
		return true, nil
	}
	return false, errors.New("verify failed, signature verify not passed.")
}

func (cg *CertGroup)Hash(data []byte) ([]byte) {
	her := crypto.SHA3_256.New()
	her.Write(data)
	return her.Sum(nil)
}
