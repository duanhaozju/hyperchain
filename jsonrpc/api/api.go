package api

import (
	"hyperchain/core/types"
	"fmt"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/hyperdb"
	"log"
	"hyperchain/core"
	"hyperchain/manager"
	"hyperchain/logger"
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

type LastestBlockShow struct{
	Number uint64
	Hash []byte
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is not enough, return false
func SendTransaction(args TxArgs) bool {

	var tx *types.Transaction

	fmt.Println(args)

	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))

	if (core.VerifyBalance(tx)) {

		// Balance is enough
		txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}


		//for i := 0; i < 500; i += 1 {
			go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		//}

		return true

	} else {
		// Balance isn't enough
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

	myLogger.GetLogger().Println(txs)
	for index, tx := range txs {

		transactions[index].Value = string(tx.Value)
		transactions[index].From = string(tx.From)
		transactions[index].To = string(tx.To)
		transactions[index].TimeStamp = tx.TimeStamp
	}

	return transactions
}

// GetAllBalances retruns all account's balance in the db,NOT CACHE DB!
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

// LastestBlock returns the number and hash of the lastest block
func LastestBlock() LastestBlockShow{
	currentChain := core.GetChainCopy()



	return LastestBlockShow{
		Number: currentChain.Height,
		Hash: currentChain.LatestBlockHash,
	}
}


