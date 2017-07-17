package hts

import (
	"crypto/x509"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto"
	"math/big"
	"encoding/asn1"
	"github.com/pkg/errors"
)

type CertGroup struct {
	enableEnroll bool

	//ECert group
	eCA []byte
	eCA_S *x509.Certificate
	eCERT []byte
	eCERT_S *x509.Certificate
	eCERTPriv []byte
	eCERTPriv_S *ecdsa.PrivateKey

	//RCert group
	rCA []byte
	rCA_S *x509.Certificate
	rCERT []byte
	rCERT_S *x509.Certificate
	rCERTPriv []byte
	rCERTPriv_S *ecdsa.PrivateKey

	//TCert group
	enableT bool

}

func(cg *CertGroup)GetECert()[]byte{
	return cg.eCERT
}



func(cg *CertGroup)GetRCert()[]byte{
	return cg.rCERT
}

func(cg *CertGroup)ESign(data []byte)([]byte,error){
	return cg.eCERTPriv_S.Sign(rand.Reader,data,crypto.SHA3_256)
}

func(cg *CertGroup)EVerify(sign []byte,hash []byte)(bool,error){
	ecdsasign := struct {
		R,S *big.Int
	}{}
	_,err := asn1.Unmarshal(sign,&ecdsasign)
	if err !=nil{
		return false,errors.New("unmarshal the signature failed. (E verify)")
	}
	return ecdsa.Verify(&cg.eCERTPriv_S.PublicKey,hash,ecdsasign.R,ecdsasign.S), nil
}

func(cg *CertGroup)RSign(data []byte)([]byte,error){
	hash := cg.Hash(data)
	return cg.rCERTPriv_S.Sign(rand.Reader,hash,crypto.SHA3_256)
}

func(cg *CertGroup)RVerify(sign []byte,data []byte)(bool,error){
	ecdsasign := struct {
		R,S *big.Int
	}{}
	_,err := asn1.Unmarshal(sign,&ecdsasign)
	if err !=nil{
		return false,errors.New("unmarshal the signature failed. (R verify)")
	}
	hash := cg.Hash(data)
	return ecdsa.Verify(&cg.rCERTPriv_S.PublicKey,hash,ecdsasign.R,ecdsasign.S),nil
}

func(cg *CertGroup)Hash(data []byte)([]byte){
	her :=crypto.SHA3_256.New()
	her.Write(data)
	return her.Sum(nil)
}
