/**
author:张珂杰
date:16-12-18
verify ecert and signature
*/

package primitives

import (
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	//"fmt"
	//"sync"
	"crypto/ecdsa"
	"github.com/pkg/errors"
)

//读取config文件
func GetConfig(path string) (string, error) {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		//fmt.Println()
		return "", err
	}

	return string(content), nil

}

//解析证书
func ParseCertificate(cert []byte) (*x509.Certificate, error) {
	block, _ := pem.Decode(cert)

	if block == nil {
		return nil, errors.New("failed to parse certificate PEM")
	}

	x509Cert, err := x509.ParseCertificate(block.Bytes)

	if err != nil {
		return nil, errors.New("faile to parse certificate")
	}

	return x509Cert, nil
}

//验证证书中的签名
func VerifyCert(cert *x509.Certificate, ca *x509.Certificate) (bool, error) {
	err := cert.CheckSignatureFrom(ca)
	if err != nil {
		log.Error("verified the cert failed", err)
		return false, err
	}
	return true, nil

}

//验证证书的来源
func VerifySignature(cert *x509.Certificate, signed string, signature string) (bool, error) {
	err := cert.CheckSignature(x509.ECDSAWithSHA256, []byte(signed), []byte(signature))
	if err != nil {
		log.Error("verified the cert failed", err)
		return false, err
	}
	return true, nil
}

func ParseKey(derPri []byte) (interface{}, error) {
	block, _ := pem.Decode(derPri)

	pri, err1 := DERToPrivateKey(block.Bytes)

	if err1 != nil {
		return nil, err1
	}

	return pri, nil
}

func ParsePubKey(pubPem string) (*ecdsa.PublicKey, error) {
	if pubPem == "" {
		return nil, errors.New("the pub pem is nil")
	}
	block, _ := pem.Decode([]byte(pubPem))
	pub, err := DERToPublicKey(block.Bytes)

	if err != nil {
		log.Error(err)
		return nil, err
	}

	pubkey := pub.(*(ecdsa.PublicKey))

	return pubkey, nil

}