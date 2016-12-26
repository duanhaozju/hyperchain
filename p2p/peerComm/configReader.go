package peerComm
//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

import (
	"io/ioutil"
	"encoding/json"
)

type ConfigReader struct {
	config PeerConfig
	nodes   map[int]Address
	maxNode int
	path string
}

// TODO return a error next to the configReader or throw a panic
func NewConfigReader(configpath string) *ConfigReader {
	content,err := ioutil.ReadFile(configpath)

	if err != nil{
		log.Error(err);
	}
	config := PeerConfig{}
	err = json.Unmarshal(content,&config);
	if err != nil{
		log.Error(err);
	}
	var configReader ConfigReader
	configReader.config  = config
	configReader.maxNode = config.Maxpeernode;
	configReader.nodes   = make(map[int]Address)
	configReader.path  = configpath

	slice := config.PeerNodes
	for _,node := range slice{
		log.Info(node.Address)
		temp_addr := Address{
			ID 	: node.ID,
			Port 	: node.Port,
			RPCPort : node.RPCPort,
			IP	:node.Address,
		}
		if _,ok:=configReader.nodes[node.ID];!ok{
			configReader.nodes[node.ID] = temp_addr
		}

	}
	return &configReader
}

func (conf *ConfigReader) GetLocalID() int{
	return conf.config.SelfConfig.NodeID
}

func (conf *ConfigReader) GetLocalIP() string {
	return conf.config.SelfConfig.LocalIP
}

func (conf *ConfigReader) GetLocalGRPCPort() int{
	return conf.config.SelfConfig.GrpcPort
}

func (conf *ConfigReader) GetLocalJsonRPCPort() int{
	return conf.config.SelfConfig.JsonrpcPort
}

func (conf *ConfigReader) GetIntroducerIP() string{
	return conf.config.SelfConfig.IntroducerIP
}

func (conf *ConfigReader) GetIntroducerID() int {
	return conf.config.SelfConfig.IntroducerID
}

func (conf *ConfigReader) GetIntroducerJSONRPCPort() int {
	return conf.config.SelfConfig.IntroducerJSONRPCPort
}

func (conf *ConfigReader) GetIntroducerPort() int{
	return conf.config.SelfConfig.IntroducerPort
}

func (conf *ConfigReader) IsOrigin() bool{
	return conf.config.SelfConfig.IsOrigin
}

func (conf *ConfigReader) GetPort(nodeID int) int {
	return conf.nodes[nodeID].Port
}

func (conf *ConfigReader) GetIP(nodeID int)string{
	return conf.nodes[nodeID].IP
}

func (conf *ConfigReader) GetMaxPeerNumber()int{
	return conf.maxNode
}