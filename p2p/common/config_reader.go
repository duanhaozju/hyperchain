package common

//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

import (
	"encoding/json"
	pb "hyperchain/p2p/message"
	"io/ioutil"
	"sync"
	"github.com/op/go-logging"
	"hyperchain/common"
)

type ConfigReader struct {
	Config    PeerConfig
	nodes     map[int]Address
	cNodes    []PeerConfigNodes
	maxNode   int
	path      string
	writeLock sync.Mutex
	logger *logging.Logger
}

// TODO return a error next to the configReader or throw a panic
func NewConfigReader(configpath, namespace string) *ConfigReader {
	content, err := ioutil.ReadFile(configpath)
	logger := common.GetLogger(namespace, "p2p/common")
	if err != nil {
		logger.Error(err)
		return nil
	}
	config := PeerConfig{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		logger.Error(err)
		return nil

	}
	var configReader ConfigReader
	configReader.Config = config
	configReader.maxNode = config.Maxpeernode
	configReader.nodes = make(map[int]Address)
	configReader.path = configpath
	configReader.cNodes = config.PeerNodes
	slice := config.PeerNodes
	configReader.logger = logger
	for _, node := range slice {
		temp_addr := Address{
			ID:      node.ID,
			Port:    node.Port,
			RPCPort: node.RPCPort,
			IP:      node.Address,
		}
		if _, ok := configReader.nodes[node.ID]; !ok {
			configReader.nodes[node.ID] = temp_addr
		}

	}
	return &configReader
}

func (conf *ConfigReader) Peers() []PeerConfigNodes {
	return conf.cNodes
}

func (conf *ConfigReader) LocalID() int {
	return conf.Config.SelfConfig.NodeID
}

func (conf *ConfigReader) LocalIP() string {
	return conf.Config.SelfConfig.LocalIP
}

func (conf *ConfigReader) LocalGRPCPort() int {
	return conf.Config.SelfConfig.GrpcPort
}

func (conf *ConfigReader) LocalJsonRPCPort() int {
	return conf.Config.SelfConfig.JsonrpcPort
}

func (conf *ConfigReader) IntroIP() string {
	return conf.Config.SelfConfig.IntroducerIP
}

func (conf *ConfigReader) IntroID() int {
	return conf.Config.SelfConfig.IntroducerID
}

func (conf *ConfigReader) IntroJSONRPCPort() int {
	return conf.Config.SelfConfig.IntroducerJSONRPCPort
}

func (conf *ConfigReader) IntroPort() int {
	return conf.Config.SelfConfig.IntroducerPort
}

func (conf *ConfigReader) IsOrigin() bool {
	return conf.Config.SelfConfig.IsOrigin
}

func (conf *ConfigReader) IsVP() bool {
	return conf.Config.SelfConfig.IsVP
}

func (conf *ConfigReader) GetPort(nodeID int) int {
	return conf.nodes[nodeID].Port
}

func (conf *ConfigReader) GetID(nodeID int) int {
	return conf.nodes[nodeID].ID
}

func (conf *ConfigReader) GetRPCPort(nodeID int) int {
	return conf.nodes[nodeID].RPCPort
}

func (conf *ConfigReader) GetIP(nodeID int) string {
	return conf.nodes[nodeID].IP
}

func (conf *ConfigReader) MaxNum() int {
	return conf.maxNode
}

func (conf *ConfigReader) persist() error {
	conf.writeLock.Lock()
	defer conf.writeLock.Unlock()
	content, err := json.Marshal(conf.Config)
	if err != nil {
		conf.logger.Error("persist the peerconfig failed, json marshal failed!")
		return err
	}
	err = ioutil.WriteFile(conf.path, content, 655)
	if err != nil {
		conf.logger.Error("persist the peerconfig failed, write file failed!")
		return err
	}
	return nil
}

func (conf *ConfigReader) addNode(addr pb.PeerAddr) {
	newAddress := NewAddress(addr.ID, addr.Port, addr.RPCPort, addr.IP)
	conf.nodes[addr.ID] = newAddress
	peerConfigNode := NewPeerConfigNodes(addr.IP, addr.RPCPort, addr.Port, addr.ID)
	conf.Config.PeerNodes = append(conf.Config.PeerNodes, *peerConfigNode)

}
func (conf *ConfigReader) updateNode(addr pb.PeerAddr) {
	if addr.ID < len(conf.Config.PeerNodes) {
		conf.Config.PeerNodes[addr.ID].ID = addr.ID
		conf.Config.PeerNodes[addr.ID].Address = addr.IP
		conf.Config.PeerNodes[addr.ID].Port = addr.Port
		conf.Config.PeerNodes[addr.ID].RPCPort = addr.RPCPort
	}
}

func (conf *ConfigReader) delNode(addr pb.PeerAddr) {
	conf.maxNode -= 1
	delete(conf.nodes, addr.ID)
	conf.Config.Maxpeernode -= 1
	conf.Config.PeerNodes = deleteElement(conf.Config.PeerNodes, addr)
}

func (conf *ConfigReader) AddNodesAndPersist(addrs map[string]pb.PeerAddr) {
	idx := 0
	for _, value := range addrs {
		if _, ok := conf.nodes[value.ID]; !ok {
			conf.logger.Debug("add a node", value.ID)
			conf.addNode(value)
		} //}else {
		//	conf.updateNode(value)
		//}
		idx++
		if idx == 1 {
			conf.Config.SelfConfig.IntroducerID = value.ID
			conf.Config.SelfConfig.IntroducerIP = value.IP
			conf.Config.SelfConfig.IntroducerPort = value.Port
			conf.Config.SelfConfig.IntroducerJSONRPCPort = value.RPCPort
		}
	}
	conf.maxNode = len(addrs) + 1
	conf.Config.Maxpeernode = len(addrs) + 1
	conf.persist()
}
func (conf *ConfigReader) DelNodesAndPersist(addrs map[string]pb.PeerAddr) {
	for _, value := range addrs {
		if _, ok := conf.nodes[value.ID]; ok {
			if value.ID < conf.Config.SelfConfig.NodeID {
				conf.Config.SelfConfig.NodeID--
			}
			conf.delNode(value)
		}
	}
	conf.persist()
}

func deleteElement(nodes []PeerConfigNodes, addr pb.PeerAddr) []PeerConfigNodes {
	result := make([]PeerConfigNodes, 0)
	index := 0
	for k, v := range nodes {
		if v.ID == addr.ID {
			result = append(result, nodes[index:k]...)
			index = k + 1
		} else if v.ID > addr.ID {
			nodes[k].ID = nodes[k].ID - 1
			result = append(result, nodes[k])
		}
	}
	return result
}
