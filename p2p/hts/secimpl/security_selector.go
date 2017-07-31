package secimpl

import (
	"hyperchain/p2p/hts"
	"fmt"
	"github.com/terasum/viper"
	"github.com/op/go-logging"
	"hyperchain/common"
)

var log *logging.Logger

func NewSecuritySelector(caconf string)(hts.Security){
	log = common.GetLogger(common.DEFAULT_NAMESPACE, "p2p")
	// read in config, and get all certs
	vip := viper.New()
	vip.SetConfigFile(caconf)
	err := vip.ReadInConfig()
	if err != nil{
		log.Error(fmt.Sprintf("cann't read in the caconfig, reason: %s ",err.Error()))
		return nil
	}
	algo := vip.GetString(common.ENCRYPTION_SECURITY_ALGO)
	switch algo{
	case "pure":
		log.Critical("Use ECDH WITH PURE ALGO!")
		return NewECDHWithPURE()
	case "sm4":
		log.Critical("Use ECDH WITH SM4 ALGO!")
		return NewECDHWithSM4()
	case "3des":
		log.Critical("Use ECDH WITH 3DES ALGO!")
		return NewECDHWithAES()
	default:
		log.Error("Unknow symmetric encryption algorithm,please modify cacondig.yaml and restart node!")
		return NewECDHWithPURE()
	}
}