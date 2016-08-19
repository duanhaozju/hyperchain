package p2p

import (
	"net/rpc"
	"log"
	"fmt"
	"strconv"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core/types"
)

//调用远程节点的提供的方法
func establishConn(serverAddress string) *rpc.Client{
	client, err := rpc.DialHTTP("tcp", serverAddress)
	if err != nil {
		log.Fatal("dialing:", err)
	}
	fmt.Println("connection has been established")
	//defer client.Close()
	return client
}

//向远程监听节点写入自己的节点数据，并获取对方的节点列表

func SaveNode(serverAddress string,localNode node.Node) node.Nodes{
	//同步调用

	var client =establishConn(serverAddress)
	defer client.Close()
	//发送的参数
	//var node1 =  model.Node{P2PAddr:"localhost1", P2PPort:8001, HTTPPORT:80001}
	//存储返回值
	var nodes node.Nodes

	err := client.Call("RemoteNode.RemoteSaveNodes", &localNode, &nodes)
	if err != nil {
		log.Fatal("Remote error:", err)
		// TODO 如果出现错误，就删除该节点
		//
	}
	fmt.Println( nodes)
	return nodes
}
//从远端取得节点信息
func GetNodes(serverAddress string,localNode node.Node) node.Nodes {
	var client = establishConn(serverAddress)
	defer client.Close()
	//存储返回值
	var nodes model.Nodes

	err := client.Call("RemoteNode.RemoteGetNodes", &localNode, &nodes)
	if err != nil {
		log.Fatal("Remote error:", err)
		// TODO 如果出现错误，则无法连接目标节点
	}
	fmt.Println( nodes)

	return nodes
}
//向远端保存交易信息
func SaveTrans(serverNode node.Node,localNode node.Node,tx types.Transaction) types.Transactions{
	//同步调用
	var trans types.Transactions
	var client =establishConn(serverNode.P2PIP+":"+strconv.Itoa(serverNode.P2PPort))
	var txTransfer  TxTransfer
	txTransfer.Node = localNode
	txTransfer.Tx = tx
	err := client.Call("RemoteNode.RemoteSaveTransaction", &txTransfer, &trans)
	if err != nil {
		log.Fatal("Remote Error:", err)
	}
	fmt.Printf("远端返回交易信息为: %v \n",trans)
	return trans
}
//与远端同步交易信息
func GetTrans(serverNode node.Node) types.Transactions{
	var trans types.Transactions
	var client =establishConn(serverNode.P2PIP+":"+strconv.Itoa(serverNode.P2PPort))
	err := client.Call("RemoteNode.RemoteGetTransaction", &LOCALNODE, &trans)
	fmt.Printf("\n从节点%v,\t同步交易数据:\n",serverNode)

	//获取交易之后自动存入数据库
	for _,tx := range trans{
		savedTx := model.SaveTransction(tx)

		if savedTx.Hash() != ""{
			fmt.Printf("\n从\t%v\t同步得到节点%v\n",serverNode,savedTx)
		}else{
			fmt.Printf("\n存储失败！%v\n",savedTx)
		}
	}
	if err != nil {
		log.Fatal("Remote Error:", err)
	}
	//fmt.Printf("RemoteNodes: %v\n ",trans)
	return trans
}

func NodeSync(serverAddress string) []node.Node
// 异步调用
//quotient := new(Quotient)
//divCall := client.Go("Arith.Divide", args, quotient, nil)
//replyCall := <-divCall.Done	// will be equal to divCall
// check errors, print, etc.