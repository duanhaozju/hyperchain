package api

import (
	"hyperchain/core/types"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/hyperdb"
	"hyperchain/core"
	"hyperchain/manager"
	"github.com/op/go-logging"
	"time"
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

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc/api")
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is not enough, return false
func SendTransaction(args TxArgs) bool {

	var tx *types.Transaction

	log.Info(args)

	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))

	if (core.VerifyBalance(tx)) {

		// Balance is enough
		txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}

		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})

		log.Infof("############# %d: start send request#############", time.Now().Unix())

		for start := time.Now().Unix() ; start < start + 600; start = time.Now().Unix() {
			for i := 0; i < 1500; i++ {
				go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
				time.Sleep(666 * time.Microsecond)
			}
		}

		log.Infof("############# %d: end send request#############", time.Now().Unix())

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

	for _, tx := range txs {
		var ts = TransactionShow{
			Value: string(tx.Value),
			From: string(tx.From),
			To: string(tx.To),
			TimeStamp: tx.TimeStamp,
		}
		transactions = append(transactions,ts)
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


