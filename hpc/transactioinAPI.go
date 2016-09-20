package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
	"hyperchain/common"
	"time"
	"github.com/op/go-logging"
	"encoding/json"
	"strconv"
	"hyperchain/core/types"
	"errors"
	"hyperchain/manager"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/vm/compiler"

	"hyperchain/core/crypto"
)

const (
	defaultGas int = 10000
	defaustGasPrice int = 10000
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc/api")
}

type PublicTransactionAPI struct {
	eventMux *event.TypeMux
	pm *manager.ProtocolManager

}


// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Gas      string  `json:"gas"`
	GasPrice string  `json:"gasPrice"`
	//Value    *jsonrpc.HexNumber  `json:"value"`
	Value    string  `json:"value"`
	Payload  string  `json:"payload"`
	//Data     string          `json:"data"`
	//Nonce    *jsonrpc.HexNumber  `json:"nonce"`
}

type TransactionResult struct {
	Hash	  common.Hash		`json:"hash"`
	//Block	  int			`json:"block"`
	From      common.Address	`json:"from"`
	To        common.Address	`json:"to"`
	Amount     string		`json:"amount"`
	Timestamp  string		`json:"timestamp"`
}

func NewPublicTransactionAPI(eventMux *event.TypeMux,pm *manager.ProtocolManager) *PublicTransactionAPI {
	return &PublicTransactionAPI{
		eventMux :eventMux,
		pm:pm,
	}
}

func prepareExcute(args SendTxArgs) SendTxArgs{
	if args.Gas == "" {
		args.Gas = strconv.Itoa(defaultGas)
	} else if args.GasPrice == "" {
		args.GasPrice = strconv.Itoa(defaustGasPrice)
	}

	return args
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash
func (tran *PublicTransactionAPI) SendTransaction(args SendTxArgs) (common.Hash, error){

	log.Info("==========SendTransaction=====,args = ",args)
	var tx *types.Transaction
	log.Info(args.Value)
	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))
	log.Info(tx.Value)
	am := tran.pm.AccountManager
	addr := common.HexToAddress(string(args.From))

	if (!core.VerifyBalance(tx)){
		return common.Hash{},errors.New("Not enough balance!")
	}else if _,found := am.Unlocked[addr];found {

		// Balance is enough
		/*txBytes, err := proto.Marshal(tx)
		if err != nil {
			log.Fatalf("proto.Marshal(tx) error: %v",err)
		}*/

		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		end:=start+6
		//end:=start+500

		for start := start ; start < end; start = time.Now().Unix() {
			for i := 0; i < 5; i++ {
				tx.TimeStamp=time.Now().UnixNano()
				txBytes, err := proto.Marshal(tx)
				if err != nil {
					log.Fatalf("proto.Marshal(tx) error: %v",err)
				}
				if manager.GetEventObject() != nil{
					go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes})
					//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
				}else{
					log.Warning("manager is Nil")
				}

			}
			time.Sleep(90 * time.Millisecond)

		}

		log.Infof("############# %d: end send request#############", time.Now().Unix())
		return tx.BuildHash(),nil

	} else {
		// Balance isn't enough
		return common.Hash{},errors.New("Not enough balance!")
	}
}

// SendTransactionOrContract deploy contract
func (tran *PublicTransactionAPI) SendTransactionOrContract(args SendTxArgs) (common.Address, error){

	var tx *types.Transaction

	realArgs := prepareExcute(args)

	gas, err := strconv.ParseInt(realArgs.Gas,10,64)
	price, err := strconv.ParseInt(realArgs.GasPrice,10,64)
	amount, err := strconv.ParseInt(realArgs.Value,10,64)
	payload, err := json.Marshal(realArgs.Payload)

	if err != nil {
		return common.Address{},err
	}

	txValue := types.NewTransactionValue(price,gas,amount,payload)

	value, err := json.Marshal(txValue)

	if err != nil {
		return common.Address{}, err
	}

	var addr common.Address
	if args.To == "" {
		tx = types.NewTransaction([]byte(realArgs.From), nil, value)
		nonce := core.GetVMEnv().Db().GetNonce(common.BytesToAddress(tx.From))
		core.GetVMEnv().Db().SetNonce(common.BytesToAddress(tx.From), nonce+1)
		addr = crypto.CreateAddress(common.BytesToAddress(tx.From), nonce)
	} else {
		tx = types.NewTransaction([]byte(realArgs.From), []byte(realArgs.To), value)
	}

	// todo 其他处理,比如存储到数据库中
	//db, err := hyperdb.GetLDBDatabase()
	//
	//if err != nil {
	//	log.Fatalf("Open database error: %v", err)
	//}
	//
	//core.PutTransaction(db, ,tx)

	return addr,nil
	//return tx.BuildHash(),nil
}

// ComplieContract complies contract to ABI
func (tran *PublicTransactionAPI) ComplieContract(ct string) ([]string, error){

	log.Debug(ct)
	abi, _, err := compiler.CompileSourcefile(ct)
	log.Debug(abi)

	if err != nil {
		return nil, err
	}

	return abi,nil
}



// GetAllTransactions return all transactions in the chain/db
func (tran *PublicTransactionAPI) GetTransactions() []*TransactionResult{
	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Fatalf("Open database error: %v", err)
	}

	txs, err := core.GetAllTransaction(db)

	if err != nil {
		log.Fatalf("GetAllTransaction error: %v", err)
	}

	var transactions []*TransactionResult


	// TODO 得到交易所在的区块哈希
	for _, tx := range txs {
		var ts = &TransactionResult{
			Hash: tx.BuildHash(),
			//Block: 1,
			Amount: string(tx.Value),
			From: common.BytesToAddress(tx.From),
			To: common.BytesToAddress(tx.To),
			Timestamp: time.Unix(tx.TimeStamp / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		}
		transactions = append(transactions,ts)
	}

	return transactions
}
