package jsonrpc

import (
	"hyperchain/core/types"
	"fmt"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/hyperdb"
	"log"
	"hyperchain/core"
	"math/big"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value string `json:"value"`
}

type Transaction struct {
	From      []byte
	To        []byte
	Value     big.Int
	TimeStamp int64
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is not enough, return false
func SendTransaction(args TxArgs) bool {

	var tx *types.Transaction

	fmt.Println(args)

	// 构造 transaction 实例
	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))

	// 判断交易余额是否足够
	if (tx.VerifyBalance()) {
		// 余额足够
		// 抛 NewTxEvent 事件
		 eventmux:=new(event.TypeMux)
		 eventmux.Post(event.NewTxEvent{Payload: proto.Marshal(tx)})

		return true

	} else {
		// 余额不足
		return false
	}
}

func GetAllTransactions()  []types.Transaction{

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}

	txs, err := core.GetAllTransaction(db)

	if err != nil {
		log.Fatalf("GetAllTransaction error: %v", err)
	}

	var val big.Int
	var transactions []Transaction

	// 将交易金额转换为整型
	for index, tx := range txs {

		val.SetString(string(tx.Value),10)

		transactions[index].Value = val
		transactions[index].From = tx.From
		transactions[index].To = tx.To
		transactions[index].TimeStamp = tx.TimeStamp
	}

	return transactions
}


func GetAllBalances() core.BalanceMap{

	balanceIns, err := core.GetBalanceIns()

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}

	balMap := balanceIns.GetAllDBBalance()

	return balMap
}


