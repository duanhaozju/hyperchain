// author: chenquan
// date: 16-9-20
// last modified: 16-9-20 13:09
// last Modified Author: chenquan
// change log:
//
package peerComm

import (
	"github.com/buger/jsonparser"
	"io/ioutil"
)

type Config interface {
	GetPort(nodeID uint64) int64
	GetIP(nodeID uint64) string
	GetMaxPeerNumber() int
}

type ConfigUtil struct {
	configs []byte
	nodes   map[uint64]Address
	maxNode int
}

func NewConfigUtil(configpath string) *ConfigUtil {
	var newConfigUtil ConfigUtil
	newConfigUtil.configs = getConfig(configpath)
	newConfigUtil.nodes = make(map[uint64]Address)
	_max_node, nodeErr := jsonparser.GetInt(newConfigUtil.configs, "maxpeernode")
	newConfigUtil.maxNode = int(_max_node)

	if nodeErr != nil {
		log.Error(nodeErr)
	}
	jsonparser.ArrayEach(newConfigUtil.configs, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		//fmt.Printf("Key: '%s'\n Value: '%s'\n Type: %s", string(key), string(value), dataType)
		_node_id, _ := jsonparser.GetInt(value, "id")
		_node_port, _ := jsonparser.GetInt(value, "port")
		_node_address, _ := jsonparser.GetString(value, "address")
		_rpc_rpcport, _ := jsonparser.GetInt(value, "rpc_port")
		temp_addr := NewAddress(_node_id, _node_port, _rpc_rpcport, _node_address)
		newConfigUtil.nodes[uint64(_node_id)] = temp_addr
	}, "nodes")
	log.Info(newConfigUtil.nodes)
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
// configs return a map[string]string
func getConfig(path string) []byte {
	content, fileErr := ioutil.ReadFile(path)
	if fileErr != nil {
		log.Fatal(fileErr)
	}
	return content
}
