package p2p

import (
	"net/rpc"
	"net"
	"log"
	"net/http"
	"fmt"
	"strconv"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core"
)

//服务器需要对外提供两个方法，RemoteGetNodes 和RemoteGetTransaction

type RemoteNode struct {
	RNodes node.Nodes
	RTransactions types.Transactions
}

// RemoteSaveNodes 远程存储节点
// 这里的返回值 retNode 在每次执行的时候都会赋值为零
func (r *RemoteNode) RemoteSaveNodes(envelope *Envelope, retNodes *node.Nodes) error {
	existFlag := false
	AllNodes,_ := core.GetAllNodeFromMEM()
	for _,inNode:= range envelope.Nodes{
		for  _,localNode := range AllNodes{
			if localNode.P2PIP== inNode.P2PIP && localNode.P2PPort == inNode.P2PPort{
				fmt.Println("节点已经存在")
				existFlag = true
			}
		}
		if !existFlag{
			fmt.Println("节点不存在")
			fmt.Println(inNode)
			//向内存中存储的节点列表添加节点
			core.PutNodeToMEM(inNode.CoinBase,inNode)
		}
	}
	var err = new(error)
	*retNodes,*err = core.GetAllNodeFromMEM()
	fmt.Println(*retNodes)
	return *err
}

//从远端取得相应的节点数据
func (r *RemoteNode) RemoteGetNodes(envelope *Envelope,retNodes *node.Nodes) error{
		// 发起请求的节点
		inNode := envelope.Nodes[0]
		existFlag := false
		AllNodes,_ := core.GetAllNodeFromMEM()
		for  _,localNode := range AllNodes{
			if localNode.P2PIP == inNode.P2PIP && localNode.P2PPort == inNode.P2PPort{
				fmt.Println("节点已经存在")
				existFlag = true
			}
		}
		if !existFlag{
			fmt.Println("节点不存在")
			fmt.Println(inNode)
			core.PutNodeToMEM(inNode.CoinBase,inNode)
		}

	fmt.Println(core.GetAllNodeFromMEM())
	*retNodes,_= core.GetAllNodeFromMEM()
	return nil
}

//Transaction部分

//RemoteSaveTransaction 远程存储交易
func (r *RemoteNode) RemoteSaveTransaction(envelope *Envelope,trans *types.Transactions) error {
	remoteNode := envelope.Nodes[0]
	for _,tx := range envelope.Transactions{
		fmt.Printf("获取远端新交易数据,请求来源：%v,\t交易数据为%v\t\n",remoteNode,tx)
		err := core.PutTransactionToLDB(tx.Hash(),tx)
		if err != nil{
			return err
		}
	}
	return nil
}

func (r *RemoteNode) RemoteGetTransaction(envelope *Envelope,trans *types.Transactions) error{
	remoteNode := envelope.Nodes[0]
	//TODO 远程取得交易信息
	fmt.Printf("远端请求同步,请求来源：%v\n",remoteNode)
	var err = new(error)
	*trans,*err = core.GetAllTransactionFromLDB()
	return *err
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
