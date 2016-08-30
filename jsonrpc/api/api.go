package hyperchain

import (
	"hyperchain/core/types"
	"fmt"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/hyperdb"
	"log"
	"hyperchain/core"
	"hyperchain/manager"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value string `json:"value"`
}

type TransactionShow struct {
	From      string
	To        string
	Value     string
	TimeStamp int64
}

type BalanceShow map[string]string

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is not enough, return false
func SendTransaction(args TxArgs) bool {

	var tx *types.Transaction

	fmt.Println(args)

	// 构造 transaction 实例
	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))

	// 判断交易余额是否足够
	if (core.VerifyBalance(tx)) {
		// 余额足够
		// 抛 NewTxEvent 事件

		txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}


		err = manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})

		if err != nil {
			log.Fatalf("Post event.NewTxEvent{Payload: txBytes} error: %v",err)
		}

		return true

	} else {
		// 余额不足
		return false
	}
}

// GetAllTransactions return all transactions in the chain/db
func GetAllTransactions()  []TransactionShow{

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}

	txs, err := core.GetAllTransaction(db)

	if err != nil {
		log.Fatalf("GetAllTransaction error: %v", err)
	}

	var transactions []TransactionShow

	// 将交易金额转换为整型
	for index, tx := range txs {

		transactions[index].Value = string(tx.Value)
		transactions[index].From = string(tx.From)
		transactions[index].To = string(tx.To)
		transactions[index].TimeStamp = tx.TimeStamp
	}

	return transactions
}

// GetAllBalances retrun all account's balance in the db,NOT CACHE DB!
func GetAllBalances() BalanceShow{

	var balances = make(BalanceShow)

	balanceIns, err := core.GetBalanceIns()

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}

	balMap := balanceIns.GetAllDBBalance()

	for key, value := range balMap {

		balances[key.Str()] = string(value)
	}

	return balances
}


