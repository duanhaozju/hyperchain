//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package main

import (
	"errors"
	"github.com/mkideal/cli"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus/controller"
	"hyperchain/core"
	"hyperchain/core/blockpool"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/jsonrpc"
	"hyperchain/manager"
	"hyperchain/membersrvc"
	"hyperchain/p2p"
	"strconv"
)

type argT struct {
	cli.Helper
	NodeID      int    `cli:"o,id" usage:"node ID" dft:"1"`
	ConfigPath  string `cli:"c,conf" usage:"配置文件所在路径" dft:"./config/global.yaml"`
	GRPCPort    int    `cli:"l,rpcport" usage:"远程连接端口" dft:"8001"`
	HTTPPort    int    `cli:"t,httpport" useage:"jsonrpc开放端口" dft:"8081"`
	IsReconnect bool  `cli:"r,isReconnect" usage:"是否重新链接" dft:"false"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		config := newconfigsImpl(argv.ConfigPath, argv.NodeID, argv.GRPCPort, argv.HTTPPort)

		membersrvc.Start(config.getMemberSRVCConfigPath(), config.getNodeID())

		//init log
		common.InitLog(config.getLogLevel(), config.getLogDumpFileDir(), config.getGRPCPort(), config.getLogDumpFileFlag())

		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(config.getPeerConfigPath(), config.getNodeID())

		//init db
		core.InitDB(config.getDatabaseDir(), config.getGRPCPort())

		//init genesis
		core.CreateInitBlock(config.getGenesisConfigPath())

		//init pbft consensus
		cs := controller.NewConsenter(uint64(config.getNodeID()), eventMux, config.getPBFTConfigPath())

		//init encryption object

		encryption := crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GenerateNodeKey(strconv.Itoa(config.getNodeID()), config.getKeyNodeDir())
		//
		am := accounts.NewAccountManager(config.getKeystoreDir(), encryption)
		am.UnlockAllAccount(config.getKeystoreDir())

		//init hash object
		kec256Hash := crypto.NewKeccak256Hash("keccak256")

		//init block pool to save block
		blockPoolConf := blockpool.BlockPoolConf{
			BlockVersion: config.getBlockVersion(),
			TransactionVersion: config.getTransactionVersion(),
		}
		blockPool := blockpool.NewBlockPool(eventMux, cs, blockPoolConf)
		if blockPool == nil {
			return errors.New("Initialize BlockPool failed")
		}

		//init manager
		exist := make(chan bool)
		syncReplicaInterval, _ := config.getSyncReplicaInterval()
		syncReplicaEnable := config.getSyncReplicaEnable()
		pm := manager.New(eventMux, blockPool, grpcPeerMgr, cs, am, kec256Hash,
			config.getNodeID(), syncReplicaInterval, syncReplicaEnable, exist, config.getLicense())

		go jsonrpc.Start(config.getHTTPPort(), eventMux, pm, config.getRateLimitConfig())
		<-exist
		return nil
	})
}
