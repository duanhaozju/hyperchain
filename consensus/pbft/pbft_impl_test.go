package pbft

import (
	"testing"
	"hyperchain/common"
	"github.com/spf13/viper"
	"path/filepath"
	"hyperchain/consensus/helper"
	"hyperchain/manager/event"
	"fmt"
	"hyperchain/consensus"
	"github.com/facebookgo/ensure"
	"hyperchain/core/db_utils"
)

func NewConfig(path ,name string)(*common.Config){
	conf := common.NewConfig(path)
	common.InitHyperLoggerManager(conf)
	conf.Set(common.NAMESPACE, name)
	common.InitHyperLogger(conf)
	// init peer configurations
	peerConfigPath := conf.GetString("global.configs.peers")
	peerViper := viper.New()
	peerViper.SetConfigFile(filepath.Join("../../build/node1/",peerConfigPath))
	err := peerViper.ReadInConfig()
	if err != nil {
		panic(err)
	}
	conf.Set(common.C_NODE_ID, peerViper.GetInt("self.node_id"))
	conf.Set(common.C_HTTP_PORT, peerViper.GetInt("self.jsonrpc_port"))
	conf.Set(common.C_REST_PORT, peerViper.GetInt("self.restful_port"))
	conf.Set(common.P2P_PORT, peerViper.GetInt("self.grpc_port"))
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, path)
	conf.Set(common.JVM_PORT, peerViper.GetInt("self.jvm_port"))
	conf.Set(common.LEDGER_PORT, peerViper.GetInt("self.ledger_port"))

	conf.Set(common.NAMESPACE, name)

	pcPath := conf.GetString(consensus.CONSENSUS_ALGO_CONFIG_PATH)
	if pcPath == "" {
		err = fmt.Errorf("Invalid consensus algorithm configuration path, %s: %s",
			consensus.CONSENSUS_ALGO_CONFIG_PATH, pcPath)
		panic(err)
	}
	conf, err = conf.MergeConfig(filepath.Join("../../build/node1/",pcPath))
	if err != nil {
		err = fmt.Errorf("Load pbft config error: %v", err)
		panic(err)
	}

	conf, err = conf.MergeConfig(filepath.Join("../../build/node1/","namespaces/global/config/pbft.yaml"))
	if err != nil {
		err = fmt.Errorf("Load pbft config error: %v", err)
		panic(err)
	}

	return conf
}


func TestPbftImpl_func1(t *testing.T) {

	//new PBFT
	conf:=NewConfig("../../build/node1/namespaces/global/config/global.yaml","global")
	_=conf
	conf.Set(common.DB_CONFIG_PATH,filepath.Join("../../build/node1/",conf.GetString(common.DB_CONFIG_PATH)))
	err := db_utils.InitDBForNamespace(conf, "global")
	if err != nil {
		t.Errorf("init db for namespace: %s error, %v", "global", err)
	}

	h:=helper.NewHelper(new(event.TypeMux))
	pbft,err:=newPBFT("global",conf,h)

	ensure.Nil(t,err)

	ensure.DeepEqual(t,pbft.namespace,"global")
	ensure.DeepEqual(t,pbft.activeView,uint32(1))
	ensure.DeepEqual(t,pbft.f,(pbft.N - 1) / 3)
	ensure.DeepEqual(t,pbft.N,conf.GetInt(PBFT_NODE_NUM))
	ensure.DeepEqual(t,pbft.h,uint64(0))
	ensure.DeepEqual(t,pbft.id,uint64(conf.GetInt64(common.C_NODE_ID)))
	ensure.DeepEqual(t,pbft.K,uint64(10))
	ensure.DeepEqual(t,pbft.logMultiplier,uint64(4))
	ensure.DeepEqual(t,pbft.L,pbft.logMultiplier * pbft.K)
	ensure.DeepEqual(t,pbft.seqNo,uint64(0))
	ensure.DeepEqual(t,pbft.view,uint64(0))
	ensure.DeepEqual(t,pbft.nvInitialSeqNo,uint64(0))

	//Test Consenter interface

	pbft.Start()

	pbft.RecvLocal()

	//pbft.RecvMsg()
	pbft.Close()

}