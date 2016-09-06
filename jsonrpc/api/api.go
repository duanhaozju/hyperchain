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
	"encoding/hex"
	"math/big"
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
	TimeStamp string
}

type BalanceShow map[string]string

type LastestBlockShow struct{
	Number uint64
	Hash string
}

type BlockShow struct{
        Height uint64
        TxCounts uint64
	WriteTime string
        Counts int64
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
		/*txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}*/

		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})

		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		end:=start+6

		//go func(){
			for start := start ; start < end; start = time.Now().Unix() {
			for i := 0; i < 5000; i++ {
				tx.TimeStamp=time.Now().UnixNano()
				txBytes, err := proto.Marshal(tx)
				if err != nil {
					log.Fatalf("proto.Marshal(tx) error: %v",err)
				}

				go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
				//time.Sleep(2 * time.Microsecond)

			}
		}
		//}()
		//go func(){
		//	for start := start ; start < end; start = time.Now().Unix() {
		//		for i := 0; i < 5000; i++ {
		//			tx.TimeStamp=time.Now().UnixNano()
		//			txBytes, err := proto.Marshal(tx)
		//			if err != nil {
		//				log.Fatalf("proto.Marshal(tx) error: %v",err)
		//			}
		//
		//			go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		//			//time.Sleep(2 * time.Microsecond)
		//
		//		}
		//	}
		//}()


		log.Infof("############# %d: end send request#############", time.Now().Unix())

		//tx.TimeStamp=time.Now().UnixNano()
		//txBytes, err := proto.Marshal(tx)
		//if err != nil {
		//	log.Fatalf("proto.Marshal(tx) error: %v",err)
		//}
		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})

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
			TimeStamp: time.Unix(tx.TimeStamp / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
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
		log.Info(key.Str())
		balances[key.Str()] = string(value)
	}

	return balances
}

// LastestBlock returns the number and hash of the lastest block
func LastestBlock() LastestBlockShow{
	currentChain := core.GetChainCopy()



	return LastestBlockShow{
		Number: currentChain.Height,
		Hash: hex.EncodeToString(currentChain.LatestBlockHash),
	}
}


func GetAllBlocks() []BlockShow{

	var blocks []BlockShow

	height := LastestBlock().Number

	for height > 0 {
		blocks = append(blocks,blockShow(height))
		height--
	}

	return blocks
}

func QueryExcuteTime(args TxArgs) int64{

	var from big.Int
	var to big.Int
	from.SetString(args.From, 10)
	to.SetString(args.To, 10)


	return core.CalcResponseAVGTime(from.Uint64(),to.Uint64())
}

func blockShow(height uint64) BlockShow{

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}

	blockHash, err := core.GetBlockHash(db,height)
	if err != nil {
		log.Fatalf("GetBlockHash error: %v", err)
	}

	block, err := core.GetBlock(db,blockHash)
	if err != nil {
		log.Fatalf("GetBlock error: %v", err)
	}

	txCounts := uint64(len(block.Transactions))

	return BlockShow{
			Height: height,
			TxCounts: txCounts,
			WriteTime: time.Unix(block.WriteTime / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
			Counts: core.CalcResponseCount(height, int64(300)),
		}

}

