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
	"hyperchain-alpha/jsonrpc/model"
	"hyperchain-alpha/jsonrpc/routers"
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
		model.InitDB(argv.LocalPort)

		//存储本地节点
		p2p.LOCALNODE = model.Node{P2PAddr:localIp,P2PPort:argv.LocalPort,HTTPPORT:argv.HttpServerPORT}

		//测试方法
		if argv.Test{
			P2P_test(p2p.LOCALNODE)
		}else{

			//将本机地址加入Nodes列表中
			p2p.GLOBALNODES = append(p2p.GLOBALNODES,p2p.LOCALNODE)

			//如果传入了对端节点地址，则首先向远程节点同步
			if argv.PeerIp != "" && argv.PeerPort !=0{
				remotesNodes := p2p.GetNodes(argv.PeerIp+":"+strconv.Itoa(argv.PeerPort),p2p.LOCALNODE)
				//检查节点是否已经存在
				for _,remoteNode := range remotesNodes{
					existFlag := false
					for  _,localNode := range p2p.GLOBALNODES{
						if localNode.P2PAddr == remoteNode.P2PAddr && localNode.P2PPort == remoteNode.P2PPort{
							existFlag = true
						}
					}
					if !existFlag{
						p2p.GLOBALNODES = append(p2p.GLOBALNODES,remoteNode)
					}
				}

				//TODO GetNodes方法只在刚刚加入时调用，需要向所有的取得节点发送自己的节点消息
				//向所有新取得的列表中的节点广播自己的信息
				for _,remoteNode := range p2p.GLOBALNODES{
					//如果是本地节点
					if remoteNode.P2PAddr == p2p.LOCALNODE.P2PAddr && remoteNode.P2PPort == p2p.LOCALNODE.P2PPort{
						continue
					}
					//如果是指定对端节点
					if remoteNode.P2PAddr == argv.PeerIp && remoteNode.P2PPort == argv.PeerPort{
						continue
					}
					//向其它节点告知
					p2p.SaveNode(remoteNode.P2PAddr+":"+strconv.Itoa(remoteNode.P2PPort),p2p.LOCALNODE)
				}

				fmt.Printf("同步对端数据节点成功，%v\n",p2p.GLOBALNODES)
				//TODO 同步交易信息
				serverNode := model.Node{P2PAddr:argv.PeerIp,P2PPort:argv.PeerPort,HTTPPORT:0}
				p2p.GetTrans(serverNode)

			}



			//启用p2p服务
			p2p.StratP2PServer(argv.LocalPort)

			//实例化路由
			router := routers.NewRouter()
			// 指定静态文件目录
			http.Handle("/static/", http.StripPrefix("/static/",http.FileServer(http.Dir("./static"))))
			//启动http服务
			ctx.String("启动http服务...\n")
			log.Fatal(http.ListenAndServe(":"+strconv.Itoa(argv.HttpServerPORT),router))
		}
		return nil
	})
}

//rpc测试方法，测试已经成功，可以取得远端数据并获取相应输出
func P2P_test(localNode model.Node){
	rnodes := p2p.SaveNode("localhost:8001",localNode)
	fmt.Printf("%v\n",rnodes)
}