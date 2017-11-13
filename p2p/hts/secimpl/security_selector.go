package secimpl

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/terasum/viper"
)

// NewSecuritySelector will select the appropriate security encryption algorithm
// according to the specified configuration.
func NewSecuritySelector(caconf string) Security {
	log := common.GetLogger(common.DEFAULT_NAMESPACE, "p2p")
	// read in config, and get all certs
	vip := viper.New()
	vip.SetConfigFile(caconf)
	err := vip.ReadInConfig()
	if err != nil {
		log.Error(fmt.Sprintf("cann't read in the caconfig, reason: %s ", err.Error()))
		return nil
	}
	algo := vip.GetString(common.ENCRYPTION_SECURITY_ALGO)
	switch algo {
	case "pure":
		log.Notice("Use ECDH WITH PURE ALGO!")
		return NewECDHWithPURE()
	case "sm4":
		log.Notice("Use ECDH WITH SM4 ALGO!")
		return NewECDHWithSM4()
	case "3des":
		log.Notice("Use ECDH WITH 3DES ALGO!")
		return NewECDHWith3DES()
	case "aes":
		log.Notice("Use ECDH WITH AES ALGO!")
		return NewECDHWithAES()
	default:
		log.Error("Unknow symmetric encryption algorithm,please modify cacondig.yaml and restart node!")
		return NewECDHWithPURE()
	}
}
