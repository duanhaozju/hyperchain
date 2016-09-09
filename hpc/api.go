package hpc

import (
	"github.com/op/go-logging"
	"hyperchain/core/types"
	"hyperchain/core"
	"time"
	//"github.com/golang/protobuf/proto"
	"hyperchain/manager"
	"hyperchain/event"
	"hyperchain/common"
	"github.com/syndtr/goleveldb/leveldb/errors"
	"hyperchain/hyperdb"
	"github.com/golang/protobuf/proto"
	"math/big"
)


var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc/api")
}

type PublicTransactionAPI struct {}


// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendArgs struct {
	From     string  `json:"from"`
	To       string `json:"to"`
	//Gas      *jsonrpc.HexNumber  `json:"gas"`
	//GasPrice *jsonrpc.HexNumber  `json:"gasPrice"`
	//Value    *jsonrpc.HexNumber  `json:"value"`
	Value    string  `json:"value"`
	//Payload     []byte          `json:"payload"`
	//Data     string          `json:"data"`
	//Nonce    *jsonrpc.HexNumber  `json:"nonce"`
}

type TransactionShow struct {
	Hash	  common.Hash		`json:"hash"`
	Block	  int			`json:"block"`
	From      common.Address	`json:"from"`
	To        common.Address	`json:"to"`
	Amount     string		`json:"amount"`
	Timestamp string		`json:"timestamp"`
}

func NewPublicTransactionAPI() *PublicTransactionAPI {
	return &PublicTransactionAPI{}
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash
func (hpc *PublicTransactionAPI) SendTransaction(args SendArgs) (common.Hash, error){

	log.Info("==========SendTransaction=====,args = ",args)

	var tx *types.Transaction

	log.Info(args.Value)
	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))

	if (core.VerifyBalance(tx)) {

		// Balance is enough
		/*txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}*/

		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		log.Info(tx.Value)
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		end:=start+1

		for start := start ; start < end; start = time.Now().Unix() {
			for i := 0; i < 5000; i++ {
				tx.TimeStamp=time.Now().UnixNano()
				txBytes, err := proto.Marshal(tx)
				if err != nil {
					log.Fatalf("proto.Marshal(tx) error: %v",err)
				}
				go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
				time.Sleep(200 * time.Microsecond)
			}
		}

		log.Infof("############# %d: end send request#############", time.Now().Unix())

		//tx.TimeStamp=time.Now().UnixNano()
		//txBytes, err := proto.Marshal(tx)
		//if err != nil {
		//	log.Fatalf("proto.Marshal(tx) error: %v",err)
		//}
		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		return tx.BuildHash(),nil

	} else {
		// Balance isn't enough
		return common.Hash{},errors.New("Not enough balance!")
	}
}

// GetAllTransactions return all transactions in the chain/db
func (hpc *PublicTransactionAPI) GetTransactions() []*TransactionShow{
	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}

	txs, err := core.GetAllTransaction(db)

	if err != nil {
		log.Fatalf("GetAllTransaction error: %v", err)
	}

	var transactions []*TransactionShow


	// TODO 得到交易所在的区块哈希
	for _, tx := range txs {
		var ts = &TransactionShow{
			Hash: tx.BuildHash(),
			Block: 1,
			Amount: string(tx.Value),
			From: common.BytesToAddress(tx.From),
			To: common.BytesToAddress(tx.To),
			Timestamp: time.Unix(tx.TimeStamp / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		}
		transactions = append(transactions,ts)
	}

	return transactions
}




type PublicNodeAPI struct{}

func NewPublicNodeAPI() *PublicNodeAPI{
	return &PublicNodeAPI{}
}

// TODO 得到节点状态
func (node *PublicNodeAPI) GetNodes() error{
	return nil
}

// TODO 节点交易平均处理时间
func QueryExcuteTime(args SendArgs) int64{

	var from big.Int
	var to big.Int
	from.SetString(args.From, 10)
	to.SetString(args.To, 10)


	return core.CalcResponseAVGTime(from.Uint64(),to.Uint64())
}


