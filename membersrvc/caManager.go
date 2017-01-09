package membersrvc

import (
	"crypto/x509"
	"hyperchain/core/crypto/primitives"
	"github.com/pkg/errors"
	"io/ioutil"
)

type CAManager struct {
	ecacert *x509.Certificate
	ecert *x509.Certificate
	rcert *x509.Certificate
	rcacert *x509.Certificate
	ecertPrivateKey interface{}
	// cert byte
	ecacertByte []byte
	ecertByte []byte
	rcertByte []byte
	rcacertByte []byte
	ecertPrivateKeyByte []byte
}

func NewCAManager(ecacertPath string,ecertPath string,rcertPath string,rcacertPath string, ecertPrivateKeyPath string) (*CAManager,error){
	var caManager CAManager
	var err error

	ecacert,rerr := ioutil.ReadFile(ecacertPath)
	if rerr != nil{
		log.Error("ecacert read failed")
		return nil,rerr
	}
	caManager.ecacertByte = ecacert
	ecert,rerr := ioutil.ReadFile(ecertPath)
	if rerr != nil{
		log.Error("ecert read failed")
		return nil,rerr
	}
	caManager.ecertByte = ecert
	rcert,rerr := ioutil.ReadFile(rcertPath)
	if rerr != nil{
		log.Error("rcert read failed")
		return nil,rerr
	}
	caManager.rcertByte = rcert
	rcacert,rerr := ioutil.ReadFile(rcacertPath)
	if rerr != nil{
		log.Error("rcacert read failed")
		return nil,rerr
	}
	caManager.rcacertByte = rcacert
	ecertPrivateKey,rerr := ioutil.ReadFile(ecertPrivateKeyPath)
	if rerr != nil{
		log.Error("ecaPrivateKey read failed")
		return nil,rerr
	}
	caManager.ecertPrivateKeyByte =ecertPrivateKey

	// TODO check the private type private key is der

	caManager.ecert,err = primitives.ParseCertificate(string(ecert))
	if err != nil{
		log.Error("cannot parse the ecert")
		return nil,errors.New("cannot parse the ecert")
	}
	caManager.ecacert,err = primitives.ParseCertificate(string(ecacert))
	if err != nil{
		log.Error("cannot parse the ecacert")
		return nil,errors.New("cannot parse the ecacert")
	}
	caManager.rcert,err = primitives.ParseCertificate(string(rcert))
	if err != nil{
		log.Error("cannot parse the rcert")
		return nil,errors.New("cannot parse the rcert")
	}
	caManager.rcacert,err = primitives.ParseCertificate(string(rcacert))
	if err != nil{
		log.Error("cannot parse the rcert")
		return nil,errors.New("cannot parse the rcacert")
	}
	//the caManager
	caManager.ecertPrivateKey,err = primitives.ParseKey(string(ecertPrivateKey))
	if err != nil{
		log.Error("cannot parse the caprivatekey")
		return nil,errors.New("cannot parse the caprivatekey")
	}
	return &caManager,nil
}

func (caManager *CAManager)SignTCert(publicKey string) (string,error){
	pubKey,err := primitives.ParsePubKey(publicKey)
	if err != nil{
		log.Error(err)
		return "",err
	}
	tcert,err := primitives.GenTCert(caManager.ecert,caManager.ecertPrivateKey,pubKey)
	if err != nil{
		log.Error(err)
		return "",err
	}
	return string(tcert),nil
}

func (caManager *CAManager)VerifySignature(tcertPEM string)(bool,error){
	tcertToVerify,err := primitives.ParseCertificate(tcertPEM)
	if err != nil {
		log.Error("cannot parse the tcert",err)
		return false,err
	}
	verifyTcert,err := primitives.VerifySignature(tcertToVerify,caManager.ecert)
	if verifyTcert==false{
		log.Error("verified falied")
		return false,err
	}
	return true,nil
}

func (caManager *CAManager) VerifyECert(ecertPEM string)(bool,error){
	ecertToVerify,err := primitives.ParseCertificate(ecertPEM)
	if err != nil {
		log.Error("cannot parse the ecert",err)
		return false,err
	}
	verifyEcert,err := primitives.VerifySignature(ecertToVerify,caManager.ecacert)
	if verifyEcert==false || err != nil{
		log.Error("verified ecert falied")
		return false,err
	}
	return true,nil
}

func (caManager *CAManager) VerifyRCert(rcertPEM string)(bool,error){
	rcertToVerify,err := primitives.ParseCertificate(rcertPEM)
	if err != nil {
		log.Error("cannot parse the rcert",err)
		return false,err
	}
	verifyRcert,err := primitives.VerifySignature(rcertToVerify,caManager.rcacert)
	if verifyRcert==false || err != nil{
		log.Error("verified rcert falied")
		return false,err
	}
	return true,nil
}

/**
	getMethods
 */


func (caManager *CAManager)  GetECACertByte() []byte{
	return caManager.ecertPrivateKeyByte
}
func (caManager *CAManager)  GetECertByte() []byte{
	return caManager.ecacertByte
}
func (caManager *CAManager)  GetRCertByte() []byte{
	return caManager.rcertByte
}
func (caManager *CAManager)  GetRCAcertByte() []byte{
	return caManager.rcacertByte
}
func (caManager *CAManager)  GetECAPrivateKeyByte() []byte{
	return caManager.ecertPrivateKeyByte
}