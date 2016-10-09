package pbft

import (
	"fmt"
	//"os"
	//"path/filepath"
	"strings"

	"hyperchain/consensus"
	"hyperchain/consensus/helper"
	"github.com/spf13/viper"
	//"path"
)

const configPrefix = "CORE_PBFT"

var pluginInstance consensus.Consenter // singleton service
var config *viper.Viper

// GetPlugin returns the handle to the Consenter singleton
func GetPlugin(id uint64, h helper.Stack, pbftConfigPath string) consensus.Consenter {

	if pluginInstance == nil {
		pluginInstance = New(id, h, pbftConfigPath)
	}

	return pluginInstance
}

// New creates a new *batch instance that provides the Consenter interface.
// Internally, it uses an opaque pbft-core instance.
func New(id uint64, h helper.Stack, pbftConfigPath string) consensus.Consenter {

	config = loadConfig(pbftConfigPath)
	switch strings.ToLower(config.GetString("general.mode")) {
	case "batch":
		return newPbft(id, config, h)
	default:
		panic(fmt.Errorf("Invalid PBFT mode: %s", config.GetString("general.mode")))
	}

}


// loadConfig load the config in the config.yaml
func loadConfig(pbftConfigPath string) (config *viper.Viper) {

	config = viper.New()

	// for environment variables
	config.SetEnvPrefix(configPrefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)

	config.SetConfigName("config")
	//config.AddConfigPath("./")
	//config.AddConfigPath("../consensus/pbft")
	//config.AddConfigPath("../../consensus/pbft")
	// Path to look for the config file in based on GOPATH
	//gopath := os.Getenv("GOPATH")
	//for _, p := range filepath.SplitList(gopath) {
	//	pbftpath := filepath.Join(p, pbftConfigPath)
	//	config.AddConfigPath(pbftpath)
	//}
	config.AddConfigPath(pbftConfigPath)


	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", configPrefix, err))
	}

	return
}
