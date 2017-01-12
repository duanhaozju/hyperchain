package membersrvc

import (
	"crypto/x509"
	"hyperchain/core/crypto/primitives"
	"github.com/pkg/errors"
	"io/ioutil"
	"crypto/ecdsa"
)

type CAManager struct {
	ecacert *x509.Certificate
	ecert *x509.Certificate
	rcert *x509.Certificate
	rcacert *x509.Certificate
	ecertPrivateKey interface{}
	tcacert *x509.Certificate

	// cert byte
	ecacertByte []byte
	ecertByte []byte
	rcertByte []byte
	rcacertByte []byte
	ecertPrivateKeyByte []byte
	tcacertByte []byte
	isUsed bool

}

var CaManager *CAManager

func GetCaManager(ecacertPath string,ecertPath string,rcertPath string,rcacertPath string, ecertPrivateKeyPath string,tcacertPath string,isUsed bool) (*CAManager,error) {
	if CaManager == nil {
		var err error
		CaManager,err = NewCAManager(ecacertPath,ecertPath,rcertPath,rcacertPath,ecertPrivateKeyPath,tcacertPath,isUsed)
		if err != nil {
			return nil,err
		}
		return CaManager,nil
	} else {
		return CaManager,nil
	}
}

func NewCAManager(ecacertPath string,ecertPath string,rcertPath string,rcacertPath string, ecertPrivateKeyPath string,tcacertPath string,isUsed bool) (*CAManager,error){


	var caManager CAManager
	var err error
	if isUsed != true{
		caManager.ecacertByte = []byte{}
		caManager.ecertByte = []byte{}
		caManager.rcertByte = []byte{}
		caManager.rcacertByte = []byte{}
		caManager.ecertPrivateKeyByte = []byte{}
		caManager.tcacertByte = []byte{}
		caManager.isUsed = isUsed
		return &caManager,nil
	}else {
		ecacert, rerr := ioutil.ReadFile(ecacertPath)
		if rerr != nil {
			log.Error("ecacert read failed")
			return nil, rerr
		}
		caManager.ecacertByte = ecacert
		ecert, rerr := ioutil.ReadFile(ecertPath)
		if rerr != nil {
			log.Error("ecert read failed")
			return nil, rerr
		}
		caManager.ecertByte = ecert
		rcert, rerr := ioutil.ReadFile(rcertPath)
		if rerr != nil {
			log.Error("rcert read failed")
			return nil, rerr
		}
		caManager.rcertByte = rcert
		rcacert, rerr := ioutil.ReadFile(rcacertPath)
		if rerr != nil {
			log.Error("rcacert read failed")
			return nil, rerr
		}
		caManager.rcacertByte = rcacert
		ecertPrivateKey, rerr := ioutil.ReadFile(ecertPrivateKeyPath)
		if rerr != nil {
			log.Error("ecaPrivateKey read failed")
			return nil, rerr
		}
		caManager.ecertPrivateKeyByte = ecertPrivateKey

		//TODO delete this part
		tcacert, tcaerr := ioutil.ReadFile(tcacertPath)
		if tcaerr != nil {
			log.Error("tcacert read failed")
			return nil, rerr
		}
		caManager.tcacertByte = tcacert

		// TODO check the private type private key is der

		caManager.ecert, err = primitives.ParseCertificate(string(ecert))
		if err != nil {
			log.Error("cannot parse the ecert")
			return nil, errors.New("cannot parse the ecert")
		}
		caManager.ecacert, err = primitives.ParseCertificate(string(ecacert))
		if err != nil {
			log.Error("cannot parse the ecacert")
			return nil, errors.New("cannot parse the ecacert")
		}
		caManager.rcert, err = primitives.ParseCertificate(string(rcert))
		if err != nil {
			log.Error("cannot parse the rcert")
			return nil, errors.New("cannot parse the rcert")
		}
		caManager.rcacert, err = primitives.ParseCertificate(string(rcacert))
		if err != nil {
			log.Error("cannot parse the rcert")
			return nil, errors.New("cannot parse the rcacert")
		}
		//the caManager
		caManager.ecertPrivateKey, err = primitives.ParseKey(string(ecertPrivateKey))
		if err != nil {
			log.Error("cannot parse the caprivatekey")
			return nil, errors.New("cannot parse the caprivatekey")
		}

		//TODO delete this part
		caManager.tcacert, err = primitives.ParseCertificate(string(tcacert))
		if err != nil {
			log.Error("cannot parse the tcacert")
			return nil, errors.New("cannot parse the tcacert")
		}

		//TODO use ca or not
		caManager.isUsed = isUsed
		return &caManager, nil
	}
}

func (caManager *CAManager)SignTCert(publicKey string) (string,error){
	if caManager.isUsed != true{
		return "",nil
	}
	pubKey, err := primitives.ParsePubKey(publicKey)
	if err != nil {
		log.Error(err)
		return "", err
	}
	tcert, err := primitives.GenTCert(caManager.ecert, caManager.ecertPrivateKey, pubKey)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return string(tcert), nil

}


//TCERT 需要用为其签发的ECert来验证，但是在没有TCERT的时候只能够用
//ECERT 进行充当TCERT 所以需要用ECA.CERT 即ECA.CA 作为根证书进行验证
func (caManager *CAManager)VerifyTCert(tcertPEM string)(bool,error){
	if caManager.isUsed != true{
		return true,nil
	}
	tcertToVerify, err := primitives.ParseCertificate(tcertPEM)
	if err != nil {
		log.Error("cannot parse the tcert", err)
		return false, err
	}
	// TODO 应该将caManager.tcacert 转为caManager.ecacert
	verifyTcert, err := primitives.VerifyCert(tcertToVerify, caManager.tcacert)
	if verifyTcert == false {
		log.Error("verified falied")
		return false, err
	}
	return true, nil

}

/**
	验证签名，验证签名需要有三个参数：
	第一个是携带的数字证书，即tcert,
	第二个是签名，
	第三个是原始数据
	这个方法用来验证签名是否来自数字证书用户
 */
func (caManager *CAManager)VerifySignature(certPEM,signature,signed string) (bool,error) {
	if caManager.isUsed != true{
		return true,nil
	}
	tcertToVerify,err := primitives.ParseCertificate(certPEM)
	if err != nil {
		log.Error("cannot parse the tcert",err)
		return false,err
	}
	verifySign,err := primitives.VerifySignature(tcertToVerify,signed,signature)
	if verifySign==false{
		log.Error("verified falied")
		return false,err
	}
	return true,nil

}

func (caManager *CAManager) VerifyECert(ecertPEM string)(bool,error){
	if caManager.isUsed != true{
		return true,nil
	}else {
		ecertToVerify,err := primitives.ParseCertificate(ecertPEM)
		if err != nil {
			log.Error("cannot parse the ecert",err)
			return false,err
		}
		verifyEcert,err := primitives.VerifyCert(ecertToVerify,caManager.ecacert)
		if verifyEcert==false || err != nil{
			log.Error("verified ecert falied")
			return false,err
		}
		return true,nil
	}
}

func (ca *CAManager) VerifyECertSignature(ecertPEM string,msg,sign []byte)(bool,error){
	if CaManager.isUsed !=true {
		return true,nil
	}
	ecertToVerify,err := primitives.ParseCertificate(ecertPEM)
	if err != nil {
		log.Error("cannot parse the ecert",err)
		return false,err
	}
	ecdsaEncry := primitives.NewEcdsaEncrypto("ecdsa")


	key := ecertToVerify.PublicKey.(*(ecdsa.PublicKey))
	result,err1 := ecdsaEncry.VerifySign(*key,msg,sign)
	if(err1 != nil){
		log.Error("fail to verify signture",err)
		return false,err1
	}

	return  result,nil;


}

func (caManager *CAManager) VerifyRCert(rcertPEM string)(bool,error){
	if caManager.isUsed!=true{
		return true,nil
	}

	rcertToVerify,err := primitives.ParseCertificate(rcertPEM)
	if err != nil {
		log.Error("cannot parse the rcert",err)
		return false,err
	}
	verifyRcert,err := primitives.VerifyCert(rcertToVerify,caManager.rcacert)
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
func (caManager *CAManager) GetIsUsed() bool{
	return caManager.isUsed
}