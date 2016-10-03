package hpc

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/core/vm/compiler"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"strconv"
	"time"
)

const (
	defaultGas      int = 10000000
	defaustGasPrice int = 10000000
)

var (
	log        *logging.Logger // package-level logger
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

func init() {
	log = logging.MustGetLogger("jsonrpc/api")
}

type PublicTransactionAPI struct {
	eventMux *event.TypeMux
	pm       *manager.ProtocolManager
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
// If type is Ptr or String, it is optional parameter
type SendTxArgs struct {
	//From     common.Address  `json:"from"`
	//To       *common.Address  `json:"to"`
	//Gas      *jsonrpc.Number  `json:"gas"`
	//GasPrice *jsonrpc.Number  `json:"gasPrice"`
	//Value    *jsonrpc.Number  `json:"value"`
	From     string `json:"from"`
	To       string `json:"to"`
	Gas      string `json:"gas"`
	GasPrice string `json:"gasPrice"`
	Value    string `json:"value"`
	Payload  string `json:"payload"`
	//Nonce    *jsonrpc.HexNumber  `json:"nonce"`
}

type TransactionResult struct {
	Hash common.Hash `json:"hash"`
	//Block	  int			`json:"block"`
	From      common.Address `json:"from"`
	To        common.Address `json:"to"`
	Amount    string         `json:"amount"`
	Timestamp string         `json:"timestamp"`
}

func NewPublicTransactionAPI(eventMux *event.TypeMux, pm *manager.ProtocolManager) *PublicTransactionAPI {
	return &PublicTransactionAPI{
		eventMux: eventMux,
		pm:       pm,
	}
}

func prepareExcute(args SendTxArgs) SendTxArgs {
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
func (tran *PublicTransactionAPI) SendTransaction(args SendTxArgs) (common.Hash, error) {
	log.Info("==========SendTransaction=====,args = ", args)
	args = prepareExcute(args)
	var tx *types.Transaction

	// (1) parse args
	payload := common.FromHex(args.Payload)
	amount, _ := strconv.ParseInt(common.HexToString(args.Value), 16, 64)
	gasLimit, _ := strconv.ParseInt(common.HexToString(args.Gas), 16, 64)
	gasPrice, _ := strconv.ParseInt(common.HexToString(args.GasPrice), 16, 64)
	tv := types.NewTransactionValue(gasPrice, gasLimit, amount, payload)
	tvData, _ := proto.Marshal(tv)
	tx = types.NewTransaction(common.HexToAddress(args.From).Bytes(), common.HexToAddress(args.To).Bytes(), tvData)

	//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
	log.Infof("############# %d: start send request#############", time.Now().Unix())
	/*
		start := time.Now().Unix()
		end := start + 6
		//end:=start+500

		for start := start; start < end; start = time.Now().Unix() {
			for i := 0; i < 10; i++ {
				tx.TimeStamp = time.Now().UnixNano()

				// calculate signature
				keydir := "./keystore/"
				encryption := crypto.NewEcdsaEncrypto("ecdsa")
				am := accounts.NewAccountManager(keydir, encryption)
				// TODO replace password with test value
				signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), "123")
				if err != nil {
					log.Errorf("Sign(tx) error :%v", err)
				}
				tx.Signature = signature
				txBytes, err := proto.Marshal(tx)
				if err != nil {
					log.Errorf("proto.Marshal(tx) error: %v", err)
				}
				if manager.GetEventObject() != nil {
					go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes})
					//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
				} else {
					log.Warning("manager is Nil")
				}
			}
			time.Sleep(20 * time.Millisecond)
		}
	*/
	tx.TimeStamp = time.Now().UnixNano()

	// TODO replace password with test value
	signature, err := tran.pm.AccountManager.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), "123")
	if err != nil {
		log.Errorf("Sign(tx) error :%v", err)
	}
	tx.Signature = signature
	txBytes, err := proto.Marshal(tx)
	if err != nil {
		log.Errorf("proto.Marshal(tx) error: %v", err)
	}
	if manager.GetEventObject() != nil {
		go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes})
		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
	} else {
		log.Warning("manager is Nil")
	}
	log.Infof("############# %d: end send request#############", time.Now().Unix())

	time.Sleep(2000 * time.Millisecond)
	receipt := core.GetReceipt(tx.BuildHash())
	fmt.Println("GasUsed", receipt.GasUsed)
	fmt.Println("PostState", receipt.PostState)
	fmt.Println("ContractAddress", receipt.ContractAddress)
	fmt.Println("CumulativeGasUsed", receipt.CumulativeGasUsed)
	fmt.Println("Ret", receipt.Ret)
	fmt.Println("TxHash", receipt.TxHash)
	fmt.Println("Status", receipt.Status)
	fmt.Println("Message", receipt.Message)
	fmt.Println("Log", receipt.Logs)
	return tx.BuildHash(), nil
}

// SendTransactionOrContract deploy contract
func (tran *PublicTransactionAPI) SendTransactionOrContract(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction
	//var amount int64

	realArgs := prepareExcute(args)

	//gas, err := strconv.ParseInt(realArgs.Gas,10,64)
	//price, err := strconv.ParseInt(realArgs.GasPrice,10,64)
	//
	//if realArgs.Value == "" {
	//	amount = 0
	//} else {
	//	amount, err = strconv.ParseInt(realArgs.Value,10,64)
	//}

	payload := common.FromHex(realArgs.Payload)
	amount, _ := strconv.ParseInt(common.HexToString(realArgs.Value), 16, 64)
	gasLimit, _ := strconv.ParseInt(common.HexToString(args.Gas), 16, 64)
	gasPrice, _ := strconv.ParseInt(common.HexToString(args.GasPrice), 16, 64)
	txValue := types.NewTransactionValue(gasPrice, gasLimit, amount, payload)
	value, _ := proto.Marshal(txValue)

	if args.To == "" {
		tx = types.NewTransaction(common.HexToAddress(realArgs.From).Bytes(), nil, value)
	} else {
		tx = types.NewTransaction(common.HexToAddress(realArgs.From).Bytes(), common.FromHex(realArgs.To), value)
	}

	am := tran.pm.AccountManager

	log.Infof("############# %d: start send request#############", time.Now().Unix())
	tx.TimeStamp = time.Now().UnixNano()

	// TODO replace password with test value
	signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), "123")
	if err != nil {
		log.Errorf("Sign(tx) error :%v", err)
	}
	tx.Signature = signature

	txBytes, err := proto.Marshal(tx)
	if err != nil {
		log.Errorf("proto.Marshal(tx) error: %v", err)
	}
	if manager.GetEventObject() != nil {
		go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes})
	} else {
		log.Warning("manager is Nil")
	}
	log.Infof("############# %d: end send request#############", time.Now().Unix())

	time.Sleep(2000 * time.Millisecond)
	receipt := core.GetReceipt(tx.BuildHash())
	fmt.Println("GasUsed", receipt.GasUsed)
	fmt.Println("PostState", receipt.PostState)
	fmt.Println("ContractAddress", receipt.ContractAddress)
	fmt.Println("CumulativeGasUsed", receipt.CumulativeGasUsed)
	fmt.Println("Ret", receipt.Ret)
	fmt.Println("TxHash", receipt.TxHash)
	fmt.Println("Status", receipt.Status)
	fmt.Println("Message", receipt.Message)
	fmt.Println("Log", receipt.Logs)

	return tx.BuildHash(), nil
}

type CompileCode struct {
	Abi []string
	Bin []string
}

// ComplieContract complies contract to ABI
func (tran *PublicTransactionAPI) ComplieContract(ct string) (*CompileCode, error) {

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
	log.Info("transactionAPI.go,", core.GetReceipt(hash).ContractAddress)
	return core.GetReceipt(hash)
}

// GetAllTransactions return all transactions in the chain/db
func (tran *PublicTransactionAPI) GetTransactions() []*TransactionResult {
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
			Amount:    string(tx.Value),
			From:      common.BytesToAddress(tx.From),
			To:        common.BytesToAddress(tx.To),
			Timestamp: time.Unix(tx.TimeStamp/int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		}
		transactions = append(transactions, ts)
	}

	return transactions
}
