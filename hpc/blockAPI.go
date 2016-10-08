package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
	"time"
	"hyperchain/common"
	"strconv"
	"hyperchain/core/types"
)

type PublicBlockAPI struct{}

type BlockResult struct{
	Number    	Number      	`json:"number"`
	Hash      	common.Hash 	`json:"hash"`
	ParentHash	common.Hash	`json:"parentHash"`
	WriteTime 	string      	`json:"writeTime"`
	AvgTime   	Number      	`json:"avgTime"`
	TxCounts  	Number      	`json:"txcounts"`
	Counts    	Number      	`json:"Counts"`
	Percents  	string      	`json:"percents"`
}

func NewPublicBlockAPI() *PublicBlockAPI{
	return &PublicBlockAPI{}
}

// GetBlocks returns all the block.
func (blk *PublicBlockAPI) GetBlocks() ([]*BlockResult, error){
	var blocks []*BlockResult

	block, err := lastestBlock()

	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	height := block.Number

	// only genesis block
	if height == 0 {
		//blocks = append(blocks, block)
		return nil, nil
	}

	for height > 0 {
		b, err := getBlockByNumber(height)
		if err != nil {
			return nil, err
			break;
		}
		blocks = append(blocks, b)
		height--
	}

	return blocks, nil
}

// LastestBlock returns the number and hash of the lastest block.
func (blk *PublicBlockAPI) LastestBlock() (*BlockResult, error){
	return lastestBlock()
}

func lastestBlock() (*BlockResult, error) {
	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	currentChain := core.GetChainCopy()

	lastestBlkHeight := currentChain.Height
	log.Infof("lastestBlkHeight: %v", lastestBlkHeight)
	block, err := core.GetBlockByNumber(db, lastestBlkHeight)

	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	return outputBlockResult(block), nil
}

// getBlockByNumber convert type Block to type BlockResult for the given block number.
func getBlockByNumber(height Number) (*BlockResult, error){

	h := height.ToUint64()

	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	block, err := core.GetBlockByNumber(db,h)
	if err != nil {
		log.Errorf("%v", err)
		return nil, err
	}

	return outputBlockResult(block), nil

}

func outputBlockResult(block *types.Block) *BlockResult{

	txCounts := int64(len(block.Transactions))
	count,percent := core.CalcResponseCount(block.Number, int64(200))

	return &BlockResult{
		Number: *NewUint64ToNumber(block.Number),
		Hash: common.BytesToHash(block.BlockHash),
		ParentHash: common.BytesToHash(block.ParentHash),
		WriteTime: time.Unix(block.WriteTime / int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		AvgTime: *NewInt64ToNumber(core.CalcResponseAVGTime(block.Number, block.Number)),
		TxCounts: *NewInt64ToNumber(txCounts),
		Counts: *NewInt64ToNumber(count),
		Percents: strconv.FormatFloat(percent*100, 'f', 2, 32)+"%",
	}
}

// GetBlockByHash returns the block for the given block hash.
func (blk *PublicBlockAPI) GetBlockByHash(hash common.Hash) (*BlockResult, error){
	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		log.Errorf("Open database error: %v", err)
		return nil, err
	}

	block, err := core.GetBlock(db, hash[:])

	if err != nil {
		return nil,err
	}

	return outputBlockResult(block), nil
}

// GetBlockByNumber returns the bock for the given block number.
func (blk *PublicBlockAPI) GetBlockByNumber(number Number) (*BlockResult, error){
	return getBlockByNumber(number)
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
// then return the avg time and the count of all the transaction.
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
	exeTime := core.CalcResponseAVGTime(args.From.ToUint64(),args.To.ToUint64())

	return &ExeTimeResult{
		Count: count,
		Time: exeTime,
	}
}

// QueryCommitAndBatchTime returns commit time and batch time between from block and to block
func (blk *PublicBlockAPI) QueryCommitAndBatchTime(args SendQueryArgs) (*BatchTimeResult,error) {

	commitTime, batchTime :=  core.CalcCommitBatchAVGTime(args.From.ToUint64(),args.To.ToUint64())

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime: batchTime,
	},nil
}

// QueryEvmAvgTime returns EVM average time between from block and to block
func (blk *PublicBlockAPI) QueryEvmAvgTime(args SendQueryArgs) (int64,error) {

	evmTime :=  core.CalcEvmAVGTime(args.From.ToUint64(),args.To.ToUint64())
	log.Info("-----evmTime----",evmTime)

	return evmTime, nil
}
//func (blk *PublicBlockAPI)QueryTransactionSum() string {
//	sum := core.CalTransactionSum()
//	return sum.String()
//}
func (blk *PublicBlockAPI)QueryTransactionSum() uint64 {
	sum := core.CalTransactionSum()
	return sum
}