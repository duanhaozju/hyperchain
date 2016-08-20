package hyper

import (
	"time"
	"hyperchain-alpha/core"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/encrypt"
	"strconv"
	"crypto/dsa"
	"hyperchain-alpha/p2p"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value int `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type TransactionPoolAPI struct{

}


var pri2pub map[dsa.PublicKey]dsa.PrivateKey

// 生成一张私钥与公钥的映射表
func init() {
	//TODO 读取公私钥对

}

// name2key是 name 与 privateKey 的 key-value
func sign(publicKey dsa.PublicKey,dataHash []byte) (encrypt.Signature,[]byte){

	privateKey := pri2pub[publicKey]

	return encrypt.Sign(privateKey,dataHash)
}

// 参数是一个json对象，
// TODO 发送一个交易
func (t *TransactionPoolAPI) SendTransaction(args TxArgs) error {

	var tx types.Transaction

	//init()

	// 构造 transaction 实例
	tx = types.NewTransaction(args.From,args.To,args.Value,args.Timestamp)

	//TODO 生成一个签名
	signature,_ := sign(args.From,[]byte(args.From + args.To + strconv.Itoa(args.Value) + args.Timestamp))

	//TODO 签名交易
	tx.Singnature = signature // 已经签名的交易

	//TODO 验证交易，发送者是否存在，余额是否足够


	// TODO  数据 存储到交易池
	core.PutTransactionToLDB(transaction.Hash(),transaction)
	//TODO 判断交易池是否满
	//TODO if ture
	// TODO 1.打包区块 2.清空交易池 MXM ====

	//TODO 3. 远程同步数据
	//TODO 判断是否有新的区块，并封装数据信封

	allNodes,_:=core.GetAllNodeFromMEM()
	for _,remoteNode := range allNodes{
		if remoteNode != p2p.LOCALNODE{
			p2p.TransSync(new p2p.Envelope{})
		}
	}

	return nil
}


// TODO 获取某个用户地址的所有交易
func (t *TransactionPoolAPI) GetTransactions(addr []byte) []types.Transaction {

	return nil
}

// TODO 获取某个指定区块
func (t *TransactionPoolAPI) GetBlcok(number int) types.Block {

	return nil
}


