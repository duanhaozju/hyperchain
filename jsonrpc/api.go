package jsonrpc

import (
	"hyperchain/core/types"
	"hyperchain/encrypt"
	"hyperchain/utils"
	"hyperchain/core"
	"hyperchain/p2p"
	"fmt"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value int `json:"value"`
	Timestamp int64 `json:"timestamp"`
}

type ResData struct{
	Data interface{}
	Code int
}

const MAXCOUNT = 5

var name2key = make(map[string]utils.KeyPair)

// 生成一张私钥与公钥的映射表
//func initial() {
//	accounts,_ := utils.GetAccount()
//
//	for _,ac := range accounts{
//		for key,value := range ac{
//			name2key[key] = value
//		}
//	}
//}

// 参数是一个json对象
func SendTransaction(args TxArgs) ResData {

	var tx *types.Transaction

	fmt.Println(args)
	//initial()

	//pubFrom := name2key[args.From].PubKey
	//pubTo := name2key[args.To].PubKey

	// 将公钥转换为string类型
	//fromPubKey := encrypt.EncodePublicKey(&pubFrom)
	//toPubKey := encrypt.EncodePublicKey(&pubTo)

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
		fmt.Println("=======================================valid")
		// 验证通过

		var envelopes *p2p.Envelope

		// 提交到交易池
		core.AddTransactionToTxPool(*tx)
		core.PutTransactionToLDB(txHash,*tx)

		transactions := make(types.Transactions,1)
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

			// 更新全局最新一个区块的HASH
			core.UpdateChain(block.BlockHash)

			// 则清空交易池
			core.ClearTxPool()

			// 将交易和区块信息传入信封
			blocks := make([]types.Block,1)
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

		// 远程同步信封数据
		p2p.BroadCast(envelopes)

		return ResData{
			Data: nil,
			Code:1,
		}

	}

	return ResData{
		Data: nil,
		Code: 0,
	}
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


