//main 入口函数，用于处理命令行入口
package main
import (
	"log"
	"net/http"
	"github.com/mkideal/cli"
	"strconv"
	"hyperchain-alpha/p2p"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/jsonrpc/routers"
	"hyperchain-alpha/core"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/encrypt"
	"time"
	"encoding/json"
	"fmt"
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
		log.Printf("本机ip地址为："+localIp+ "\n")

		//初始化数据库,传入数据库地址自动生成数据库文件
		core.InitDB(argv.LocalPort)

		//存储本地节点
		p2p.LOCALNODE = node.NewNode(localIp,argv.LocalPort,argv.HttpServerPORT)
		// 初始化keystore
		//utils.GenKeypair("#"+localIp+"&"+strconv.Itoa(argv.LocalPort))

		//将本机地址加入Nodes列表中
		core.PutNodeToMEM(p2p.LOCALNODE.CoinBase,p2p.LOCALNODE)
		allNodes,_ := core.GetAllNodeFromMEM()
		log.Println("本机初始化拥有节点",allNodes)

		//如果传入了对端节点地址，则首先向远程节点同步
		if argv.PeerIp != "" && argv.PeerPort !=0{
				peerNode := node.NewNode(argv.PeerIp,argv.PeerPort,0)
				// 同步节点
				p2p.NodeSync(&peerNode)
				//从远端节点同步区块
				p2p.BlockSync(&peerNode)
				// TODO 同步最新区块的Hash 即同步chain
				//p2p.ChainSync(&peerNode)
				// 同步交易池
				p2p.TxPoolSync(&peerNode)
				//p2p.TransSync(peerNode)
		//}else{
			//未传入地址，则自己需要初始化创世区块
			// TODO 构造创世区块
			CreateInitBlock()
		}

		//启用p2p服务
		p2p.StratP2PServer(argv.LocalPort)

		//实例化路由
		router := routers.NewRouter()
		// 指定静态文件目录
		router.PathPrefix("/").Handler(http.FileServer(http.Dir("./jsonrpc/")))

		//启动http服务
		ctx.String("启动http服务...\n")
		log.Fatal(http.ListenAndServe(":"+strconv.Itoa(argv.HttpServerPORT),router))

		return nil
	})
}

//-- 创建初始块
func CreateInitBlock()  {
	//-- 获取创世快from用户
	godAccount := utils.GetGodAccount()[0]
	//-- 获取所有的提前生成好的account
	accounts, err := utils.GetAccount()
	if err != nil {
		log.Fatal(err)
		return
	}
	//-- 为account分别构造transaction 五个)
	//-- from 为上面的from用户
	//-- value 为10000, 20000, 30000 ...
	var transactions types.Transactions
	for i, _ := range accounts {
		fmt.Printf("%#v\n",godAccount["god"].PubKey)
		t := types.Transaction {
			//From: encrypt.EncodePublicKey(&(godAccount["god"].PubKey)),
			From: "1222",
			//To: encrypt.EncodePublicKey(&account[0].PubKey),
			To: "1212",
			Value: (i+1) * 10000,
			TimeStamp: time.Now().Unix(),
		}
		by := []byte(t.From + t.To + strconv.Itoa(t.Value) + strconv.FormatInt(t.TimeStamp, 10))
		signature, _ := encrypt.Sign(godAccount["god"].PriKey, by)
		t.Signature = signature
		transactions = append(transactions, t)
	}

	//-- 打包创世块
	block := types.Block{
		ParentHash: string(encrypt.GetHash([]byte("0"))),
		Transactions: transactions,
		TimeStramp: time.Now().Unix(),
		CoinBase: p2p.LOCALNODE,
		MerkleRoot: "root",
	}
	txBStr, _ := json.Marshal(block.Transactions)
	coinbaseBStr , _ := json.Marshal(block.CoinBase)
	block.BlockHash = string(encrypt.GetHash([]byte(block.ParentHash + string(txBStr) + strconv.FormatInt(block.TimeStramp, 10) + string(coinbaseBStr) + block.MerkleRoot)))
	//-- 将创世块存入数据库
	core.PutBlockToLDB(block.BlockHash, block)
	//-- 将初始block的BlockHash存如Chain
	core.UpdateChain(block.BlockHash)

	//-- 初始初始化balance
	core.UpdateBalance(block)
}