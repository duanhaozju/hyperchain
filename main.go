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
	"hyperchain/consensus/csmgr"
	"hyperchain/core/executor"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/manager"
	"hyperchain/p2p"
	"strconv"
	"hyperchain/core/db_utils"
)

type argT struct {
	cli.Helper
	//NodeID     int    `cli:"o,id" usage:"node ID" dft:"1"`
	ConfigPath string `cli:"c,conf" usage:"config file path" dft:"./config/global.yaml"`
	//GRPCPort   int    `cli:"l,rpcport" usage:"inner grpc connect port" dft:"8001"`
	//HTTPPort   int    `cli:"t,httpport" useage:"jsonrpc open port" dft:"8081"`
	//RESTPort   int    `cli:"f,restport" useage:"restful api port" dft:"9000"`
}

const (
	DefaultNamespace = "Global"
)

func initConf(argv *argT) *common.Config {
	conf := common.NewConfig(argv.ConfigPath)
	//conf.Set(common.HYPERCHAIN_ID, argv.NodeID)
	//conf.Set(common.HTTP_PORT, argv.HTTPPort)
	//conf.Set(common.REST_PORT, argv.RESTPort)
	//conf.Set(common.GRPC_PORT, argv.GRPCPort)
	// read the global peers path
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

		//TODO:remove this config later
		config := newconfigsImpl(conf)
		//LOG
		common.InitLog(conf)
		//DB

		conf.MergeConfig(config.getDbConfig())//todo:refactor it

		db_utils.InitDBForNamespace(conf, DefaultNamespace, config.getDbConfig(),conf.GetInt(common.C_NODE_ID))

		eventMux := new(event.TypeMux)

		/**
		 *传入true则开启所有验证，false则为取消ca以及签名的所有验证
		 */
		globalConfig := viper.New()
		globalConfig.SetConfigFile(conf.GetString(common.C_GLOBAL_CONFIG_PATH))
		err := globalConfig.ReadInConfig()
		if err != nil {
			panic(err)
		}
		cm, cmerr := admittance.GetCaManager(globalConfig)
		if cmerr != nil {
			panic("cannot initliazied the camanager")
		}

		//init peer manager to start grpc server and client
		//grpcPeerMgr := p2p.NewGrpcManager(config.getPeerConfigPath())
		grpcPeerMgr := p2p.NewGrpcManager(conf)

		//init genesis
		//init pbft consensus
		consenter := csmgr.Consenter(DefaultNamespace, conf, eventMux)
		consenter.Start()

		//init encryption object
		encryption := crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GenerateNodeKey(strconv.Itoa(config.getNodeID()), config.getKeyNodeDir())
		am := accounts.NewAccountManager(config.getKeystoreDir(), encryption)
		am.UnlockAllAccount(config.getKeystoreDir())

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
		go jsonrpc.Start(config.getHTTPPort(), config.getRESTPort(), config.getLogDumpFileDir(), eventMux, pm, cm, conf)
		go CheckLicense(exist, conf)
		<-exist
		return nil
	})
}
