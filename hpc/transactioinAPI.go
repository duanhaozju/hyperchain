package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
	"hyperchain/common"
	"time"
	"github.com/op/go-logging"
	"strconv"
	"hyperchain/core/types"
	"errors"
	"hyperchain/manager"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/vm/compiler"
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
// If type is Ptr or String, it is optional parameter
type SendTxArgs struct {
	//From     common.Address  `json:"from"`
	//To       *common.Address  `json:"to"`
	//Gas      *jsonrpc.Number  `json:"gas"`
	//GasPrice *jsonrpc.Number  `json:"gasPrice"`
	//Value    *jsonrpc.Number  `json:"value"`
	From     string  `json:"from"`
	To       string  `json:"to"`
	Gas      string  `json:"gas"`
	GasPrice string  `json:"gasPrice"`
	Value    string  `json:"value"`
	Payload  string  `json:"payload"`
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
	}
	if args.GasPrice == "" {
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
	//tx = types.NewTransaction(args.From[:], (*args.To)[:], []byte(args.Value))
	log.Info(tx.Value)
	am := tran.pm.AccountManager
	addr := common.HexToAddress(string(args.From))

	if (!core.VerifyBalance(tx)){
		return common.Hash{},errors.New("Not enough balance!")
	}else if _,found := am.Unlocked[addr];found {

		// Balance is enough
		//txBytes, err := proto.Marshal(tx)
		//if err != nil {
		//	log.Fatalf("proto.Marshal(tx) error: %v",err)
		//}

		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		end:=start+6
		//end:=start+500

		for start := start ; start < end; start = time.Now().Unix() {
			for i := 0; i < 100; i++ {
				tx.TimeStamp=time.Now().UnixNano()
				txBytes, err := proto.Marshal(tx)
				if err != nil {
					log.Errorf("proto.Marshal(tx) error: %v",err)
				}
				if manager.GetEventObject() != nil{
					go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes})
					//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
				}else{
					log.Warning("manager is Nil")
				}

			}
			time.Sleep(20 * time.Millisecond)

		}

		log.Infof("############# %d: end send request#############", time.Now().Unix())
		return tx.BuildHash(),nil

	} else {
		return common.Hash{}, errors.New("account isn't unlock")
	}
}

// SendTransactionOrContract deploy contract
func (tran *PublicTransactionAPI) SendTransactionOrContract(args SendTxArgs) (common.Hash, error){

	var tx *types.Transaction
	var amount int64

	realArgs := prepareExcute(args)

	gas, err := strconv.ParseInt(realArgs.Gas,10,64)
	price, err := strconv.ParseInt(realArgs.GasPrice,10,64)

	if realArgs.Value == "" {
		amount = 0
	} else {
		amount, err = strconv.ParseInt(realArgs.Value,10,64)
	}

	payload := common.FromHex(realArgs.Payload)

	if err != nil {
		return common.Hash{},err
	}

	txValue := types.NewTransactionValue(price,gas,amount,payload)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}

	if args.To == "" {
		tx = types.NewTransaction(common.HexToAddress(realArgs.From).Bytes(), nil, value)

	} else {
		tx = types.NewTransaction(common.HexToAddress(realArgs.From).Bytes(), common.HexToAddress(realArgs.To).Bytes(), value)
	}

	am := tran.pm.AccountManager
	addr := common.HexToAddress(args.From)

	if _,found := am.Unlocked[addr];found {
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		end:=start+6
		//end:=start+500

		for start := start ; start < end; start = time.Now().Unix() {
			for i := 0; i < 50; i++ {
				tx.TimeStamp=time.Now().UnixNano()
				txBytes, err := proto.Marshal(tx)
				if err != nil {
					log.Errorf("proto.Marshal(tx) error: %v",err)
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
	} else {
		return common.Hash{}, errors.New("account isn't unlock")
	}

	return tx.BuildHash(),nil
}

type CompileCode struct{
	Abi []string
	Bin []string
}

// ComplieContract complies contract to ABI
func (tran *PublicTransactionAPI) ComplieContract(ct string) (*CompileCode,error){

	abi, bin, err := compiler.CompileSourcefile(ct)

	if err != nil {
		return nil, err
	}

	return &CompileCode{
		Abi: abi,
		Bin: bin,
	}, nil
}

// GetTransactionReceipt returns transaction's receipt for given transaction hash
func (tran *PublicTransactionAPI) GetTransactionReceipt(hash common.Hash) *types.ReceiptTrans {
	return core.GetReceipt(hash)
}

// GetAllTransactions return all transactions in the chain/db
func (tran *PublicTransactionAPI) GetTransactions() []*TransactionResult{
	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
	}

	txs, err := core.GetAllTransaction(db)

	if err != nil {
		log.Errorf("GetAllTransaction error: %v", err)
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
