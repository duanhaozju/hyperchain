package pbft

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hyperchain-alpha/consensus"
	"hyperchain-alpha/event"
	"hyperchain-alpha/consensus/helper"

	"github.com/spf13/viper"
)

const configPrefix = "CORE_PBFT"

var pluginInstance consensus.Consenter // singleton service
var config *viper.Viper

func init() {
	config = loadConfig()
}

// GetPlugin returns the handle to the Consenter singleton
<<<<<<< HEAD
func GetPlugin(id uint64, msgQ event.TypeMux) consensus.Consenter {
=======
func GetPlugin(id uint64, h helper.Stack) consensus.Consenter {
>>>>>>> e0e36f6ba3e89a1ec153f1f17441eca934d83e77
	if pluginInstance == nil {
		pluginInstance = New(id, h)
	}
	return pluginInstance
}

// New creates a new Obc* instance that provides the Consenter interface.
// Internally, it uses an opaque pbft-core instance.
<<<<<<< HEAD
func New(id uint64, msgQ event.TypeMux) consensus.Consenter {
=======
func New(id uint64, h *helper.Stack) consensus.Consenter {
>>>>>>> e0e36f6ba3e89a1ec153f1f17441eca934d83e77
	switch strings.ToLower(config.GetString("general.mode")) {
	case "batch":
		return newBatch(id, config, h)
	default:
		panic(fmt.Errorf("Invalid PBFT mode: %s", config.GetString("general.mode")))
	}
}

func loadConfig() (config *viper.Viper) {
	config = viper.New()

	// for environment variables
	config.SetEnvPrefix(configPrefix)
	config.AutomaticEnv()
	replacer := strings.NewReplacer(".", "_")
	config.SetEnvKeyReplacer(replacer)

	config.SetConfigName("config")
	config.AddConfigPath("./")
	config.AddConfigPath("../consensus/pbft")
	config.AddConfigPath("../../consensus/pbft")
	// Path to look for the config file in based on GOPATH
	gopath := os.Getenv("GOPATH")
	for _, p := range filepath.SplitList(gopath) {
		pbftpath := filepath.Join(p, "src/github.com/hyperchain/hyperchain/consensus/pbft")
		config.AddConfigPath(pbftpath)
	}

	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error reading %s plugin config: %s", configPrefix, err))
	}
	return
}
