//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
)

type Block struct {
	namespace string
	height    uint64
	log       *logging.Logger
}

type BlockResult struct {
	Version      string        `json:"version"`
	Number       *BlockNumber  `json:"number"`
	Hash         common.Hash   `json:"hash"`
	ParentHash   common.Hash   `json:"parentHash"`
	WriteTime    int64         `json:"writeTime"`
	AvgTime      *Number       `json:"avgTime"`
	TxCounts     *Number       `json:"txcounts"`
	MerkleRoot   common.Hash   `json:"merkleRoot"`
	Transactions []interface{} `json:"transactions,omitempty"`
}

type StatisticResult struct {
	BPS      string   `json:"BPS"`
	TimeList []string `json:"TimeList"`
}

func NewPublicBlockAPI(namespace string) *Block {
	return &Block{
		namespace: namespace,
		log:       common.GetLogger(namespace, "api"),
	}
}

type IntervalArgs struct {
	From         *BlockNumber    `json:"from"`
	To           *BlockNumber    `json:"to"`
	ContractAddr *common.Address `json:"address"`
	MethodID     string          `json:"methodID"`
}

type IntervalTime struct {
	StartTime int64 `json:"startTime"`
	Endtime   int64 `json:"endTime"`
}

type intArgs struct {
	from uint64
	to   uint64
}

// GetBlocks returns all the block for given block number.
func (blk *Block) GetBlocks(args IntervalArgs) ([]*BlockResult, error) {
	trueArgs, err := prepareIntervalArgs(args, blk.namespace)
	if err != nil {
		return nil, err
	} else {
		return getBlocks(trueArgs, blk.namespace, false)
	}
}

// GetPlainBlocks returns all the block for given block number without transactions.
func (blk *Block) GetPlainBlocks(args IntervalArgs) ([]*BlockResult, error) {
	trueArgs, err := prepareIntervalArgs(args, blk.namespace)
	if err != nil {
		return nil, err
	} else {
		return getBlocks(trueArgs, blk.namespace, true)
	}
}

// LatestBlock returns the lastest block.
func (blk *Block) LatestBlock() (*BlockResult, error) {
	return latestBlock(blk.namespace)
}

// GetBlockByHash returns the block for the given block hash.
func (blk *Block) GetBlockByHash(hash common.Hash) (*BlockResult, error) {
	return getBlockByHash(blk.namespace, hash, false)
}

// GenesisBlock returns current genesis block number.
func (blk *Block) GenesisBlock() (uint64, error) {
	err, genesis := edb.GetGenesisTag(blk.namespace)
	return genesis, &common.CallbackError{Message: err.Error()}
}

// GetPlainBlockByHash returns the block for the given block hash.
func (blk *Block) GetPlainBlockByHash(hash common.Hash) (*BlockResult, error) {
	return getBlockByHash(blk.namespace, hash, true)
}

// GetBlockByNumber returns the block for the given block number.
func (blk *Block) GetBlockByNumber(n BlockNumber) (*BlockResult, error) {
	number, err := prepareBlockNumber(n, blk.namespace)
	if err != nil {
		return nil, err
	} else {
		return getBlockByNumber(blk.namespace, number, false)
	}
}

// GetPlainBlockByNumber returns the block for the given block number without transactions.
func (blk *Block) GetPlainBlockByNumber(n BlockNumber) (*BlockResult, error) {
	number, err := prepareBlockNumber(n, blk.namespace)
	if err != nil {
		return nil, err
	} else {
		return getBlockByNumber(blk.namespace, number, true)
	}
}

type BlocksIntervalResult struct {
	SumOfBlocks *Number      `json:"sumOfBlocks"`
	StartBlock  *BlockNumber `json:"startBlock"`
	EndBlock    *BlockNumber `json:"endBlock"`
}

// GetBlocksByTime returns the number of blocks, starting block and ending block for the given time duration.
func (blk *Block) GetBlocksByTime(args IntervalTime) (*BlocksIntervalResult, error) {

	if args.StartTime > args.Endtime {
		return nil, &common.InvalidParamsError{Message: "invalid params"}
	}

	sumOfBlocks, startBlock, endBlock, err := getBlocksByTime(blk.namespace, args.StartTime, args.Endtime)
	if err != nil {
		return nil, err
	}

	return &BlocksIntervalResult{
		SumOfBlocks: NewUint64ToNumber(sumOfBlocks),
		StartBlock:  startBlock,
		EndBlock:    endBlock,
	}, nil
}

// GetAvgGenerateTimeByBlockNumber calculates the average generation time of all blocks for the given block number.
func (blk *Block) GetAvgGenerateTimeByBlockNumber(args IntervalArgs) (Number, error) {
	intargs, err := prepareIntervalArgs(args, blk.namespace)
	if err != nil {
		return 0, err
	}

	if t, err := edb.CalBlockGenerateAvgTime(blk.namespace, intargs.from, intargs.to); err != nil && err.Error() == leveldb_not_found_error {
		return 0, &common.LeveldbNotFoundError{Message: "block"}
	} else if err != nil {
		return 0, &common.CallbackError{Message: err.Error()}
	} else {
		return *NewInt64ToNumber(t), nil
	}
}

// GetChainHeight returns latest block height.
func (blk *Block) GetChainHeight() (*BlockNumber, error) {
	chain, err := edb.GetChain(blk.namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	} else if chain.Height == 0 {
		return nil, &common.NoBlockGeneratedError{Message: "There is no block generated!"}
	}
	return Uint64ToBlockNumber(chain.Height), nil
}

func latestBlock(namespace string) (*BlockResult, error) {
	chain, err := edb.GetChain(namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}
	lastestBlkHeight := chain.Height

	if lastestBlkHeight == 0 {
		return nil, &common.NoBlockGeneratedError{Message: "There is no block generated!"}
	}

	return getBlockByNumber(namespace, lastestBlkHeight, false)
}


func getBlockByNumber(namespace string, number uint64, isPlain bool) (*BlockResult, error) {
	if blk, err := edb.GetBlockByNumber(namespace, number); err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %d", number)}
	} else if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	} else {
		if edb.GetHeightOfChain(namespace) == 0 {
			return nil, &common.NoBlockGeneratedError{Message: "There is no block generated!"}
		}
		return outputBlockResult(namespace, blk, isPlain)
	}
}


func getBlocksByTime(namespace string, startTime, endTime int64) (sumOfBlocks uint64, startBlock, endBlock *BlockNumber, err error) {
	currentChain, err := edb.GetChain(namespace)
	if err != nil {
		return 0, nil, nil, &common.CallbackError{Message: err.Error()}
	}
	height := currentChain.Height

	var i uint64
	for i := height; i >= uint64(1); i-- {
		block, _ := getBlockByNumber(namespace, i, false)
		if block.WriteTime > endTime {
			continue
		}
		if block.WriteTime < startTime {
			if i != height {
				startBlock = Uint64ToBlockNumber(i + 1)
			}
			return sumOfBlocks, startBlock, endBlock, nil
		}
		if block.WriteTime >= startTime && block.WriteTime <= endTime {
			sumOfBlocks += 1
			if sumOfBlocks == 1 {
				endBlock = Uint64ToBlockNumber(i)
			}
		}
	}
	if i != height {
		startBlock = Uint64ToBlockNumber(i + 1)
	}
	return sumOfBlocks, startBlock, endBlock, nil
}

// outputBlockResult makes type conversion.
func outputBlockResult(namespace string, block *types.Block, isPlain bool) (*BlockResult, error) {

	var err error

	txCounts := int64(len(block.Transactions))
	transactions := make([]interface{}, txCounts)

	for i, tx := range block.Transactions {
		if transactions[i], err = outputTransaction(tx, namespace); err != nil {
			return nil, err
		}
	}

	if isPlain {
		return &BlockResult{
			Version:    string(block.Version),
			Number:     Uint64ToBlockNumber(block.Number),
			Hash:       common.BytesToHash(block.BlockHash),
			ParentHash: common.BytesToHash(block.ParentHash),
			WriteTime:  block.WriteTime,
			AvgTime:    NewInt64ToNumber(edb.CalcResponseAVGTime(namespace, block.Number, block.Number)),
			TxCounts:   NewInt64ToNumber(txCounts),
			MerkleRoot: common.BytesToHash(block.MerkleRoot),
		}, nil
	}

	return &BlockResult{
		Version:    string(block.Version),
		Number:     Uint64ToBlockNumber(block.Number),
		Hash:       common.BytesToHash(block.BlockHash),
		ParentHash: common.BytesToHash(block.ParentHash),
		WriteTime: block.WriteTime,
		AvgTime:   NewInt64ToNumber(edb.CalcResponseAVGTime(namespace, block.Number, block.Number)),
		TxCounts:  NewInt64ToNumber(txCounts),
		MerkleRoot:   common.BytesToHash(block.MerkleRoot),
		Transactions: transactions,
	}, nil
}

func getBlockByHash(namespace string, hash common.Hash, isPlain bool) (*BlockResult, error) {

	if common.EmptyHash(hash) == true {
		return nil, &common.InvalidParamsError{Message: "invalid hash"}
	}

	block, err := edb.GetBlock(namespace, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}

	return outputBlockResult(namespace, block, isPlain)
}

func getBlocks(args *intArgs, namespace string, isPlain bool) ([]*BlockResult, error) {
	var blocks []*BlockResult

	from := args.from
	to := args.to

	for from <= to {
		b, err := getBlockByNumber(namespace, to, isPlain)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, b)
		if to == 1 {
			break
		}
		to--
	}

	return blocks, nil
}

type BatchTimeResult struct {
	CommitTime int64
	BatchTime  int64
}

// QueryCommitAndBatchTime returns commit time and batch time between from block and to block.
func (blk *Block) QueryCommitAndBatchTime(args IntervalArgs) (*BatchTimeResult, error) {
	trueArgs, err := prepareIntervalArgs(args, blk.namespace)
	if err != nil {
		return nil, err
	}

	commitTime, batchTime := edb.CalcCommitBatchAVGTime(blk.namespace, trueArgs.from, trueArgs.to)

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime:  batchTime,
	}, nil
}

// QueryEvmAvgTime returns EVM average time between from block and to block.
func (blk *Block) QueryEvmAvgTime(args IntervalArgs) (int64, error) {
	trueArgs, err := prepareIntervalArgs(args, blk.namespace)
	if err != nil {
		return 0, err
	}

	evmTime := edb.CalcEvmAVGTime(blk.namespace, trueArgs.from, trueArgs.to)
	blk.log.Info("-----evmTime----", evmTime)

	return evmTime, nil
}

func (blk *Block) QueryTPS(args IntervalTime) (string, error) {
	err, ret := edb.CalBlockGPS(blk.namespace, int64(args.StartTime), int64(args.Endtime))
	if err != nil {
		return "", &common.CallbackError{Message: err.Error()}
	}
	return ret, nil
}

func (blk *Block) QueryWriteTime(args IntervalArgs) (*StatisticResult, error) {
	trueArgs, err := prepareIntervalArgs(args, blk.namespace)
	if err != nil {
		return nil, err
	}
	err, ret := edb.GetBlockWriteTime(blk.namespace, int64(trueArgs.from), int64(trueArgs.to))
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}
	return &StatisticResult{
		TimeList: ret,
	}, nil
}
