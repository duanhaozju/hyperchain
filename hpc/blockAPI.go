package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
	"time"
	"hyperchain/common"
	"strconv"

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
		log.Fatalf("%v", err)
	}

	return blockResult
}

// blockResult convert type Block to type BlockResult according to height of the block
func blockResult(height uint64) (*BlockResult){

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatalf("%v", err)
	}

	block, err := core.GetBlockByNumber(db,height)
	if err != nil {
		log.Fatalf("%v", err)
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

// QueryExcuteTime computes excute time of transactions fo all the block
func (blk *PublicBlockAPI) QueryExcuteTime() int64{

	lastestHeight := core.GetHeightOfChain()

	var from uint64 = 1

	return core.CalcResponseAVGTime(from,lastestHeight)
}

