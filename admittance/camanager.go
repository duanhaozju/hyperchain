package admittance

import (
	"crypto/ecdsa"
	"crypto/x509"
	"github.com/pkg/errors"
	"hyperchain/crypto/primitives"
	"io/ioutil"
	"github.com/op/go-logging"
	"hyperchain/common"
	"github.com/spf13/viper"
	"fmt"
	"strings"
	"hyperchain/hyperdb"
	"encoding/asn1"
	"os"
)

var (
	errDecPubKey = errors.New("cannot decode the public string,please encode the public string as right hex string")
	errParsePubKey = errors.New("cannot parse the request publickey, please check the public string.")
	errParseCert = errors.New("cannot parse the cert pem, please check your cert pem string.")
	errGenTCert = errors.New("cannot generate the tcert, please check your public key, if not work, please contract the hyperchain maintainer")
	errDERToPEM = errors.New("cannot convert the der format cert into pem format.")
	errFailedVerifySign = errors.New("Verify the Cert Signature failed, please use the correctly certificate")
)

const CertKey string = "tcerts"


type RegisterTcerts struct {
	Tcerts []string
}

type cert struct {
	x509cert *x509.Certificate
	certByte []byte
}
type key struct {
	priKey     interface{}
	prikeybyte []byte
}
//CAManager this struct is for Certificate Auth manager
type CAManager struct {
	eCaCert               *cert
	eCert                 *cert
	rCaCert               *cert
	rCert                 *cert
	tCacert               *cert
	eCertPri              *key
	rCertPri	      *key
	//check flags
	checkERCert           bool
	checkTCert            bool
	checkCertSign         bool

	logger     *logging.Logger
}

//NewCAManager get a new ca manager instance
func NewCAManager(conf *common.Config) (*CAManager, error) {
	namespace := conf.GetString(common.NAMESPACE)
	logger := common.GetLogger(namespace, "ca")


	caconfPath := common.GetPath(namespace, conf.GetString("config.path.caconfig"))
	if caconfPath == "" {
		return nil, errors.New("cannot get the ca config file path.")
	}
	config := viper.New()
	config.SetConfigFile(caconfPath)
	err := config.ReadInConfig()
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot read ca conf,reason: %s",err.Error()))
	}
	eca, err := readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_ECERT_ECA)))
	if err != nil {
		return nil, err
	}
	ecert, err := readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_ECERT_ECERT)))
	if err != nil {
		return nil, err
	}
	rca, err := readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_RCERT_RCA)))
	if err != nil {
		return nil, err
	}
	rcert, err := readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_RCERT_RCERT)))
	if err != nil {
		rcert = &cert{}
	}
	ecertpriv, err := readKey(common.GetPath(namespace, config.GetString(common.ENCRYPTION_ECERT_PRIV)))
	if err != nil {
		return nil, err
	}
	rcertpriv, err := readKey(common.GetPath(namespace, config.GetString(common.ENCRYPTION_RCERT_PRIV)))
	if err != nil {
		return nil, err
	}

	whiteList := config.GetBool(common.ENCRYPTION_TCERT_WHITELIST)
	if whiteList {
		whiteListDir := common.GetPath(namespace, config.GetString(common.ENCRYPTION_TCERT_WHITELIST_DIR))
		tcertFiles,err := ListDir(whiteListDir,"cert")
		if err != nil{
			logger.Error("Init tcert white list failed :",err)
			return  nil,&common.CertError{Message: "Init tcert white list failed"};
		}
		for _,tcertPath := range tcertFiles{
			cert, err := ioutil.ReadFile(tcertPath)
			if err != nil {
				return nil, err
			}
			err = RegisterCert(cert)
			if err != nil{
				if err != nil {
					logger.Error(err)
					return nil, &common.CertError{Message: "Init tcert white list failed"}
				}
			}
		}
	}

	return &CAManager{
		eCaCert:eca,
		eCert:ecert,
		eCertPri:ecertpriv,
		rCertPri:rcertpriv,
		rCaCert:rca,
		rCert:rcert,
		tCacert:ecert,
		checkCertSign:config.GetBool(common.ENCRYPTION_CHECK_SIGN),
		checkERCert:config.GetBool(common.ENCRYPTION_CHECK_ENABLE),
		checkTCert:config.GetBool(common.ENCRYPTION_CHECK_ENABLE_T),
		logger:logger,

	},nil

}
//Generate a TCert for SDK client.
func (cm *CAManager) GenTCert(publicKey string) (string, error) {
	pubpem := common.TransportDecode(publicKey)
	if pubpem == "" {
		cm.logger.Error(errDecPubKey.Error())
		return "", errDecPubKey
	}
	pubKey, err := primitives.ParsePubKey(pubpem)
	if err != nil {
		cm.logger.Error(errParsePubKey.Error())
		return "", errParsePubKey
	}
	tcert, err := primitives.GenTCert(cm.eCert.x509cert, cm.eCertPri.priKey, pubKey)
	if err != nil {
		cm.logger.Error(errGenTCert.Error())
		return "", errGenTCert
	}
	tcertPem := primitives.DERCertToPEM(tcert)
	if err != nil {
		cm.logger.Error(errDERToPEM.Error())
		return "", errDERToPEM
	}
	return string(tcertPem), nil

}

//TCert 需要用为其签发的ECert来验证，但是在没有TCERT的时候只能够用
//ECert 进行充当TCERT 所以需要用ECA.CERT 即ECA.CA 作为根证书进行验证
//VerifyTCert verify the TCert is valid or not
func (cm *CAManager)VerifyTCert(tcertPEM string,method string) (bool, error) {
	log := common.GetLogger(common.NAMESPACE, "ca")
	// if check TCert flag is false, default return true
	if !cm.checkTCert {
		return true, nil
	}
	tcert, err := primitives.ParseCertificate([]byte(tcertPEM))
	if err != nil {
		cm.logger.Error(errParseCert.Error())
		return false, errParseCert
	}
	if tcert.IsCA == true{
		log.Error("tcert is CA !ERROE!")
		return false,errFailedVerifySign
	}

	//生成TCERT METHOD
	if strings.EqualFold("getTCert",method) {
		ef,_ := primitives.VerifyCert(tcert, cm.eCaCert.x509cert)
		if ef {
			return true,nil
		}else {
			return false, errFailedVerifySign

		}
	}

	//其他METHOD
	db,err := hyperdb.GetDBDatabase()
	if err!=nil {
		log.Error(err)
		return false,&common.CertError{Message: "Get Database failed"};
	}
	certs,err := db.Get([]byte(CertKey))
	if err!=nil {
		log.Error("This node has not gen tcert:",err)
		return  false,&common.CertError{Message: "This node has not gen tcert!"};
	}
	regs := struct {
		Tcerts []string
	}{}
	_,err = asn1.Unmarshal(certs,&regs)
	if err!=nil {
		log.Error(err)
		return  false,&common.CertError{Message: "UnMarshal cert lists failed"};
	}
	for _,v := range regs.Tcerts  {
		if strings.EqualFold(v,tcertPEM) {
			tf,_:= primitives.VerifyCert(tcert, cm.tCacert.x509cert)
			if tf {
				return true,nil
			}else {
				return false, errFailedVerifySign

			}
		}
	}
	log.Error("Node has not gen this Tcert!Please check it")
	return false, errFailedVerifySign
}

// verify the ecert is valid or not
func (cm *CAManager) VerifyECert(ecertPEM string) (bool, error) {
	if !cm.checkERCert {
		return true,nil
	}
	// if SDK hasn't TCert it can use the ecert to send the transaction
	// but if the switch is off, this will not check the ecert is valid or not.
	ecertToVerify, err := primitives.ParseCertificate([]byte(ecertPEM))
	if err != nil {
		cm.logger.Error(errParseCert.Error())
		return false, errParseCert
	}
	return primitives.VerifyCert(ecertToVerify, cm.eCaCert.x509cert)
}

/**
验证签名，验证签名需要有三个参数：
第一个是携带的数字证书，即tcert,
第二个是签名，
第三个是原始数据
这个方法用来验证签名是否来自数字证书用户
*/
//VerifyCertSignature Verify the Signature of Cert
func (cm *CAManager) VerifyCertSign(certPEM string, msg, sign []byte) (bool, error) {
	// if checkCertSign == false, return true and nil
	if !cm.checkCertSign {
		return true, nil
	}
	ecert, err := primitives.ParseCertificate([]byte(certPEM))
	if err != nil {
		cm.logger.Error(errParseCert.Error())
		return false, errParseCert
	}
	ecdsaEncrypto := primitives.NewEcdsaEncrypto("ecdsa")

	key := ecert.PublicKey.(*(ecdsa.PublicKey))
	result, err := ecdsaEncrypto.VerifySign(*key, msg, sign)
	if err != nil {
		cm.logger.Critical(err)
		cm.logger.Error(errFailedVerifySign.Error())
		return false, errFailedVerifySign
	}
	return result, nil
}
//VerifyRCert verify the rcert is valid or not
func (cm *CAManager) VerifyRCert(rcertPEM string) (bool, error) {
	if cm.checkERCert {
		return true,nil
	}
	rcert, err := primitives.ParseCertificate([]byte(rcertPEM))
	if err != nil {
		cm.logger.Error(errParseCert)
		return false, errParseCert
	}
	return primitives.VerifyCert(rcert, cm.rCaCert.x509cert)
}

/**
getMethods
*/

func (caManager *CAManager) GetECACertByte() []byte {
	return caManager.eCaCert.certByte
}
func (caManager *CAManager) GetECertByte() []byte {
	return caManager.eCert.certByte
}
func (caManager *CAManager) GetRCertByte() []byte {
	return caManager.rCert.certByte
}
func (caManager *CAManager) GetRCAcertByte() []byte {
	return caManager.rCaCert.certByte
}
func (caManager *CAManager) GetECertPrivateKeyByte() []byte {
	return caManager.eCertPri.prikeybyte
}
func (caManager *CAManager) GetECertPrivKey() interface{} {
	return caManager.eCertPri.priKey
}
func (caManager *CAManager) IsCheckSign() bool {
	return caManager.checkCertSign
}
func (caManager *CAManager) IsCheckTCert() bool {
	return caManager.checkTCert
}

// tool method for read cert file
func readCert(path string) (*cert, error) {
	certb, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	certs, err := primitives.ParseCertificate(certb)
	if err != nil {
		return nil, err
	}
	return &cert{
		x509cert:certs,
		certByte:certb,
	}, nil
}

//tool method for read key
func readKey(path string) (*key, error) {
	keyb, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	priKey, err := primitives.ParseKey(keyb)
	if err != nil {
		return nil, errors.New("cannot parse the caprivatekey")
	}
	return &key{
		priKey:priKey,
		prikeybyte:keyb,
	}, nil
}

func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix) //忽略后缀匹配的大小写
	for _, fi := range dir {
		if fi.IsDir() { // 忽略目录
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) { //匹配文件
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

func RegisterCert(tcert []byte) error {
	log := common.GetLogger(common.DEFAULT_NAMESPACE, "api")
	db,err := hyperdb.GetDBDatabase()
	if err!=nil {
		log.Error(err)
		return &common.CertError{Message: "Get Database failed"};
	}
	certs,err := db.Get([]byte(CertKey))
	tcertStr := string(tcert)
	//First to Save CertList
	if err != nil{
		//log.Critical("Register TCERT:",tcertStr)
		regLists := RegisterTcerts{[]string{tcertStr}}
		lists,err := asn1.Marshal(regLists)
		if err!= nil{
			log.Error(err)
			return &common.CertError{Message: "Marshal cert lists failed"};
		}
		err = db.Put([]byte(CertKey),lists)
		if err!= nil{
			log.Error(err)
			return &common.CertError{Message: "Save cert lists failed"};
		}
		return nil
	}
	//log.Critical("GET CERT LIST FROM DB:",certs)
	Regs := struct {
		Tcerts []string
	}{}
	_,err = asn1.Unmarshal(certs,&Regs)
	if err!=nil {
		log.Error(err)
		return  &common.CertError{Message: "UnMarshal cert lists failed"};
	}
	Regs.Tcerts = append(Regs.Tcerts,tcertStr)
	lists,err := asn1.Marshal(Regs)
	if err!= nil{
		log.Error(err)
		return &common.CertError{Message: "Marshal cert lists failed"};
	}
	err = db.Put([]byte(CertKey),lists)
	if err!= nil{
		log.Error(err)
		return &common.CertError{Message: "Save cert lists failed"};
	}
	return nil
}