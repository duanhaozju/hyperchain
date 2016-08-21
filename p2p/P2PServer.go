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
	NowAllNodes,_ := core.GetAllNodeFromMEM()
	fmt.Println("当前拥有节点：",NowAllNodes)
	*retNodes,_= core.GetAllNodeFromMEM()
	return nil
}

//Transaction部分

//RemoteSaveTransaction 远程存储交易
func (r *RemoteNode) RemoteSaveTransaction(envelope *Envelope,trans *types.Transactions) error {
	remoteNode := envelope.Nodes[0]
	for _,tx := range envelope.Transactions{
		fmt.Printf("获取远端新交易数据,请求来源：%s,\t交易数据为%v\t\n",remoteNode,tx)
		err := core.PutTransactionToLDB(tx.Hash(),tx)
		if err != nil{
			return err
		}
	}
	return nil
}

//让客户端远程调用的方法，用于客户端同步现有交易信息
func (r *RemoteNode) RemoteGetTransactions(envelope *Envelope,trans *types.Transactions) error{
	remoteNode := envelope.Nodes[0]
	//远程取得交易信息
	fmt.Printf("远端请求同步,请求来源：%s\n",remoteNode)
	var err = new(error)
	*trans,*err = core.GetAllTransactionFromLDB()
	return *err
}

//客户端请求block信息，需要携带自己的节点信息
func (r *RemoteNode) RemoteGetBlocks(envelop *Envelope,retEnvelop *Envelope) error{
	// 将blocks信息返回给请求节点，如果block数据太多的话，数据的传输由底层rpc实现控制
	//请求的节点
	peerNode := envelop.Nodes[0]
	log.Println("请求来自对端：",peerNode)
	blocks,err := core.GetAllBlockFromLDB()
	(*retEnvelop).Blocks = blocks
	return err

}
//客户端请求chain信息，需要携带自己的节点信息
func (r *RemoteNode) RemoteGetChain(envelop *Envelope,retEnvelop *Envelope) error{
	//请求的节点
	peerNode := envelop.Nodes[0]
	log.Println("请求来自对端：",peerNode)
	chain := core.GetChain()
	(*retEnvelop).Chain = *chain
	return nil
}

func (r *RemoteNode) RemoteGetPoolTransactions(envelop *Envelope,retEnvelop *Envelope) error{
	peerNode := envelop.Nodes[0]
	log.Println("请求来自对端：",peerNode)
	var transactions types.Transactions
	transactions = core.GetTransactionsFromTxPool()
	(*retEnvelop).Transactions = transactions
	return nil
}


//接收广播数据并进行处理的方法
func (r *RemoteNode)RemoteDataTransfer(envelope *Envelope,retEnvelope *Envelope) error{
	blocks := envelope.Blocks
	trans    := envelope.Transactions
	//首先处理transaction
	for _,tran := range trans{
		//先验证
		if tran.VerifyTransaction(core.GetBalanceFromMEM(tran.From),core.GetTransactionsFromTxPool()){
			//先存储 tran
			core.AddTransactionToTxPool(tran)
			// 判断txPool是否满
			if core.TxPoolIsFull(){
				//如果满
				//1. 取出trans 池子中所有数据\
				transaction := core.GetTransactionsFromTxPool()
				//2. 打包成Block
				// 打包成一个区块
				newBlock := types.NewBlock(transaction,core.GetLashestBlockHash(),LOCALNODE)
				//3. 清空txPool
				core.ClearTxPool()
				//4. 存储 Block
				core.PutBlockToLDB(newBlock.BlockHash,*newBlock)
				//5.  更新整个chain
				core.UpdateChain(newBlock.BlockHash)
				//6. TODO 更新balance表
				//review core.UpdateBalance(newBlock)
				//review 是否需要对外广播新打包的区块
				//review 需要防止广播风暴
				//var envelope Envelope
				//envelope.Blocks = append(envelope.Blocks,newBlock)
				// review 服务端调用客户端方法，这是会导致循环的，正确处理方式是，将block数据返回给客户端，让客户端再发起一次广播
				//BroadCast(envelope)

			}else{
				//如果非满，直接判断结束
			}
		}else{
			log.Fatalln("交易验证失败，签名无效",tran)
		}
	}

	// 然后处理block
	for _,block := range blocks {
		//需要验证block是否存在
		if _,ok := core.GetBlockFromLDB(block.BlockHash); ok != nil{
			//block不存在
			core.PutBlockToLDB(block.BlockHash,block)
			// review 更新整个chain
			core.UpdateChain(block.BlockHash)
			// TODO 更新balance表
			//review core.UpdateBalance(block)
		}else{
			//block存在
		}
	}

	//处理返回值
	retEnvelope.Chain = *core.GetChain()
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
