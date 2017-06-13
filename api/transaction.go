//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"github.com/op/go-logging"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"time"
)

const (
	defaultGas              int64 = 10000
	defaustGasPrice         int64 = 10000
	leveldb_not_found_error       = "leveldb: not found"
)

var (
	kec256Hash = crypto.NewKeccak256Hash("keccak256")
)

type Transaction struct {
	namespace   string
	eh          *manager.EventHub
	tokenBucket *ratelimit.Bucket
	config      *common.Config
	log         *logging.Logger
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
	Request    *Number `json:"request"`
	Simulate   bool    `json:"simulate"`
	Opcode     int32   `json:"opcode"`
	Nonce      int64   `json:"nonce"`
	SnapshotId string  `json:"snapshotId"`
}

type TransactionResult struct {
	Version     string         `json:"version"`
	Hash        common.Hash    `json:"hash"`
	BlockNumber *BlockNumber   `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash   `json:"blockHash,omitempty"`
	TxIndex     *Number        `json:"txIndex,omitempty"`
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Amount      *Number        `json:"amount,omitempty"`
	//Gas         *Number        `json:"gas"`
	//GasPrice    *Number        `json:"gasPrice"`
	Timestamp   int64   `json:"timestamp"`
	Nonce       int64   `json:"nonce"`
	ExecuteTime *Number `json:"executeTime,omitempty"`
	Payload     string  `json:"payload,omitempty"`
	Invalid     bool    `json:"invalid,omitempty"`
	InvalidMsg  string  `json:"invalidMsg,omitempty"`
}

func NewPublicTransactionAPI(namespace string, eh *manager.EventHub, config *common.Config) *Transaction {
	log := common.GetLogger(namespace, "api")
	fillrate, err := getFillRate(namespace, config, TRANSACTION)
	if err != nil {
		log.Errorf("invalid ratelimit fill rate parameters.")
		fillrate = 10 * time.Millisecond
	}
	peak := getRateLimitPeak(namespace, config, TRANSACTION)
	if peak == 0 {
		log.Errorf("got invalid ratelimit peak parameters as 0. use default peak parameters 500")
		peak = 500
	}
	return &Transaction{
		namespace:   namespace,
		eh:          eh,
		config:      config,
		tokenBucket: ratelimit.NewBucket(fillrate, peak),
		log:         log,
	}
}

// txType 0 represents send normal tx, txType 1 represents deploy contract, txType 2 represents invoke contract, txType 3 represents signHash, txType 4 represents maintain contract.
func prepareExcute(args SendTxArgs, txType int) (SendTxArgs, error) {
	if args.Gas == nil {
		args.Gas = NewInt64ToNumber(defaultGas)
	}
	if args.GasPrice == nil {
		args.GasPrice = NewInt64ToNumber(defaustGasPrice)
	}
	if args.From.Hex() == (common.Address{}).Hex() {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "address 'from' is invalid"}
	}
	if (txType == 0 || txType == 2 || txType == 4) && args.To == nil {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "address 'to' is invalid"}
	}
	if args.Timestamp <= 0 || (5*int64(time.Minute)+time.Now().UnixNano()) < args.Timestamp {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "'timestamp' is invalid"}
	}
	if txType != 3 && args.Signature == "" {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "'signature' can't be empty"}
	}
	if args.Nonce <= 0 {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "'nonce' is invalid"}
	}
	if txType == 4 && args.Opcode == 1 && (args.Payload == "" || args.Payload == "0x") {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "contract code is empty"}
	}
	if txType == 1 && (args.Payload == "" || args.Payload == "0x") {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "contract code is empty"}
	}
	if args.SnapshotId != "" && args.Simulate != true {
		return SendTxArgs{}, &common.InvalidParamsError{Message: "can not query history ledger without `simulate` mode"}
	}
	return args, nil
}

// SendTransaction is to build a transaction object,and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash
func (tran *Transaction) SendTransaction(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(tran.config) && tran.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{Message: "system is too busy to response "}
	}
	var tx *types.Transaction

	realArgs, err := prepareExcute(args, 0)
	if err != nil {
		return common.Hash{}, err
	}

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(),
		realArgs.Value.ToInt64(), nil, 0)

	value, err := proto.Marshal(txValue)

	if err != nil {
		return common.Hash{}, &common.CallbackError{err.Error()}
	}

	//tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, common.FromHex(args.Signature))
	tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	tx.Id = uint64(tran.eh.GetPeerManager().GetNodeId())
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()

	//delete repeated tx
	var exist, _ = edb.JudgeTransactionExist(tran.namespace, tx.TransactionHash)

	if exist {
		return common.Hash{}, &common.RepeadedTxError{Message: "repeated tx"}
	}

	if args.Request != nil {

		// ** For Hyperboard **
		for i := 0; i < (*args.Request).ToInt(); i++ {
			// Unsign Test
			if !tx.ValidateSign(tran.eh.GetAccountManager().Encryption, kec256Hash) {
				tran.log.Error("invalid signature")
				// ATTENTION, return invalid transactino directly
				return common.Hash{}, &common.SignatureInvalidError{Message: "invalid signature"}
			}
			go tran.eh.GetEventObject().Post(event.NewTxEvent{
				Transaction: tx,
				Simulate:    args.Simulate,
				SnapshotId:  args.SnapshotId,
			})
		}
	} else {
		// ** For Hyperchain **
		if !tx.ValidateSign(tran.eh.GetAccountManager().Encryption, kec256Hash) {
			tran.log.Error("invalid signature")
			// ATTENTION, return invalid transactino directly
			return common.Hash{}, &common.SignatureInvalidError{Message: "invalid signature"}
		}

		go tran.eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
		})
	}
	return tx.GetHash(), nil
}

type ReceiptResult struct {
	TxHash          string        `json:"txHash"`
	PostState       string        `json:"postState"`
	ContractAddress string        `json:"contractAddress"`
	Ret             string        `json:"ret"`
	Log             []interface{} `json:"log"`
}

// GetTransactionReceipt returns transaction's receipt for given transaction hash.
func (tran *Transaction) GetTransactionReceipt(hash common.Hash) (*ReceiptResult, error) {
	if errType, err := edb.GetInvaildTxErrType(tran.namespace, hash.Bytes()); errType == -1 {
		receipt := edb.GetReceipt(tran.namespace, hash)
		if receipt == nil {
			//return nil, nil
			return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("receipt by %#x", hash)}
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
		return nil, &common.CallbackError{Message: err.Error()}
	} else {
		if errType == types.InvalidTransactionRecord_SIGFAILED {
			return nil, &common.SignatureInvalidError{Message: errType.String()}
		} else if errType == types.InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED {
			return nil, &common.ContractDeployError{Message: errType.String()}
		} else if errType == types.InvalidTransactionRecord_INVOKE_CONTRACT_FAILED {
			return nil, &common.ContractInvokeError{Message: errType.String()}
		} else if errType == types.InvalidTransactionRecord_OUTOFBALANCE {
			return nil, &common.OutofBalanceError{Message: errType.String()}
		} else if errType == types.InvalidTransactionRecord_INVALID_PERMISSION {
			return nil, &common.ContractPermissionError{Message: errType.String()}
		} else {
			return nil, &common.CallbackError{Message: errType.String()}
		}
	}

}

// GetTransactions return all transactions in the chain/db
func (tran *Transaction) GetTransactions(args IntervalArgs) ([]*TransactionResult, error) {
	trueArgs, err := prepareIntervalArgs(args, tran.namespace)
	if err != nil {
		return nil, err
	}

	var transactions []*TransactionResult

	if blocks, err := getBlocks(trueArgs, tran.namespace, false); err != nil {
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
func (tran *Transaction) GetDiscardTransactions() ([]*TransactionResult, error) {

	reds, err := edb.GetAllDiscardTransaction(tran.namespace)
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: "discard transactions"}
	} else if err != nil {
		tran.log.Errorf("GetAllDiscardTransaction error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	var transactions []*TransactionResult

	for _, red := range reds {
		if ts, err := outputTransaction(red, tran.namespace, tran.log); err != nil {
			return nil, err
		} else {
			transactions = append(transactions, ts)
		}
	}

	return transactions, nil
}

// getDiscardTransactionByHash returns the invalid transaction for the given transaction hash.
func (tran *Transaction) getDiscardTransactionByHash(hash common.Hash) (*TransactionResult, error) {

	red, err := edb.GetDiscardTransaction(tran.namespace, hash.Bytes())
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("discard transaction by %#x", hash)}
	} else if err != nil {
		tran.log.Errorf("GetDiscardTransaction error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return outputTransaction(red, tran.namespace, tran.log)
}

// GetTransactionByHash returns the transaction for the given transaction hash.
func (tran *Transaction) GetTransactionByHash(hash common.Hash) (*TransactionResult, error) {

	tx, err := edb.GetTransaction(tran.namespace, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return tran.getDiscardTransactionByHash(hash)
	} else if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return outputTransaction(tx, tran.namespace, tran.log)
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and index.
func (tran *Transaction) GetTransactionByBlockHashAndIndex(hash common.Hash, index Number) (*TransactionResult, error) {
	//return nil, errors.New("hahaha")
	if common.EmptyHash(hash) == true {
		return nil, &common.InvalidParamsError{Message: "Invalid hash"}
	}

	block, err := edb.GetBlock(tran.namespace, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= txCount {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("transaction, this block contains %v transactions, but the index %v is out of range", txCount, index)}
	}

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.namespace, tran.log)
	}

	return nil, nil
}

// GetTransactionsByBlockNumberAndIndex returns the transaction for the given block number and index.
func (tran *Transaction) GetTransactionByBlockNumberAndIndex(n BlockNumber, index Number) (*TransactionResult, error) {
	latest := edb.GetChainCopy(tran.namespace).Height
	blknumber, err := n.BlockNumberToUint64(latest)
	if err != nil {
		return nil, err
	}

	block, err := edb.GetBlockByNumber(tran.namespace, blknumber)
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %d", n)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	if index.ToInt() >= txCount {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("transaction, this block contains %v transactions, but the index %v is out of range", txCount, index)}
	}

	if index.ToInt() >= 0 && index.ToInt() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.namespace, tran.log)
	}

	return nil, nil
}

// GetTransactionsByTime returns the transactions for the given time duration.
func (tran *Transaction) GetTransactionsByTime(args IntervalTime) ([]*TransactionResult, error) {

	if args.StartTime > args.Endtime || args.StartTime < 0 || args.Endtime < 0 {
		return nil, &common.InvalidParamsError{Message: "Invalid params, both startTime and endTime must be positive, startTime is less than endTime"}
	}

	currentChain := edb.GetChainCopy(tran.namespace)
	height := currentChain.Height

	var txs = make([]*TransactionResult, 0)

	for i := height; i >= uint64(1); i-- {
		block, _ := edb.GetBlockByNumber(tran.namespace, i)
		if block.WriteTime > args.Endtime {
			continue
		}
		if block.WriteTime < args.StartTime {
			return txs, nil
		}
		if block.WriteTime >= args.StartTime && block.WriteTime <= args.Endtime {
			trans := block.GetTransactions()

			for _, t := range trans {
				tx, err := outputTransaction(t, tran.namespace, tran.log)
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
func (tran *Transaction) GetBlockTransactionCountByHash(hash common.Hash) (*Number, error) {

	if common.EmptyHash(hash) == true {
		return nil, &common.InvalidParamsError{Message: "Invalid hash"}
	}

	block, err := edb.GetBlock(tran.namespace, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	return NewIntToNumber(txCount), nil
}

// GetBlockTransactionCountByNumber returns the number of block transactions for given block number.
func (tran *Transaction) GetBlockTransactionCountByNumber(n BlockNumber) (*Number, error) {
	latest := edb.GetChainCopy(tran.namespace).Height
	blknumber, err := n.BlockNumberToUint64(latest)
	if err != nil {
		return nil, err
	}

	block, err := edb.GetBlockByNumber(tran.namespace, blknumber)
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by number %#x", n)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	return NewIntToNumber(txCount), nil
}

// GetSignHash returns the hash for client signature.
func (tran *Transaction) GetSignHash(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction

	realArgs, err := prepareExcute(args, 3) // Allow the param "to" and "signature" is empty
	if err != nil {
		return common.Hash{}, err
	}

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(realArgs.GasPrice.ToInt64(), realArgs.Gas.ToInt64(), realArgs.Value.ToInt64(), payload, args.Opcode)

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &common.CallbackError{Message: err.Error()}
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
func (tran *Transaction) GetTransactionsCount() (interface{}, error) {

	chain := edb.GetChainCopy(tran.namespace)

	return struct {
		Count     *Number `json:"count,"`
		Timestamp int64   `json:"timestamp"`
	}{
		Count:     NewUint64ToNumber(chain.CurrentTxSum),
		Timestamp: time.Now().UnixNano(),
	}, nil
}

// GetTxAvgTimeByBlockNumber returns tx execute avg time.
func (tran *Transaction) GetTxAvgTimeByBlockNumber(args IntervalArgs) (Number, error) {
	intargs, err := prepareIntervalArgs(args, tran.namespace)
	if err != nil {
		return 0, err
	}

	exeTime := edb.CalcResponseAVGTime(tran.namespace, intargs.from, intargs.to)

	if exeTime <= 0 {
		return 0, nil
	}

	return *NewInt64ToNumber(exeTime), nil
}

func (tran *Transaction) GetTransactionsCountByContractAddr(args IntervalArgs) (interface{}, error) {
	if args.ContractAddr == nil {
		return nil, &common.InvalidParamsError{"'address' can't be empty"}
	} else if args.MethodID != "" {
		return nil, &common.InvalidParamsError{"invalid params, 'methodID' is unrecognized"}
	}
	return tran.getTransactionsCountByBlockNumber(args)
}

func (tran *Transaction) GetTransactionsCountByMethodID(args IntervalArgs) (interface{}, error) {
	if args.ContractAddr == nil {
		return nil, &common.InvalidParamsError{"'address' can't be empty"}
	} else if args.MethodID == "" {
		return nil, &common.InvalidParamsError{"'methodID' can't be empty"}
	}
	return tran.getTransactionsCountByBlockNumber(args)
}

func (tran *Transaction) getTransactionsCountByBlockNumber(args IntervalArgs) (interface{}, error) {

	realArgs, err := prepareIntervalArgs(args, tran.namespace)
	if err != nil {
		return 0, err
	}

	from := realArgs.from
	txCounts := 0
	lastIndex := 0
	contractAddr := args.ContractAddr.Hex()
	var lastBlockNum uint64

	for from <= realArgs.to {

		block, err := getBlockByNumber(tran.namespace, from, false)
		if err != nil {
			return 0, err
		}

		for _, tx := range block.Transactions {
			txResult := tx.(*TransactionResult)

			to := txResult.To.Hex()
			txIndex := txResult.TxIndex.ToInt()
			blockNum, err := prepareBlockNumber(*txResult.BlockNumber, tran.namespace)

			if err != nil {
				return nil, &common.CallbackError{Message: err.Error()}
			}

			if to == contractAddr && to != "0x0000000000000000000000000000000000000000" {
				if args.MethodID != "" {
					if substr(txResult.Payload, 2, 10) == args.MethodID {
						txCounts++
						lastIndex = txIndex
						lastBlockNum = blockNum
					}
				} else {
					txCounts++
					lastIndex = txIndex
					lastBlockNum = blockNum
				}

			}

			if args.MethodID == "" && to == "0x0000000000000000000000000000000000000000" {
				if receipt, err := tran.GetTransactionReceipt(txResult.Hash); err != nil {
					return 0, err
				} else if receipt.ContractAddress == contractAddr {
					txCounts++
					lastIndex = txIndex
					lastBlockNum = blockNum
				}
			}
		}

		from++
	}

	return struct {
		Count        *Number      `json:"count,"`
		LastIndex    *Number      `json:"lastIndex"`
		LastBlockNum *BlockNumber `json:"lastBlockNum"`
	}{
		Count:        NewIntToNumber(txCounts),
		LastIndex:    NewIntToNumber(lastIndex),
		LastBlockNum: Uint64ToBlockNumber(lastBlockNum),
	}, nil

}

type PagingArgs struct {
	BlkNumber      BlockNumber     `json:"blkNumber"`
	MaxBlkNumber   BlockNumber     `json:"maxBlkNumber"`
	MinBlkNumber   BlockNumber     `json:"minBlkNumber"`
	TxIndex        Number          `json:"txIndex"`
	Separated      Number          `json:"separated"`
	PageSize       Number          `json:"pageSize"`
	ContainCurrent bool            `json:"containCurrent"`
	ContractAddr   *common.Address `json:"address"`
	MethodID       string          `json:"methodID"`
}

type pagingArgs struct {
	pageSize     int
	minBlkNumber uint64
	maxBlkNumber uint64
	contractAddr *common.Address
	methodId     string
}

func preparePagingArgs(args PagingArgs) (PagingArgs, error) {
	if args.PageSize == 0 {
		return PagingArgs{}, &common.InvalidParamsError{"'pageSize' can't be zero or empty"}
	} else if args.Separated%args.PageSize != 0 {
		return PagingArgs{}, &common.InvalidParamsError{"invalid 'pageSize' or 'separated'"}
	} else if args.BlkNumber < args.MinBlkNumber || args.BlkNumber > args.MaxBlkNumber {
		return PagingArgs{}, &common.InvalidParamsError{fmt.Sprintf("'blkNumber' is out of range, it must be in the range %d to %d", args.MinBlkNumber, args.MaxBlkNumber)}
	} else if args.MaxBlkNumber == BlockNumber(0) || args.MinBlkNumber == BlockNumber(0) {
		return PagingArgs{}, &common.InvalidParamsError{"'minBlkNumber' or 'maxBlkNumber' can't be zero or empty"}
	} else if args.ContractAddr == nil {
		return PagingArgs{}, &common.InvalidParamsError{"'address' can't be empty"}
	}

	return args, nil
}

func (tran *Transaction) GetNextPageTransactions(args PagingArgs) ([]interface{}, error) {

	realArgs, err := preparePagingArgs(args)
	if err != nil {
		return nil, err
	}

	txs := make([]interface{}, 0)

	// to comfirm start position
	blkNumber, err := prepareBlockNumber(realArgs.BlkNumber, tran.namespace) // 3
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}
	index := realArgs.TxIndex.ToInt() // 10
	separated := realArgs.Separated.ToInt()
	contractAddr := realArgs.ContractAddr
	txCounts := 0
	txCounts_temp := 0
	filteredTxs := make([]interface{}, 0)
	isFirstTx := true

	for txCounts < separated {
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}

		blockTxCount := block.TxCounts.ToInt()
		if blockTxCount <= index {
			return nil, &common.InvalidParamsError{fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		}

		// filter
		if filteredTxsByAddr, err := tran.filterTransactionsByAddress(block.Transactions[index:], contractAddr); err != nil {
			return nil, err
		} else if realArgs.MethodID != "" {

			if filteredTxs, err = tran.filterTransactionsByMethodID(filteredTxsByAddr, realArgs.MethodID); err != nil {
				return nil, err
			}

		} else {
			filteredTxs = filteredTxsByAddr
		}
		filtedTxsCount := len(filteredTxs)
		//log.Noticef("blkNumber = %d, index = %d, txCounts = %d, blockTxCount = %d, filtedTxsCount = %d",blkNumber, index, txCounts, blockTxCount,filtedTxsCount)
		if filtedTxsCount != blockTxCount {
			blockTxCount = filtedTxsCount
			index = 0
		}

		if !isFirstTx && index == 0 {
			txCounts_temp = txCounts + blockTxCount
		} else {
			// if the tx is the first tx and index is equal to 0, exclude this tx.
			txCounts_temp = txCounts + (blockTxCount - (index + 1))
		}

		if txCounts_temp < separated {
			txCounts = txCounts_temp
			blkNumber++
			index = 0
		} else if txCounts_temp == separated {
			index = blockTxCount - 1
			txCounts = txCounts_temp
		} else {
			index = separated - txCounts + index - 1
			txCounts += index + 1
		}

		isFirstTx = false
	}

	if !args.ContainCurrent {
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}

		blockTxCount := block.TxCounts.ToInt()

		if index < blockTxCount-1 {
			index++
		} else if index == blockTxCount-1 {
			blkNumber++
			index = 0
		} else {
			return nil, &common.InvalidParamsError{fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		}
	}

	//log.Noticef("当前区块号 %v: \n", blkNumber)
	//log.Noticef("当前交易索引 %v: \n", index)
	//log.Noticef("当前pageSize %v: \n", realArgs.PageSize.ToInt())
	//log.Noticef("最大区块号 %v: \n", realArgs.MaxBlkNumber.ToUint64())

	min, err := prepareBlockNumber(realArgs.MinBlkNumber, tran.namespace)
	max, err := prepareBlockNumber(realArgs.MaxBlkNumber, tran.namespace)
	if err != nil {
		return nil, &common.InvalidParamsError{Message: err.Error()}
	}

	return tran.getNextPagingTransactions(txs, blkNumber, index, pagingArgs{
		pageSize:     realArgs.PageSize.ToInt(),
		minBlkNumber: min,
		maxBlkNumber: max,
		contractAddr: realArgs.ContractAddr,
		methodId:     realArgs.MethodID,
	})
}

func (tran *Transaction) GetPrevPageTransactions(args PagingArgs) ([]interface{}, error) {

	realArgs, err := preparePagingArgs(args)
	if err != nil {
		return nil, err
	}

	txs := make([]interface{}, 0)

	// to comfirm end position
	blkNumber, err := prepareBlockNumber(realArgs.BlkNumber, tran.namespace) // 3
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}
	index := realArgs.TxIndex.ToInt() // 40
	separated := realArgs.Separated.ToInt()
	txCounts := 0
	contractAddr := realArgs.ContractAddr
	filteredTxs := make([]interface{}, 0)

	for txCounts < separated {
		if blkNumber == 0 {
			break
		}
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}
		blockTxCount := block.TxCounts.ToInt()
		if blockTxCount <= index {
			return nil, &common.InvalidParamsError{fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		}

		if index == -1 {
			index = blockTxCount
		}

		// filter
		if filteredTxsByAddr, err := tran.filterTransactionsByAddress(block.Transactions[:index], contractAddr); err != nil {
			return nil, err
		} else if realArgs.MethodID != "" {

			if filteredTxs, err = tran.filterTransactionsByMethodID(filteredTxsByAddr, realArgs.MethodID); err != nil {
				return nil, err
			}

		} else {
			filteredTxs = filteredTxsByAddr
		}
		filtedTxsCount := len(filteredTxs)
		//log.Noticef("blkNumber = %d, index = %d, txCounts = %d, blockTxCount = %d, filtedTxsCount = %d",blkNumber, index, txCounts, blockTxCount,filtedTxsCount)
		if filtedTxsCount != blockTxCount {
			index = filtedTxsCount
		}

		txCounts += index
		if txCounts < separated {
			blkNumber--
			index = -1
		} else if txCounts == separated {
			index = 0
		} else {
			index = txCounts - separated
		}
	}

	if !args.ContainCurrent {
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}

		blockTxCount := block.TxCounts.ToInt()

		if index >= blockTxCount {
			return nil, &common.InvalidParamsError{fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		} else if index == 0 {
			blkNumber--
			blk, err := getBlockByNumber(tran.namespace, blkNumber, false)
			if err != nil {
				return nil, err
			}
			index = blk.TxCounts.ToInt() - 1
		} else {
			index--
		}
	}

	//log.Noticef("当前区块号 %v: \n", blkNumber)
	//log.Noticef("当前交易索引 %v: \n", index)
	//log.Noticef("当前pageSize %v: \n", realArgs.PageSize)
	//log.Noticef("最小区块号 %v: \n", realArgs.MinBlkNumber)

	min, err := prepareBlockNumber(realArgs.MinBlkNumber, tran.namespace)
	max, err := prepareBlockNumber(realArgs.MaxBlkNumber, tran.namespace)
	if err != nil {
		return nil, &common.InvalidParamsError{Message: err.Error()}
	}

	return tran.getPrevPagingTransactions(txs, blkNumber, index, pagingArgs{
		pageSize:     realArgs.PageSize.ToInt(),
		minBlkNumber: min,
		maxBlkNumber: max,
		contractAddr: realArgs.ContractAddr,
		methodId:     realArgs.MethodID,
	})
}

func (tran *Transaction) getNextPagingTransactions(txs []interface{}, currentNumber uint64, currentIndex int, constant pagingArgs) ([]interface{}, error) {
	//log.Notice("===== enter getNextPagingTransactions =======\n")
	//log.Noticef("当前交易量 %v: \n", len(txs))
	//log.Noticef("当前区块号 %v: \n", currentNumber)
	//log.Noticef("当前交易索引 %v: \n", currentIndex)

	if len(txs) == constant.pageSize || currentNumber > constant.maxBlkNumber {
		return txs, nil
	}

	blk, err := getBlockByNumber(tran.namespace, currentNumber, false)
	if err != nil {
		return nil, err
	}

	blockTxCount := blk.TxCounts.ToInt()

	if currentIndex >= blockTxCount {
		return nil, &common.InvalidParamsError{fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", currentIndex, currentNumber, blockTxCount)}
	}

	var flag bool
	if currentIndex == 0 {
		flag = blockTxCount <= constant.pageSize-len(txs)
	} else {
		flag = blockTxCount-(currentIndex+1) <= constant.pageSize-len(txs)
	}

	if flag {

		if filteredTxByAddr, err := tran.filterTransactionsByAddress(blk.Transactions[currentIndex:], constant.contractAddr); err != nil {
			return nil, err
		} else {
			if constant.methodId != "" {
				if filteredTx, err := tran.filterTransactionsByMethodID(filteredTxByAddr, constant.methodId); err != nil {
					return nil, err
				} else {
					txs = append(txs, filteredTx...)
					currentNumber++
				}
			} else {
				txs = append(txs, filteredTxByAddr...)
				currentNumber++
			}
		}

		return tran.getNextPagingTransactions(txs, currentNumber, 0, constant)
	} else {
		index := currentIndex + constant.pageSize - len(txs)

		if filteredTxByAddr, err := tran.filterTransactionsByAddress(blk.Transactions[currentIndex:index], constant.contractAddr); err != nil {
			return nil, err
		} else {
			if constant.methodId != "" {
				if filteredTx, err := tran.filterTransactionsByMethodID(filteredTxByAddr, constant.methodId); err != nil {
					return nil, err
				} else {
					txs = append(txs, filteredTx...)
				}
			} else {
				txs = append(txs, filteredTxByAddr...)
			}
		}

		return tran.getNextPagingTransactions(txs, currentNumber, index, constant)
	}
}

func (tran *Transaction) getPrevPagingTransactions(txs []interface{}, currentNumber uint64, currentIndex int, constant pagingArgs) ([]interface{}, error) {

	//log.Notice("===== enter getPrevPagingTransactions =======\n")
	//log.Noticef("当前交易量 %v: \n", len(txs))
	//log.Noticef("当前区块号 %v: \n", currentNumber)
	//log.Noticef("当前交易索引 %v: \n", currentIndex)

	if len(txs) == constant.pageSize || currentNumber < constant.minBlkNumber || currentNumber == 0 {
		return txs, nil
	}

	blk, err := getBlockByNumber(tran.namespace, currentNumber, false)
	if err != nil {
		return nil, err
	}

	if currentIndex == -1 {
		currentIndex = blk.TxCounts.ToInt() - 1
	}

	if currentIndex+1 <= constant.pageSize-len(txs) {

		if filteredTxByAddr, err := tran.filterTransactionsByAddress(blk.Transactions[:currentIndex+1], constant.contractAddr); err != nil {
			return nil, err
		} else {
			if constant.methodId != "" {
				if filteredTx, err := tran.filterTransactionsByMethodID(filteredTxByAddr, constant.methodId); err != nil {
					return nil, err
				} else {

					txs = append(txs, filteredTx...)
					currentNumber--
				}
			} else {
				txs = append(txs, filteredTxByAddr...)
				currentNumber--
			}
		}

		return tran.getPrevPagingTransactions(txs, currentNumber, -1, constant) // -1 represent the last trasaction of block
	} else {

		index := currentIndex - (constant.pageSize - len(txs)) + 1

		if filteredTxByAddr, err := tran.filterTransactionsByAddress(blk.Transactions[index:currentIndex+1], constant.contractAddr); err != nil {
			return nil, err
		} else {
			if constant.methodId != "" {
				if filteredTx, err := tran.filterTransactionsByMethodID(filteredTxByAddr, constant.methodId); err != nil {
					return nil, err
				} else {
					txs = append(txs, filteredTx...)
				}
			} else {
				txs = append(txs, filteredTxByAddr...)
			}
		}

		return tran.getPrevPagingTransactions(txs, currentNumber, index-1, constant)
	}
}

func (tran *Transaction) filterTransactionsByMethodID(txs []interface{}, methodID string) ([]interface{}, error) {

	result := make([]interface{}, 0)
	for _, tx := range txs {
		txResult := tx.(*TransactionResult)
		if substr(txResult.Payload, 2, 10) == methodID {
			result = append(result, tx)
		}
	}
	return result, nil
}

func (tran *Transaction) filterTransactionsByAddress(txs []interface{}, address *common.Address) ([]interface{}, error) {

	result := make([]interface{}, 0)
	contractAddr := *address
	for _, tx := range txs {
		txResult := tx.(*TransactionResult)
		if txResult.To == contractAddr {
			result = append(result, tx)
		} else if txResult.To.Hex() == "0x0000000000000000000000000000000000000000" {
			if receipt, err := tran.GetTransactionReceipt(txResult.Hash); err != nil {
				return nil, err
			} else if receipt.ContractAddress == contractAddr.Hex() {
				result = append(result, tx)
			}
		}
	}
	return result, nil
}

func outputTransaction(trans interface{}, namespace string, log *logging.Logger) (*TransactionResult, error) {

	var txValue types.TransactionValue

	var txRes *TransactionResult

	switch t := trans.(type) {
	case *types.Transaction:
		if err := proto.Unmarshal(t.Value, &txValue); err != nil {
			log.Errorf("%v", err)
			return nil, &common.CallbackError{Message: err.Error()}
		}

		txHash := t.GetHash()
		bn, txIndex := edb.GetTxWithBlock(namespace, txHash[:])

		if blk, err := edb.GetBlockByNumber(namespace, bn); err == nil {
			bHash := common.BytesToHash(blk.BlockHash)
			txRes = &TransactionResult{
				Version:     string(t.Version),
				Hash:        txHash,
				BlockNumber: Uint64ToBlockNumber(bn),
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
			}
		} else if err != nil && err.Error() == leveldb_not_found_error {
			return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %d", bn)}
		} else if err != nil {
			return nil, &common.CallbackError{Message: err.Error()}
		}

	case *types.InvalidTransactionRecord:
		if err := proto.Unmarshal(t.Tx.Value, &txValue); err != nil {
			log.Errorf("%v", err)
			return nil, &common.CallbackError{Message: err.Error()}
		}
		txHash := t.Tx.GetHash()
		txRes = &TransactionResult{
			Version: string(t.Tx.Version),
			Hash:    txHash,
			From:    common.BytesToAddress(t.Tx.From),
			To:      common.BytesToAddress(t.Tx.To),
			Amount:  NewInt64ToNumber(txValue.Amount),
			Nonce:   t.Tx.Nonce,
			//Gas: 		NewInt64ToNumber(txValue.GasLimit),
			//GasPrice: 	NewInt64ToNumber(txValue.Price),
			Timestamp:  t.Tx.Timestamp,
			Payload:    common.ToHex(txValue.Payload),
			Invalid:    true,
			InvalidMsg: t.ErrType.String(),
		}
	}

	return txRes, nil
}
