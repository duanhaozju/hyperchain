package hyperchain

import (
	"hyperchain/core/types"
	"fmt"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/hyperdb"
	"log"
	"hyperchain/core"
	"math/big"
	"hyperchain/common"
	"hyperchain/manager"
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

type Balance map[common.Address]big.Int

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
		//eventmux:=new(event.TypeMux)

		txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}


		fmt.Println(txBytes)

		manager.GetEventObject().Post(event.NewTxEvent{Payload: []byte{0x00, 0x00, 0x03, 0xe8}})

		return true

	} else {
		// 余额不足
		return false
	}
}

// GetAllTransactions return all transactions in the chain/db
func GetAllTransactions()  []Transaction{

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

// GetAllBalances retrun all account's balance in the db,NOT CACHE DB!
func GetAllBalances() Balance{

	var val big.Int
	var balances Balance

	balanceIns, err := core.GetBalanceIns()
	//fmt.Println("get balance")

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}

	balMap := balanceIns.GetAllDBBalance()

	if balMap==nil{
		for key, value := range balMap {

			val.SetString(string(value),10)

			balances[key] = val
		}
	}


	return balances
}


