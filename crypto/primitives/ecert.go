/**
author:ZhangKejie
date:16-12-18
verify ecert and signature
*/

package primitives

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/x509"
	"encoding/pem"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"io/ioutil"
	"time"
)

var (
	log          = logging.MustGetLogger("crypto")
	defaultCurve elliptic.Curve
)

// GetDefaultCurve returns the default elliptic curve used by the crypto layer
func GetDefaultCurve() elliptic.Curve {
	return defaultCurve
}

func GetConfig(path string) (string, error) {
	content, err := ioutil.ReadFile(path)

	if err != nil {
		return "", err
	}

	return string(content), nil

}

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

func VerifyCert(cert *x509.Certificate, ca *x509.Certificate) (bool, error) {
	err := cert.CheckSignatureFrom(ca)

	endTime := cert.NotAfter
	startTime := cert.NotBefore

	today := time.Now()

	if startTime.After(today) || endTime.Before(today) {
		log.Error("This Cert is overdue.Please check it!")
		return false, errors.New("This Cert is overdue.Please check it!")
	}

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