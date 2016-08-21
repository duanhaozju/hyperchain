package p2p

import (
	"net/rpc"
	"log"
	"fmt"
	"strconv"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core"
	"encoding/hex"
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
func SaveNode(remoteNode node.Node) node.Nodes{
	//同步调用
	serverAddress := remoteNode.P2PIP+":"+strconv.Itoa(remoteNode.P2PPort)
	log.Println("向远程节点同步自己的节点信息...")
	var client =establishConn(serverAddress)
	defer client.Close()

	//存储返回值
	var nodes node.Nodes
	// 实例化信息信封，放入自己的节点信息
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	err := client.Call("RemoteNode.RemoteSaveNodes", &messageEnvelope, &nodes)
	if err != nil {
		//review 如果出现错误，就删除该节点
		log.Fatal("Remote error:", err)
	}
	fmt.Println( nodes)
	return nodes
}
//从远端取得节点信息
func getNodes(serverAddress string) node.Nodes {
	var client = establishConn(serverAddress)
	defer client.Close()
	//存储返回值
	var nodes node.Nodes
	//实例化信息传输信封
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	err := client.Call("RemoteNode.RemoteGetNodes", &messageEnvelope, &nodes)
	if err != nil {
		//review 如果出现错误，则无法连接目标节点
		log.Fatal("Remote error:", err)

	}
	fmt.Println( nodes)
//GetBasePath
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
func getTrans(serverNode node.Node) types.Transactions{
	var client =establishConn(serverNode.P2PIP+":"+strconv.Itoa(serverNode.P2PPort))
	defer client.Close()
	// 存储交易的容器
	var trans types.Transactions
	//信息传输信封
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	err := client.Call("RemoteNode.RemoteGetTransactions", &messageEnvelope, &trans)
	fmt.Printf("\n从节点%s,同步交易数据:\n",serverNode)

	//获取交易之后自动存入数据库
	for _,tx := range trans{
		core.PutTransactionToLDB(tx.Hash(),tx)
		if tx.VerifyTransaction(core.GetBalanceFromMEM(tx.From),core.GetTransactionsFromTxPool()) {
			fmt.Printf("\n从%s同步得到交易数据%v\n",serverNode,tx)
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

//从对端节点同步取得相应信息
func NodeSync(peerNode *node.Node) ([]node.Node,error){
	log.Println("同步节点...")
	serverAddress :=string(peerNode.P2PIP +":"+ strconv.Itoa(peerNode.P2PPort))
	//取得所有远程节点
	remotesNodes := getNodes(serverAddress)
	fmt.Println("对端返回节点数据：",remotesNodes)
	//将对端节点数据进行更新,对端存储的第一个节点都是对端节点的完整信息
	*peerNode = remotesNodes[0]
	fmt.Println("交换之后的对端节点信息",peerNode)
	//取得所有本地节点
	AllNodes,_ := core.GetAllNodeFromMEM()
	//检查节点是否已经存在
	for _,remoteNode := range remotesNodes{
		existFlag := false
		for  _,localNode := range AllNodes{
			if localNode.P2PIP == remoteNode.P2PIP && localNode.P2PPort == remoteNode.P2PPort{
				existFlag = true
			}
		}
		if !existFlag{
			core.PutNodeToMEM(remoteNode.CoinBase,remoteNode)
		}
	}
	// 重新取得节点信息
	AllNodes,_ = core.GetAllNodeFromMEM()
	//review GetNodes方法只在刚刚加入时调用，需要向所有的取得节点发送自己的节点消息
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
		SaveNode(remoteNode)
	}

	fmt.Printf("同步对端数据节点成功，%s\n",AllNodes)
	return AllNodes,nil
}

//从对端节点同步取得相应信息
func TransSync(peerNode node.Node){
	getTrans(peerNode)
}

//向全网节点广播信息
func BroadCast(envelope *Envelope)(int,error){
	log.Println("===============广播数据=============")
	e := *envelope
	log.Println("需要广播的数据：",e)
	allNodes,_:=core.GetAllNodeFromMEM()
	for _,remoteNode := range allNodes{
		if remoteNode.P2PIP == LOCALNODE.P2PIP && remoteNode.P2PPort == LOCALNODE.P2PPort && remoteNode.HttpPort == LOCALNODE.HttpPort{
			log.Println("节点忽略..",remoteNode,LOCALNODE)
		}else{
			returnEnvelope,err := dataTransfer(envelope,remoteNode)
			if err != nil{
				panic(err)
			}else{
				log.Println("节点返回数据：",returnEnvelope)
			}
		}
	}
	return 0,nil
}

func dataTransfer(envelop *Envelope, peerNode node.Node)(Envelope,error){
	serverAddress :=string(peerNode.P2PIP +":"+ strconv.Itoa(peerNode.P2PPort))
	var client =establishConn(serverAddress)
	defer client.Close()
	//用于存储返回信息
	var returnMessage Envelope
	log.Println("客户端发起trans len",len(envelop.Transactions))
	err := client.Call("RemoteNode.RemoteDataTransfer", &envelop, &returnMessage)
	return returnMessage,err
}

// block同步方法，从服务端节点将服务端的block同步下来
func BlockSync(peerNode *node.Node) ([]types.Block,error){
	log.Println("同步区块...")
	//review 区块同步，由于没有顺序，所以只是将区块信息从对端节点同步回来
	// 连接对端
	serverAddress :=string(peerNode.P2PIP +":"+ strconv.Itoa(peerNode.P2PPort))
	var client =establishConn(serverAddress)
	defer client.Close()
	//用于发送信息
	var envelop Envelope
	envelop.Nodes = append(envelop.Nodes,LOCALNODE)
	//用于存储返回信息
	var returnEnvelop Envelope
	err := client.Call("RemoteNode.RemoteGetBlocks", &envelop, &returnEnvelop)
	//review 完成balance表的更新
	for _,block := range returnEnvelop.Blocks{
		core.UpdateBalance(block)
		//将block中的交易存储起来
		for _,tran :=range block.Transactions{
			log.Println("存储交易："+hex.EncodeToString([]byte(tran.Hash()))+"\n")
			core.PutTransactionToLDB(tran.Hash(),tran)
		}
	}

	if len(returnEnvelop.Blocks) >0{
		log.Println("同步区块个数：",len(returnEnvelop.Blocks))
	}
	return returnEnvelop.Blocks,err
}

// block Header 同步方法，用于将Chain数据同步回来，用于实现数据回溯
func ChainSync(peerNode *node.Node)(types.Chain,error){
	log.Println("同步chain信息...")
	//review 将Chain信息同步回来
	// 连接对端
	serverAddress :=string(peerNode.P2PIP +":"+ strconv.Itoa(peerNode.P2PPort))
	var client =establishConn(serverAddress)
	defer client.Close()
	//用于发送信息
	var envelop Envelope
	envelop.Nodes = append(envelop.Nodes,LOCALNODE)
	//用于存储返回信息
	var returnEnvelop Envelope
	err := client.Call("RemoteNode.RemoteGetChain", &envelop, &returnEnvelop)

	log.Println("chain信息：",returnEnvelop.Chain)
	//更新本地chain信息
	core.UpdateChain(returnEnvelop.Chain.LastestBlockHash)
	return returnEnvelop.Chain,err
}

func TxPoolSync(peerNode *node.Node)(types.Transactions,error){
	log.Println("同步交易池信息...")
	// review 将对端交易池中的数据同步回来
	serverAddress :=string(peerNode.P2PIP +":"+ strconv.Itoa(peerNode.P2PPort))
	var client =establishConn(serverAddress)
	defer client.Close()
	// 存储交易的容器
	var trans types.Transactions
	//信息传输信封
	var messageEnvelope = new (Envelope)
	messageEnvelope.Nodes = append(messageEnvelope.Nodes,LOCALNODE)

	var retEnvelope Envelope

	err := client.Call("RemoteNode.RemoteGetPoolTransactions", &messageEnvelope, &retEnvelope)

	//从信封中取出数据
	trans = retEnvelope.Transactions
	//获取交易之后自动存入数据库
	for _,tx := range trans{
		//验证
		if tx.VerifyTransaction(core.GetBalanceFromMEM(tx.From),core.GetTransactionsFromTxPool()) {
			//先放入池子中
			core.AddTransactionToTxPool(tx)
			//

			//判断是否满
			if (core.TxPoolIsFull()){
				//panic(error.Error("同步区块错误，远端交易池已满，在单线程条件下不允许发生"))
				log.Fatalln("同步区块错误，远端交易池已满，在单线程条件下不允许发生")
				//transactions := core.GetTransactionsFromTxPool()
				// review 是否需要打包区块？

				// review 是否需要广播区块？

			}else{
				log.Println("同步得到交易数据并存储到交易池",tx)
			}
		}else{
			log.Fatalf("\n交易验证失败，无法存储！%v\n",tx)
		}
	}
	if err != nil {
		log.Fatal("服务器错误:", err)
	}
	log.Println("交易池信息同步结束，共有：",len(trans),"条信息")
	return trans,err
}


// 当异步调用时候可以采用以下方法，可以大幅提高效率
//quotient := new(Quotient)
//divCall := client.Go("Arith.Divide", args, quotient, nil)
//replyCall := <-divCall.Done	// will be equal to divCall
// check errors, print, etc.