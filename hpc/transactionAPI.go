package hpc

import (
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
	"time"
	//"hyperchain/accounts"
	"errors"
)

const (
	defaultGas int64 = 10000
	defaustGasPrice int64 = 10000
)

var (
	log        *logging.Logger // package-level logger
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

func init() {
	log = logging.MustGetLogger("hpc")
}

type PublicTransactionAPI struct {
	eventMux *event.TypeMux
	pm *manager.ProtocolManager

}


// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
// If type is Ptr or String, it is optional parameter
type SendTxArgs struct {
	From      common.Address	`json:"from"`
	To        *common.Address	`json:"to"`
	Gas       *Number  		`json:"gas"`
	GasPrice  *Number		`json:"gasPrice"`
	Value     *Number		`json:"value"`
	Payload   string		`json:"payload"`
	Signature string		`json:"signature"`
	//Nonce    *jsonrpc.HexNumber  `json:"nonce"`
}

type TransactionResult struct {
	Hash	  	common.Hash	`json:"hash"`
	BlockNumber	Number		`json:"blockNumber"`
	BlockHash	common.Hash	`json:"blockHash"`
	TxIndex	  	Number		`json:"txIndex"`
	From      	common.Address	`json:"from"`
	To        	common.Address	`json:"to"`
	Amount     	Number		`json:"amount"`
	Gas	   	Number		`json:"gas"`
	GasPrice   	Number		`json:"gasPrice"`
	Timestamp  	string		`json:"timestamp"`
	ExecuteTime	Number		`json:"executeTime"`
}

func NewPublicTransactionAPI(eventMux *event.TypeMux,pm *manager.ProtocolManager) *PublicTransactionAPI {
	return &PublicTransactionAPI{
		eventMux :eventMux,
		pm:pm,
	}
}

func prepareExcute(args SendTxArgs) SendTxArgs{
	if args.Gas == nil {
		args.Gas = NewInt64ToNumber(defaultGas)
	}
	if args.GasPrice == nil {
		args.GasPrice = NewInt64ToNumber(defaustGasPrice)
	}
	return args
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash
func (tran *PublicTransactionAPI) SendTransaction(args SendTxArgs) (common.Hash, error){

	var tx *types.Transaction
	var found bool

	realArgs := prepareExcute(args)
	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), nil)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}
	tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, []byte(args.Signature))


	if tran.pm == nil {

		// Test environment
		found = true
	} else {

		// Development environment
		am := tran.pm.AccountManager
		_, found = am.Unlocked[args.From]
	}
	//am := tran.pm.AccountManager

	//if (!core.VerifyBalance(tx)){
	//	return common.Hash{},errors.New("Not enough balance!")
	//}else
	if found == true {

		// Balance is enough

		//go manager.GetEventObject().Post(event.NewTxEvent{Payload: txBytes})
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		end := start + 6
		//end:=start+500

		for start := start; start < end; start = time.Now().Unix() {
			for i := 0; i < 100; i++ {
				tx.TimeStamp = time.Now().UnixNano()

				// calculate signature
				/*keydir := "./keystore/"
				encryption := crypto.NewEcdsaEncrypto("ecdsa")
				am := accounts.NewAccountManager(keydir, encryption)
				// TODO replace password with test value
				signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), "123")
				if err != nil {
					log.Errorf("Sign(tx) error :%v", err)
				}
				tx.Signature = signature*/
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
			time.Sleep(90 * time.Millisecond)
		}
	/*tx.TimeStamp = time.Now().UnixNano()

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
	time.Sleep(2000 * time.Millisecond)*/
		log.Infof("############# %d: end send request#############", time.Now().Unix())
	/*
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
	*/
	return tx.BuildHash(),nil

	} else {
		return common.Hash{}, errors.New("account don't unlock")
	}
}


// SendTransactionOrContract deploy contract
func (tran *PublicTransactionAPI) SendTransactionOrContract(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction
	var found bool

	realArgs := prepareExcute(args)

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(),realArgs.Gas.ToInt64(),realArgs.Value.ToInt64(),payload)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}

	if args.To == nil {

		// 部署合约
		tx = types.NewTransaction(realArgs.From[:], nil, value, []byte(args.Signature))

	} else {

		// 调用合约或者普通交易(普通交易还需要加检查余额)
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, []byte(args.Signature))
	}

	if tran.pm == nil {

		// Test environment
		found = true
	} else {

		// Development environment
		am := tran.pm.AccountManager
		_, found = am.Unlocked[args.From]

		//// TODO replace password with test value
		//signature, err := am.SignWithPassphrase(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes(), "123")
		//if err != nil {
		//	log.Errorf("Sign(tx) error :%v", err)
		//}
		//tx.Signature = signature

	}
	//am := tran.pm.AccountManager

	if found == true {
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		tx.TimeStamp = time.Now().UnixNano()

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
	} else {
		return common.Hash{}, errors.New("account don't unlock")
	}

	time.Sleep(2000 * time.Millisecond)
	/*
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
	*/
	return tx.BuildHash(), nil
}

type CompileCode struct {
	Abi []string
	Bin []string
}

// ComplieContract complies contract to ABI ---------------- (该方法已移到 contractAPI.go 中, 后期不再使用这里)
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

func outputTransaction(tx *types.Transaction) (*TransactionResult, error) {

	var txValue types.TransactionValue
	var bh common.Hash
	var bn , txIndex uint64
	var blk *types.Block

	txHash := tx.BuildHash()

	if err := proto.Unmarshal(tx.Value,&txValue); err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	if db, err := hyperdb.GetLDBDatabase(); err != nil {
		log.Errorf("Open database error: %v", err)
		return nil, err
	} else {
		bh, bn, txIndex = core.GetTxWithBlock(db, txHash[:])

		if blk, err = core.GetBlockByNumber(db, bn);err != nil {
			return nil, err
		}
	}



	return &TransactionResult{
		Hash: 		txHash,
		BlockNumber: 	*NewUint64ToNumber(bn),
		BlockHash: 	bh,
		TxIndex: 	*NewUint64ToNumber(txIndex),
		From: 		common.BytesToAddress(tx.From),
		To: 		common.BytesToAddress(tx.To),
		Amount: 	*NewInt64ToNumber(txValue.Amount),
		Gas: 		*NewInt64ToNumber(txValue.GasLimit),
		GasPrice: 	*NewInt64ToNumber(txValue.Price),
		Timestamp: 	time.Unix(tx.TimeStamp / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		ExecuteTime:	(blk.WriteTime - tx.TimeStamp) / int64(time.Millisecond),
	}, nil
}

// GetTransactionReceipt returns transaction's receipt for given transaction hash
func (tran *PublicTransactionAPI) GetTransactionReceipt(hash common.Hash) *types.ReceiptTrans {
	return core.GetReceipt(hash)
}

// GetAllTransactions return all transactions in the chain/db
//func (tran *PublicTransactionAPI) GetTransactions() []*TransactionResult{
//	db, err := hyperdb.GetLDBDatabase()
//
//	if err != nil {
//		log.Errorf("Open database error: %v", err)
//	}
//
//	txs, err := core.GetAllTransaction(db)
//
//	if err != nil {
//		log.Errorf("GetAllTransaction error: %v", err)
//	}
//
//	var transactions []*TransactionResult
//
//
//	// TODO 1.得到交易所在的区块哈希 2.取出 tx.Value 中的 amount
//	for _, tx := range txs {
//		var ts = &TransactionResult{
//			Hash: tx.BuildHash(),
//			//Block: 1,
//			Amount: string(tx.Value),
//			From: common.BytesToAddress(tx.From),
//			To: common.BytesToAddress(tx.To),
//			Timestamp: time.Unix(tx.TimeStamp / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
//		}
//		transactions = append(transactions,ts)
//	}
//
//	return transactions
//}

// GetTransactionByHash returns the transaction for the given transaction hash.
func (tran *PublicTransactionAPI) GetTransactionByHash(hash common.Hash) (*TransactionResult, error){

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
		return nil, err
	}

	tx, err := core.GetTransaction(db, hash[:])

	if tx.From == nil {
		return nil, errors.New("Not found this transaction")
	}

	return outputTransaction(tx)
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (tran *PublicTransactionAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, index Number) (*TransactionResult, error) {

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
		return nil, err
	}

	block, err := core.GetBlock(db, hash[:])
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx)
	}

	return nil, nil
}

// GetTransactionsByBlockNumberAndIndex returns the transaction for the given block number and index.
func (tran *PublicTransactionAPI) GetTransactionsByBlockNumberAndIndex(n Number, index Number) (*TransactionResult, error){

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
		return nil, err
	}

	block, err := core.GetBlockByNumber(db, n.ToUint64())
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx)
	}

	return nil, nil
}


