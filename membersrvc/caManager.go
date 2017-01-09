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
	ecaPrivateKey interface{}

}

func NewCAManager(ecacertPath string,ecertPath string,rcertPath string,ecaPrivateKeyPath string) (*CAManager,error){
	ecacert,rerr := ioutil.ReadFile(ecacertPath)
	if rerr != nil{
		log.Error("ecacert read failed")
		return nil,rerr
	}
	ecert,rerr := ioutil.ReadFile(ecertPath)
	if rerr != nil{
		log.Error("ecert read failed")
		return nil,rerr
	}
	rcert,rerr := ioutil.ReadFile(rcertPath)
	if rerr != nil{
		log.Error("rcert read failed")
		return nil,rerr
	}
	ecaPrivateKey,rerr := ioutil.ReadFile(ecaPrivateKeyPath)
	if rerr != nil{
		log.Error("ecaPrivateKey read failed")
		return nil,rerr
	}

	// TODO check the private type private key is der
	var caManager CAManager
	var err error
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
	//the caManager
	caManager.ecaPrivateKey,err = primitives.ParseKey(string(ecaPrivateKey))
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
	tcert,err := primitives.GenTcert(caManager.ecacert,caManager.ecaPrivateKey,pubKey)
	if err != nil{
		log.Error(err)
		return "",err
	}
	return string(tcert),nil
}
