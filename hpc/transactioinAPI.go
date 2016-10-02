package hpc

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/accounts"
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
	defaultGas      int = 10000
	defaustGasPrice int = 10000
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
	var tx *types.Transaction
	amount, _ := strconv.Atoi(common.HexToString(args.Value))
	tv := types.NewTransactionValue(0, 0, int64(amount), nil)
	tvData, _ := proto.Marshal(tv)
	tx = types.NewTransaction(common.HexToAddress(args.From).Bytes(), common.HexToAddress(args.To).Bytes(), tvData)
	log.Info(tx.Value)

	// TODO check balance

	//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
	log.Infof("############# %d: start send request#############", time.Now().Unix())
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

	log.Infof("############# %d: end send request#############", time.Now().Unix())
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

	if args.To == "" {
		//payload := common.FromHex("60606040526000600060006101000a81548163ffffffff021916908302179055507f6162636465666768696a6b6c6d6e6f707172737475767778797a0000000000006001600050556101aa806100556000396000f360606040526000357c0100000000000000000000000000000000000000000000000000000000900480633ad14af31461005a578063569c5f6d1461007b5780638da9b772146100a4578063d09de08a146100cb57610058565b005b61007960048080359060200190919080359060200190919050506100f4565b005b610088600480505061012a565b604051808263ffffffff16815260200191505060405180910390f35b6100b16004805050610149565b604051808260001916815260200191505060405180910390f35b6100d8600480505061015b565b604051808263ffffffff16815260200191505060405180910390f35b8082600060009054906101000a900463ffffffff160101600060006101000a81548163ffffffff021916908302179055505b5050565b6000600060009054906101000a900463ffffffff169050610146565b90565b60006001600050549050610158565b90565b60006001600060009054906101000a900463ffffffff1601600060006101000a81548163ffffffff02191690830217905550600060009054906101000a900463ffffffff1690506101a7565b9056")
		//var code = "0x6000805463ffffffff1916815560a0604052600b6060527f68656c6c6f20776f726c6400000000000000000000000000000000000000000060805260018054918190527f68656c6c6f20776f726c6400000000000000000000000000000000000000001681559060be907fb10e2d527612073b26eecdfd717e6a320cf44b4afac2b0732d9fcbe2b7fa0cf66020600261010084871615026000190190931692909204601f01919091048101905b8082111560ce576000815560010160ac565b50506101b8806100d26000396000f35b509056606060405260e060020a60003504633ad14af3811461003c578063569c5f6d146100615780638da9b77214610071578063d09de08a146100da575b005b6000805460043563ffffffff8216016024350163ffffffff199190911617905561003a565b6100f960005463ffffffff165b90565b604080516020818101835260008252600180548451600261010083851615026000190190921691909104601f81018490048402820184019095528481526101139490928301828280156101ac5780601f10610181576101008083540402835291602001916101ac565b61003a6000805463ffffffff19811663ffffffff909116600101179055565b6040805163ffffffff929092168252519081900360200190f35b60405180806020018281038252838181518152602001915080519060200190808383829060006004602084601f0104600302600f01f150905090810190601f1680156101735780820380516001836020036101000a031916815260200191505b509250505060405180910390f35b820191906000526020600020905b81548152906001019060200180831161018f57829003601f168201915b5050505050905061006e56"
		//payload := common.FromHex(code)
		payload := common.FromHex(realArgs.Payload)
		txValue := types.NewTransactionValue(100000, 100000, 100, payload)
		value, _ := proto.Marshal(txValue)
		tx = types.NewTransaction(common.HexToAddress(realArgs.From).Bytes(), nil, value)
		//tx = types.NewTestCreateTransaction()
		//log.Info("the to is null")
	} else {
		payload := common.FromHex(realArgs.Payload)
		txValue := types.NewTransactionValue(100000, 100000, 100, payload)
		value, _ := proto.Marshal(txValue)
		tx = types.NewTransaction(common.HexToAddress(realArgs.From).Bytes(), common.FromHex(realArgs.To), value)
		//log.Info("the old value is",tx.Payload())
		//tx = types.NewTestCallTransaction()
		//tx.From = common.HexToAddress(realArgs.From).Bytes()
		//tx.To = common.FromHex(realArgs.To)
		//log.Info("the transactionapi real to byte is",realArgs.To)
		//log.Info("the real from addr is",tx.From)
		//log.Info("the real to addr is %#V",tx.To)
		//log.Info("the new value is",tx.Payload())
	}

	am := tran.pm.AccountManager
	addr := common.HexToAddress(args.From)

	if _, found := am.Unlocked[addr]; found {
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		//start := time.Now().Unix()
		//end:=start+6
		//end:=start+500

		//for start := start ; start < end; start = time.Now().Unix() {
		//	for i := 0; i < 50; i++ {
		tx.TimeStamp = time.Now().UnixNano()
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
		//}
		//time.Sleep(90 * time.Millisecond)
		//}

		log.Infof("############# %d: end send request#############", time.Now().Unix())
	} else {
		return common.Hash{}, errors.New("account isn't unlock")
	}

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
