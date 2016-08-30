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
	"hyperchain/jsonrpc"
)

type argT struct {
	cli.Helper
	//NodePath string `cli:"o,hostport" usage:"本地RPC监听端口" dft:"8001"`
	NodeId int `cli:"o,nodeId" usage:"本地RPC监听端口" dft:"8001"`

	LocalPort int `cli:"l,LocalPort" usage:"本地RPC监听端口" dft:"8001"`
	//HttpServerPORT int `cli:"s,httpport" usage:"启动本地http服务的端口，默认值为8003" dft:"8003"`

}



func main(){
	cli.Run(new(argT), func(ctx *cli.Context) error {





		/*init peer manager object,consensus object*/

		argv := ctx.Argv().(*argT)

		eventMux := new(event.TypeMux)

		//init peer manager to start grpc server and client
		grpcPeerMgr:=new(p2p.GrpcPeerManager)

		//init fetcher to accept block
		fetcher := core.NewFetcher()




		//init pbft consensus
		//cs:=controller.NewConsenter(argv.ConsensusNum)

		// init http server for web call
		go jsonrpc.StartHttp(argv.LocalPort,eventMux)

		core.InitDB(123)
		core.CreateInitBlock("./core/genesis.json")



		//init encryption object


		encryption :=crypto.NewEcdsaEncrypto("ecdsa")
		encryption.GeneralKey(string(argv.LocalPort))



		//init hash object
		kec256Hash:=crypto.NewKeccak256Hash("keccak256")

		 nodePath:="./p2p/peerconfig.json"
		//init db
		core.InitDB(argv.LocalPort)


		//init manager

		manager.New(eventMux,grpcPeerMgr,nil,fetcher,encryption,kec256Hash,
			nodePath,argv.NodeId)



		return nil
	})
}




