package hpc

import (
	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"time"
	//"hyperchain/accounts"
	"encoding/hex"
	"errors"
)

const (
	defaultGas int64 = 10000
	defaustGasPrice int64 = 10000
)

var (
	log        *logging.Logger // package-level logger
	encryption = crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

func init() {
	log = logging.MustGetLogger("hpc")
}

type PublicTransactionAPI struct {
	eventMux *event.TypeMux
	pm       *manager.ProtocolManager
	db       *hyperdb.LDBDatabase
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
// If type is Ptr or String, it is optional parameter
type SendTxArgs struct {
	From      common.Address  `json:"from"`
	To        *common.Address `json:"to"`
	Gas       *Number         `json:"gas"`
	GasPrice  *Number         `json:"gasPrice"`
	Value     *Number         `json:"value"`
	Payload   string          `json:"payload"`
	Signature string                `json:"signature"`
	Timestamp int64                 `json:"timestamp"`
	//Nonce    *jsonrpc.HexNumber  `json:"nonce"`
	// --- test -----
	PrivKey   string `json:"privKey"`
	Request   *Number `json:"request"`
}

type TransactionResult struct {
	Hash        common.Hash    `json:"hash"`
	BlockNumber *Number        `json:"blockNumber"`
	BlockHash   common.Hash    `json:"blockHash"`
	TxIndex     *Number        `json:"txIndex"`
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Amount      *Number        `json:"amount"`
	//Gas         *Number        `json:"gas"`
	//GasPrice    *Number        `json:"gasPrice"`
	Timestamp   int64         `json:"timestamp"`
	ExecuteTime *Number        `json:"executeTime"`
	Invalid     bool           `json:"invalid"`
	InvalidMsg  string           `json:"invalidMsg"`
}

// ----- 性能测试参数 ---------
const (
	//DURATION int64 = 230400
	//COUNT = 25
	//SLEEPTIME int = 300
	DURATION int64 = 30
	COUNT int = 100
	SLEEPTIME time.Duration = 90
)

const TIMEOUT int64 = 2

func NewPublicTransactionAPI(eventMux *event.TypeMux, pm *manager.ProtocolManager, hyperDb *hyperdb.LDBDatabase) *PublicTransactionAPI {
	return &PublicTransactionAPI{
		eventMux: eventMux,
		pm:       pm,
		db:          hyperDb,
	}
}

func prepareExcute(args SendTxArgs) SendTxArgs {
	if args.Gas == nil {
		args.Gas = NewInt64ToNumber(defaultGas)
	}
	if args.GasPrice == nil {
		args.GasPrice = NewInt64ToNumber(defaustGasPrice)
	}
	return args
}

//func (tran *PublicTransactionAPI) SendRawTransaction(args SendTxArgs) (common.Hash, error) {
//
//	tx := new(types.Transaction)
//	//var tx *types.Transaction
//	if err := rlp.DecodeBytes(common.FromHex(args.Signature), tx); err != nil {
//		log.Info("rlp.DecodeBytes error: ", err)
//		return common.Hash{}, err
//	}
//	log.Infof("tx: %#v", tx)
//	log.Info(tx.Signature)
//	log.Infof("tx sign: %#v", tx.Signature)
//
//	return tx.BuildHash(), nil
//}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash
func (tran *PublicTransactionAPI) SendTransaction(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction
	//var found bool

	realArgs := prepareExcute(args)

	// ################################# 测试代码 START ####################################### //
	var v int64 = realArgs.Value.ToInt64()
	if realArgs.Timestamp == 0 {
		realArgs.Timestamp = time.Now().UnixNano()
	}
	// ################################# 测试代码 END ####################################### //

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), nil)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}

	//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, common.FromHex(args.Signature))
	tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp)

	if args.Request != nil {

		// ** For Hyperboard Test **
		for i := 0; i < (*args.Request).ToInt(); i++ {
			// ################################# 测试代码 START ####################################### // (用不同的value代替之前不同的timestamp以标志不同的transaction)
			txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), v, nil)

			value, err := proto.Marshal(txValue)

			if err != nil {
				return common.Hash{}, err
			}
			tx.Value = value
			v++
			// ################################# 测试代码 END ####################################### //
			tx.Id = uint64(tran.pm.Peermanager.GetNodeId())

			if realArgs.PrivKey == "" {
				// For Hyperchain test signature

				// TODO replace password with test value
				signature, err := tran.pm.AccountManager.Sign(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes())
				if err != nil {
					log.Errorf("Sign(tx) error :%v", err)
				}
				tx.Signature = signature
			} else {
				// For Hyperboard test signature

				key, err := hex.DecodeString(args.PrivKey)
				if err != nil {
					return common.Hash{}, err
				}
				pri := crypto.ToECDSA(key)

				hash := tx.SighHash(kec256Hash).Bytes()
				sig, err := encryption.Sign(hash, pri)
				if err != nil {
					return common.Hash{}, err
				}

				tx.Signature = sig
			}

			tx.TransactionHash = tx.BuildHash().Bytes()
			// Unsign Test
			if !tx.ValidateSign(tran.pm.AccountManager.Encryption, kec256Hash) {
				log.Error("invalid signature")
				// 不要返回，因为要将失效交易存到db中
				//return common.Hash{}, errors.New("invalid signature")
			}

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

			start_getErr := time.Now().Unix()
			end_getErr := start_getErr + TIMEOUT
			var errMsg string
			for start_getErr := start_getErr; start_getErr < end_getErr; start_getErr = time.Now().Unix() {
				errType, _ := core.GetInvaildTxErrType(tran.db, tx.GetTransactionHash().Bytes());

				if errType != -1 {
					errMsg = errType.String()
					break;
				} else if rept := core.GetReceipt(tx.GetTransactionHash()); rept != nil {
					break
				}

			}
			if start_getErr != end_getErr && errMsg != "" {
				return common.Hash{}, errors.New(errMsg)
			} else if start_getErr == end_getErr {
				return common.Hash{}, errors.New("Sending return timeout,may be something wrong.")
			}
		}

	} else {

		// ** For Hyperchain Test **
		log.Infof("############# %d: start send request#############", time.Now().Unix())
		start := time.Now().Unix()
		//end:=start+1
		end := start + DURATION

		for start := start; start < end; start = time.Now().Unix() {
			for i := 0; i < COUNT; i++ {
				// ################################# 测试代码 START ####################################### // (用不同的value代替之前不同的timestamp以标志不同的transaction)
				txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), v, nil)

				value, err := proto.Marshal(txValue)

				if err != nil {
					return common.Hash{}, err
				}
				tx.Value = value
				v++
				// ################################## 测试代码 END ####################################### //
				tx.Id = uint64(tran.pm.Peermanager.GetNodeId())

				if realArgs.PrivKey == "" {
					// For Hyperchain test

					// TODO replace password with test value
					signature, err := tran.pm.AccountManager.Sign(common.BytesToAddress(tx.From), tx.SighHash(kec256Hash).Bytes())
					if err != nil {
						log.Errorf("Sign(tx) error :%v", err)
					}
					tx.Signature = signature
				} else {
					// For Hyperboard test

					key, err := hex.DecodeString(args.PrivKey)
					if err != nil {
						return common.Hash{}, err
					}
					pri := crypto.ToECDSA(key)

					hash := tx.SighHash(kec256Hash).Bytes()
					sig, err := encryption.Sign(hash, pri)
					if err != nil {
						return common.Hash{}, err
					}

					tx.Signature = sig
				}

				tx.TransactionHash = tx.BuildHash().Bytes()
				// Unsign Test
				if !tx.ValidateSign(tran.pm.AccountManager.Encryption, kec256Hash) {
					log.Error("invalid signature")
					// 不要返回，因为要将失效交易存到db中
					//return common.Hash{}, errors.New("invalid signature")
				}

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

				start_getErr := time.Now().Unix()
				end_getErr := start_getErr + TIMEOUT
				var errMsg string
				for start_getErr := start_getErr; start_getErr < end_getErr; start_getErr = time.Now().Unix() {
					errType, _ := core.GetInvaildTxErrType(tran.db, tx.GetTransactionHash().Bytes());

					if errType != -1 {
						errMsg = errType.String()
						break;
					} else if rept := core.GetReceipt(tx.GetTransactionHash()); rept != nil {
						break
					}

				}
				if start_getErr != end_getErr && errMsg != "" {
					return common.Hash{}, errors.New(errMsg)
				} else if start_getErr == end_getErr {
					return common.Hash{}, errors.New("Sending return timeout,may be something wrong.")
				}
			}
			time.Sleep(SLEEPTIME * time.Millisecond)
		}

		log.Infof("############# %d: end send request#############", time.Now().Unix())
	}

	return tx.GetTransactionHash(), nil

}

// GetTransactionReceipt returns transaction's receipt for given transaction hash.
func (tran *PublicTransactionAPI) GetTransactionReceipt(hash common.Hash) (*types.ReceiptTrans, error) {

	if errType, err := core.GetInvaildTxErrType(tran.db, hash.Bytes()); errType == -1 {
		return core.GetReceipt(hash), nil
	} else if err != nil {
		return nil, err
	} else {
		return nil, errors.New(errType.String())
	}

}

// GetTransactions return all transactions in the chain/db
func (tran *PublicTransactionAPI) GetTransactions() ([]*TransactionResult, error) {

	txs, err := core.GetAllTransaction(tran.db)

	if err != nil {
		log.Errorf("GetAllTransaction error: %v", err)
		return nil, err
	}

	var transactions []*TransactionResult

	for _, tx := range txs {
		if ts, err := outputTransaction(tx, tran.db); err != nil {
			return nil, err
		} else {
			transactions = append(transactions, ts)
		}
	}

	return transactions, nil
}

// GetDiscardTransactions return all invalid transaction that dont be saved on the blockchain.
func (tran *PublicTransactionAPI) GetDiscardTransactions() ([]*TransactionResult, error) {

	reds, err := core.GetAllDiscardTransaction(tran.db)

	if err != nil {
		log.Errorf("GetAllDiscardTransaction error: %v", err)
		return nil, err
	}

	var transactions []*TransactionResult

	for _, red := range reds {
		log.Notice(red.Tx.TransactionHash)
		if ts, err := outputTransaction(red, tran.db); err != nil {
			return nil, err
		} else {
			transactions = append(transactions, ts)
		}
	}

	return transactions, nil
}

// GetTransactionByHash returns the transaction for the given transaction hash.
func (tran *PublicTransactionAPI) GetTransactionByHash(hash common.Hash) (*TransactionResult, error) {

	tx, err := core.GetTransaction(tran.db, hash[:])

	if err != nil {
		return nil, err
	}

	if tx.From == nil {
		return nil, errors.New("Not found this transaction")
	}

	return outputTransaction(tx, tran.db)
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (tran *PublicTransactionAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, index Number) (*TransactionResult, error) {

	block, err := core.GetBlock(tran.db, hash[:])
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.db)
	}

	return nil, nil
}

// GetTransactionsByBlockNumberAndIndex returns the transaction for the given block number and index.
func (tran *PublicTransactionAPI) GetTransactionByBlockNumberAndIndex(n Number, index Number) (*TransactionResult, error) {

	block, err := core.GetBlockByNumber(tran.db, uint64(n))
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.db)
	}

	return nil, nil
}

// GetBlockTransactionCountByHash returns the number of block transactions for given block hash.
func (tran *PublicTransactionAPI) GetBlockTransactionCountByHash(hash common.Hash) (*Number, error) {

	block, err := core.GetBlock(tran.db, hash[:])
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	txCount := len(block.Transactions)

	return NewIntToNumber(txCount), nil
}

// 这个方法先保留
func (tran *PublicTransactionAPI) GetSighHash(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction

	realArgs := prepareExcute(args)

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), payload)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, err
	}

	if args.To == nil {

		// 部署合约
		//tx = types.NewTransaction(realArgs.From[:], nil, value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp)

	} else {

		// 调用合约或者普通交易(普通交易还需要加检查余额)
		//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, []byte(args.Signature))
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp)
	}

	return tx.SighHash(kec256Hash), nil
}

//func outputTransaction(tx *types.Transaction, db *hyperdb.LDBDatabase) (*TransactionResult, error) {
func outputTransaction(trans interface{}, db *hyperdb.LDBDatabase) (*TransactionResult, error) {

	var txValue types.TransactionValue
	var blk *types.Block

	var tx *types.Transaction
	var red *types.InvalidTransactionRecord

	tx, found := trans.(*types.Transaction)

	if found == false {
		red = trans.(*types.InvalidTransactionRecord)
		tx = red.Tx
	}

	txHash := tx.GetTransactionHash()

	if err := proto.Unmarshal(tx.Value, &txValue); err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	bn, txIndex := core.GetTxWithBlock(db, txHash[:])

	blk, err := core.GetBlockByNumber(db, bn)
	if err != nil {
		return nil, err
	}

	txRes := &TransactionResult{
		Hash:                txHash,
		BlockNumber:        NewUint64ToNumber(bn),
		BlockHash:        common.BytesToHash(blk.BlockHash),
		TxIndex:        NewInt64ToNumber(txIndex),
		From:                common.BytesToAddress(tx.From),
		To:                common.BytesToAddress(tx.To),
		Amount:        NewInt64ToNumber(txValue.Amount),
		//Gas: 		NewInt64ToNumber(txValue.GasLimit),
		//GasPrice: 	NewInt64ToNumber(txValue.Price),
		Timestamp:        tx.Timestamp / 1e6,
		ExecuteTime:        NewInt64ToNumber((blk.WriteTime - tx.Timestamp) / int64(time.Millisecond)),
		Invalid:        false,
	}

	if red == nil {
		txRes.Invalid = false
	} else {
		txRes.Invalid = true
		txRes.InvalidMsg = red.ErrType.String()
	}

	return txRes, nil
}