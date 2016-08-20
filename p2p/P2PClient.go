package p2p

import (
	"net/rpc"
	"log"
	"fmt"
	"strconv"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core"
)

// 全局变量
var LOCALNODE node.Node


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
func SaveNode(serverAddress string) node.Nodes{
	//同步调用

	var client =establishConn(serverAddress)
	defer client.Close()

	//存储返回值
	var nodes node.Nodes
	// 实例化信息信封，放入自己的节点信息
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	err := client.Call("RemoteNode.RemoteSaveNodes", &messageEnvelope, &nodes)
	if err != nil {
		log.Fatal("Remote error:", err)
		// TODO 如果出现错误，就删除该节点
		//
	}
	fmt.Println( nodes)
	return nodes
}
//从远端取得节点信息
func GetNodes(serverAddress string) node.Nodes {
	var client = establishConn(serverAddress)
	defer client.Close()
	//存储返回值
	var nodes node.Nodes
	//实例化信息传输信封
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	err := client.Call("RemoteNode.RemoteGetNodes", &messageEnvelope, &nodes)
	if err != nil {
		log.Fatal("Remote error:", err)
		// TODO 如果出现错误，则无法连接目标节点
	}
	fmt.Println( nodes)

	return nodes
}
//向远端保存交易信息
func SaveTrans(serverNode node.Node,localNode node.Node,tx types.Transaction) types.Transactions{
	//建立连接
	var client =establishConn(serverNode.P2PIP+":"+strconv.Itoa(serverNode.P2PPort))
	defer client.Close()
	//用于保存返回信息
	var trans types.Transactions
	//实例化信息传输信封
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,localNode)
	messageEnvelope.Transactions = append(messageEnvelope.Transactions,tx)

	err := client.Call("RemoteNode.RemoteSaveTransaction", &messageEnvelope, &trans)
	if err != nil {
		log.Fatal("Remote Error:", err)
	}
	fmt.Printf("远端返回交易信息为: %v \n",trans)
	return trans
}
//与远端同步交易信息
func GetTrans(serverNode node.Node) types.Transactions{
	var client =establishConn(serverNode.P2PIP+":"+strconv.Itoa(serverNode.P2PPort))
	defer client.Close()
	// 存储交易的容器
	var trans types.Transactions
	//信息传输信封
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	err := client.Call("RemoteNode.RemoteGetTransaction", &messageEnvelope, &trans)
	fmt.Printf("\n从节点%v,\t同步交易数据:\n",serverNode)

	//获取交易之后自动存入数据库
	for _,tx := range trans{
		core.PutTransactionToLDB(tx.Hash(),tx)
		if tx.Verify() {
			fmt.Printf("\n从\t%v\t同步得到节点%v\n",serverNode,tx)
		}else{
			fmt.Printf("\n存储失败！%v\n",tx)
		}
	}
	if err != nil {
		log.Fatal("Remote Error:", err)
	}
	//fmt.Printf("RemoteNodes: %v\n ",trans)
	return trans
}

func NodeSync(peerNode node.Node) ([]node.Node,error){
	serverAddress :=string(peerNode.P2PIP +":"+ strconv.Itoa(peerNode.P2PPort))
	//取得所有远程节点
	remotesNodes := GetNodes(serverAddress)
	//取得所有本地节点
	AllNodes,_ := core.GetAllNodeFromMEM()
	//检查节点是否已经存在
	for _,remoteNode := range remotesNodes{
		existFlag := false
		for  _,localNode := range AllNodes{
			if LOCALNODE.P2PIP == remoteNode.P2PIP && localNode.P2PPort == remoteNode.P2PPort{
				existFlag = true
			}
		}
		if !existFlag{
			core.PutNodeToMEM(remoteNode.CoinBase,remoteNode)
		}
	}

	//TODO GetNodes方法只在刚刚加入时调用，需要向所有的取得节点发送自己的节点消息
	//向所有新取得的列表中的节点广播自己的信息
	for _,remoteNode := range AllNodes{
		//如果是本地节点
		if remoteNode.P2PIP == LOCALNODE.P2PIP && remoteNode.P2PPort == LOCALNODE.P2PPort{
			continue
		}
		//如果是指定对端节点
		if remoteNode.P2PIP == peerNode.P2PIP && remoteNode.P2PPort == peerNode.P2PPort{
			continue
		}
		//向其它节点告知
		SaveNode(remoteNode.P2PIP+":"+strconv.Itoa(remoteNode.P2PPort))
	}

	fmt.Printf("同步对端数据节点成功，%v\n",AllNodes)
	return AllNodes,nil
}


func TransSync(peerNode node.Node){
	GetTrans(peerNode)
}

// 异步调用
//quotient := new(Quotient)
//divCall := client.Go("Arith.Divide", args, quotient, nil)
//replyCall := <-divCall.Done	// will be equal to divCall
// check errors, print, etc.