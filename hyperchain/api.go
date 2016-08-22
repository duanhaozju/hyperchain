package hyperchain

import (
	"time"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/encrypt"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/core"
	"hyperchain-alpha/p2p"
	"fmt"
	"errors"
	"log"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value int `json:"value"`
	Timestamp int64 `json:"timestamp"`
}

//type TransactionPoolAPI struct{
//
//}

const MAXCOUNT = 5

var name2key = make(map[string]utils.KeyPair)

// 生成一张私钥与公钥的映射表
func initial() {
	accounts,_ := utils.GetAccount()

	for _,ac := range accounts{
		for key,value := range ac{
			name2key[key] = value
		}
	}
}

// 参数是一个json对象
func SendTransaction(args TxArgs) error {

	var tx *types.Transaction

	fmt.Println(args)
	initial()

	pubFrom := name2key[args.From].PubKey
	pubTo := name2key[args.To].PubKey

	// 将公钥转换为string类型
	fromPubKey := encrypt.EncodePublicKey(&pubFrom)
	toPubKey := encrypt.EncodePublicKey(&pubTo)

	// 构造 transaction 实例
	tx = types.NewTransaction(fromPubKey,toPubKey,args.Value)

	// 生成一个签名
	txHash := tx.Hash()
	signature,_ := encrypt.Sign(name2key[args.From].PriKey,[]byte(txHash))

	// 已经签名的交易
	tx.Signature = signature


	// 验证用户余额，交易是否合法
	balance := core.GetBalanceFromMEM(tx.From)
	txPoolsTrans := core.GetTransactionsFromTxPool()

	if (tx.VerifyTransaction(balance,txPoolsTrans)) {
		fmt.Println(time.Now().Format("2006/01/02 15:04:05") + " 用户提交交易验证有效...")
		// 验证通过

		var envelopes *p2p.Envelope

		// 提交到交易池
		core.AddTransactionToTxPool(*tx)
		//存到交易池的不存到数据库
		//core.PutTransactionToLDB(txHash,*tx)

		var transactions types.Transactions

		transactions = append(transactions,*tx)

		// 判断交易池是否已满
		if(core.GetTxPoolCapacity() == MAXCOUNT){

			// 若已满，生成一个新的区块
			trans := core.GetTransactionsFromTxPool()

			block := types.NewBlock(trans,core.GetLashestBlockHash(),p2p.LOCALNODE)

			// 更新balance表
			core.UpdateBalance(*block)

			// （没有验证区块）区块存进数据库
			core.PutBlockToLDB(block.BlockHash,*block)
			// review 遍历最新区块并存储
			for _,trans := range block.Transactions{
				core.PutTransactionToLDB(trans.Hash(),trans)
			}

			// 更新全局最新一个区块的HASH
			core.UpdateChain(block.BlockHash)

			// 则清空交易池
			core.ClearTxPool()

			// 将交易和区块信息传入信封
			var blocks []types.Block
			blocks = append(blocks,*block)

			envelopes = &p2p.Envelope{
				Transactions: transactions,
				Blocks: blocks,
			}

		} else {
			// 将交易信息传入信封
			envelopes = &p2p.Envelope{
				Transactions: transactions,
			}

		}

		//log.Println("网页封装rans lengs",len(envelopes.Transactions))
		//输出封装的信息
		for _,tran := range envelopes.Transactions{
			log.Println("即将广播的交易信息：",tran)
		}
		for _,blk := range envelopes.Blocks{
			log.Println("即将广播的区块信息：",blk)
		}
		// 远程同步信封数据
		p2p.BroadCast(envelopes)

		return nil

	}

	return errors.New("余额不足")
}

func GetAllTransactions() (types.Transactions,error) {

	var txs types.Transactions

	txs,err := core.GetAllTransactionFromLDB()

	if (err != nil) {
		return nil,err
	}

	return txs,nil

}

func GetAllAccountBalances() ([]types.Balance,error) {
	var bals []types.Balance
	bals = core.GetAllBalanceFromMEM()
	return bals,nil
}



//
//// TODO 获取某个用户地址的所有交易
//func GetTransactions(addr []byte) []types.Transaction {
//
//	return nil
//}
//
//// TODO 获取某个指定区块
//func GetBlcok(number int) types.Block {
//
//	return nil
//}


