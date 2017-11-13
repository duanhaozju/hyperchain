package admittance

import (
	"crypto/ecdsa"
	"crypto/x509"
	"encoding/asn1"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto/primitives"
	"github.com/hyperchain/hyperchain/hyperdb"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"io/ioutil"
	"os"
	"strings"
)

var (
	errDecPubKey        = errors.New("cannot decode the public string,please encode the public string as right hex string")
	errParsePubKey      = errors.New("cannot parse the request publickey, please check the public string.")
	errParseCert        = errors.New("cannot parse the cert pem, please check your cert pem string.")
	errGenTCert         = errors.New("cannot generate the tcert, please check your public key, if not work, please contract the hyperchain maintainer")
	errDERToPEM         = errors.New("cannot convert the der format cert into pem format.")
	errFailedVerifySign = errors.New("Verify the Cert Signature failed, please use the correctly certificate")
)

const CertKey string = "tcerts"

// RegisterTcerts Lists
type RegisterTcerts struct {
	Tcerts []string
}

// Cert struct
type cert struct {
	x509cert *x509.Certificate
	certByte []byte
}

// Key struct
type key struct {
	priKey     interface{}
	prikeybyte []byte
}

// CAManager this struct is for Certificate Auth manager
type CAManager struct {
	eCaCert       *cert
	eCert         *cert
	rCaCert       *cert
	rCert         *cert
	tCacert       *cert
	eCertPri      *key
	rCertPri      *key
	checkERCert   bool
	checkTCert    bool
	checkCertSign bool

	logger *logging.Logger
}

// NewCAManager get a new ca manager instance.
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
		return nil, errors.New(fmt.Sprintf("cannot read ca conf,reason: %s", err.Error()))
	}
	eca, err := readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_ECERT_ECA)))
	if err != nil {
		return nil, err
	}
	ecert, err := readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_ECERT_ECERT)))
	if err != nil {
		return nil, err
	}
	ecertpriv, err := readKey(common.GetPath(namespace, config.GetString(common.ENCRYPTION_ECERT_PRIV)))
	if err != nil {
		return nil, err
	}

	enable := config.GetBool(common.ENCRYPTION_CHECK_ENABLE)

	var rca *cert
	var rcert *cert
	var rcertpriv *key
	if enable {
		rca, err = readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_RCERT_RCA)))
		if err != nil {
			return nil, err
		}
		rcert, err = readCert(common.GetPath(namespace, config.GetString(common.ENCRYPTION_RCERT_RCERT)))
		if err != nil {
			rcert = &cert{}
		}
		rcertpriv, err = readKey(common.GetPath(namespace, config.GetString(common.ENCRYPTION_RCERT_PRIV)))
		if err != nil {
			return nil, err
		}
	}

	//Init tcert white list.
	whiteList := config.GetBool(common.ENCRYPTION_TCERT_WHITELIST)
	if whiteList {
		logger.Critical("Init Tcert White list!")
		whiteListDir := common.GetPath(namespace, config.GetString(common.ENCRYPTION_TCERT_WHITELIST_DIR))
		tcertFiles, err := ListDir(whiteListDir, "cert")
		if err != nil {
			logger.Error("Init tcert white list failed :", err)
			return nil, &common.CertError{Message: "Init tcert white list failed"}
		}
		for _, tcertPath := range tcertFiles {
			cert, err := ioutil.ReadFile(tcertPath)
			if err != nil {
				return nil, err
			}
			err = RegisterCert(cert)
			if err != nil {
				if err != nil {
					logger.Error(err)
					return nil, &common.CertError{Message: "Init tcert white list failed"}
				}
			}
		}
	}

	return &CAManager{
		eCaCert:       eca,
		eCert:         ecert,
		eCertPri:      ecertpriv,
		rCertPri:      rcertpriv,
		rCaCert:       rca,
		rCert:         rcert,
		tCacert:       ecert,
		checkCertSign: config.GetBool(common.ENCRYPTION_CHECK_SIGN),
		checkERCert:   config.GetBool(common.ENCRYPTION_CHECK_ENABLE),
		checkTCert:    config.GetBool(common.ENCRYPTION_CHECK_ENABLE_T),
		logger:        logger,
	}, nil

}

// Generate a TCert for SDK client.
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

// VerifyTCert verify the TCert is valid or not.
func (cm *CAManager) VerifyTCert(tcertPEM string, method string) (bool, error) {
	// if check TCert flag is false, default return true
	if !cm.checkTCert {
		return true, nil
	}
	tcert, err := primitives.ParseCertificate([]byte(tcertPEM))
	if err != nil {
		cm.logger.Error(errParseCert.Error())
		return false, errParseCert
	}
	if tcert.IsCA == true {
		cm.logger.Error("tcert is CA !ERROE!")
		return false, errFailedVerifySign
	}

	if strings.EqualFold("getTCert", method) {
		ef, _ := primitives.VerifyCert(tcert, cm.eCaCert.x509cert)
		if ef {
			return true, nil
		} else {
			return false, errFailedVerifySign

		}
	}

	// Verify this tcert is in this node's tcert list.
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		cm.logger.Error(err)
		return false, &common.CertError{Message: "Get Database failed"}
	}
	certs, err := db.Get([]byte(CertKey))
	if err != nil {
		cm.logger.Error("This node has not gen tcert:", err)
		return false, &common.CertError{Message: "This node has not gen tcert!"}
	}
	regs := struct {
		Tcerts []string
	}{}
	_, err = asn1.Unmarshal(certs, &regs)
	if err != nil {
		cm.logger.Error(err)
		return false, &common.CertError{Message: "UnMarshal cert lists failed"}
	}
	for _, v := range regs.Tcerts {
		if strings.EqualFold(v, tcertPEM) {
			tf, _ := primitives.VerifyCert(tcert, cm.tCacert.x509cert)
			if tf {
				return true, nil
			} else {
				return false, errFailedVerifySign

			}
		}
	}
	cm.logger.Error("Node has not gen this Tcert!Please check it")
	return false, errFailedVerifySign
}

// Verify the ecert is valid or not.
func (cm *CAManager) VerifyECert(ecertPEM string) (bool, error) {
	if !cm.checkERCert {
		return true, nil
	}
	// If SDK hasn't TCert it can use the ecert to send the transaction
	// But if the switch is off, this will not check the ecert is valid or not.
	ecertToVerify, err := primitives.ParseCertificate([]byte(ecertPEM))
	if err != nil {
		cm.logger.Error(errParseCert.Error())
		return false, errParseCert
	}
	return primitives.VerifyCert(ecertToVerify, cm.eCaCert.x509cert)
}

// VerifyCertSignature Verify the Signature of Cert.
func (cm *CAManager) VerifyCertSign(certPEM string, msg, sign []byte) (bool, error) {
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

// VerifyRCert verify the rcert is valid or not.
func (cm *CAManager) VerifyRCert(rcertPEM string) (bool, error) {
	if !cm.checkERCert {
		return true, nil
	}
	rcert, err := primitives.ParseCertificate([]byte(rcertPEM))
	if err != nil {
		cm.logger.Error(errParseCert)
		return false, errParseCert
	}
	return primitives.VerifyCert(rcert, cm.rCaCert.x509cert)
}

// GetECACertByte get eca.cert in bytes.
func (caManager *CAManager) GetECACertByte() []byte {
	return caManager.eCaCert.certByte
}

// GetECertByte get ecert.cert in bytes.
func (caManager *CAManager) GetECertByte() []byte {
	return caManager.eCert.certByte
}

// GetRCertByte get rcert.cert in bytes.
func (caManager *CAManager) GetRCertByte() []byte {
	return caManager.rCert.certByte
}

// GetRCAcertByte get rca.cert in bytes.
func (caManager *CAManager) GetRCAcertByte() []byte {
	return caManager.rCaCert.certByte
}

// GetECertPrivateKeyByte get ecert.priv in bytes.
func (caManager *CAManager) GetECertPrivateKeyByte() []byte {
	return caManager.eCertPri.prikeybyte
}

// GetECertPrivKey get ecert.priv in interface.
func (caManager *CAManager) GetECertPrivKey() interface{} {
	return caManager.eCertPri.priKey
}

// IsCheckSign get checkCertSign value.
func (caManager *CAManager) IsCheckSign() bool {
	return caManager.checkCertSign
}

// IsCheckTCert get checkTCert value.
func (caManager *CAManager) IsCheckTCert() bool {
	return caManager.checkTCert
}

// ReadCert is a tool method for read cert file.
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
		x509cert: certs,
		certByte: certb,
	}, nil
}

// ReadKey is tool method for read key.
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
		priKey:     priKey,
		prikeybyte: keyb,
	}, nil
}

// ListDir is tool method for traversing files in a folder.
func ListDir(dirPth string, suffix string) (files []string, err error) {
	files = make([]string, 0, 10)
	dir, err := ioutil.ReadDir(dirPth)
	if err != nil {
		return nil, err
	}
	PthSep := string(os.PathSeparator)
	suffix = strings.ToUpper(suffix)
	for _, fi := range dir {
		if fi.IsDir() {
			continue
		}
		if strings.HasSuffix(strings.ToUpper(fi.Name()), suffix) {
			files = append(files, dirPth+PthSep+fi.Name())
		}
	}
	return files, nil
}

// Register tcert in tcerts list.
func RegisterCert(tcert []byte) error {
	log := common.GetLogger(common.DEFAULT_NAMESPACE, "api")
	db, err := hyperdb.GetDBDatabase()
	if err != nil {
		log.Error(err)
		return &common.CertError{Message: "Get Database failed"}
	}

	certs, err := db.Get([]byte(CertKey))
	tcertStr := string(tcert)

	// First to save tcert list in db.
	if err != nil {
		regLists := RegisterTcerts{[]string{tcertStr}}
		lists, err := asn1.Marshal(regLists)
		if err != nil {
			log.Error(err)
			return &common.CertError{Message: "Marshal cert lists failed"}
		}
		err = db.Put([]byte(CertKey), lists)
		if err != nil {
			log.Error(err)
			return &common.CertError{Message: "Save cert lists failed"}
		}
		return nil
	}

	Regs := struct {
		Tcerts []string
	}{}
	_, err = asn1.Unmarshal(certs, &Regs)
	if err != nil {
		log.Error(err)
		return &common.CertError{Message: "UnMarshal cert lists failed"}
	}
	Regs.Tcerts = append(Regs.Tcerts, tcertStr)
	lists, err := asn1.Marshal(Regs)
	if err != nil {
		log.Error(err)
		return &common.CertError{Message: "Marshal cert lists failed"}
	}
	err = db.Put([]byte(CertKey), lists)
	if err != nil {
		log.Error(err)
		return &common.CertError{Message: "Save cert lists failed"}
	}
	return nil
}
