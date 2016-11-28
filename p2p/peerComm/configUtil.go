//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package peerComm

import (
	"fmt"
	"reflect"

	"github.com/spf13/viper"
)

type Config interface {
	GetPort(nodeID uint64) int64
	GetIP(nodeID uint64) string
	GetMaxPeerNumber() int
}

type ConfigUtil struct {
	configs *viper.Viper
	nodes   map[uint64]Address
	maxNode int
}

func NewConfigUtil(configDir string) *ConfigUtil {
	var newConfigUtil ConfigUtil
	newConfigUtil.configs = getConfig(configDir)
	newConfigUtil.maxNode = newConfigUtil.configs.GetInt("maxpeernode")
	log.Info(newConfigUtil.maxNode)
	newConfigUtil.nodes = make(map[uint64]Address)
	slice := newConfigUtil.configs.Get("nodes")
	s := reflect.ValueOf(slice)
	if s.Kind() != reflect.Slice {
		panic("InterfaceSlice() given a non-slice type")
	}

	for i := 0; i < s.Len(); i++ {
		_tmp_var := s.Index(i).Interface()
		if _tmp_map, ok := _tmp_var.(map[string]interface{}); ok {
			log.Info(_tmp_map, reflect.TypeOf(_tmp_map))
			_node_id := (int64)(_tmp_map["id"].(float64))
			_node_port := (int64)(_tmp_map["port"].(float64))
			_node_address := _tmp_map["address"].(string)
			_rpc_rpcport := (int64)(_tmp_map["rpc_port"].(float64))
			temp_addr := NewAddress(_node_id, _node_port, _rpc_rpcport, _node_address)
			newConfigUtil.nodes[uint64(_node_id)] = temp_addr
		}

	}

	//log.Info(newConfigUtil.nodes[2])

	return &newConfigUtil
}

//
func (confutil *ConfigUtil) GetPort(nodeID uint64) int64 {
	return confutil.nodes[nodeID].Port
}

func (confutil *ConfigUtil) GetIP(nodeID uint64) string {
	return confutil.nodes[nodeID].IP
}

//
func (confutil *ConfigUtil) GetMaxPeerNumber() int {
	return confutil.maxNode

}

// GetConfig this is a tool function for get the json file config
// configs return a viper instance
func getConfig(path string) (config *viper.Viper) {
	config = viper.New()
	config.SetEnvPrefix("P2P")
	//config.SetConfigName("peerconfig")
	//config.SetConfigType("json")
	//config.AddConfigPath(path)
	//
	config.SetConfigFile(path)
	err := config.ReadInConfig()
	if err != nil {
		panic(fmt.Errorf("Error %s reading %s", "P2P", err))
	}

	return
}
