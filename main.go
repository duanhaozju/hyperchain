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
	"hyperchain/jsonrpc"
	"hyperchain/common"
	"github.com/op/go-logging"
	"hyperchain/accounts"
)

type argT struct {
	cli.Helper
	//NodePath string `cli:"o,hostport" usage:"本地RPC监听端口" dft:"8001"`
	NodeId    int `cli:"o,nodeId" usage:"本地RPC监听端口" dft:"8001"`

	LocalPort int `cli:"l,LocalPort" usage:"本地RPC监听端口" dft:"8001"`
	//HttpServerPORT int `cli:"s,httpport" usage:"启动本地http服务的端口，默认值为8003" dft:"8003"`
}

func main() {
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)

		//init log
		common.InitLog(logging.INFO, "./logs/", argv.LocalPort)
		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr := new(p2p.GrpcPeerManager)

		//init fetcher to accept block
		fetcher := core.NewFetcher()


		//init db
		core.InitDB(argv.LocalPort)

		//init genesis
		core.CreateInitBlock("./core/genesis.json")

		//init pbft consensus
		cs := controller.NewConsenter(uint64(argv.NodeId), eventMux)

		//init encryption object
		keydir := "./keystore/"

		encryption := crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GenerateNodeKey(strconv.Itoa(argv.LocalPort),keydir)

		am := accounts.NewAccountManager(keydir,encryption)
		//am.NewAccount("123")


		//init hash object
		kec256Hash := crypto.NewKeccak256Hash("keccak256")
		nodePath := "./p2p/peerconfig.json"

		//init block pool to save block
		blockPool := core.NewBlockPool(eventMux)

		//start http server
		go jsonrpc.StartHttp(argv.LocalPort, eventMux)

		//init manager

		manager.New(eventMux,blockPool,grpcPeerMgr,cs,fetcher,am,kec256Hash,
			nodePath,argv.NodeId)

		////init manager
		//manager.New(eventMux,blockPool,grpcPeerMgr,cs,fetcher,encryption,kec256Hash,
		//	nodePath,argv.NodeId)



		return nil
	})
}




