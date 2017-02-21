//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
)

type PublicBlockAPI struct {
	db hyperdb.Database
}

type BlockResult struct {
	Version    string       `json:"version"`
	Number     *BlockNumber `json:"number"`
	Hash       common.Hash  `json:"hash"`
	ParentHash common.Hash  `json:"parentHash"`
	WriteTime  int64        `json:"writeTime"`
	AvgTime    *Number      `json:"avgTime"`
	TxCounts   *Number      `json:"txcounts"`
	//Counts       *Number       `json:"counts"`
	//Percents     string        `json:"percents"`
	MerkleRoot   common.Hash   `json:"merkleRoot"`
	Transactions []interface{} `json:"transactions"`
}

type StatisticResult struct {
	BPS      string   `json:"BPS"`
	TimeList []string `json:"TimeList"`
}

func NewPublicBlockAPI(hyperDb hyperdb.Database) *PublicBlockAPI {
	return &PublicBlockAPI{
		db: hyperDb,
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
		return IntervalArgs{}, &InvalidParamsError{"missing params 'from' or 'to'"}
	} else {
		from = args.From
		to = args.To
	}

	if *from > *to || *from < 1 || *to < 1 {
		return IntervalArgs{}, &InvalidParamsError{"invalid params"}
	}

	return IntervalArgs{
		From: from,
		To:   to,
	}, nil
}

// GetBlocks returns all the block.
func (blk *PublicBlockAPI) GetBlocks(args IntervalArgs) ([]*BlockResult, error) {
	return getBlocks(args, blk.db)
}

// LastestBlock returns the number and hash of the lastest block.
func (blk *PublicBlockAPI) LatestBlock() (*BlockResult, error) {
	return latestBlock(blk.db)
}

// GetBlockByHash returns the block for the given block hash.
func (blk *PublicBlockAPI) GetBlockByHash(hash common.Hash) (*BlockResult, error) {
	return getBlockByHash(hash, blk.db)
}

// GetBlockByNumber returns the block for the given block number.
func (blk *PublicBlockAPI) GetBlockByNumber(number BlockNumber) (*BlockResult, error) {
	return getBlockByNumber(number, blk.db)
}

type BlocksIntervalResult struct {
	SumOfBlocks *Number      `json:"sumOfBlocks"`
	StartBlock  *BlockNumber `json:"startBlock"`
	EndBlock    *BlockNumber `json:"endBlock"`
}

// GetBlocksByTime returns the block for the given block time duration.
func (blk *PublicBlockAPI) GetBlocksByTime(args IntervalTime) (*BlocksIntervalResult, error) {

	if args.StartTime > args.Endtime {
		return nil, &InvalidParamsError{"invalid params"}
	}

	sumOfBlocks, startBlock, endBlock := getBlocksByTime(args.StartTime, args.Endtime, blk.db)

	return &BlocksIntervalResult{
		SumOfBlocks: NewUint64ToNumber(sumOfBlocks),
		StartBlock:  startBlock,
		EndBlock:    endBlock,
	}, nil
}

func (blk *PublicBlockAPI) GetAvgGenerateTimeByBlockNumber(args IntervalArgs) (Number, error) {
	realArgs, err := prepareIntervalArgs(args)
	if err != nil {
		return 0, err
	}

	if t, err := core.CalBlockGenerateAvgTime(realArgs.From.ToUint64(), realArgs.To.ToUint64()); err != nil && err.Error() == leveldb_not_found_error {
		return 0, &LeveldbNotFoundError{"block"}
	} else if err != nil {
		return 0, &CallbackError{err.Error()}
	} else {
		return *NewInt64ToNumber(t), nil
	}
}

func latestBlock(db hyperdb.Database) (*BlockResult, error) {

	currentChain := core.GetChainCopy()

	lastestBlkHeight := currentChain.Height

	if lastestBlkHeight == 0 {
		return nil, nil
	}

	return getBlockByNumber(*NewUint64ToBlockNumber(lastestBlkHeight), db)
}

// getBlockByNumber convert type Block to type BlockResult for the given block number.
func getBlockByNumber(n BlockNumber, db hyperdb.Database) (*BlockResult, error) {

	m := n.ToUint64()
	if blk, err := core.GetBlockByNumber(db, m); err != nil && err.Error() == leveldb_not_found_error {
		return nil, &LeveldbNotFoundError{fmt.Sprintf("block by %d", n)}
	} else if err != nil {
		return nil, &CallbackError{err.Error()}
	} else {
		return outputBlockResult(blk, db)
	}
}

// GetBlockByNumber returns the bolck for the given block time duration.
func getBlocksByTime(startTime, endTime int64, db hyperdb.Database) (sumOfBlocks uint64, startBlock, endBlock *BlockNumber) {
	currentChain := core.GetChainCopy()
	height := currentChain.Height

	var i uint64
	for i := height; i >= uint64(1); i-- {
		block, _ := getBlockByNumber(*NewUint64ToBlockNumber(i), db)
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

func outputBlockResult(block *types.Block, db hyperdb.Database) (*BlockResult, error) {

	txCounts := int64(len(block.Transactions))
	//count, percent :=types.go.CalcResponseCount(block.Number, int64(200))

	transactions := make([]interface{}, txCounts)
	var err error
	for i, tx := range block.Transactions {
		if transactions[i], err = outputTransaction(tx, db); err != nil {
			return nil, err
		}
	}

	return &BlockResult{
		Version:    string(block.Version),
		Number:     NewUint64ToBlockNumber(block.Number),
		Hash:       common.BytesToHash(block.BlockHash),
		ParentHash: common.BytesToHash(block.ParentHash),
		//WriteTime:    time.Unix(block.WriteTime/int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		WriteTime: block.WriteTime,
		AvgTime:   NewInt64ToNumber(core.CalcResponseAVGTime(block.Number, block.Number)),
		TxCounts:  NewInt64ToNumber(txCounts),
		//Counts:       NewInt64ToNumber(count),
		//Percents:     strconv.FormatFloat(percent*100, 'f', 2, 32) + "%",
		MerkleRoot:   common.BytesToHash(block.MerkleRoot),
		Transactions: transactions,
	}, nil
}

func getBlockByHash(hash common.Hash, db hyperdb.Database) (*BlockResult, error) {

	if common.EmptyHash(hash) == true {
		return nil, &InvalidParamsError{"invalid hash"}
	}

	block, err := core.GetBlock(db, hash[:])
	if err != nil && err.Error() == leveldb_not_found_error {
		return nil, &LeveldbNotFoundError{fmt.Sprintf("block by %#x", hash)}
	} else if err != nil {
		return nil, &CallbackError{err.Error()}
	}

	return outputBlockResult(block, db)
}

func getBlocks(args IntervalArgs, hyperDb hyperdb.Database) ([]*BlockResult, error) {
	var blocks []*BlockResult

	realArgs, err := prepareIntervalArgs(args)
	if err != nil {
		return nil, err
	}

	from := *realArgs.From
	to := *realArgs.To

	for from <= to {
		b, err := getBlockByNumber(to, hyperDb)
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

// 测试用
type BatchTimeResult struct {
	CommitTime int64
	BatchTime  int64
}

// QueryCommitAndBatchTime returns commit time and batch time between from block and to block
func (blk *PublicBlockAPI) QueryCommitAndBatchTime(args IntervalArgs) (*BatchTimeResult, error) {

	commitTime, batchTime := core.CalcCommitBatchAVGTime(args.From.ToUint64(), args.To.ToUint64())

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime:  batchTime,
	}, nil
}

// QueryEvmAvgTime returns EVM average time between from block and to block
func (blk *PublicBlockAPI) QueryEvmAvgTime(args IntervalArgs) (int64, error) {

	evmTime := core.CalcEvmAVGTime(args.From.ToUint64(), args.To.ToUint64())
	log.Info("-----evmTime----", evmTime)

	return evmTime, nil
}

func (blk *PublicBlockAPI) QueryTPS(args IntervalTime) (string, error) {
	err, ret := core.CalBlockGPS(int64(args.StartTime), int64(args.Endtime))
	if err != nil {
		return "", &CallbackError{err.Error()}
	}
	return ret, nil
}

func (blk *PublicBlockAPI) QueryWriteTime(args IntervalArgs) (*StatisticResult, error) {
	err, ret := core.GetBlockWriteTime(int64(args.From.ToUint64()), int64(args.To.ToUint64()))
	if err != nil {
		return nil, &CallbackError{err.Error()}
	}
	return &StatisticResult{
		TimeList: ret,
	}, nil
}
