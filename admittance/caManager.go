package admittance

import (
	"crypto/ecdsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"hyperchain/core/crypto/primitives"
	"io/ioutil"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc"
	"github.com/terasum/viper"
	"github.com/op/go-logging"
	"hyperchain/common"
)
// Init the log setting
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("admittance")
}

type CAManager struct {
	ecacert         *x509.Certificate
	ecert           *x509.Certificate
	rcert           *x509.Certificate
	rcacert         *x509.Certificate
	ecertPrivateKey interface{}
	tcacert         *x509.Certificate

	// cert byte
	ecacertByte         []byte
	ecertByte           []byte
	rcertByte           []byte
	rcacertByte         []byte
	ecertPrivateKeyByte []byte
	tcacertByte         []byte
	isUsed              bool
	checkTCert          bool
	//
	config *viper.Viper
}

var caManager *CAManager

func GetCaManager(global_config *viper.Viper) (*CAManager, error) {
	if caManager == nil {
		config := viper.New()
		config.SetConfigFile(global_config.GetString("global.configs.caconfig"))
		err := config.ReadInConfig()
		if err != nil {
			log.Error("cannot read the ca config file!")
			panic(err)
		}
		ecacertPath := config.GetString("ecert.ca")
		ecertPath := config.GetString("ecert.cert")
		ecertPrivateKeyPath :=config.GetString("ecert.priv")
		rcacertPath:=config.GetString("rcert.ca")
		rcertPath:=config.GetString("rcert.cert")
		checkERCert := config.GetBool("check.checkercert")
		checkTCert := config.GetBool("check.checktcert")

		caManager, err = newCAManager(ecacertPath, ecertPath, rcertPath, rcacertPath, ecertPrivateKeyPath, ecertPath, checkERCert, checkTCert)
		if err != nil {
			return nil, err
		}
		caManager.config = config
		return caManager, nil
	} else {
		return caManager, nil
	}
}

func newCAManager(ecacertPath string, ecertPath string, rcertPath string, rcacertPath string, ecertPrivateKeyPath string, tcacertPath string, isUsed bool, checkTCert bool) (*CAManager, error) {
	var caManager CAManager
	var err error
	caManager.checkTCert = checkTCert
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

func (caManager *CAManager) SignTCert(publicKey string) (string, error) {
	//if caManager.isUsed != true {
	//	log.Critical("NOT USE TCERT")
	//	return "", nil
	//}
	caManager.GetIsCheckTCert()
	pubPem := common.TransportDecode(publicKey)

	pubKey,err := primitives.ParsePubKey(pubPem)
	if err != nil {
		log.Error(err)
		return "", err
	}
	tcert, err := primitives.GenTCert(caManager.ecert, caManager.ecertPrivateKey, pubKey)
	tcertPem := primitives.DERCertToPEM(tcert)
	if err != nil {
		log.Error(err)
		return "", err
	}
	return string(tcertPem), nil

}

//TCERT 需要用为其签发的ECert来验证，但是在没有TCERT的时候只能够用
//ECERT 进行充当TCERT 所以需要用ECA.CERT 即ECA.CA 作为根证书进行验证
func (caManager *CAManager)VerifyTCert(tcertPEM string)(bool,error){
	if caManager.checkTCert != true{
		return true,nil
	}
	tcertToVerify, err := primitives.ParseCertificate(tcertPEM)
	if err != nil {
		log.Error("cannot parse the tcert", err)
		return false, err
	}
	// TODO 应该将caManager.tcacert 转为caManager.ecacert
	isCa := tcertToVerify.IsCA
	var verifyTcert bool
	if isCa == true {
		verifyTcert, err = primitives.VerifyCert(tcertToVerify, caManager.ecacert)
	} else {
		verifyTcert, err = primitives.VerifyCert(tcertToVerify, caManager.tcacert)

	}
	if verifyTcert == false {
		log.Error("verified falied")
		return false, err
	}
	return true, nil

}


func (caManager *CAManager) VerifyECert(ecertPEM string) (bool, error) {
	if caManager.isUsed != true {
		return true, nil
	}
	ecertToVerify, err := primitives.ParseCertificate(ecertPEM)
	if err != nil {
		log.Error("cannot parse the ecert", err)
		return false, err
	}
	verifyEcert, err := primitives.VerifyCert(ecertToVerify, caManager.ecacert)
	if verifyEcert == false || err != nil {
		log.Error("verified ecert falied")
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
//VerifyCertSignature Verify the Signature of Cert
func (ca *CAManager) VerifyCertSignature(certPEM string, msg, sign []byte) (bool, error) {
	if caManager.isUsed != true {
		return true, nil
	}
	ecertToVerify, err := primitives.ParseCertificate(certPEM)
	if err != nil {
		log.Error("cannot parse the ecert", err)
		return false, err
	}
	ecdsaEncry := primitives.NewEcdsaEncrypto("ecdsa")

	key := ecertToVerify.PublicKey.(*(ecdsa.PublicKey))
	result, err := ecdsaEncry.VerifySign(*key, msg, sign)
	if err != nil {
		log.Error("fail to verify signture", err)
		return false, err
	}

	return result, nil

}

func (caManager *CAManager) VerifyRCert(rcertPEM string) (bool, error) {
	if caManager.isUsed != true {
		return true, nil
	}

	rcertToVerify, err := primitives.ParseCertificate(rcertPEM)
	if err != nil {
		log.Error("cannot parse the rcert", err)
		return false, err
	}
	verifyRcert, err := primitives.VerifyCert(rcertToVerify, caManager.rcacert)
	if verifyRcert == false || err != nil {
		log.Error("verified rcert falied")
		return false, err
	}
	return true, nil

}

/**
  tls ca get dial opts and server opts part
 */

//获取客户端ca配置opt
func (caManager *CAManager) GetGrpcClientOpts() []grpc.DialOption {
	var opts []grpc.DialOption
	creds, err := credentials.NewClientTLSFromFile(caManager.config.GetString("tlscert.ca"), caManager.config.GetString("tlscert.serverhostoverride"))
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return opts
}

//获取服务器端ca配置opts
func (caManager *CAManager) GetGrpcServerOpts() []grpc.ServerOption {
	var opts []grpc.ServerOption
	creds, err := credentials.NewServerTLSFromFile(caManager.config.GetString("tlscert.cert"), caManager.config.GetString("tlscert.priv"))
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	opts = []grpc.ServerOption{grpc.Creds(creds)}
	return opts
}

/**
getMethods
*/

func (caManager *CAManager) GetECACertByte() []byte {
	return caManager.ecacertByte
}
func (caManager *CAManager) GetECertByte() []byte {
	return caManager.ecertByte
}
func (caManager *CAManager) GetRCertByte() []byte {
	return caManager.rcertByte
}
func (caManager *CAManager) GetRCAcertByte() []byte {
	return caManager.rcacertByte
}
func (caManager *CAManager) GetECertPrivateKeyByte() []byte {
	return caManager.ecertPrivateKeyByte
}
func (caManager *CAManager) GetECertPrivKey() interface{} {
	return caManager.ecertPrivateKey
}
func (caManager *CAManager) GetIsUsed() bool {
	return caManager.isUsed
}
func (caManager *CAManager) GetIsCheckTCert() bool {
	return caManager.checkTCert
}
