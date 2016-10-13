// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-23
// last modified:2016-08-29
package main

import (
	"github.com/mkideal/cli"
	"hyperchain/manager"
	"hyperchain/p2p"

	"hyperchain/core"

	"hyperchain/crypto"

	"hyperchain/event"

	"strconv"

	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/consensus/controller"
	"hyperchain/jsonrpc"
	"hyperchain/membersrvc"
	"os"
	"path"
)

type argT struct {
	cli.Helper
	//NodePath string `cli:"o,hostport" usage:"本地RPC监听端口" dft:"8001"`
	NodeID         int    `cli:"o,id" usage:"node ID 默认 1" dft:"1"`
	LocalPort      int    `cli:"l,port" usage:"本地RPC监听端口" dft:"8001"`
	JsonConfigPath string `cli:"j,json" usage:"节点信息的端口，默认值为./peerconfig.json" dft:"./peerconfig.json"`
	YamlConfigPath string `cli:"y,yaml" usage:"pbft配置文件, 默认值为./" dft:"./"`
	//GenesisPath    string `cli:"g,genesisconfig" usage:"genesis配置文件，用于创建创世块, 默认值是./genesis.json" dft:"./core/genesis.json"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		// read config
		_nodeID := argv.NodeID
		_localGRPCPort := argv.LocalPort
		_jsonConfigPath := argv.JsonConfigPath
		_yamlConfigPath := argv.YamlConfigPath
		_workingPath, _ := os.Getwd()
		_logOutputDir := path.Join(_workingPath, "./logs/")
		_keyStorePath := path.Join(_workingPath, "./keystore/")

		membersrvc.Start(argv.YamlConfigPath, argv.NodeID)

		//init log
		common.InitLog(logging.INFO, "./logs/", argv.LocalPort)
		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(argv.JsonConfigPath, argv.NodeID)

		//init fetcher to accept block
		fetcher := core.NewFetcher()

		//init db
		core.InitDB(argv.LocalPort)
		//core.TxSum = core.CalTransactionSum()

		core.InitEnv()
		//init genesis
		core.CreateInitBlock(argv.JsonConfigPath)

		//init pbft consensus
		cs := controller.NewConsenter(uint64(argv.NodeID), eventMux, argv.YamlConfigPath)

		//init encryption object
		keydir := "./keystore/"

		encryption := crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GenerateNodeKey(strconv.Itoa(argv.LocalPort), keydir)

		am := accounts.NewAccountManager(keydir, encryption)
		am.UnlockAllAccount(keydir)

		//init hash object
		kec256Hash := crypto.NewKeccak256Hash("keccak256")
		//nodePath := "./p2p/peerconfig.json"
		nodePath := argv.JsonConfigPath

		//init block pool to save block
		blockPool := core.NewBlockPool(eventMux)

		//start http server
		//go jsonrpc.StartHttp(argv.LocalPort, eventMux)

		//go jsonrpc.Start(argv.LocalPort, eventMux)

		//init manager

		exist := make(chan bool)
		pm := manager.New(eventMux, blockPool, grpcPeerMgr, cs, fetcher, am, kec256Hash,
			nodePath, argv.NodeID)
		go jsonrpc.Start(argv.LocalPort, eventMux, pm)

		<-exist
		////init manager
		//manager.New(eventMux,blockPool,grpcPeerMgr,cs,fetcher,encryption,kec256Hash,
		//	nodePath,argv.NodeId)

		return nil
	})
}
