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
	"runtime"
)

type argT struct {
	cli.Helper
	NodeId int `cli:"o,nodeid" usage:"current node ID" dft:"1"`
	LocalPort      int    `cli:"l,localport" usage:"gRPC data transport port" dft:"8001"`
	PeerConfigPath string `cli:"p,peerconfig" usage:"peerconfig.json path" dft:"./peerconfig.json"`
	PbftConfigPath string `cli:"f,pbftconfig" usage:"pbft config file folder path,ensure this is a valid dir path" dft:"./consensus/pbft/"`
	GenesisPath    string `cli:"g,genesisconfig" usage:"genesis.json config file path " dft:"./core/genesis.json"`
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
		blockPool := core.NewBlockPool(eventMux, cs)

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
