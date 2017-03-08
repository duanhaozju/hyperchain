//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"fmt"
	"strings"

	"hyperchain/consensus"
	"hyperchain/consensus/helper"

	"github.com/spf13/viper"
)

const configPrefix = "CORE_PBFT"

var pluginInstance consensus.Consenter // singleton service
var config *viper.Viper

// GetPlugin returns the handle to the Consenter singleton
func GetPlugin(namespace string, id uint64, h helper.Stack, pbftConfigPath string) consensus.Consenter {

	if pluginInstance == nil {
		pluginInstance = New(namespace, id, h, pbftConfigPath)
	}

	return pluginInstance
}

// New creates a new *batch instance that provides the Consenter interface.
// Internally, it uses an opaque pbft-core instance.
func New(namespace string, id uint64, h helper.Stack, pbftConfigPath string) consensus.Consenter {

	config = loadConfig(pbftConfigPath)
	return newPbft(namespace, id, config, h)

}

// loadConfig load the config in the config.yaml
func loadConfig(pbftConfigPath string) (config *viper.Viper) {

	config = viper.New()

	// for environment variables
	config.SetEnvPrefix(configPrefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)
	config.SetConfigFile(pbftConfigPath)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", configPrefix, err))
	}

	return
}
