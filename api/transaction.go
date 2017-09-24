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
	"hyperchain/hyperdb/db"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"time"
)

const (
	DEFAULT_GAS      int64 = 100000000
	DEFAULT_GAS_PRICE int64 = 10000
)

var (
	kec256Hash              = crypto.NewKeccak256Hash("keccak256")
	leveldb_not_found_error = db.DB_NOT_FOUND.Error()
)

type Transaction struct {
	namespace   string
	eh          *manager.EventHub
	tokenBucket *ratelimit.Bucket
	config      *common.Config
	log         *logging.Logger
}

// SendTxArgs represents the arguments to sumbit a new transaction into the transaction pool.
type SendTxArgs struct {
	From       common.Address  `json:"from"`		// transaction sender address
	To         *common.Address `json:"to"`		    // transaction receiver address
	Value      Number         `json:"value"`		// transaction amount
	Payload    string          `json:"payload"`		// contract payload
	Signature  string          `json:"signature"`	// signature of sender for the transaction
	Timestamp  int64           `json:"timestamp"`	// timestamp of the transaction happened
	Simulate   bool            `json:"simulate"`	// Simulate determines if the transaction requires consensus, if true, no consensus.
	Nonce      int64           `json:"nonce"`		// 16-bit random decimal number, for example 5956491387995926
	VmType     string          `json:"type"`		// specify which engine executes contract

	// 1 value for Opcode means upgrades contract, 2 means freezes contract,
	// 3 means unfreezes contract, 100 means archives data.
	Opcode     int32           `json:"opcode"`

	// Snapshot saves the state of ledger at a moment.
	// SnapshotId specifies the based ledger when client sends transaction or invokes contract with Simulate=true.
	SnapshotId string          `json:"snapshotId"`
}

type TransactionResult struct {
	Version     string         `json:"version"`					// hyperchain version when the transaction is executed
	Hash        common.Hash    `json:"hash"`
	BlockNumber *BlockNumber   `json:"blockNumber,omitempty"`
	BlockHash   *common.Hash   `json:"blockHash,omitempty"`
	TxIndex     *Number        `json:"txIndex,omitempty"`
	From        common.Address `json:"from"`
	To          common.Address `json:"to"`
	Amount      *Number        `json:"amount,omitempty"`
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

// SendTransaction is to build a transaction object, and then post event NewTxEvent,
// if the sender's balance is enough, return tx hash.
func (tran *Transaction) SendTransaction(args SendTxArgs) (common.Hash, error) {
	if getRateLimitEnable(tran.config) && tran.tokenBucket.TakeAvailable(1) <= 0 {
		return common.Hash{}, &common.SystemTooBusyError{Message: "system is too busy to response "}
	}
	var tx *types.Transaction

	// 1. verify if the parameters are valid
	realArgs, err := prepareExcute(args, 0)
	if err != nil {
		return common.Hash{}, err
	}

	// 2. create a new transaction instance
	txValue := types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS,
		realArgs.Value.Int64(), nil, 0, types.TransactionValue_EVM)

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &common.CallbackError{err.Error()}
	}

	tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	if tran.eh.NodeIdentification() == manager.IdentificationVP {
		tx.Id = uint64(tran.eh.GetPeerManager().GetNodeId())
	} else {
		hash := tran.eh.GetPeerManager().GetLocalNodeHash()
		if err := tx.SetNVPHash(hash); err != nil {
			tran.log.Errorf("set NVP hash failed! err Msg: %v.", err.Error())
			return common.Hash{}, &common.CallbackError{Message: "marshal nvp hash error"}
		}
	}
	tx.Signature = common.FromHex(realArgs.Signature)
	tx.TransactionHash = tx.Hash().Bytes()

	// 3. check if there is duplicated transaction
	var exist bool
	if err, exist = edb.LookupTransaction(tran.namespace, tx.GetHash()); err != nil || exist == true {
		if exist, _ = edb.JudgeTransactionExist(tran.namespace, tx.TransactionHash); exist {
			return common.Hash{}, &common.RepeadedTxError{Message: "repeated tx " + common.ToHex(tx.TransactionHash)}
		}
	}

	// 4. verify transaction signature
	if !tx.ValidateSign(tran.eh.GetAccountManager().Encryption, kec256Hash) {
		tran.log.Error("invalid signature")
		return common.Hash{}, &common.SignatureInvalidError{Message: "invalid signature, tx hash " + common.ToHex(tx.TransactionHash)}
	}

	// 5. post transaction event
	if tran.eh.NodeIdentification() == manager.IdentificationNVP {
		ch := make(chan bool)
		go tran.eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			Ch:          ch,
		})
		res := <-ch
		close(ch)
		if res == false {
			// nvp node fails to forward tx to vp node
			return common.Hash{}, &common.CallbackError{Message: "send tx to nvp failed."}
		}
	} else {
		go tran.eh.GetEventObject().Post(event.NewTxEvent{
			Transaction: tx,
			Simulate:    args.Simulate,
			SnapshotId:  args.SnapshotId,
		})
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
			return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("receipt by %#x", hash)}
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

// GetTransactions return all transactions in the given block number.
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

// GetDiscardTransactions returns all invalid transactions that dont be saved in the ledger.
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
		return nil, &common.InvalidParamsError{Message: "Invalid params, both startTime and endTime must be positive, startTime is less than endTime"}
	}

	reds, err := edb.GetAllDiscardTransaction(tran.namespace)
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: "discard transactions"}
	} else if err != nil {
		tran.log.Errorf("GetDiscardTransactionsByTime error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
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
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("discard transaction by %#x", hash)}
	} else if err != nil {
		tran.log.Errorf("GetDiscardTransaction error: %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return outputTransaction(red, tran.namespace)
}

// GetTransactionByHash returns the transaction for the given transaction hash. The method
func (tran *Transaction) GetTransactionByHash(hash common.Hash) (*TransactionResult, error) {
	tx, err := edb.GetTransaction(tran.namespace, hash[:])
	if err != nil && err == edb.NotFindTxMetaErr {
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
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	if index.Int() >= txCount {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("transaction, this block contains %v transactions, but the index %v is out of range", txCount, index)}
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
		return nil,  &common.InvalidParamsError{Message: err.Error()}
	}

	block, err := edb.GetBlockByNumber(tran.namespace, blknumber)
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %d", n)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	if index.Int() >= txCount {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("transaction, this block contains %v transactions, but the index %v is out of range", txCount, index)}
	}

	if index.Int() >= 0 && index.Int() < txCount {

		tx := block.Transactions[index]

		return outputTransaction(tx, tran.namespace)
	}

	return nil, nil
}

// GetTransactionsByTime returns the transactions in the given time duration.
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
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by number %#x", n)}
	} else if err != nil {
		tran.log.Errorf("%v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}

	txCount := len(block.Transactions)

	return intToNumber(txCount), nil
}

// GetSignHash returns hash of transaction content for client signature.
func (tran *Transaction) GetSignHash(args SendTxArgs) (common.Hash, error) {

	var tx *types.Transaction

	realArgs, err := prepareExcute(args, 3) // empty contract address and empty transaction signature
	if err != nil {
		return common.Hash{}, err
	}

	payload := common.FromHex(realArgs.Payload)

	txValue := types.NewTransactionValue(DEFAULT_GAS_PRICE, DEFAULT_GAS, realArgs.Value.Int64(), payload, args.Opcode, types.TransactionValue_EVM)

	value, err := proto.Marshal(txValue)
	if err != nil {
		return common.Hash{}, &common.CallbackError{Message: err.Error()}
	}

	if args.To == nil {
		// deploy contract
		tx = types.NewTransaction(realArgs.From[:], nil, value, realArgs.Timestamp, realArgs.Nonce)

	} else {
		// send contract or send transaction
		tx = types.NewTransaction(realArgs.From[:], (*realArgs.To)[:], value, realArgs.Timestamp, realArgs.Nonce)
	}

	return tx.SighHash(kec256Hash), nil
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

// GetTransactionsCountByContractAddr returns the number of eligible transaction, the latest block number and transaction index
// of eligible transaction in the block for the given block number, contract address.
func (tran *Transaction) GetTransactionsCountByContractAddr(args IntervalArgs) (interface{}, error) {
	if args.ContractAddr == nil {
		return nil, &common.InvalidParamsError{"'address' can't be empty"}
	} else if args.MethodID != "" {
		return nil, &common.InvalidParamsError{"invalid params, 'methodID' is unrecognized"}
	}
	return tran.getTransactionsCountByBlockNumber(args)
}

// GetTransactionsCountByMethodID returns the number of eligible transaction, the latest block number and transaction index
// of eligible transaction in the block for the given block number, methodID. Method id is contract method identifier in a contract.
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
	//contractAddr := args.ContractAddr.Hex()
	contractAddr := args.ContractAddr
	var lastBlockNum uint64

	for from <= realArgs.to {

		block, err := getBlockByNumber(tran.namespace, from, false)
		if err != nil {
			return 0, err
		}

		for _, tx := range block.Transactions {
			txResult := tx.(*TransactionResult)

			//to := txResult.To.Hex()
			to := txResult.To
			txIndex := txResult.TxIndex.Int()
			blockNum, err := prepareBlockNumber(*txResult.BlockNumber, tran.namespace)

			if err != nil {
				return nil, &common.CallbackError{Message: err.Error()}
			}

			//if to == contractAddr && to != "0x0000000000000000000000000000000000000000" {
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

			//if args.MethodID == "" && to == "0x0000000000000000000000000000000000000000" {
			if args.MethodID == "" && to.IsZero() {
				if receipt, err := tran.GetTransactionReceipt(txResult.Hash); err != nil {
					return 0, err
				//} else if receipt.ContractAddress == contractAddr {
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

// PagingArgs specifies conditions that transactions filter for transaction paging. PagingArgs determines starting position including
// current block number and index of the transaction in the current block, quantity of returned transactions and filter conditions.
//
// For example, pageSize is 10. From page 1 to page 2, so "separated" value is 0. From page 1 to page 3, so "separated" value is 10.
type PagingArgs struct {
	BlkNumber      BlockNumber     `json:"blkNumber"`		// the current block number
	MaxBlkNumber   BlockNumber     `json:"maxBlkNumber"`    // the maximum block number of allowing to query
	MinBlkNumber   BlockNumber     `json:"minBlkNumber"`	// the minimum block number of allowing to query
	TxIndex        Number          `json:"txIndex"`			// index of the transaction in the current block
	Separated      Number          `json:"separated"`		// specify how many transactions to skip.
	PageSize       Number          `json:"pageSize"`		// specify the number of transaction returned
	ContainCurrent bool            `json:"containCurrent"`  // specify if the returned transactions contain current transaction
	ContractAddr   *common.Address `json:"address"`			// specify which contract transactions belong to
	MethodID       string          `json:"methodID"`		// specify which contract method transactions belong to
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
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("'blkNumber' %v is out of range, it must be in the range %v to %v", blkNumber, min, max)}
	}

	txs := make([]interface{}, 0)

	// to comfirm starting position
	index := realArgs.TxIndex.Int()
	separated := realArgs.Separated.Int()
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

		blockTxCount := block.TxCounts.Int()
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

		blockTxCount := block.TxCounts.Int()

		if index < blockTxCount-1 {
			index++
		} else if index == blockTxCount-1 {
			blkNumber++
			index = 0
		} else {
			return nil, &common.InvalidParamsError{Message: fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
		}
	}

	tran.log.Debugf("current transaction index = %v\n", index)
	tran.log.Debugf("     current block number = %v\n", blkNumber)
	tran.log.Debugf("     minimum block number = %v\n", min)
	tran.log.Debugf("     maximum block number = %v\n", max)
	tran.log.Debugf("				 pageSize = %v\n", realArgs.PageSize.Int())

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
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("'blkNumber' %v is out of range, it must be in the range %v to %v", blkNumber, min, max)}
	}

	txs := make([]interface{}, 0)

	// to comfirm end position
	index := realArgs.TxIndex.Int() // 40
	separated := realArgs.Separated.Int()
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
		blockTxCount := block.TxCounts.Int()
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

		blockTxCount := block.TxCounts.Int()

		if index >= blockTxCount {
			return nil, &common.InvalidParamsError{fmt.Sprintf("'txIndex' %d is out of range, and now the number of transactions of block %d is %d", index, blkNumber, blockTxCount)}
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

	tran.log.Debugf("current transaction index = %v\n", index)
	tran.log.Debugf("     current block number = %v\n", blkNumber)
	tran.log.Debugf("     minimum block number = %v\n", min)
	tran.log.Debugf("     maximum block number = %v\n", max)
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
		//} else if txResult.To.Hex() == "0x0000000000000000000000000000000000000000" {
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
				Timestamp:   t.Timestamp,
				ExecuteTime: int64ToNumber((blk.WriteTime - blk.Timestamp) / int64(time.Millisecond)),
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
			Amount:  int64ToNumber(txValue.Amount),
			Nonce:   t.Tx.Nonce,
			Timestamp:  t.Tx.Timestamp,
			Payload:    common.ToHex(txValue.Payload),
			Invalid:    true,
			InvalidMsg: t.ErrType.String(),
		}
	}

	return txRes, nil
}
