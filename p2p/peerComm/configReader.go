package peerComm
//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

import (
	"io/ioutil"
	"encoding/json"
	"fmt"
	pb "hyperchain/p2p/peermessage"
)

type ConfigReader struct {
	config PeerConfig
	nodes   map[uint64]Address
	maxNode int
	path string
}

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
	configReader.nodes   = make(map[uint64]Address)
	configReader.path  = configpath

	slice := config.PeerNodes
	for _,node := range slice{
		log.Info(node.Address)
		temp_addr := Address{
			ID 	: uint64(node.ID),
			Port 	: int64(node.Port),
			RPCPort : int64(node.RPCPort),
			IP	:node.Address,
		}
		if _,ok:=configReader.nodes[uint64(node.ID)];!ok{
			configReader.nodes[uint64(node.ID)] = temp_addr
		}

	}
	return &configReader
}

//
func (conf *ConfigReader) GetPort(nodeID uint64) int64 {
	return conf.nodes[nodeID].Port
}

func (conf *ConfigReader) GetIP(nodeID uint64) string {
	return conf.nodes[nodeID].IP
}

//
func (conf *ConfigReader) GetMaxPeerNumber() int {
	return conf.maxNode

}




func (conf *ConfigReader) SaveAddress(addr Address) error{
	conf.nodes[addr.ID]=Address{
		ID:addr.ID,
		IP:addr.IP,
		Port:addr.Port,
		RPCPort:addr.RPCPort,
	}

	pcNode := PeerConfigNodes{
		ExternalAddress : addr.IP,
		RPCPort :int(addr.RPCPort),
		Port : int(addr.Port),
		ID :int(addr.ID),
		Address :addr.IP,
	}
	conf.config.Maxpeernode +=1
	conf.config.PeerNodes = append(conf.config.PeerNodes,pcNode)

	confSerlized,err := json.Marshal(conf.config)
	if err != nil{
		log.Warning(err)
	}


	fmt.Println("asda",string(confSerlized))

	ioutil.WriteFile(conf.path,confSerlized,0777)

	return nil
}

func persist(addr pb.PeerAddress) (address Address) {
	return Address{
		ID:addr.ID,
		IP:addr.IP,
		Port:addr.Port,
		RPCPort:addr.Port,
	}
}