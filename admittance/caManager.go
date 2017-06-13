package admittance

import (
	"crypto/ecdsa"
	"crypto/x509"
	"fmt"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
	"hyperchain/common"
	"hyperchain/core/crypto/primitives"
	"io/ioutil"
	"time"
)

var (
	errDecPubKey        = errors.New("cannot decode the public string,please encode the public string as right hex string")
	errParsePubKey      = errors.New("cannot parse the request publickey, please check the public string.")
	errParseCert        = errors.New("cannot parse the cert pem, please check your cert pem string.")
	errGenTCert         = errors.New("cannot generate the tcert, please check your public key, if not work, please contract the hyperchain maintainer")
	errDERToPEM         = errors.New("cannot convert the der format cert into pem format.")
	errFailedVerifySign = errors.New("Verify the Cert Signature failed, please use the correctly certificate")
)

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
	eCaCert  *cert
	eCert    *cert
	rCaCert  *cert
	rCert    *cert
	tCacert  *cert
	eCertPri *key
	//check flags
	checkERCert   bool
	checkTCert    bool
	checkCertSign bool

	// TLS part
	tlsCA                 string
	tlsCert               string
	tlsCertPriv           string
	tlsServerHostOverride string
	//security options
	enableTls         bool
	EnableSymmetrical bool

	//those options just put here temporary
	RetryTimeLimit     int
	RecoveryTimeLimit  int
	KeepAliveTimeLimit int
	KeepAliveInterval  time.Duration
	RetryTimeout       time.Duration
	RecoveryTimeout    time.Duration
	logger             *logging.Logger
}

var caManager *CAManager

//NewCAManager get a new ca manager instance
func NewCAManager(conf *common.Config) (*CAManager, error) {
	logger := common.GetLogger(conf.GetString(common.NAMESPACE), "ca")
	caconfPath := conf.GetString("global.configs.caconfig")
	logger.Debug(caconfPath)
	enableTLS := conf.GetBool("global.security.enabletls")
	enableSymmetrical := conf.GetBool("global.security.enablesymmetrical")

	//temp options
	retryTimeLimit := conf.GetInt("global.connection.retryTimeLimit")
	recoveryTimeLimit := conf.GetInt("global.connection.recoveryTimeLimit")
	keepAliveTimeLimit := conf.GetInt("global.connection.keepAliveTimeLimit")
	keepAliveIntervals := conf.GetString("global.connection.keepAliveInterval")
	keepAliveInterval, err := time.ParseDuration(keepAliveIntervals)
	if err != nil {
		return nil, errors.New(fmt.Sprintf("cannot parse the keep alive time duration,val: %v", keepAliveInterval))
	}
	retryTimeouts := conf.GetString("global.connection.retryTimeout")
	retryTimeout, err := time.ParseDuration(retryTimeouts)
	if err != nil {
		return nil, errors.New("cannot parse the retry timeout time duration")
	}

	recoveryTimeouts := conf.GetString("global.connection.retryTimeout")
	recoveryTimeout, err := time.ParseDuration(recoveryTimeouts)
	if err != nil {
		return nil, errors.New("cannot parse the recovery timeout time duration")
	}

	if caconfPath == "" {
		return nil, errors.New("cannot get the ca config file path.")
	}
	config := viper.New()
	config.SetConfigFile(caconfPath)
	err = config.ReadInConfig()
	if err != nil {
		return nil, errors.New("cannot read ca conf")
	}
	eca, err := readCert(config.GetString("ecert.ca"))
	if err != nil {
		return nil, err
	}
	ecert, err := readCert(config.GetString("ecert.cert"))
	if err != nil {
		return nil, err
	}
	rca, err := readCert(config.GetString("rcert.ca"))
	if err != nil {
		return nil, err
	}
	rcert, err := readCert(config.GetString("rcert.cert"))
	if err != nil {
		rcert = &cert{}
	}
	ecertpriv, err := readKey(config.GetString("ecert.priv"))
	if err != nil {
		return nil, err
	}

	return &CAManager{
		eCaCert:               eca,
		eCert:                 ecert,
		eCertPri:              ecertpriv,
		rCaCert:               rca,
		rCert:                 rcert,
		tCacert:               ecert,
		checkCertSign:         config.GetBool("check.certsign"),
		checkERCert:           config.GetBool("check.ercert"),
		checkTCert:            config.GetBool("check.tcert"),
		tlsCA:                 config.GetString("tlscert.ca"),
		tlsCert:               config.GetString("tlscert.cert"),
		tlsCertPriv:           config.GetString("tlscert.priv"),
		tlsServerHostOverride: config.GetString("tlscert.serverhostoverride"),
		enableTls:             enableTLS,
		EnableSymmetrical:     enableSymmetrical,
		RetryTimeLimit:        retryTimeLimit,
		RecoveryTimeLimit:     recoveryTimeLimit,
		KeepAliveTimeLimit:    keepAliveTimeLimit,
		KeepAliveInterval:     keepAliveInterval,
		RetryTimeout:          retryTimeout,
		RecoveryTimeout:       recoveryTimeout,
		logger:                logger,
	}, nil

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
func (cm *CAManager) VerifyTCert(tcertPEM string) (bool, error) {
	// if check TCert flag is false, default return true
	if !cm.checkTCert {
		return true, nil
	}
	tcert, err := primitives.ParseCertificate(tcertPEM)
	if err != nil {
		cm.logger.Error(errParseCert.Error())
		return false, errParseCert
	}
	if tcert.IsCA == true {
		return primitives.VerifyCert(tcert, cm.eCaCert.x509cert)
	} else {
		return primitives.VerifyCert(tcert, cm.tCacert.x509cert)

	}
}

// verify the ecert is valid or not
func (cm *CAManager) VerifyECert(ecertPEM string) (bool, error) {
	if !cm.checkERCert {
		return true, nil
	}
	// if SDK hasn't TCert it can use the ecert to send the transaction
	// but if the switch is off, this will not check the ecert is valid or not.
	ecertToVerify, err := primitives.ParseCertificate(ecertPEM)
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
	ecert, err := primitives.ParseCertificate(certPEM)
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
		return true, nil
	}
	rcert, err := primitives.ParseCertificate(rcertPEM)
	if err != nil {
		cm.logger.Error(errParseCert)
		return false, errParseCert
	}
	return primitives.VerifyCert(rcert, cm.rCaCert.x509cert)
}

/**
  tls ca get dial opts and server opts part
*/

//GetGrpcClientOpts get GrpcClient options
func (cm *CAManager) GetGrpcClientOpts() []grpc.DialOption {
	var opts []grpc.DialOption
	if !cm.enableTls {
		cm.logger.Warning("disable Client TLS")
		opts = append(opts, grpc.WithInsecure())
		return opts
	}
	creds, err := credentials.NewClientTLSFromFile(cm.tlsCA, cm.tlsServerHostOverride)
	if err != nil {
		panic("cannot get the TLS Cert")
	}
	opts = append(opts, grpc.WithTransportCredentials(creds))
	return opts
}

//GetGrpcServerOpts get server grpc options
func (cm *CAManager) GetGrpcServerOpts() []grpc.ServerOption {
	var opts []grpc.ServerOption
	if !cm.enableTls {
		cm.logger.Warning("disable Server TLS")
		return opts
	}
	creds, err := credentials.NewServerTLSFromFile(cm.tlsCert, cm.tlsCertPriv)
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
	certs, err := primitives.ParseCertificate(string(certb))
	if err != nil {
		return nil, err
	}
	return &cert{
		x509cert: certs,
		certByte: certb,
	}, nil
}

//tool method for read key
func readKey(path string) (*key, error) {
	keyb, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	priKey, err := primitives.ParseKey(string(keyb))
	if err != nil {
		return nil, errors.New("cannot parse the caprivatekey")
	}
	return &key{
		priKey:     priKey,
		prikeybyte: keyb,
	}, nil
}
