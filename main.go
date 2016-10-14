// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-23
// last modified:2016-08-29
package main

import (
	"github.com/mkideal/cli"
	"hyperchain/p2p"
	"hyperchain/manager"

	"hyperchain/core"

	"hyperchain/crypto"

	"hyperchain/event"

	"strconv"

	"hyperchain/consensus/controller"
	"hyperchain/common"
	"github.com/op/go-logging"
	"hyperchain/accounts"
	"hyperchain/jsonrpc"
	"hyperchain/membersrvc"
	"runtime"
)

type argT struct {
	cli.Helper
	//NodePath string `cli:"o,hostport" usage:"本地RPC监听端口" dft:"8001"`
	NodeId         int `cli:"o,nodeId" usage:"本地RPC监听端口" dft:"8001"`

	LocalPort      int `cli:"l,LocalPort" usage:"本地RPC监听端口" dft:"8001"`
	PeerConfigPath string `cli:"p,peerconfig" usage:"节点信息的端口，默认值为./peerconfig.json" dft:"./p2p/peerconfig.json"`
	PbftConfigPath string `cli:"f,pbftconfig" usage:"pbft配置文件, 默认值为./" dft:"./consensus/pbft/"`
	GenesisPath    string `cli:"g,genesisconfig" usage:"genesis配置文件，用于创建创世块, 默认值是./genesis.json" dft:"./core/genesis.json"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		runtime.GOMAXPROCS(-1)
		membersrvc.Start("./", argv.NodeId)

		//init log
		common.InitLog(logging.INFO, "./logs/", argv.LocalPort)
		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr := p2p.NewGrpcManager(argv.PeerConfigPath, argv.NodeId)

		//init fetcher to accept block
		fetcher := core.NewFetcher()


		//init db
		core.InitDB(argv.LocalPort)
		//core.TxSum = core.CalTransactionSum()

		core.InitEnv()
		//init genesis
		core.CreateInitBlock(argv.GenesisPath)

		//init pbft consensus
		cs := controller.NewConsenter(uint64(argv.NodeId), eventMux, argv.PbftConfigPath)

		//init encryption object
		keydir := "./keystore/"

		encryption := crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GenerateNodeKey(strconv.Itoa(argv.LocalPort), keydir)

		am := accounts.NewAccountManager(keydir, encryption)
		am.UnlockAllAccount(keydir)


		//init hash object
		kec256Hash := crypto.NewKeccak256Hash("keccak256")
		//nodePath := "./p2p/peerconfig.json"
		nodePath := argv.PeerConfigPath

		//init block pool to save block
		blockPool := core.NewBlockPool(eventMux,cs)

		//start http server
		//go jsonrpc.StartHttp(argv.LocalPort, eventMux)

		//go jsonrpc.Start(argv.LocalPort, eventMux)


		//init manager

		exist := make(chan bool)
		pm := manager.New(eventMux, blockPool, grpcPeerMgr, cs, fetcher, am, kec256Hash,
			nodePath, argv.NodeId)
		go jsonrpc.Start(argv.LocalPort, eventMux, pm)

		<-exist
		////init manager
		//manager.New(eventMux,blockPool,grpcPeerMgr,cs,fetcher,encryption,kec256Hash,
		//	nodePath,argv.NodeId)


		return nil
	})
}




