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
	//Number    uint64      `json:"number"`
	Number    Number      `json:"number"`
	Hash      common.Hash `json:"hash"`
	WriteTime string      `json:"writeTime"`
	AvgTime   Number      `json:"avgTime"`
	TxCounts  Number      `json:"txcounts"`
	Counts    Number      `json:"Counts"`
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

	blockResult := blockResult(*NewUint64ToNumber(block.Number))

	if err != nil {
		log.Errorf("%v", err)
	}

	return blockResult
}

// blockResult convert type Block to type BlockResult according to height of the block
//func blockResult(height uint64) *BlockResult{
func blockResult(height Number) *BlockResult{

	h := height.ToUnit64()

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Errorf("%v", err)
	}

	block, err := core.GetBlockByNumber(db,h)
	if err != nil {
		log.Errorf("%v", err)
	}

	txCounts := int64(len(block.Transactions))
	count,percent := core.CalcResponseCount(h, int64(200))

	return &BlockResult{
		Number: height,
		Hash: common.BytesToHash(block.BlockHash),
		WriteTime: time.Unix(block.WriteTime / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		AvgTime: *NewInt64ToNumber(core.CalcResponseAVGTime(h,h)),
		TxCounts: *NewInt64ToNumber(txCounts),
		Counts: *NewInt64ToNumber(count),
		Percents: strconv.FormatFloat(percent*100, 'f', 2, 32)+"%",
	}

}

// 测试用
type SendQueryArgs struct {
	From Number
	To Number
}
type BatchTimeResult struct {
	CommitTime int64
	BatchTime int64
}
type ExeTimeResult struct {
	Count int       `json:"count"`
	Time  int64	`json:"time"`
}

// QueryExecuteTime computes execute time of transactions fo all the block,
// then return the avg time and the count of all the transaction
func (blk *PublicBlockAPI) QueryExecuteTime(args SendQueryArgs) *ExeTimeResult{

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Errorf("%v", err)
	}
	txs,err := core.GetAllTransaction(db)

	if err != nil {
		log.Errorf("%v", err)
	}

	count := len(txs)
	exeTime := core.CalcResponseAVGTime(args.From.ToUnit64(),args.To.ToUnit64())

	return &ExeTimeResult{
		Count: count,
		Time: exeTime,
	}
}

// QueryCommitAndBatchTime returns commit time and batch time between from block and to block
func (blk *PublicBlockAPI) QueryCommitAndBatchTime(args SendQueryArgs) (*BatchTimeResult,error) {

	commitTime, batchTime :=  core.CalcCommitBatchAVGTime(args.From.ToUnit64(),args.To.ToUnit64())

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime: batchTime,
	},nil
}

// QueryEvmAvgTime returns EVM average time between from block and to block
func (blk *PublicBlockAPI) QueryEvmAvgTime(args SendQueryArgs) (int64,error) {

	evmTime :=  core.CalcEvmAVGTime(args.From.ToUnit64(),args.To.ToUnit64())

	return evmTime, nil
}
func (blk *PublicBlockAPI)QueryTransactionSum() string {
	//return strconv.FormatUint(core.CalTransactionSum(),10)
	currentChain := core.GetChainCopy()
	sum := currentChain.CurrentTxSum
	return strconv.FormatUint(sum,10)
}
func (blk *PublicBlockAPI)QueryBlockAvgTime(args SendQueryArgs)(int64,error)  {
	from, err := strconv.ParseUint(args.From, 10, 64)
	to, err := strconv.ParseUint(args.To, 10, 64)

	if err != nil {
		return 0,err
	}

	evmTime :=  core.CalcBlockAVGTime(from,to)

	return evmTime, nil
}
func (blk *PublicBlockAPI)QueryBlockGPS()(error)  {
	return core.CalBlockGPS()
}