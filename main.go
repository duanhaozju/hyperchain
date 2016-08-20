//main 入口函数，用于处理命令行入口
package main
import (
	"log"
	"net/http"
	"github.com/mkideal/cli"
	"fmt"
	"strconv"
	"hyperchain-alpha/p2p"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/jsonrpc/routers"
	"hyperchain-alpha/core"
	"hyperchain-alpha/core/node"
)

type argT struct {
	cli.Helper
	PeerIp string `cli:"i,peerip" usage:"对端节点地址" dft:""`
	PeerPort int  `cli:"p,peerport" usage:"对端节点监听端口" dft:"0"`
	LocalPort int `cli:"o,hostport" usage:"本地RPC监听端口" dft:"8001"`
	HttpServerPORT int `cli:"s,httpport" usage:"启动本地http服务的端口，默认值为8003" dft:"8003"`
	Test bool `cli:"t" usage:"用于测试" dft:"false"`
}

func main(){
	cli.Run(new(argT), func(ctx *cli.Context) error {
		argv := ctx.Argv().(*argT)
		ctx.String("P2P_ip=%s, \tP2P_port=%d,\t本地http端口:%d,\t本地P2P监听端口:%d,\tTest Flag:%v\n", argv.PeerIp, argv.PeerPort,argv.HttpServerPORT,argv.LocalPort,argv.Test)
		//正常启动相应服务
		localIp := utils.GetIpAddr()
		fmt.Printf("本机ip地址为："+localIp+ "\n")

		//初始化数据库,传入数据库地址自动生成数据库文件
		core.InitDB(argv.LocalPort)

		//存储本地节点
		p2p.LOCALNODE = node.NewNode(argv.PeerIp,argv.PeerPort,argv.HttpServerPORT)

		//将本机地址加入Nodes列表中
		core.PutNodeToMEM(p2p.LOCALNODE.CoinBase,p2p.LOCALNODE)

		//如果传入了对端节点地址，则首先向远程节点同步
			if argv.PeerIp != "" && argv.PeerPort !=0{
				peerNode := node.NewNode(argv.PeerIp,argv.PeerPort,0)
				p2p.NodeSync(peerNode)
				p2p.TransSync(peerNode)
			}

			//启用p2p服务
			p2p.StratP2PServer(argv.LocalPort)

			//实例化路由
			router := routers.NewRouter()
			// 指定静态文件目录
			router.PathPrefix("/").Handler(http.FileServer(http.Dir(".")))

			//启动http服务
			ctx.String("启动http服务...\n")
			log.Fatal(http.ListenAndServe(":"+strconv.Itoa(argv.HttpServerPORT),router))

		return nil
	})
}
