package p2p

import (
	"net/rpc"
	"net"
	"log"
	"net/http"
	"fmt"
	"hyperchain.cn/app/jsonrpc/model"
	"strconv"
)

//服务器需要对外提供两个方法，RemoteGetNodes 和RemoteGetTransaction

type RemoteNode struct {
	RNodes model.Nodes
	RTransactions model.Transactions
}

// RemoteSaveNodes 远程存储节点
// 这里的返回值 retNode 在每次执行的时候都会赋值为零
func (r *RemoteNode) RemoteSaveNodes(inNode *model.Node, retNodes *model.Nodes) error {
	existFlag := false
	for  _,localNode := range GLOBALNODES{
		if localNode.P2PAddr == inNode.P2PAddr && localNode.P2PPort == inNode.P2PPort{
			fmt.Println("节点已经存在")
			existFlag = true
		}
	}
	if !existFlag{
		fmt.Println("节点不存在")
		fmt.Println(*inNode)
		//向内存中存储的节点列表添加节点
		GLOBALNODES = append(GLOBALNODES,*inNode)
	}
	*retNodes = GLOBALNODES
	fmt.Println(*inNode)
	fmt.Println(*retNodes)

	return nil
}

//从远端取得相应的节点数据
func (r *RemoteNode) RemoteGetNodes(inNode *model.Node,retNodes *model.Nodes) error{
		existFlag := false
		for  _,localNode := range GLOBALNODES{
			if localNode.P2PAddr == inNode.P2PAddr && localNode.P2PPort == inNode.P2PPort{
				fmt.Println("节点已经存在")
				existFlag = true
			}
		}
		if !existFlag{
			fmt.Println("节点不存在")
			fmt.Println(*inNode)
			GLOBALNODES = append(GLOBALNODES,*inNode)
		}

	fmt.Println(GLOBALNODES)
	*retNodes = GLOBALNODES
	return nil
}

//Transaction部分

//RemoteSaveTransaction 远程存储交易
func (r *RemoteNode) RemoteSaveTransaction(txtrans*TxTransfer,trans *model.Transactions) error {
	fmt.Printf("获取远端新交易数据,请求来源：%v,\t交易数据为%v\t\n",txtrans.Node,txtrans.Tx)
	retTx := model.SaveTransction(txtrans.Tx)
	*trans = append(*trans,retTx)
	//TODO 远程存储交易
	return nil
}

func (r *RemoteNode) RemoteGetTransaction(inNode *model.Node,trans *model.Transactions) error{
	//TODO 远程取得交易信息
	fmt.Printf("远端请求同步,请求来源：%v\n",inNode)
	*trans,_ = model.GetAllTransaction()
return nil
}

func StratP2PServer(p2pServerPort int){

	remoteNode := new(RemoteNode)
	rpc.Register(remoteNode)
	rpc.HandleHTTP()
	l, e := net.Listen("tcp", ":"+strconv.Itoa(p2pServerPort))
	if e != nil {
		log.Fatal("listen error:", e)
	}
	fmt.Println("启动P2P远程调用服务...")
	go http.Serve(l, nil)
}
