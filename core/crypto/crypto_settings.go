//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"github.com/op/go-logging"
	"github.com/spf13/viper"
	"hyperchain/core/crypto/primitives"
)

var (
	log = logging.MustGetLogger("crypto")
)

// Init initializes the crypto layer. It load from viper the security level
// and the logging setting.
func Init() (err error) {
	// Init security level
	securityLevel := 256
	if viper.IsSet("security.level") {
		ovveride := viper.GetInt("security.level")
		if ovveride != 0 {
			securityLevel = ovveride
		}
	}

	hashAlgorithm := "SHA3"
	if viper.IsSet("security.hashAlgorithm") {
		ovveride := viper.GetString("security.hashAlgorithm")
		if ovveride != "" {
			hashAlgorithm = ovveride
		}
	}

	log.Debugf("Working at security level [%d]", securityLevel)
	if err = primitives.InitSecurityLevel(hashAlgorithm, securityLevel); err != nil {
		log.Errorf("Failed setting security level: [%s]", err)

		return
	}

	return
}
