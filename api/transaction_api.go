//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/event"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"time"
)

const (
	defaultGas              int64 = 10000
	defaustGasPrice         int64 = 10000
	leveldb_not_found_error       = "leveldb: not found"
)

var (
	log        *logging.Logger // package-level logger
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

func init() {
	log = logging.MustGetLogger("hpc")
}

type PublicTransactionAPI struct {
	eventMux    *event.TypeMux
	pm          *manager.ProtocolManager
	db          hyperdb.Database
	tokenBucket *ratelimit.Bucket
	config      *common.Config
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
	Signature string          `json:"signature"`
	Timestamp int64           `json:"timestamp"`
	// --- test -----
	Request   *Number     `json:"request"`
	Simulate  bool        `json:"simulate"`
	Update    bool        `json:"update"`
	Nonce     int64       `json:"nonce"`
}

type TransactionResult struct {
	Version     string         `json:"version"`
	Hash        common.Hash    `json:"hash"`
	BlockNumber *BlockNumber   `json:"blockNumber"`
	BlockHash   *common.Hash   `json:"blockHash"`
	TxIndex     *Number        `json:"txIndex"`
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Amount      *Number        `json:"amount"`
	//Gas         *Number        `json:"gas"`
	//GasPrice    *Number        `json:"gasPrice"`
	Timestamp   int64   `json:"timestamp"`
	Nonce       int64   `json:"nonce"`
	ExecuteTime *Number `json:"executeTime"`
	Payload     string  `json:"payload"`
	Invalid     bool    `json:"invalid"`
	InvalidMsg  string  `json:"invalidMsg"`
}

func NewPublicTransactionAPI(eventMux *event.TypeMux, pm *manager.ProtocolManager, hyperDb hyperdb.Database, config *common.Config) *PublicTransactionAPI {
	fillrate, err := getFillRate(config, TRANSACTION)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillrate = 10 * time.Millisecond
	}
	peak := getRateLimitPeak(config, TRANSACTION)
	if peak == 0 {
		log.Errorf("got invalid ratelimit peak parameters as 0. use default peak parameters 500")
		peak = 500
	}
	return &PublicTransactionAPI{
		eventMux:    eventMux,
		pm:          pm,
		db:          hyperDb,
		config:      config,
		tokenBucket: ratelimit.NewBucket(fillrate, peak),
	}
}

// txType 0 represents send normal tx, txType 1 represents deploy contract, txType 2 represents invoke contract, txType 3 represents signHash.
func prepareExcute(args SendTxArgs, txType int) (SendTxArgs, error) {
	if args.Gas == nil {
		args.Gas = NewInt64ToNumber(defaultGas)
	}
	if args.GasPrice == nil {
		args.GasPrice = NewInt64ToNumber(defaustGasPrice)
	}
	if args.From.Hex() == (common.Address{}).Hex() {
		return SendTxArgs{}, &invalidParamsError{"address 'from' is invalid"}
	}
	if (txType == 0 || txType == 2) && args.To == nil {
		return SendTxArgs{}, &invalidParamsError{"address 'to' is invalid"}
	}
	if args.Timestamp <= 0 || (5*int64(time.Minute)+time.Now().UnixNano()) < args.Timestamp {
		return SendTxArgs{}, &invalidParamsError{"'timestamp' is invalid"}
	}
	if txType != 3 && args.Signature == "" {
		return SendTxArgs{}, &invalidParamsError{"'signature' can't be empty"}
	}
	if args.Nonce <= 0 {
		return SendTxArgs{}, &invalidParamsError{"'nonce' is invalid"}
	}
	return args, nil
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash
func (tran *PublicTransactionAPI) SendTransaction(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(tran.config) && tran.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &systemTooBusyError{"system is too busy to response "}
	}
	var tx *types.Transaction

	realArgs, err := prepareExcute(args, 0)
	if err != nil {
		return common.Hash{}, err
	}

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), nil, false)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, &callbackError{err.Error()}
	}

	//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, common.FromHex(args.Signature))
	tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	tx.Id = uint64(tran.pm.Peermanager.GetNodeId())
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.BuildHash().Bytes()

	//delete repeated tx
	var exist, _ = core.JudgeTransactionExist(tran.db, tx.TransactionHash)

	if exist {
		return common.Hash{}, &repeadedTxError{"repeated tx"}
	}

	if args.Request != nil {

		// ** For Hyperboard **
		for i := 0; i < (*args.Request).ToInt(); i++ {
			// Unsign Test
			if !tx.ValidateSign(tran.pm.AccountManager.Encryption, kec256Hash) {
				log.Error("invalid signature")
				// ATTENTION, return invalid transactino directly
				return common.Hash{}, &signatureInvalidError{"invalid signature"}
			}

			if txBytes, err := proto.Marshal(tx); err != nil {
				log.Errorf("proto.Marshal(tx) error: %v", err)
				return common.Hash{}, &callbackError{"proto.Marshal(tx) happened error"}
			} else if manager.GetEventObject() != nil {
				go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes, Simulate: args.Simulate})
			} else {
				log.Error("manager is Nil")
				return common.Hash{}, &callbackError{"EventObject is nil"}
			}
		}
	} else {
		// ** For Hyperchain **
		if !tx.ValidateSign(tran.pm.AccountManager.Encryption, kec256Hash) {
			log.Error("invalid signature")
			// ATTENTION, return invalid transactino directly
			return common.Hash{}, &signatureInvalidError{"invalid signature"}
		}

		if txBytes, err := proto.Marshal(tx); err != nil {
			log.Errorf("proto.Marshal(tx) error: %v", err)
			return common.Hash{}, &callbackError{"proto.Marshal(tx) happened error"}
		} else if manager.GetEventObject() != nil {
			go tran.eventMux.Post(event.NewTxEvent{Payload: txBytes, Simulate: args.Simulate})
		} else {
			log.Error("manager is Nil")
			return common.Hash{}, &callbackError{"EventObject is nil"}
		}
	}
	return tx.GetTransactionHash(), nil
}

type ReceiptResult struct {
	TxHash          string        `json:"txHash"`
	PostState       string        `json:"postState"`
	ContractAddress string        `json:"contractAddress"`
	Ret             string        `json:"ret"`
	Log             []interface{} `json:"log"`
}

// GetTransactionReceipt returns transaction's receipt for given transaction hash.
func (tran *PublicTransactionAPI) GetTransactionReceipt(hash common.Hash) (*ReceiptResult, error) {
	if errType, err := core.GetInvaildTxErrType(tran.db, hash.Bytes()); errType == -1 {
		receipt := core.GetReceipt(hash)
		if receipt == nil {
			//return nil, nil
			return nil, &leveldbNotFoundError{fmt.Sprintf("receipt by %#x", hash)}
		}
		logs := make([]interface{}, len(receipt.Logs))
		for idx := range receipt.Logs {
			logs[idx] = receipt.Logs[idx]
		}
		return &ReceiptResult{
			TxHash:          receipt.TxHash,
			PostState:       receipt.PostState,
			ContractAddress: receipt.ContractAddress,
			Ret:             receipt.Ret,
			Log:             logs,
		}, nil
	} else if err != nil {
		return nil, err
	} else {
		if errType == types.InvalidTransactionRecord_SIGFAILED {
			return nil, &signatureInvalidError{errType.String()}
		} else if errType == types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED {
			return nil, &contractDeployError{errType.String()}
		} else if errType == types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED {
			return nil, &contractInvokeError{errType.String()}
		} else if errType == types.InvalidTransactionRecord_OUTOFBALANCE {
			return nil, &outofBalanceError{errType.String()}
		} else {
			return nil, &callbackError{errType.String()}
		}
	}

}

// GetTransactions return all transactions in the chain/db
func (tran *PublicTransactionAPI) GetTransactions(args IntervalArgs) ([]*TransactionResult, error) {

	var transactions []*TransactionResult

	if blocks, err := getBlocks(args, tran.db); err != nil {
		return nil, err
	} else {
		for _, block := range blocks {
			txs := block.Transactions

			for _, t := range txs {
				tx, _ := t.(*TransactionResult)
				transactions = append(transactions, tx)
			}
		}
	}

	return transactions, nil
}

// GetDiscardTransactions returns all invalid transaction that dont be saved on the blockchain.
func (tran *PublicTransactionAPI) GetDiscardTransactions() ([]*TransactionResult, error) {

	reds, err := core.GetAllDiscardTransaction(tran.db)
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &leveldbNotFoundError{"discard transactions"}
	} else if err != nil {
		log.Errorf("GetAllDiscardTransaction error: %v", err)
		return nil, &callbackError{err.Error()}
	}

	var transactions []*TransactionResult

	for _, red := range reds {
		if ts, err := outputTransaction(red, tran.db); err != nil {
			return nil, err
		} else {
			transactions = append(transactions, ts)
		}
	}

	return transactions, nil
}

// getDiscardTransactionByHash returns the invalid transaction for the given transaction hash.
func (tran *PublicTransactionAPI) getDiscardTransactionByHash(hash common.Hash) (*TransactionResult, error) {

	red, err := core.GetDiscardTransaction(tran.db, hash.Bytes())
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &leveldbNotFoundError{fmt.Sprintf("discard transaction by %#x", hash)}
	} else if err != nil {
		log.Errorf("GetDiscardTransaction error: %v", err)
		return nil, &callbackError{err.Error()}
	}

	return outputTransaction(red, tran.db)
}

// GetTransactionByHash returns the transaction for the given transaction hash.
func (tran *PublicTransactionAPI) GetTransactionByHash(hash common.Hash) (*TransactionResult, error) {

	tx, err := core.GetTransaction(tran.db, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return tran.getDiscardTransactionByHash(hash)
	} else if err != nil {
		return nil, &callbackError{err.Error()}
	}

	return outputTransaction(tx, tran.db)
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (tran *PublicTransactionAPI) GetTransactionByBlockHashAndIndex(hash common.Hash, index Number) (*TransactionResult, error) {

	if common.EmptyHash(hash) == true {
		return nil, &invalidParamsError{"Invalid hash"}
	}

	block, err := core.GetBlock(tran.db, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &leveldbNotFoundError{fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		log.Errorf("%v", err)
		return nil, &callbackError{err.Error()}
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.db)
	}

	return nil, nil
}

// GetTransactionsByBlockNumberAndIndex returns the transaction for the given block number and index.
func (tran *PublicTransactionAPI) GetTransactionByBlockNumberAndIndex(n BlockNumber, index Number) (*TransactionResult, error) {

	block, err := core.GetBlockByNumber(tran.db, n.ToUint64())
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &leveldbNotFoundError{fmt.Sprintf("block by %d", n)}
	} else if err != nil {
		log.Errorf("%v", err)
		return nil, &callbackError{err.Error()}
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.db)
	}

	return nil, nil
}

// GetTransactionsByTime returns the transactions for the given time duration.
func (tran *PublicTransactionAPI) GetTransactionsByTime(args IntervalTime) ([]*TransactionResult, error) {

	if args.StartTime > args.Endtime {
		return nil, &invalidParamsError{"invalid params"}
	}

	currentChain := core.GetChainCopy()
	height := currentChain.Height

	var txs = make([]*TransactionResult, 0)

	for i := height; i >= uint64(1); i-- {
		block, _ := core.GetBlockByNumber(tran.db, i)
		if block.WriteTime > args.Endtime {
			continue
		}
		if block.WriteTime < args.StartTime {
			return txs, nil
		}
		if block.WriteTime >= args.StartTime && block.WriteTime <= args.Endtime {
			trans := block.GetTransactions()
			log.Error(len(trans))
			for _, t := range trans {
				tx, err := outputTransaction(t, tran.db)
				if err != nil {
					return nil, err
				}
				txs = append(txs, tx)
			}
		}
	}
	return txs, nil
}

// GetBlockTransactionCountByHash returns the number of block transactions for given block hash.
func (tran *PublicTransactionAPI) GetBlockTransactionCountByHash(hash common.Hash) (*Number, error) {

	if common.EmptyHash(hash) == true {
		return nil, &invalidParamsError{"Invalid hash"}
	}

	block, err := core.GetBlock(tran.db, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &leveldbNotFoundError{fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		log.Errorf("%v", err)
		return nil, &callbackError{err.Error()}
	}

	txCount := len(block.Transactions)

	return NewIntToNumber(txCount), nil
}

// GetSignHash returns the hash for client signature.
func (tran *PublicTransactionAPI) GetSignHash(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction

	realArgs, err := prepareExcute(args, 3) // Allow the param "to" and "signature" is empty
	if err != nil {
		return common.Hash{}, err
	}

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), payload, args.Update)

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &callbackError{err.Error()}
	}

	if args.To == nil {
		// deploy contract
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)

	} else {

		// contract invocation or normal transfer transaction
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}

	return tx.SighHash(kec256Hash), nil
}

// GetTransactionsCount returns the number of transaction in hyperchain.
func (tran *PublicTransactionAPI) GetTransactionsCount() (interface{}, error) {

	chain := core.GetChainCopy()

	return struct {
		Count     *Number `json:"count,"`
		Timestamp int64   `json:"timestamp"`
	}{
		Count:     NewUint64ToNumber(chain.CurrentTxSum),
		Timestamp: time.Now().UnixNano(),
	}, nil
}

// GetTxAvgTimeByBlockNumber returns tx execute avg time.
func (tran *PublicTransactionAPI) GetTxAvgTimeByBlockNumber(args IntervalArgs) (Number, error) {

	realArgs, err := prepareIntervalArgs(args)
	if err != nil {
		return 0, err
	}

	from := realArgs.From.ToUint64()
	to := realArgs.To.ToUint64()

	exeTime := core.CalcResponseAVGTime(from, to)

	if exeTime <= 0 {
		return 0, nil
	}

	return *NewInt64ToNumber(exeTime), nil
}

func outputTransaction(trans interface{}, db hyperdb.Database) (*TransactionResult, error) {

	var txValue types.TransactionValue

	var txRes *TransactionResult

	switch t := trans.(type) {
	case *types.Transaction:
		if err := proto.Unmarshal(t.Value, &txValue); err != nil {
			log.Errorf("%v", err)
			return nil, &callbackError{err.Error()}
		}

		txHash := t.GetTransactionHash()
		bn, txIndex := core.GetTxWithBlock(db, txHash[:])

		if blk, err := core.GetBlockByNumber(db, bn); err == nil {
			bHash := common.BytesToHash(blk.BlockHash)
			txRes = &TransactionResult{
				Version:     string(t.Version),
				Hash:        txHash,
				BlockNumber: NewUint64ToBlockNumber(bn),
				BlockHash:   &bHash,
				TxIndex:     NewInt64ToNumber(txIndex),
				From:        common.BytesToAddress(t.From),
				To:          common.BytesToAddress(t.To),
				Amount:      NewInt64ToNumber(txValue.Amount),
				Nonce:       t.Nonce,
				//Gas: 		NewInt64ToNumber(txValue.GasLimit),
				//GasPrice: 	NewInt64ToNumber(txValue.Price),
				Timestamp:   t.Timestamp,
				ExecuteTime: NewInt64ToNumber((blk.WriteTime - blk.Timestamp) / int64(time.Millisecond)),
				Payload:     common.ToHex(txValue.Payload),
				Invalid:     false,
			}
		} else if err != nil && err.Error() == leveldb_not_found_error {
			return nil, &leveldbNotFoundError{fmt.Sprintf("block by %d", bn)}
		} else if err != nil {
			return nil, &callbackError{err.Error()}
		}

	case *types.InvalidTransactionRecord:
		if err := proto.Unmarshal(t.Tx.Value, &txValue); err != nil {
			log.Errorf("%v", err)
			return nil, &callbackError{err.Error()}
		}
		txHash := t.Tx.GetTransactionHash()
		txRes = &TransactionResult{
			Version:     string(t.Tx.Version),
			Hash:        txHash,
			BlockNumber: nil,
			BlockHash:   nil,
			TxIndex:     nil,
			From:        common.BytesToAddress(t.Tx.From),
			To:          common.BytesToAddress(t.Tx.To),
			Amount:      NewInt64ToNumber(txValue.Amount),
			Nonce:       t.Tx.Nonce,
			//Gas: 		NewInt64ToNumber(txValue.GasLimit),
			//GasPrice: 	NewInt64ToNumber(txValue.Price),
			Timestamp:   t.Tx.Timestamp,
			ExecuteTime: nil,
			Payload:     common.ToHex(txValue.Payload),
			Invalid:     true,
			InvalidMsg:  t.ErrType.String(),
		}
	}

	return txRes, nil
}
