// init ProtcolManager
// author: Lizhong kuang
// date: 2016-08-23
// last modified:2016-08-29
package main

import (
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus/controller"
	"hyperchain/core"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/jsonrpc"
	"hyperchain/manager"
	"hyperchain/membersrvc"
	"hyperchain/p2p"
	"strconv"

	"github.com/mkideal/cli"
)

type argT struct {
	cli.Helper
	NodeID     int    `cli:"o,id" usage:"node ID" dft:"1"`
	ConfigPath string `cli:"c,conf" usage:"配置文件所在路径" dft:"./config/global.yaml"`
	GRPCPort   int    `cli:"l,rpcport" usage:"远程连接端口" dft:"8001"`
	HTTPPort   int    `cli:"t,httpport" useage:"jsonrpc开放端口" dft:"8081"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		config := newconfigsImpl(argv.ConfigPath)
		// add condition judge if need the command line param override the config file config
		config.initConfig(argv.NodeID, argv.GRPCPort, argv.HTTPPort)

		membersrvc.Start(config.getMemberSRVCConfigPath(), config.getNodeID())

		//init log
		common.InitLog(config.getLogLevel(), config.getLogDumpFileDir(), config.getGRPCPort(), config.getLogDumpFileFlag())

		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(config.getPeerConfigPath(), config.getNodeID())

		//init fetcher to accept block
		fetcher := core.NewFetcher()

		//init db
		core.InitDB(config.getDatabaseDir(), config.getGRPCPort())

		core.InitEnv()
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
		blockPool := core.NewBlockPool(eventMux,cs)

		//init manager
		exist := make(chan bool)
		pm := manager.New(eventMux, blockPool, grpcPeerMgr, cs, fetcher, am, kec256Hash,
			config.getNodeID())

		go jsonrpc.Start(config.getHTTPPort(), eventMux, pm)

		<-exist

		return nil
	})
}
