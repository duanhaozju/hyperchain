// init ProtocolManager
// author: Lizhong kuang
// date: 2016-08-23
// last modified:2016-08-26
package main
import (
	"log"
	"github.com/mkideal/cli"
	"hyperchain/p2p"
	"hyperchain/manager"
	"hyperchain/jsonrpc"
	"hyperchain/core"
	"hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/consensus/controller"

)

type argT struct {
	cli.Helper
	NodePath string
	IsFirst bool
	ConsensusNum int
	from common.Hash
	to common.Hash
	PeerIp string `cli:"i,peerip" usage:"对端节点地址" dft:""`
	PeerPort int  `cli:"p,peerport" usage:"对端节点监听端口" dft:"0"`
	LocalPort int `cli:"o,hostport" usage:"本地RPC监听端口" dft:"8001"`
	HttpServerPORT int `cli:"s,httpport" usage:"启动本地http服务的端口，默认值为8003" dft:"8003"`
	Test bool `cli:"t" usage:"用于测试" dft:"false"`
}

func main(){
	cli.Run(new(argT), func(ctx *cli.Context) error {



		/*init peer manager object,consensus object*/

		argv := ctx.Argv().(*argT)

		//init peer manager to start grpc server and client
		grpcPeerMgr:=new(p2p.GrpcPeerManager)

		//init fetcher to accept block
		fetcher := core.NewFetcher()


		//init pbft consensus
		cs:=controller.NewConsenter(argv.ConsensusNum)

		// init http server for web call
		jsonrpc.StartHttp(argv.LocalPort)


		//init encryption object
		encryption :=crypto.NewEcdsaEncrypto("ecdsa")


		//init hash object
		kec256Hash:=crypto.NewKeccak256Hash("keccak256")


		//init manager
		manager.New(grpcPeerMgr,cs,fetcher,encryption,kec256Hash,
			argv.NodePath,argv.IsFirst)



		/*eventmux:=new(event.TypeMux)
		eventmux.Post(event.ConsensusEvent{[]byte{0x00, 0x00, 0x03, 0xe8}})
*/

		//

		/*jobs := make(chan int, 100)
		<-jobs*/
		/*argv := ctx.Argv().(*argT)
		ctx.String("P2P_ip=%s, \tP2P_port=%d,\t本地http端口:%d,\t本地P2P监听端口:%d,\tTest Flag:%v\n", argv.PeerIp, argv.PeerPort,argv.HttpServerPORT,argv.LocalPort,argv.Test)
		//正常启动相应服务
		localIp := utils.GetIpAddr()
		log.Printf("本机ip地址为："+localIp+ "\n")*/

		//初始化数据库,传入数据库地址自动生成数据库文件
		core.InitDB(argv.LocalPort)

		//存储本地节点
		//p2p.LOCALNODE = node.NewNode(localIp,argv.LocalPort,argv.HttpServerPORT)
		// 初始化keystore
		//utils.GenKeypair("#"+localIp+"&"+strconv.Itoa(argv.LocalPort))

		//将本机地址加入Nodes列表中
		//core.PutNodeToMEM(p2p.LOCALNODE.CoinBase,p2p.LOCALNODE)
		//allNodes,_ := core.GetAllNodeFromMEM()
		//log.Println("本机初始化拥有节点",allNodes)

		//如果传入了对端节点地址，则首先向远程节点同步
		/*if argv.PeerIp != "" && argv.PeerPort !=0{
				peerNode := node.NewNode(argv.PeerIp,argv.PeerPort,0)
				// 同步节点
				p2p.NodeSync(&peerNode)
				//从远端节点同步区块
				p2p.BlockSync(&peerNode)
				// TODO 同步最新区块的Hash 即同步chain
				p2p.ChainSync(&peerNode)
				// 同步交易池
				p2p.TxPoolSync(&peerNode)
				//p2p.TransSync(peerNode)
		}else{
			//未传入地址，则自己需要初始化创世区块
			// TODO 构造创世区块
			CreateInitBlock()
		}*/

		/*//启用p2p服务
		p2p.StratP2PServer(argv.LocalPort)

		//实例化路由
		router := routers.NewRouter()
		// 指定静态文件目录
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("./jsonrpc/")))
*/
		//启动http服务

		log.Println("启动http服务...")
		/*log.Fatal(http.ListenAndServe(":"+strconv.Itoa(argv.HttpServerPORT),router))
*/
		return nil
	})
}

//-- 创建初始块



