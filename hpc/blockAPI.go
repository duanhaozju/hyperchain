package hpc

import (
	"hyperchain/hyperdb"
	"hyperchain/core"
	"time"
	"hyperchain/common"
	"strconv"
	"hyperchain/core/types"
	"hyperchain/core/state"
	"errors"
)

type PublicBlockAPI struct{
	db *hyperdb.LDBDatabase
}

type BlockResult struct {
	Number       *Number      	`json:"number"`
	Hash         common.Hash 	`json:"hash"`
	ParentHash   common.Hash 	`json:"parentHash"`
	WriteTime    string      	`json:"writeTime"`
	AvgTime      *Number      	`json:"avgTime"`
	TxCounts     *Number      	`json:"txcounts"`
	Counts       *Number      	`json:"Counts"`
	Percents     string      	`json:"percents"`
	MerkleRoot   common.Hash	`json:"merkleRoot"`
	Transactions []interface{}	`json:"transactions"`
}

func NewPublicBlockAPI(hyperDb *hyperdb.LDBDatabase) *PublicBlockAPI {
	return &PublicBlockAPI{
		db: hyperDb,
	}
}

type BlockArgsTest struct {
	From *Number	`json:"from"`
	To *Number	`json:"to"`
}

// GetBlocks returns all the block.
func (blk *PublicBlockAPI) GetBlocks(args BlockArgsTest) ([]*BlockResult, error) {
	var blocks []*BlockResult

	if args.From == nil && args.To == nil {
		block, err := blk.lastestBlock()

		if err != nil {
			log.Errorf("%v", err)
			return nil, err
		}

		height := *block.Number

		// only genesis block
		if height == 0 {
			//blocks = append(blocks, block)
			return nil, nil
		}

		for height > 0 {
			b, err := getBlockByNumber(height, blk.db)
			if err != nil {
				return nil, err
				break
			}
			blocks = append(blocks, b)
			height--
		}
	} else if args.From == nil || args.To == nil || *args.From > *args.To || *args.From < 0 || *args.To < 0 {
		return nil, errors.New("Invalid params")
	} else if *args.From == *args.To {
		if block, err := blk.GetBlockByNumber(*args.From); err != nil {
			return nil, err
		} else {
			blocks = append(blocks, block)
		}
	}  else {
		from := *args.From
		to := *args.To
		for from <= to {
			b, err := blk.GetBlockByNumber(to)
			if err != nil {
				log.Errorf("%v", err)
				return nil, err
			}
			blocks = append(blocks, b)
			to--
		}
	}

	return blocks, nil
}

// LastestBlock returns the number and hash of the lastest block.
func (blk *PublicBlockAPI) LastestBlock() (*BlockResult, error) {
	return blk.lastestBlock()
}

// GetBlockByHash returns the block for the given block hash.
func (blk *PublicBlockAPI) GetBlockByHash(hash common.Hash) (*BlockResult, error) {
	return getBlockByHash(hash, blk.db)
}

// GetBlockByNumber returns the bock for the given block number.
func (blk *PublicBlockAPI) GetBlockByNumber(number Number) (*BlockResult, error) {
	block, err := getBlockByNumber(number, blk.db)
	return block, err
}

func (blk *PublicBlockAPI) lastestBlock() (*BlockResult, error) {

	currentChain := core.GetChainCopy()

	lastestBlkHeight := currentChain.Height

	return getBlockByNumber(*NewUint64ToNumber(lastestBlkHeight) ,blk.db)
}

// getBlockByNumber convert type Block to type BlockResult for the given block number.
func getBlockByNumber(n Number, db *hyperdb.LDBDatabase) (*BlockResult, error) {

	if n == latestBlockNumber {
		chain := core.GetChainCopy()
		return getBlockByHash(common.BytesToHash(chain.LatestBlockHash), db)
		//if block, err := getBlockByHash(common.BytesToHash(chain.LatestBlockHash), db); err != nil {
		//	return nil, err
		//} else {
		//	return block, nil
		//}
	} else {
		m := n.ToUint64()
		if blk, err := core.GetBlockByNumber(db, m); err != nil {
			return nil, err
		} else {
			return outputBlockResult(blk, db)
		}
	}

}

func getBlockStateDb(n Number, db *hyperdb.LDBDatabase) (*state.StateDB, error) {

	block, err := getBlockByNumber(n, db)

	if err != nil {
		return nil, err
	}

	stateDB, err := state.New(block.MerkleRoot, db)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return nil, err
	}

	return stateDB, nil
}

func outputBlockResult(block *types.Block, db *hyperdb.LDBDatabase) (*BlockResult, error) {

	txCounts := int64(len(block.Transactions))
	count, percent := core.CalcResponseCount(block.Number, int64(200))

	transactions := make([]interface{}, txCounts)
	var err error
	for i, tx := range block.Transactions {
		if transactions[i], err = outputTransaction(tx, db); err != nil {
			return nil, err
		}
	}


	return &BlockResult{
		Number:       NewUint64ToNumber(block.Number),
		Hash:         common.BytesToHash(block.BlockHash),
		ParentHash:   common.BytesToHash(block.ParentHash),
		WriteTime:    time.Unix(block.WriteTime/int64(time.Second), 0).Format("2006-01-02 15:04:05"),
		AvgTime:      NewInt64ToNumber(core.CalcResponseAVGTime(block.Number, block.Number)),
		TxCounts:     NewInt64ToNumber(txCounts),
		Counts:       NewInt64ToNumber(count),
		Percents:     strconv.FormatFloat(percent*100, 'f', 2, 32) + "%",
		MerkleRoot:   common.BytesToHash(block.MerkleRoot),
		Transactions: transactions,
	}, nil
}

func getBlockByHash(hash common.Hash, db *hyperdb.LDBDatabase) (*BlockResult, error) {

	block, err := core.GetBlock(db, hash[:])

	if err != nil {
		return nil, err
	}

	return outputBlockResult(block, db)
}

// 测试用
type SendQueryArgs struct {
	From Number
	To   Number
}
type BatchTimeResult struct {
	CommitTime int64
	BatchTime  int64
}
type ExeTimeResult struct {
	Count int   `json:"count"`
	Time  int64 `json:"time"`
}

// QueryExecuteTime computes execute time of transactions fo all the block,
// then return the avg time and the count of all the transaction.
func (blk *PublicBlockAPI) QueryExecuteTime(args SendQueryArgs) *ExeTimeResult {

	//db, err := hyperdb.GetLDBDatabase()
	//if err != nil {
	//	log.Errorf("%v", err)
	//}
	txs, err := core.GetAllTransaction(blk.db)

	if err != nil {
		log.Errorf("%v", err)
	}

	count := len(txs)
	exeTime := core.CalcResponseAVGTime(args.From.ToUint64(), args.To.ToUint64())

	return &ExeTimeResult{
		Count: count,
		Time:  exeTime,
	}
}

// QueryCommitAndBatchTime returns commit time and batch time between from block and to block
func (blk *PublicBlockAPI) QueryCommitAndBatchTime(args SendQueryArgs) (*BatchTimeResult, error) {

	commitTime, batchTime := core.CalcCommitBatchAVGTime(args.From.ToUint64(), args.To.ToUint64())

	return &BatchTimeResult{
		CommitTime: commitTime,
		BatchTime:  batchTime,
	}, nil
}

// QueryEvmAvgTime returns EVM average time between from block and to block
func (blk *PublicBlockAPI) QueryEvmAvgTime(args SendQueryArgs) (int64, error) {

	evmTime := core.CalcEvmAVGTime(args.From.ToUint64(), args.To.ToUint64())
	log.Info("-----evmTime----", evmTime)

	return evmTime, nil
}

//func (blk *PublicBlockAPI)QueryTransactionSum() string {
//	sum := core.CalTransactionSum()
//	return sum.String()
//}
func (blk *PublicBlockAPI) QueryTransactionSum() uint64 {
	sum := core.CalTransactionSum()
	return sum
}
