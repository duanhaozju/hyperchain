//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/juju/ratelimit"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/ledger/bloom"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb/db"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"strings"
	"time"
)

// This file implements the handler of Transaction service API which
// can be invoked by client in JSON-RPC request.

var (
	kec256Hash         = crypto.NewKeccak256Hash("keccak256")
	db_not_found_error = db.DB_NOT_FOUND.Error()
)

type Transaction struct {
	namespace   string
	eh          *manager.EventHub
	tokenBucket *ratelimit.Bucket
	config      *common.Config
	log         *logging.Logger
}

// SendTxArgs represents the arguments to submit a new transaction into the transaction pool.
type SendTxArgs struct {
	From      common.Address  `json:"from"`      // transaction sender address
	To        *common.Address `json:"to"`        // transaction receiver address
	Value     Number          `json:"value"`     // transaction amount
	Payload   string          `json:"payload"`   // contract payload
	Signature string          `json:"signature"` // signature of sender for the transaction
	Timestamp int64           `json:"timestamp"` // timestamp of the transaction happened
	Simulate  bool            `json:"simulate"`  // Simulate determines if the transaction requires consensus, if true, no consensus.
	Nonce     int64           `json:"nonce"`     // 16-bit random decimal number, for example 5956491387995926
	Extra     string          `json:"extra"`     // extra data stored in transaction
	VmType    string          `json:"type"`      // specify which engine executes contract

	// 1 value for Opcode means upgrading contract, 2 means freezing contract,
	// 3 means unfreezing contract, 4 means vm skipping, 100 means archiving data.
	Opcode int32 `json:"opcode"`

	// Snapshot saves the state of ledger at a moment.
	// SnapshotId specifies the based ledger when client sends transaction or invokes contract with Simulate=true.
	SnapshotId string `json:"snapshotId"`
}

type TransactionResult struct {
	Version     string         `json:"version"`               // hyperchain version when the transaction is executed
	Hash        common.Hash    `json:"hash"`                  // transaction hash
	BlockNumber *BlockNumber   `json:"blockNumber,omitempty"` // block number that transaction belongs to
	BlockHash   *common.Hash   `json:"blockHash,omitempty"`   // block hash that transaction belongs to
	TxIndex     *Number        `json:"txIndex,omitempty"`     // the index of transaction in the block
	From        common.Address `json:"from"`                  // transaction sender
	To          common.Address `json:"to"`                    // transaction receiver
	Amount      *Number        `json:"amount,omitempty"`      // the amount of transaction
	Timestamp   int64          `json:"timestamp"`
	Nonce       int64          `json:"nonce"`
	Extra       string         `json:"extra"`
	ExecuteTime *Number        `json:"executeTime,omitempty"` // the time it takes to execute the transaction
	Payload     string         `json:"payload,omitempty"`
	Invalid     bool           `json:"invalid,omitempty"`    // indicate whether it is invalid or not
	InvalidMsg  string         `json:"invalidMsg,omitempty"` // if Invalid is true, printing invalid message
}

// NewPublicTransactionAPI creates and returns a new Transaction instance for given namespace name.
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

// SendTransaction is to create a transaction object, and then post event NewTxEvent,
// if the sender's balance is not enough, account transfer will fail.
func (tran *Transaction) SendTransaction(args SendTxArgs) (common.Hash, error) {
	consentor := tran.eh.GetConsentor()
	normal, full := consentor.GetStatus()
	if !normal || full {
		return common.Hash{}, &common.SystemTooBusyError{}
	}

	if getRateLimitEnable(tran.config) && tran.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{}
	}

	// 1. create a new transaction instance
	tx, err := prepareTransaction(args, 0, tran.namespace, tran.eh)
	if err != nil {
		return common.Hash{}, err
	}

	// 2. post a event.NewTxEvent event
	err = postNewTxEvent(args, tx, tran.eh)
	if err != nil {
		return common.Hash{}, err
	}

	return tx.GetHash(), nil
}

type ReceiptResult struct {
	Version         string        `json:"version"`
	TxHash          string        `json:"txHash"`
	VmType          string        `json:"vmType"`
	ContractAddress string        `json:"contractAddress"`
	Ret             string        `json:"ret"`
	Log             []interface{} `json:"log"`
}

// GetTransactionReceipt returns transaction's receipt for given transaction hash.
func (tran *Transaction) GetTransactionReceipt(hash common.Hash) (*ReceiptResult, error) {
	if errType, err := edb.GetInvaildTxErrType(tran.namespace, hash.Bytes()); errType == -1 {
		receipt := edb.GetReceipt(tran.namespace, hash)
		if receipt == nil {
			return nil, &common.DBNotFoundError{Type: RECEIPT, Id: hash.Hex()}
		}
		logs := make([]interface{}, len(receipt.Logs))
		for idx := range receipt.Logs {
			logs[idx] = receipt.Logs[idx]
		}
		return &ReceiptResult{
			Version:         receipt.Version,
			TxHash:          receipt.TxHash,
			VmType:          receipt.VmType,
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

// GetTransactions returns all transactions in the given block number. Parameters include
// start block number and end block number.
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

// GetDiscardTransactions returns all invalid transactions that don't be saved in the ledger.
func (tran *Transaction) GetDiscardTransactions() ([]*TransactionResult, error) {

	reds, err := edb.GetAllDiscardTransaction(tran.namespace)
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: DISCARDTXS}
	} else if err != nil {
		tran.log.Errorf("GetAllDiscardTransaction error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	} else if len(reds) == 0 {
		return nil, &common.DBNotFoundError{Type: DISCARDTXS}
	}

	var transactions []*TransactionResult

	for _, red := range reds {
		if ts, err := outputTransaction(red, tran.namespace); err != nil {
			return nil, err
		} else {
			transactions = append(transactions, ts)
		}
	}

	return transactions, nil
}

// GetDiscardTransactionsByTime returns the invalid transactions in the given time duration.
func (tran *Transaction) GetDiscardTransactionsByTime(args IntervalTime) ([]*TransactionResult, error) {

	if args.StartTime > args.Endtime || args.StartTime < 0 || args.Endtime < 0 {
		return nil, &common.InvalidParamsError{Message: "Invalid params, both startTime and endTime must be positive, startTime must be less than endTime"}
	}

	reds, err := edb.GetAllDiscardTransaction(tran.namespace)
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: DISCARDTXS}
	} else if err != nil {
		tran.log.Errorf("GetDiscardTransactionsByTime error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	} else if len(reds) == 0 {
		return nil, &common.DBNotFoundError{Type: DISCARDTXS}
	}

	var transactions []*TransactionResult

	for _, red := range reds {
		if red.Tx.Timestamp <= args.Endtime && red.Tx.Timestamp >= args.StartTime {
			if ts, err := outputTransaction(red, tran.namespace); err != nil {
				return nil, err
			} else {
				transactions = append(transactions, ts)
			}
		}
	}

	return transactions, nil
}

// getDiscardTransactionByHash returns the invalid transaction for the given transaction hash.
func (tran *Transaction) getDiscardTransactionByHash(hash common.Hash) (*TransactionResult, error) {

	red, err := edb.GetDiscardTransaction(tran.namespace, hash.Bytes())
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: TRANSACTION, Id: hash.Hex()}
	} else if err != nil {
		tran.log.Errorf("GetDiscardTransaction error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return outputTransaction(red, tran.namespace)
}

// GetTransactionByHash returns the transaction in the ledger for the given transaction hash.
func (tran *Transaction) GetTransactionByHash(hash common.Hash) (*TransactionResult, error) {
	tx, err := edb.GetTransaction(tran.namespace, hash[:])
	if err != nil && err == edb.ErrNotFindTxMeta {
		return tran.getDiscardTransactionByHash(hash)
	} else if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return outputTransaction(tx, tran.namespace)
}

// GetTransactionByBlockHashAndIndex returns the transaction for the given block hash and transaction index.
func (tran *Transaction) GetTransactionByBlockHashAndIndex(hash common.Hash, index Number) (*TransactionResult, error) {
	if common.EmptyHash(hash) == true {
		return nil, &common.InvalidParamsError{Message: "Invalid hash"}
	}

	block, err := edb.GetBlock(tran.namespace, hash[:])
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: BLOCK, Id: hash.Hex()}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	if index.Int() >= txCount {
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("Invalid params. This block contains %v transactions, but the index %v is out of range", txCount, index)}
	}

	if index.Int() >= 0 && index.Int() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.namespace)
	}

	return nil, nil
}

// GetTransactionsByBlockNumberAndIndex returns the transaction for the given block number and transaction index.
func (tran *Transaction) GetTransactionByBlockNumberAndIndex(n BlockNumber, index Number) (*TransactionResult, error) {
	chain, err := edb.GetChain(tran.namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	latest := chain.Height
	blknumber, err := n.BlockNumberToUint64(latest)
	if err != nil {
		return nil, &common.InvalidParamsError{Message: err.Error()}
	}

	block, err := edb.GetBlockByNumber(tran.namespace, blknumber)
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: BLOCK, Id: fmt.Sprintf("%#x", blknumber)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	if index.Int() >= txCount {
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("Invalid params. This block contains %v transactions, but the index %v is out of range", txCount, index)}
	}

	if index.Int() >= 0 && index.Int() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.namespace)
	}

	return nil, nil
}

// GetTransactionsByTime returns the transactions in the ledger in the given time duration.
func (tran *Transaction) GetTransactionsByTime(args IntervalTime) ([]*TransactionResult, error) {

	if args.StartTime > args.Endtime || args.StartTime < 0 || args.Endtime < 0 {
		return nil, &common.InvalidParamsError{Message: "Invalid params, both startTime and endTime must be positive, startTime is less than endTime"}
	}

	currentChain, err := edb.GetChain(tran.namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

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
				tx, err := outputTransaction(t, tran.namespace)
				if err != nil {
					return nil, err
				}
				txs = append(txs, tx)
			}
		}
	}

	if len(txs) == 0 {
		return nil, &common.DBNotFoundError{Type: TRANSACTIONS}
	}

	return txs, nil
}

// GetBlockTransactionCountByHash returns the number of block transactions for given block hash.
func (tran *Transaction) GetBlockTransactionCountByHash(hash common.Hash) (*Number, error) {

	if common.EmptyHash(hash) == true {
		return nil, &common.InvalidParamsError{Message: "Invalid hash"}
	}

	block, err := edb.GetBlock(tran.namespace, hash[:])
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: BLOCK, Id: hash.Hex()}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	return intToNumber(txCount), nil
}

// GetBlockTransactionCountByNumber returns the number of block transactions for given block number.
func (tran *Transaction) GetBlockTransactionCountByNumber(n BlockNumber) (*Number, error) {
	chain, err := edb.GetChain(tran.namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	latest := chain.Height
	blknumber, err := n.BlockNumberToUint64(latest)
	if err != nil {
		return nil, &common.InvalidParamsError{Message: err.Error()}
	}

	block, err := edb.GetBlockByNumber(tran.namespace, blknumber)
	if err != nil && err.Error() == db_not_found_error {
		return nil, &common.DBNotFoundError{Type: BLOCK, Id: fmt.Sprintf("number %#x", n)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	return intToNumber(txCount), nil
}

// GetSignHash returns transaction content hash used for the client signature.
func (tran *Transaction) GetSignHash(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction

	realArgs, err := prepareExcute(args, 3) // empty contract address and empty transaction signature
	if err != nil {
		return common.Hash{}, err
	}

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS, realArgs.Value.Int64(), payload, args.Opcode, []byte(args.Extra), types.TransactionValue_EVM)

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &common.CallbackError{Message: err.Error()}
	}

	if args.To == nil {
		// deploy contract
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)

	} else {
		// invoke contract or send transaction
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}

	return tx.SignHash(kec256Hash), nil
}

// GetTransactionsCount returns the number of transaction in ledger.
func (tran *Transaction) GetTransactionsCount() (interface{}, error) {

	chain, err := edb.GetChain(tran.namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return struct {
		Count     *Number `json:"count,"`
		Timestamp int64   `json:"timestamp"`
	}{
		Count:     uint64ToNumber(chain.CurrentTxSum),
		Timestamp: time.Now().UnixNano(),
	}, nil
}

// GetTxAvgTimeByBlockNumber calculates the average execution time of all transactions in the given block number.
// Parameters include start block number and end block number.
func (tran *Transaction) GetTxAvgTimeByBlockNumber(args IntervalArgs) (Number, error) {
	intargs, err := prepareIntervalArgs(args, tran.namespace)
	if err != nil {
		return 0, err
	}

	exeTime := edb.CalcResponseAVGTime(tran.namespace, intargs.from, intargs.to)

	if exeTime <= 0 {
		return 0, nil
	}

	return *int64ToNumber(exeTime), nil
}

// GetTransactionsCountByContractAddr returns the number of eligible transaction, the latest block number and
// index of eligible last transaction in the latest block. Parameters include start block number, end block
// number and contract address.
func (tran *Transaction) GetTransactionsCountByContractAddr(args IntervalArgs) (interface{}, error) {
	if args.ContractAddr == nil {
		return nil, &common.InvalidParamsError{"Invalid params. 'address' can't be empty"}
	} else if args.MethodID != "" {
		return nil, &common.InvalidParamsError{"Invalid params. 'methodID' is unrecognized"}
	}
	return tran.getTransactionsCountByBlockNumber(args)
}

// GetTransactionsCountByMethodID returns the number of eligible transaction, the latest block number and
// index of eligible last transaction in the latest block. Parameters include start block number, end block
// number and method ID. Method ID is contract method identifier in a contract.
func (tran *Transaction) GetTransactionsCountByMethodID(args IntervalArgs) (interface{}, error) {
	if args.ContractAddr == nil {
		return nil, &common.InvalidParamsError{"Invalid params. 'address' can't be empty"}
	} else if args.MethodID == "" {
		return nil, &common.InvalidParamsError{"Invalid params. 'methodID' can't be empty"}
	}
	mid := strings.TrimSpace(args.MethodID)
	if strings.HasPrefix(mid, "0x") {
		args.MethodID = substr(mid, 2, len(mid))
	}
	return tran.getTransactionsCountByBlockNumber(args)
}

func (tran *Transaction) getTransactionsCountByBlockNumber(args IntervalArgs) (interface{}, error) {

	realArgs, err := prepareIntervalArgs(args, tran.namespace)
	if err != nil {
		return nil, err
	}

	from := realArgs.from
	txCounts := 0
	lastIndex := 0
	contractAddr := args.ContractAddr
	var lastBlockNum uint64

	for from <= realArgs.to {

		block, err := getBlockByNumber(tran.namespace, from, false)
		if err != nil {
			return 0, err
		}

		for _, tx := range block.Transactions {
			txResult := tx.(*TransactionResult)

			to := txResult.To
			txIndex := txResult.TxIndex.Int()
			blockNum, err := prepareBlockNumber(*txResult.BlockNumber, tran.namespace)

			if err != nil {
				return nil, &common.CallbackError{Message: err.Error()}
			}

			if to == *contractAddr && !to.IsZero() {
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

			if args.MethodID == "" && to.IsZero() {
				if receipt, err := tran.GetTransactionReceipt(txResult.Hash); err != nil {
					return 0, err
				} else if receipt.ContractAddress == contractAddr.Hex() {
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
		Count:        intToNumber(txCounts),
		LastIndex:    intToNumber(lastIndex),
		LastBlockNum: uint64ToBlockNumber(lastBlockNum),
	}, nil

}

// PagingArgs specifies filter conditions for transaction paging. PagingArgs determines starting
// position including current block number and index of the transaction in the current block,
// quantity of returned transactions and filter conditions.
//
// For example, pageSize is 10. From page 1 to page 2, so "separated" value is 0. From page 1 to page 3,
// "separated" value is 10.
type PagingArgs struct {
	MaxBlkNumber BlockNumber `json:"maxBlkNumber"` // the maximum block number of allowing to query
	MinBlkNumber BlockNumber `json:"minBlkNumber"` // the minimum block number of allowing to query
	BlkNumber    BlockNumber `json:"blkNumber"`    // the current block number
	TxIndex      Number      `json:"txIndex"`      // index of the transaction in the current block
	Separated    Number      `json:"separated"`    // specify how many transactions to skip.
	PageSize     Number      `json:"pageSize"`     // specify the number of transaction returned

	// specify if the returned transactions contain current transaction(BlkNumber and TxIndex)
	// or if contain current transaction(BlkNumber and TxIndex) when calculating the number of transactions.
	ContainCurrent bool            `json:"containCurrent"`
	ContractAddr   *common.Address `json:"address"`  // specify which contract transactions belong to
	MethodID       string          `json:"methodID"` // specify which contract method transactions belong to
}

type pagingArgs struct {
	pageSize     int
	minBlkNumber uint64
	maxBlkNumber uint64
	contractAddr *common.Address
	methodId     string
}

// GetNextPageTransactions returns next page data.
func (tran *Transaction) GetNextPageTransactions(args PagingArgs) ([]interface{}, error) {

	// check paging parameters
	realArgs, err := preparePagingArgs(args)
	if err != nil {
		return nil, err
	}

	blkNumber, err := prepareBlockNumber(realArgs.BlkNumber, tran.namespace)
	if err != nil {
		return nil, err
	}
	min, err := prepareBlockNumber(realArgs.MinBlkNumber, tran.namespace)
	if err != nil {
		return nil, err
	}
	max, err := prepareBlockNumber(realArgs.MaxBlkNumber, tran.namespace)
	if err != nil {
		return nil, err
	}

	if blkNumber < min || blkNumber > max {
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("Invalid params. 'blkNumber' %v is out of range, it must be in the range %v to %v", blkNumber, min, max)}
	}

	// find the starting position
	txs := make([]interface{}, 0)
	index := realArgs.TxIndex.Int()
	separated := realArgs.Separated.Int()
	contractAddr := realArgs.ContractAddr
	txCounts := 0 // how many transactions have been skipped
	txCounts_temp := 0
	filteredBlkTxs := make([]interface{}, 0)

	if !args.ContainCurrent {

		// skip current transaction, to choose the next transaction as the first transaction.
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}

		blockTxCount := block.TxCounts.Int()

		if index < blockTxCount-1 {
			index++
		} else if index == blockTxCount-1 {
			blkNumber++
			index = 0
		} else {
			return nil, &common.InvalidParamsError{Message: fmt.Sprintf("Invalid params. 'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		}
	}

	// if separated value is not equal to 0, reset starting position
	for txCounts < separated {
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}

		blkTxsCount := block.TxCounts.Int()
		if blkTxsCount <= index {
			return nil, &common.InvalidParamsError{Message: fmt.Sprintf("Invalid params. 'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blkTxsCount)}
		}

		// starting with the specified index of the block transactions, filter all the eligible transaction
		if filteredBlkTxsByAddr, err := tran.filterTransactionsByAddress(block.Transactions[index:], contractAddr); err != nil {
			return nil, err
		} else if realArgs.MethodID != "" {

			if filteredBlkTxs, err = tran.filterTransactionsByMethodID(filteredBlkTxsByAddr, realArgs.MethodID); err != nil {
				return nil, err
			}

		} else {
			filteredBlkTxs = filteredBlkTxsByAddr
		}

		txCounts_temp = txCounts + len(filteredBlkTxs)

		if txCounts_temp <= separated {

			// all transactions in the current block should be skipped
			txCounts = txCounts_temp
			blkNumber++
			index = 0
		} else {

			// part of transactions in the current block should be skipped
			tx := filteredBlkTxs[separated-txCounts].(*TransactionResult)
			index = tx.TxIndex.Int()
			txCounts = separated
		}

	}

	tran.log.Debugf("current transaction index = %v", index)
	tran.log.Debugf("     current block number = %v", blkNumber)
	tran.log.Debugf("     minimum block number = %v", min)
	tran.log.Debugf("     maximum block number = %v", max)
	tran.log.Debugf("				 pageSize = %v", realArgs.PageSize.Int())

	return tran.getNextPagingTransactions(txs, blkNumber, index, pagingArgs{
		pageSize:     realArgs.PageSize.Int(),
		minBlkNumber: min,
		maxBlkNumber: max,
		contractAddr: realArgs.ContractAddr,
		methodId:     realArgs.MethodID,
	})
}

// GetNextPageTransactions returns previous page data.
func (tran *Transaction) GetPrevPageTransactions(args PagingArgs) ([]interface{}, error) {

	// check paging parameters
	realArgs, err := preparePagingArgs(args)
	if err != nil {
		return nil, err
	}

	blkNumber, err := prepareBlockNumber(realArgs.BlkNumber, tran.namespace)
	if err != nil {
		return nil, err
	}
	min, err := prepareBlockNumber(realArgs.MinBlkNumber, tran.namespace)
	if err != nil {
		return nil, err
	}
	max, err := prepareBlockNumber(realArgs.MaxBlkNumber, tran.namespace)
	if err != nil {
		return nil, err
	}

	if blkNumber < min || blkNumber > max {
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("Invalid params. 'blkNumber' %v is out of range, it must be in the range %v to %v", blkNumber, min, max)}
	}

	// find the starting position
	txs := make([]interface{}, 0)
	index := realArgs.TxIndex.Int()
	separated := realArgs.Separated.Int()
	txCounts := 0
	txCounts_temp := 0
	contractAddr := realArgs.ContractAddr
	filteredBlkTxs := make([]interface{}, 0)

	if !args.ContainCurrent {

		// skip current transaction, to choose the last transaction as the first transaction.
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}

		blockTxCount := block.TxCounts.Int()

		if index >= blockTxCount {
			return nil, &common.InvalidParamsError{fmt.Sprintf("Invalid params. 'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		} else if index == 0 {
			blkNumber--
			blk, err := getBlockByNumber(tran.namespace, blkNumber, false)
			if err != nil {
				return nil, err
			}
			index = blk.TxCounts.Int() - 1
		} else {
			index--
		}
	}

	// if separated value is not equal to 0, reset the starting position
	for txCounts <= separated && separated != 0 {
		if blkNumber == 0 || txCounts == separated {
			break
		}
		block, err := getBlockByNumber(tran.namespace, blkNumber, false)
		if err != nil {
			return nil, err
		}
		blkTxsCount := block.TxCounts.Int()
		if blkTxsCount <= index {
			return nil, &common.InvalidParamsError{fmt.Sprintf("Invalid params. 'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blkTxsCount)}
		}

		if index == -1 {
			index = blkTxsCount - 1
		}

		// starting with the specified index of the block transactions, filter all the eligible transaction
		if filteredTxsByAddr, err := tran.filterTransactionsByAddress(block.Transactions[:index+1], contractAddr); err != nil {
			return nil, err
		} else if realArgs.MethodID != "" {

			if filteredBlkTxs, err = tran.filterTransactionsByMethodID(filteredTxsByAddr, realArgs.MethodID); err != nil {
				return nil, err
			}

		} else {
			filteredBlkTxs = filteredTxsByAddr
		}
		filtedBlkTxsCount := len(filteredBlkTxs)
		txCounts_temp = txCounts + filtedBlkTxsCount

		if txCounts_temp <= separated {

			// all transactions in the current block should be skipped
			blkNumber--
			index = -1
			txCounts = txCounts_temp
		} else {

			// part of transactions in the current block should be skipped
			tx := filteredBlkTxs[filtedBlkTxsCount-(separated-txCounts)-1].(*TransactionResult)
			index = tx.TxIndex.Int()
			txCounts = separated
		}
	}

	tran.log.Debugf("current transaction index = %v", index)
	tran.log.Debugf("     current block number = %v", blkNumber)
	tran.log.Debugf("     minimum block number = %v", min)
	tran.log.Debugf("     maximum block number = %v", max)
	tran.log.Debugf("				 pageSize = %v\n", realArgs.PageSize.Int())

	return tran.getPrevPagingTransactions(txs, blkNumber, index, pagingArgs{
		pageSize:     realArgs.PageSize.Int(),
		minBlkNumber: min,
		maxBlkNumber: max,
		contractAddr: realArgs.ContractAddr,
		methodId:     realArgs.MethodID,
	})
}

func (tran *Transaction) getNextPagingTransactions(txs []interface{}, currentNumber uint64, currentIndex int, constant pagingArgs) ([]interface{}, error) {
	tran.log.Debugf("===== enter getNextPagingTransactions =======\n")
	tran.log.Debugf("current transaction index = %v\n", currentIndex)
	tran.log.Debugf("     current block number = %v\n", currentNumber)
	tran.log.Debugf("         current len(txs) =  %v\n", len(txs))

	if len(txs) == constant.pageSize || currentNumber > constant.maxBlkNumber {
		return txs, nil
	}

	blk, err := getBlockByNumber(tran.namespace, currentNumber, false)
	if err != nil {
		return nil, err
	}

	blockTxCount := blk.TxCounts.Int()

	if currentIndex >= blockTxCount {
		return nil, &common.InvalidParamsError{fmt.Sprintf("Invalid params. 'txIndex' %d is out of range, and now the number of transactions of block %d is %d", currentIndex, currentNumber, blockTxCount)}
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

	tran.log.Debug("===== enter getPrevPagingTransactions =======\n")
	tran.log.Debugf("current transaction index = %v\n", currentIndex)
	tran.log.Debugf("     current block number = %v\n", currentNumber)
	tran.log.Debugf("         current len(txs) =  %v\n", len(txs))

	if len(txs) == constant.pageSize || currentNumber < constant.minBlkNumber || currentNumber == 0 {
		return txs, nil
	}

	blk, err := getBlockByNumber(tran.namespace, currentNumber, false)
	if err != nil {
		return nil, err
	}

	if currentIndex == -1 {
		currentIndex = blk.TxCounts.Int() - 1
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
		} else if txResult.To.IsZero() {
			if receipt, err := tran.GetTransactionReceipt(txResult.Hash); err != nil {
				return nil, err
			} else if receipt.ContractAddress == contractAddr.Hex() {
				result = append(result, tx)
			}
		}
	}
	return result, nil
}

func prepareTransaction(args SendTxArgs, txType int, namespace string, eh *manager.EventHub) (*types.Transaction, error) {

	var tx *types.Transaction
	var txValue *types.TransactionValue

	log := common.GetLogger(namespace, "api")

	// 1. verify if the parameters are valid
	realArgs, err := prepareExcute(args, txType)
	if err != nil {
		return nil, err
	}

	// 2. create a new transaction instance
	if txType == 0 {
		txValue = types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS,
			realArgs.Value.Int64(), nil, 0, []byte(args.Extra), types.TransactionValue_EVM)
	} else {
		payload := common.FromHex(realArgs.Payload)
		txValue = types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS,
			realArgs.Value.Int64(), payload, args.Opcode, []byte(args.Extra), parseVmType(realArgs.VmType))
	}

	value, err := proto.Marshal(txValue)
	if err != nil {
		return nil, &common.CallbackError{err.Error()}
	}

	if args.To == nil {
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)
	} else {
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}

	if eh.NodeIdentification() == manager.IdentificationVP {
		tx.Id = uint64(eh.GetPeerManager().GetNodeId())
	} else {
		hash := eh.GetPeerManager().GetLocalNodeHash()
		if err := tx.SetNVPHash(hash); err != nil {
			log.Errorf("set NVP hash failed! err Msg: %v.", err.Error())
			return nil, &common.CallbackError{Message: "Marshal nvp hash error"}
		}
	}
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()

	// 3. check if there is duplicated transaction
	var exist bool
	if exist, err = bloom.LookupTransaction(namespace, tx.GetHash()); err != nil || exist == true {
		if exist, _ = edb.IsTransactionExist(namespace, tx.TransactionHash); exist {
			return nil, &common.RepeadedTxError{TxHash: common.ToHex(tx.TransactionHash)}
		}
	}

	// 4. verify transaction signature
	encryp := crypto.NewEcdsaEncrypto("ecdsa")
	if !tx.ValidateSign(encryp, kec256Hash) {
		log.Errorf("invalid signature, tx hash %v", common.ToHex(tx.TransactionHash))
		return nil, &common.SignatureInvalidError{Message: "Invalid signature, tx hash " + common.ToHex(tx.TransactionHash)}
	}

	return tx, nil
}

func postNewTxEvent(args SendTxArgs, tx *types.Transaction, eh *manager.EventHub) error {

	// post transaction event
	if eh.NodeIdentification() == manager.IdentificationNVP {
		ch := make(chan bool)
		go eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
			Ch:          ch,
		})
		res := <-ch
		close(ch)
		if res == false {
			// nvp node fails to forward tx to vp node
			return &common.CallbackError{Message: "Send tx to nvp failed."}
		}
	} else {
		go eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
		})
	}
	return nil
}

// outputTransaction makes type conversion.
func outputTransaction(trans interface{}, namespace string) (*TransactionResult, error) {
	log := common.GetLogger(namespace, "api")

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
				BlockNumber: uint64ToBlockNumber(bn),
				BlockHash:   &bHash,
				TxIndex:     int64ToNumber(txIndex),
				From:        common.BytesToAddress(t.From),
				To:          common.BytesToAddress(t.To),
				Amount:      int64ToNumber(txValue.Amount),
				Nonce:       t.Nonce,
				Extra:       string(t.GetTransactionValue().GetExtra()),
				Timestamp:   t.Timestamp,
				ExecuteTime: int64ToNumber((blk.WriteTime - blk.Timestamp) / int64(time.Millisecond)),
				Payload:     common.ToHex(txValue.Payload),
			}
		} else if err != nil && err.Error() == db_not_found_error {
			return nil, &common.DBNotFoundError{Type: BLOCK, Id: fmt.Sprintf("number %#x", bn)}
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
			Version:    string(t.Tx.Version),
			Hash:       txHash,
			From:       common.BytesToAddress(t.Tx.From),
			To:         common.BytesToAddress(t.Tx.To),
			Amount:     int64ToNumber(txValue.Amount),
			Nonce:      t.Tx.Nonce,
			Extra:      string(t.Tx.GetTransactionValue().GetExtra()),
			Timestamp:  t.Tx.Timestamp,
			Payload:    common.ToHex(txValue.Payload),
			Invalid:    true,
			InvalidMsg: t.ErrType.String(),
		}
	}

	return txRes, nil
}
