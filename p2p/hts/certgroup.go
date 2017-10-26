package hts

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rand"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"github.com/pkg/errors"
	"github.com/terasum/viper"
	"hyperchain/common"
	crypto2 "hyperchain/crypto"
	"hyperchain/crypto/primitives"
	"io/ioutil"
	"math/big"
)

type CertGroup struct {
	enableEnroll bool
	sign         bool

	//ECert group
	eCA         []byte
	eCA_S       *x509.Certificate
	eCERT       []byte
	eCERT_S     *x509.Certificate
	eCERTPriv   []byte
	eCERTPriv_S *ecdsa.PrivateKey

	//RCert group
	rCA         []byte
	rCA_S       *x509.Certificate
	rCERT       []byte
	rCERT_S     *x509.Certificate
	rCERTPriv   []byte
	rCERTPriv_S *ecdsa.PrivateKey
}

// NewCertGroup creates and returns a new CertGroup instance.
func NewCertGroup(namespace string, vip *viper.Viper) (*CertGroup, error) {

	var cg *CertGroup = new(CertGroup)

	// read in ECA
	ecap := common.GetPath(namespace, vip.GetString(common.ENCRYPTION_ECERT_ECA))
	if !common.FileExist(ecap) {
		return nil, errors.New(fmt.Sprintf("cannot read in eca, reason: file not exist (%s)", ecap))
	}
	eca, err := ioutil.ReadFile(ecap)
	if err != nil {
		return nil, err
	}
	cg.eCA = eca
	ecas, err := primitives.ParseCertificate(eca)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse the eca certificate, reason: %s", err.Error()))
	}
	cg.eCA_S = ecas

	// read in ECERT
	ecertp := common.GetPath(namespace, vip.GetString(common.ENCRYPTION_ECERT_ECERT))
	if !common.FileExist(ecertp) {
		return nil, errors.New(fmt.Sprintf("cannot read in ecert,reason: file not exist (%s)", ecertp))
	}
	ecert, err := ioutil.ReadFile(ecertp)
	if err != nil {
		return nil, err
	}
	cg.eCERT = ecert
	ecerts, err := primitives.ParseCertificate(ecert)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse the e certificate, reason %s", err.Error()))
	}
	cg.eCERT_S = ecerts

	// read in ECERT private key
	eprivp := common.GetPath(namespace, vip.GetString(common.ENCRYPTION_ECERT_PRIV))
	if !common.FileExist(eprivp) {
		return nil, errors.New(fmt.Sprintf("cannot read in ecert priv,reason: file not exist (%s)", eprivp))
	}
	epriv, err := ioutil.ReadFile(eprivp)
	if err != nil {
		return nil, err
	}
	cg.eCERTPriv = epriv
	eps, err := primitives.ParseKey(epriv)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse the private key, reason %s", err.Error()))
	}
	ep_s, ok := eps.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New(fmt.Sprintf("cannot parse the private key, reason %s", "cannot convert private key into *ecdsa.PrivateKey"))
	}
	cg.eCERTPriv_S = ep_s

	// if enable RCERT, enEnroll is true
	enEnroll := vip.GetBool(common.ENCRYPTION_CHECK_ENABLE)
	enSign := vip.GetBool(common.ENCRYPTION_CHECK_SIGN)
	cg.enableEnroll = enEnroll
	cg.sign = enSign

	if enEnroll {

		// read in RCA
		rcap := common.GetPath(namespace, vip.GetString(common.ENCRYPTION_RCERT_RCA))
		if !common.FileExist(rcap) {
			return nil, errors.New(fmt.Sprintf("cannot read in rca,reason: file not exist (%s)", rcap))
		}
		rca, err := ioutil.ReadFile(rcap)
		if err != nil {
			return nil, err
		}
		cg.rCA = rca
		rcas, err := primitives.ParseCertificate(rca)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("cannot parse the rca certificate, reason: %s", err.Error()))
		}
		cg.rCA_S = rcas

		// read in RCERT
		rcertp := common.GetPath(namespace, vip.GetString(common.ENCRYPTION_RCERT_RCERT))
		if !common.FileExist(rcertp) {
			return nil, errors.New(fmt.Sprintf("cannot read in rcert,reason: file not exist (%s)", rcertp))
		}
		rcert, err := ioutil.ReadFile(rcertp)
		if err != nil {
			return nil, err
		}
		cg.rCERT = []byte(rcert)
		rcerts, err := primitives.ParseCertificate(rcert)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("cannot parse the r certificate, reason %s", err.Error()))
		}
		cg.rCERT_S = rcerts

		// read in RCERT private key
		rprivp := common.GetPath(namespace, vip.GetString(common.ENCRYPTION_RCERT_PRIV))
		if !common.FileExist(rprivp) {
			return nil, errors.New(fmt.Sprintf("cannot read in rcert priv,reason: file not exist (%s)", eprivp))
		}
		rpriv, err := ioutil.ReadFile(rprivp)
		if err != nil {
			return nil, err
		}
		cg.rCERTPriv = rpriv
		rps, err := primitives.ParseKey(rpriv)
		if err != nil {
			return nil, errors.New(fmt.Sprintf("cannot parse the r private key, reason %s", err.Error()))
		}
		rp_s, ok := rps.(*ecdsa.PrivateKey)
		if !ok {
			return nil, errors.New(fmt.Sprintf("cannot parse the r private key, reason %s", "cannot convert private key into *ecdsa.PrivateKey"))
		}
		cg.rCERTPriv_S = rp_s
	}

	return cg, nil
}

func (cg *CertGroup) GetECert() []byte {
	return cg.eCERT
}

func (cg *CertGroup) GetRCert() []byte {
	return cg.rCERT
}

// ESign returns a signature signed by ecert private key.
func (cg *CertGroup) ESign(data []byte) ([]byte, error) {
	if !cg.sign {
		return nil, nil
	}
	return cg.eCERTPriv_S.Sign(rand.Reader, data, crypto.SHA3_256)
}

// EVerify verifies whether a signature signed by ecert private key is valid.
func (cg *CertGroup) EVerify(rawcert, sign []byte, hash []byte) (bool, error) {
	if !cg.sign {
		return true, nil
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

	// verify parent certificate
	eca, _ := primitives.ParseCertificate(cg.eCA)
	err = x509cert.CheckSignatureFrom(eca)
	if err != nil {
		return false, errors.New("verified the parent cert failed!")
	}

	if b {
		return true, nil
	}
	return false, errors.New("verify failed, signature verify not passed.")
}

// RSign returns a signature signed by rcert private key.
func (cg *CertGroup) RSign(data []byte) ([]byte, error) {
	if !cg.sign || !cg.enableEnroll {
		return nil, nil
	}
	hash := cg.Hash(data)
	return cg.rCERTPriv_S.Sign(rand.Reader, hash, crypto.SHA3_256)
}

// RVerify verifies whether a signature signed by rcert private key is valid.
func (cg *CertGroup) RVerify(rawcert, sign []byte, data []byte) (bool, error) {
	if !cg.sign || !cg.enableEnroll {
		return true, nil
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

	// verify parent certificate
	err = x509cert.CheckSignatureFrom(cg.rCA_S)
	if err != nil {
		return false, errors.New("verified the parent cert failed!")
	}

	if b {
		return true, nil
	}
	return false, errors.New("verify failed, signature verify not passed.")
}

func (cg *CertGroup) Hash(data []byte) []byte {
	return crypto2.Keccak256(data)
}
