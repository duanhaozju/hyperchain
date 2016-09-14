package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
	"time"
	"hyperchain/common"
	"strconv"

	"fmt"
)

type PublicBlockAPI struct{}

type BlockResult struct{
	Number    uint64      `json:"number"`
	Hash      common.Hash `json:"hash"`
	WriteTime string      `json:"writeTime"`
	AvgTime   int64      `json:"avgTime"`
	TxCounts  uint64      `json:"txcounts"`
	Counts    int64      `json:"Counts"`
	Percents  string      `json:"percents"`
}

func NewPublicBlockAPI() *PublicBlockAPI{
	return &PublicBlockAPI{}
}

// GetBlocks returns all the block
func (blk *PublicBlockAPI) GetBlocks() []*BlockResult{
	var blocks []*BlockResult

	block := lastestBlock()
	height := block.Number

	fmt.Println(height)

	for height > 0 {
		blocks = append(blocks,blockResult(height))
		height--
	}

	return blocks
}

// LastestBlock returns the number and hash of the lastest block
func (blk *PublicBlockAPI) LastestBlock() *BlockResult{
	return lastestBlock()
}

func lastestBlock() *BlockResult {
	db, err := hyperdb.GetLDBDatabase()

	currentChain := core.GetChainCopy()

	lastestBlkHeight := currentChain.Height

	block, err := core.GetBlockByNumber(db, lastestBlkHeight)

	if err != nil {
		log.Errorf("%v", err)
	}

	blockResult := blockResult(block.Number)

	if err != nil {
		log.Errorf("%v", err)
	}

	return blockResult
}

// blockResult convert type Block to type BlockResult according to height of the block
func blockResult(height uint64) *BlockResult{

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatalf("%v", err)
	}

	block, err := core.GetBlockByNumber(db,height)
	if err != nil {
		log.Errorf("%v", err)
	}

	txCounts := uint64(len(block.Transactions))
	count,percent := core.CalcResponseCount(height, int64(200))

	return &BlockResult{
		Number: height,
		Hash: common.BytesToHash(block.BlockHash),
		WriteTime: time.Unix(block.WriteTime / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		AvgTime: core.CalcResponseAVGTime(height,height),
		TxCounts: txCounts,
		Counts: count,
		Percents: strconv.FormatFloat(percent*100, 'f', 2, 32)+"%",
	}

}

type ExeTimeResult struct {
	Count int       `json:"count"`
	Time  int64	`json:"time"`
}

// QueryExcuteTime computes excute time of transactions fo all the block,
// then return the avg time and the count of all the transaction
func (blk *PublicBlockAPI) QueryExcuteTime() *ExeTimeResult{

	var from uint64 = 1

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatalf("%v", err)
	}

	txs,err := core.GetAllTransaction(db)

	if err != nil {
		log.Errorf("%v", err)
	}

	count := len(txs)
	exeTime := core.CalcResponseAVGTime(from,core.GetHeightOfChain())

	return &ExeTimeResult{
		Count: count,
		Time: exeTime,
	}
}

// 测试用
type SendQueryArgs struct {
	From string
	To string
}
type BatchTimeResult struct {
	CommitTime int64
	BatchTime int64
}
func (blk *PublicBlockAPI) QueryCommitAndBatchTime(args SendQueryArgs) (*BatchTimeResult,error) {

	from, err := strconv.ParseUint(args.From, 10, 64)
	to, err := strconv.ParseUint(args.To, 10, 64)

	if err != nil {
		return nil,err
	}

	commitTime, batchTime :=  core.CalcCommitBatchAVGTime(from,to)

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime: batchTime,
	},nil
}

