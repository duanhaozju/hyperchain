//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
	"hyperchain/common"
	edb "hyperchain/core/db_utils"
	"hyperchain/core/types"
)

type Block struct {
	namespace string
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
	}
}

type IntervalArgs struct {
	From *BlockNumber `json:"from"`
	To   *BlockNumber `json:"to"`
}

type IntervalTime struct {
	StartTime int64 `json:"startTime"`
	Endtime   int64 `json:"endTime"`
}

// If the client send BlockNumber "",it will convert to 0.If client send BlockNumber 0,it will return error
func prepareIntervalArgs(args IntervalArgs) (IntervalArgs, error) {
	var from, to *BlockNumber

	if args.From == nil || args.To == nil {
		return IntervalArgs{}, &common.InvalidParamsError{Message: "missing params 'from' or 'to'"}
	} else {
		from = args.From
		to = args.To
	}

	if *from > *to || *from < 1 || *to < 1 {
		return IntervalArgs{}, &common.InvalidParamsError{Message: "invalid params"}
	}

	return IntervalArgs{
		From: from,
		To:   to,
	}, nil
}

// GetBlocks returns all the block for given block number.
func (blk *Block) GetBlocks(args IntervalArgs) ([]*BlockResult, error) {
	return getBlocks(args, blk.namespace, false)
}

// GetPlainBlocks returns all the block for given block number.
func (blk *Block) GetPlainBlocks(args IntervalArgs) ([]*BlockResult, error) {
	return getBlocks(args, blk.namespace, true)
}

// LatestBlock returns the number and hash of the lastest block.
func (blk *Block) LatestBlock() (*BlockResult, error) {
	return latestBlock(blk.namespace)
}

// GetBlockByHash returns the block for the given block hash.
func (blk *Block) GetBlockByHash(hash common.Hash) (*BlockResult, error) {
	return getBlockByHash(blk.namespace, hash, false)
}

// GetPlainBlockByHash returns the block for the given block hash.
func (blk *Block) GetPlainBlockByHash(hash common.Hash) (*BlockResult, error) {
	return getBlockByHash(blk.namespace, hash, true)
}

// GetBlockByNumber returns the block for the given block number.
func (blk *Block) GetBlockByNumber(number BlockNumber) (*BlockResult, error) {

	return getBlockByNumber(blk.namespace, number, false)
}

// GetPlainBlockByNumber returns the block for the given block number.
func (blk *Block) GetPlainBlockByNumber(number BlockNumber) (*BlockResult, error) {
	return getBlockByNumber(blk.namespace, number, true)
}

type BlocksIntervalResult struct {
	SumOfBlocks *Number      `json:"sumOfBlocks"`
	StartBlock  *BlockNumber `json:"startBlock"`
	EndBlock    *BlockNumber `json:"endBlock"`
}

// GetBlocksByTime returns the block for the given block time duration.
func (blk *Block) GetBlocksByTime(args IntervalTime) (*BlocksIntervalResult, error) {

	if args.StartTime > args.Endtime {
		return nil, &common.InvalidParamsError{Message: "invalid params"}
	}

	sumOfBlocks, startBlock, endBlock := getBlocksByTime(blk.namespace, args.StartTime, args.Endtime)

	return &BlocksIntervalResult{
		SumOfBlocks: NewUint64ToNumber(sumOfBlocks),
		StartBlock:  startBlock,
		EndBlock:    endBlock,
	}, nil
}

func (blk *Block) GetAvgGenerateTimeByBlockNumber(args IntervalArgs) (Number, error) {
	realArgs, err := prepareIntervalArgs(args)
	if err != nil {
		return 0, err
	}

	if t, err := edb.CalBlockGenerateAvgTime(blk.namespace, realArgs.From.ToUint64(), realArgs.To.ToUint64()); err != nil && err.Error() == leveldb_not_found_error {
		return 0, &common.LeveldbNotFoundError{Message: "block"}
	} else if err != nil {
		return 0, &common.CallbackError{Message: err.Error()}
	} else {
		return *NewInt64ToNumber(t), nil
	}
}

func latestBlock(namespace string) (*BlockResult, error) {
	currentChain := edb.GetChainCopy(namespace)

	lastestBlkHeight := currentChain.Height

	if lastestBlkHeight == 0 {
		return nil, nil
	}

	return getBlockByNumber(namespace, *NewUint64ToBlockNumber(lastestBlkHeight), false)
}

// getBlockByNumber convert type Block to type BlockResult for the given block number.
func getBlockByNumber(namespace string, n BlockNumber, isPlain bool) (*BlockResult, error) {

	m := n.ToUint64()
	if blk, err := edb.GetBlockByNumber(namespace, m); err != nil && err.Error() == leveldb_not_found_error {
		return nil, &common.LeveldbNotFoundError{Message: fmt.Sprintf("block by %d", n)}
	} else if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	} else {
		if edb.GetHeightOfChain(namespace) == 0 {
			return nil, nil
		}
		return outputBlockResult(namespace, blk, isPlain)
	}
}

// getBlocksByTime returns the bolck for the given block time duration.
func getBlocksByTime(namespace string, startTime, endTime int64) (sumOfBlocks uint64, startBlock, endBlock *BlockNumber) {
	currentChain := edb.GetChainCopy(namespace)
	height := currentChain.Height

	var i uint64
	for i := height; i >= uint64(1); i-- {
		block, _ := getBlockByNumber(namespace, *NewUint64ToBlockNumber(i), false)
		if block.WriteTime > endTime {
			continue
		}
		if block.WriteTime < startTime {
			if i != height {
				startBlock = NewUint64ToBlockNumber(i + 1)
			}
			return sumOfBlocks, startBlock, endBlock
		}
		if block.WriteTime >= startTime && block.WriteTime <= endTime {
			sumOfBlocks += 1
			if sumOfBlocks == 1 {
				endBlock = NewUint64ToBlockNumber(i)
			}
		}
	}
	if i != height {
		startBlock = NewUint64ToBlockNumber(i + 1)
	}
	return sumOfBlocks, startBlock, endBlock
}

func outputBlockResult(namespace string, block *types.Block, isPlain bool) (*BlockResult, error) {

	txCounts := int64(len(block.Transactions))
	//count, percent :=types.go.CalcResponseCount(block.Number, int64(200))

	transactions := make([]interface{}, txCounts)
	var err error
	for i, tx := range block.Transactions {
		if transactions[i], err = outputTransaction(tx, namespace); err != nil {
			return nil, err
		}
	}

	if isPlain {
		return &BlockResult{
			Version:    string(block.Version),
			Number:     NewUint64ToBlockNumber(block.Number),
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
		Number:     NewUint64ToBlockNumber(block.Number),
		Hash:       common.BytesToHash(block.BlockHash),
		ParentHash: common.BytesToHash(block.ParentHash),
		//WriteTime:    time.Unix(block.WriteTime/int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		WriteTime: block.WriteTime,
		AvgTime:   NewInt64ToNumber(edb.CalcResponseAVGTime(namespace, block.Number, block.Number)),
		TxCounts:  NewInt64ToNumber(txCounts),
		//Counts:       NewInt64ToNumber(count),
		//Percents:     strconv.FormatFloat(percent*100, 'f', 2, 32) + "%",
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

func getBlocks(args IntervalArgs, namespace string, isPlain bool) ([]*BlockResult, error) {
	var blocks []*BlockResult

	realArgs, err := prepareIntervalArgs(args)
	if err != nil {
		return nil, err
	}

	from := *realArgs.From
	to := *realArgs.To

	for from <= to {
		b, err := getBlockByNumber(namespace, to, isPlain)
		if err != nil {
			log.Errorf("%v", err)
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

// QueryCommitAndBatchTime returns commit time and batch time between from block and to block
func (blk *Block) QueryCommitAndBatchTime(args IntervalArgs) (*BatchTimeResult, error) {

	commitTime, batchTime := edb.CalcCommitBatchAVGTime(blk.namespace, args.From.ToUint64(), args.To.ToUint64())

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime:  batchTime,
	}, nil
}

// QueryEvmAvgTime returns EVM average time between from block and to block
func (blk *Block) QueryEvmAvgTime(args IntervalArgs) (int64, error) {

	evmTime := edb.CalcEvmAVGTime(blk.namespace, args.From.ToUint64(), args.To.ToUint64())
	log.Info("-----evmTime----", evmTime)

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
	err, ret := edb.GetBlockWriteTime(blk.namespace, int64(args.From.ToUint64()), int64(args.To.ToUint64()))
	if err != nil {
		return nil, &common.CallbackError{Message: err.Error()}
	}
	return &StatisticResult{
		TimeList: ret,
	}, nil
}
