package hyperchain

import (
	"time"
	//"hyperchain-alpha/core"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/encrypt"
	"strconv"
	"crypto/dsa"
	//"hyperchain-alpha/p2p"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/core"
	"hyperchain-alpha/p2p"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value int `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

//type TransactionPoolAPI struct{
//
//}

type key struct {
	privateKey dsa.PrivateKey
	publicKey dsa.PublicKey
}

const MAXCOUNT = 5

var name2key map[string]key

// 生成一张私钥与公钥的映射表
func initial() {
	//TODO 读取公私钥对
	accounts,_ := utils.GetAccount()
	for ac,_ := range accounts{
		for key,value := range ac{
			name2key[key] = value
		}
	}
}

// 参数是一个json对象
// TODO 发送一个交易
func SendTransaction(args TxArgs) error {

	var tx types.Transaction


	initial()

	// 将公钥转换为string类型
	fromPubKey := encrypt.EncodePublicKey(name2key[args["from"]])
	toPubKey := encrypt.EncodePublicKey(name2key[args["to"]])

	// 构造 transaction 实例
	tx = types.NewTransaction(fromPubKey,toPubKey,strconv.Atoi(args["value"]),time.Now())

	// 生成一个签名
	signature,_ := encrypt.Sign(name2key[args["from"]].privateKey,tx.Hash())

	// 已经签名的交易
	tx.Singnature = signature


	// 验证交易
	if (tx.VerifyTransaction()) {

		// 验证通过
		var envelopes p2p.Envelope
		// 提交到交易池
		core.AddTransactionToTxPool(tx)

		transactions := make(types.Transactions,1)
		transactions = append(transactions,tx)

		// 判断交易池是否已满
		if(core.GetTxPoolCapacity() == MAXCOUNT){

			// 若已满，打包区块
			trans := core.GetTransactionsFromTxPool()
			block := types.NewBlock(trans,p2p.LOCALNODE,time.Now())


			// 则清空交易池
			core.ClearTxPool()

			// 将交易和区块信息传入信封
			blocks := make([]types.Block,1)
			blocks = append(blocks,block)

			envelopes = p2p.Envelope{
				Transactions: transactions,
				Blocks: blocks,
			}

		} else {
			// 将交易信息传入信封
			envelopes = p2p.Envelope{
				Transactions: transactions,
			}

		}

		// 远程同步信封数据
		p2p.BroadCast(envelopes)

	}
	return nil
}


// TODO 获取某个用户地址的所有交易
func GetTransactions(addr []byte) []types.Transaction {

	return nil
}

// TODO 获取某个指定区块
func GetBlcok(number int) types.Block {

	return nil
}


