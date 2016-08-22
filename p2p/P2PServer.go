package p2p

import (
	"net/rpc"
	"net"
	"log"
	"net/http"
	"strconv"
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core"
	"encoding/hex"
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
				log.Println("节点已经存在",inNode)
				existFlag = true
			}
		}
		if !existFlag{
			log.Println("节点不存在",inNode)
			//向内存中存储的节点列表添加节点
			core.PutNodeToMEM(inNode.CoinBase,inNode)
		}
	}
	var err = new(error)
	*retNodes,*err = core.GetAllNodeFromMEM()
	log.Println(*retNodes)
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
				log.Println("节点已经存在",inNode)
				existFlag = true
			}
		}
		if !existFlag{
			log.Println("节点不存在:",inNode)
			log.Println(inNode)
			core.PutNodeToMEM(inNode.CoinBase,inNode)
		}
	NowAllNodes,_ := core.GetAllNodeFromMEM()
	log.Println("当前拥有节点：",NowAllNodes)
	*retNodes,_= core.GetAllNodeFromMEM()
	return nil
}

//Transaction部分

//RemoteSaveTransaction 远程存储交易
func (r *RemoteNode) RemoteSaveTransaction(envelope *Envelope,trans *types.Transactions) error {
	remoteNode := envelope.Nodes[0]
	for _,tx := range envelope.Transactions{
		log.Printf("获取远端新交易数据,请求来源：%s,\t交易数据为%v\t\n",remoteNode,tx)
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
	log.Printf("远端请求同步,请求来源：%s\n",remoteNode)
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
	//log.Println("接收到的trans len ",len(trans))
	for _,tran := range trans{
		log.Println("<<<<<<<<<<<<<<<<<<有新交易>>>>>>>>>>>>>>>>>>")
		//先验证
		if tran.VerifyTransaction(core.GetBalanceFromMEM(tran.From),core.GetTransactionsFromTxPool()){

			log.Println("交易验证成功，签名有效,处理交易...")
			//先存储 tran
			core.AddTransactionToTxPool(tran)
			// 判断txPool是否满
			if core.TxPoolIsFull(){
				log.Println("交易池已满,需要存储交易:")
				//如果满
				//1. 取出trans 池子中所有数据
				transaction := core.GetTransactionsFromTxPool()
				//2. 打包成Block
				// 打包成一个区块
				newBlock := types.NewBlock(transaction,core.GetLashestBlockHash(),LOCALNODE)
				for _,tran :=range newBlock.Transactions{
					log.Println("存储交易："+hex.EncodeToString([]byte(tran.Hash())))
					core.PutTransactionToLDB(tran.Hash(),tran)
				}
				//3. 清空txPool
				core.ClearTxPool()
				//4. 存储 Block
				core.PutBlockToLDB(newBlock.BlockHash,*newBlock)
				//5.  更新整个chain
				core.UpdateChain(newBlock.BlockHash)
				//6. REVIEW 更新balance表 无条件接受对方打包出来的区块,自己的区块抛弃不用
				// core.UpdateBalance(*newBlock)
				//log.Println("当前本地打包区块(抛弃)hash:"+hex.EncodeToString([]byte(core.GetChain().LastestBlockHash)))
				log.Println("当前本地打包出来的区块(抛弃)\n",newBlock)
				//review 是否需要对外广播新打包的区块
				//review 需要防止广播风暴
				//var envelope Envelope
				//envelope.Blocks = append(envelope.Blocks,newBlock)
				// review 服务端调用客户端方法，这是会导致循环的，正确处理方式是，将block数据返回给客户端，让客户端再发起一次广播
				//BroadCast(envelope)

			}else{

				//如果非满，直接判断结束
				log.Println("交易池未满，大小为：",core.GetTxPoolCapacity())
				log.Println("将交易数据存储到交易池:")
				log.Println(tran)
			}
		}else{
			log.Fatalln("交易验证失败，签名无效",tran)
		}
		log.Println("<<<<<<<<<<<<<<<<<<交易处理结束>>>>>>>>>>>>>>>")
	}

	// 然后处理block
	for _,block := range blocks {
		//需要验证block是否存在
		if _,ok := core.GetBlockFromLDB(block.BlockHash); ok != nil{
			//block不存在
			core.PutBlockToLDB(block.BlockHash,block)
			// REVIEW 更新整个chain 只接受对方广播过来的区块，不采用本身打包出来的区块，这里采用一致性算法解决
			core.UpdateChain(block.BlockHash)
			//log.Println("当前最新区块hash:"+hex.EncodeToString([]byte(core.GetChain().LastestBlockHash)))
			log.Println("别的节点广播过来的区块(无条件接受)\n",block)
			// REVIEW 更新balance表 无条件接受对方发送的block信息，并更新balance表
			core.UpdateBalance(block)
		}else{
			//block存在
		}
	}

	//处理返回值
	retEnvelope.Blocks,_ = core.GetAllBlockFromLDB()
	retEnvelope.Nodes,_  = core.GetAllNodeFromMEM()
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
	log.Println("开启P2P远程调用服务...")
	go http.Serve(l, nil)
}
