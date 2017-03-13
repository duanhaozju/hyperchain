//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"errors"

	"github.com/mkideal/cli"
	"github.com/terasum/viper"

	"hyperchain/accounts"
	"hyperchain/admittance"
	"hyperchain/api/jsonrpc/core"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/event"
	"hyperchain/p2p"
	"hyperchain/consensus/csmgr"
	"hyperchain/core/db_utils"
	"hyperchain/core/executor"
	"hyperchain/manager"
)

type argT struct {
	cli.Helper
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./config/global.yaml"`
	//ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./configuration/global.yaml"`
}

const (
	DefaultNamespace = "Global"
)

func initGloableConfig(argv *argT) *common.Config {
	//default path: ./configuration/global.yaml
	conf := common.NewConfig(argv.ConfigPath)
	return conf
}

func initConf(argv *argT) *common.Config {
	conf := common.NewConfig(argv.ConfigPath)
	// init peer configurations
	peerConfigPath := conf.GetString("global.configs.peers")
	peerViper := viper.New()
	peerViper.SetConfigFile(peerConfigPath)
	err := peerViper.ReadInConfig()
	if err != nil {
		panic("read in the peer config failed")
	}
	nodeID := peerViper.GetInt("self.node_id")
	grpcPort := peerViper.GetInt("self.grpc_port")
	jsonrpcPort := peerViper.GetInt("self.jsonrpc_port")
	restfulPort := peerViper.GetInt("self.restful_port")

	conf.Set(common.C_NODE_ID, nodeID)
	conf.Set(common.C_HTTP_PORT, jsonrpcPort)
	conf.Set(common.C_REST_PORT, restfulPort)
	conf.Set(common.C_GRPC_PORT, grpcPort)
	conf.Set(common.C_PEER_CONFIG_PATH, peerConfigPath)
	conf.Set(common.C_GLOBAL_CONFIG_PATH, argv.ConfigPath)

	return conf
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		conf := initConf(argv)

		common.InitLog(conf)

		db_utils.InitDBForNamespace(conf, DefaultNamespace)
		eventMux := new(event.TypeMux)

		cm, cmerr := admittance.GetCaManager(conf)
		if cmerr != nil {
			panic("cannot initliazied the camanager")
		}

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(conf)

		//init pbft consensus
		consenter := csmgr.Consenter(DefaultNamespace, conf, eventMux)
		consenter.Start()

		am := accounts.NewAccountManager(conf)
		am.UnlockAllAccount(conf.GetString(common.KEY_STORE_DIR))

		//init block pool to save block
		executor := executor.NewExecutor(DefaultNamespace, conf, eventMux)
		if executor == nil {
			return errors.New("Initialize BlockPool failed")
		}

		executor.CreateInitBlock(conf)
		executor.Initialize()
		//init manager
		exist := make(chan bool)

		pm := manager.New(DefaultNamespace, eventMux, executor, grpcPeerMgr, consenter, am, cm)
		go jsonrpc.Start(eventMux, pm, cm, conf)
		go CheckLicense(exist, conf)

		<-exist
		return nil
	})
}
