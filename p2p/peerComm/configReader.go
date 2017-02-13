package peerComm

//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

import (
	"encoding/json"
	pb "hyperchain/p2p/peermessage"
	"io/ioutil"
	"sync"
)

type ConfigReader struct {
	Config    PeerConfig
	nodes     map[int]Address
	maxNode   int
	path      string
	writeLock sync.Mutex
}

// TODO return a error next to the configReader or throw a panic
func NewConfigReader(configpath string) *ConfigReader {
	content, err := ioutil.ReadFile(configpath)

	if err != nil {
		log.Error(err)
	}
	config := PeerConfig{}
	err = json.Unmarshal(content, &config)
	if err != nil {
		log.Error(err)
	}
	var configReader ConfigReader
	configReader.Config = config
	configReader.maxNode = config.Maxpeernode
	configReader.nodes = make(map[int]Address)
	configReader.path = configpath

	slice := config.PeerNodes
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

func (conf *ConfigReader) GetLocalID() int {
	return conf.Config.SelfConfig.NodeID
}

func (conf *ConfigReader) GetLocalIP() string {
	return conf.Config.SelfConfig.LocalIP
}

func (conf *ConfigReader) GetLocalGRPCPort() int {
	return conf.Config.SelfConfig.GrpcPort
}

func (conf *ConfigReader) GetLocalJsonRPCPort() int {
	return conf.Config.SelfConfig.JsonrpcPort
}

func (conf *ConfigReader) GetIntroducerIP() string {
	return conf.Config.SelfConfig.IntroducerIP
}

func (conf *ConfigReader) GetIntroducerID() int {
	return conf.Config.SelfConfig.IntroducerID
}

func (conf *ConfigReader) GetIntroducerJSONRPCPort() int {
	return conf.Config.SelfConfig.IntroducerJSONRPCPort
}

func (conf *ConfigReader) GetIntroducerPort() int {
	return conf.Config.SelfConfig.IntroducerPort
}

func (conf *ConfigReader) IsOrigin() bool {
	return conf.Config.SelfConfig.IsOrigin
}

func (conf *ConfigReader)IsVP()bool{
	return conf.Config.SelfConfig.IsVP
}

func (conf *ConfigReader) GetPort(nodeID int) int {
	return conf.nodes[nodeID].Port
}

func (conf *ConfigReader) GetID(nodeID int) int{
	return conf.nodes[nodeID].ID
}

func (conf *ConfigReader) GetRPCPort(nodeID int) int{
	return conf.nodes[nodeID].RPCPort
}

func (conf *ConfigReader) GetIP(nodeID int) string {
	return conf.nodes[nodeID].IP
}

func (conf *ConfigReader) GetMaxPeerNumber() int {
	return conf.maxNode
}


func (conf *ConfigReader) persist() error {
	conf.writeLock.Lock()
	defer conf.writeLock.Unlock()
	content, err := json.Marshal(conf.Config)
	if err != nil {
		log.Error("persist the peerconfig failed, json marshal failed!")
		return err
	}
	err = ioutil.WriteFile(conf.path, content, 655)
	if err != nil {
		log.Error("persist the peerconfig failed, write file failed!")
		return err
	}
	return nil
}

func (conf *ConfigReader) addNode(addr pb.PeerAddr) {
	conf.maxNode += 1
	newAddress := NewAddress(addr.ID, addr.Port, addr.RPCPort, addr.IP)
	conf.nodes[addr.ID] = newAddress
	conf.Config.Maxpeernode += 1
	peerConfigNode := NewPeerConfigNodes(addr.IP, addr.RPCPort, addr.Port, addr.ID)
	conf.Config.PeerNodes = append(conf.Config.PeerNodes, *peerConfigNode)

}

func (conf *ConfigReader) delNode(addr pb.PeerAddr) {
	conf.maxNode -= 1
	delete(conf.nodes, addr.ID)
	conf.Config.Maxpeernode -= 1
	conf.Config.PeerNodes = deleteElement(conf.Config.PeerNodes, addr)
}

func (conf *ConfigReader) AddNodesAndPersist(addrs map[string]pb.PeerAddr) {
	for _, value := range addrs {
		if _, ok := conf.nodes[value.ID]; !ok {
			conf.addNode(value)
		}
	}
	conf.persist()
}
func (conf *ConfigReader) DelNodesAndPersist(addrs map[string]pb.PeerAddr) {
	for _, value := range addrs {
		if _, ok := conf.nodes[value.ID]; ok {
			conf.delNode(value)
		}
	}
	conf.persist()
}

func deleteElement(nodes []PeerConfigNodes, addr pb.PeerAddr) []PeerConfigNodes {
	index := 0
	endIndex := len(nodes) - 1
	result := make([]PeerConfigNodes, 0)
	for k, v := range nodes {
		if v.ID == addr.ID {
			result = append(result, nodes[index:k]...)
			index = k + 1
		} else if k == endIndex {
			result = append(result, nodes[index:endIndex+1]...)
		}
	}
	return result
}
