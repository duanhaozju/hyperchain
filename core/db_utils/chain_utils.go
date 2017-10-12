package db_utils

import (
	"errors"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
	"sync"
)

var ChainNotExistErr = errors.New("chain doesn't exist")

// memChain manage safe chain
type memChain struct {
	data    types.Chain      // chain
	lock    sync.RWMutex     // the lock of chain
	cpChan  chan types.Chain // when data.Height reach check point, will be writed
	txDelta uint64
}

type memChains struct {
	chains map[string]*memChain
	lock   sync.RWMutex
}

func NewMemChains() *memChains {
	return &memChains{
		chains: make(map[string]*memChain),
	}
}

func (chains *memChains) AddChain(namespace string, chain *memChain) error {
	chains.lock.Lock()
	defer chains.lock.Unlock()
	chains.chains[namespace] = chain
	return nil
}

func (chains *memChains) GetChain(namespace string) *memChain {
	chains.lock.RLock()
	defer chains.lock.RUnlock()
	return chains.chains[namespace]
}

var (
	chains         *memChains
	chainsInitOnce sync.Once
)

// newMenChain new a memChain instance
// it read from db firstly, if not exist, create a empty chain
func InitializeChain(namespace string) *memChain {
	chainsInitOnce.Do(func() {
		chains = NewMemChains()
	})
	chain, err := getChain(namespace)
	if err == nil {
		memchain := &memChain{
			data:   *chain,
			cpChan: make(chan types.Chain),
		}
		chains.AddChain(namespace, memchain)
		return memchain
	}
	memchain := &memChain{
		data: types.Chain{
			Height: 0,
		},
		cpChan: make(chan types.Chain),
	}
	chains.AddChain(namespace, memchain)
	return memchain
}

// GetHeightOfChain get height of chain.
func GetHeightOfChain(namespace string) uint64 {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return 0
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return chain.data.Height
	}
}

// SetHeightOfChain set height of chain.
func SetHeightOfChain(namespace string, height uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.Height = height
		return nil
	}
}

// GetTxSumOfChain get totally transaction sum of chain.
func GetTxSumOfChain(namespace string) uint64 {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return 0
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return chain.data.CurrentTxSum
	}
}

// SetHeightOfChain set transaction sum of chain.
func SetTxSumOfChain(namespace string, txNum uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.CurrentTxSum = txNum
		return nil
	}
}

// IncreaseTxSumOfChain - increase transaction sum with specify delta.
func IncreaseTxSumOfChain(namespace string, delta uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.CurrentTxSum += delta
		return nil
	}
}

// decreaseTxSumOfChain - decrease transaction sum with specify delta.
func DecreaseTxSumOfChain(namespace string) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.CurrentTxSum -= uint64(chain.txDelta)
		chain.txDelta = 0
		return nil
	}
}

// GetLatestBlockHash get latest blockHash
func GetLatestBlockHash(namespace string) []byte {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return nil
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return chain.data.LatestBlockHash
	}
}

// SetLatestBlockHash set latest blockHash
func SetLatestBlockHash(namespace string, hash []byte) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.LatestBlockHash = hash
		return nil
	}
}

// GetParentBlockHash get the latest block's parentHash
func GetParentBlockHash(namespace string) []byte {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return nil
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return chain.data.ParentBlockHash
	}
}

// SetParentBlockHash set the latest block's parentHash
func SetParentBlockHash(namespace string, hash []byte) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.ParentBlockHash = hash
		return nil
	}
}

// GetTxDeltaOfMemChain get the memChain's txDelta
func GetTxDeltaOfMemChain(namespace string) uint64 {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return 0
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return chain.txDelta
	}
}

// SetTxDeltaOfMemChain set the memChain's txDelta
func AddTxDeltaOfMemChain(namespace string, txDelta uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return EmptyPointerErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.txDelta += txDelta
		return nil
	}
}

// GetChainCopy get copy of chain
func GetChainCopy(namespace string) *types.Chain {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return nil
	} else {
		return &types.Chain{
			LatestBlockHash: chain.data.LatestBlockHash,
			ParentBlockHash: chain.data.ParentBlockHash,
			Height:          chain.data.Height,
			CurrentTxSum:    chain.data.CurrentTxSum,
		}
	}
}

// UpdateChain update latest blockHash as given blockHash
// and the height of chain add 1
func UpdateChain(namespace string, batch db.Batch, block *types.Block, genesis bool, flush bool, sync bool) error {
	SetLatestBlockHash(namespace, block.BlockHash)
	SetParentBlockHash(namespace, block.ParentHash)
	if genesis {
		SetHeightOfChain(namespace, 0)
		SetTxSumOfChain(namespace, 0)
	} else {
		if GetHeightOfChain(namespace) < block.Number {
			err := SetHeightOfChain(namespace, block.Number)
			if err != nil {
				return err
			}
			err = IncreaseTxSumOfChain(namespace, uint64(len(block.Transactions)))
			if err != nil {
				return err
			}
		} else if GetHeightOfChain(namespace) > block.Number {
			err := SetHeightOfChain(namespace, block.Number)
			if err != nil {
				return err
			}
			err = DecreaseTxSumOfChain(namespace)
			if err != nil {
				return err
			}
		}
	}
	chain := chains.GetChain(namespace)
	return putChain(batch, &chain.data, flush, sync)
}

// update chain according block number, set chain current height to the block number
// return error if correspondent block missing
func UpdateChainByBlcokNum(namespace string, batch db.Batch, blockNumber uint64, flush bool, sync bool) error {
	block, err := GetBlockByNumber(namespace, blockNumber)
	if err != nil {
		logger(namespace).Warning("no required block number")
		return err
	}
	return UpdateChain(namespace, batch, block, block.Number == 0, flush, sync)
}

func PutChain(batch db.Batch, chain *types.Chain, flush, sync bool) error {
	return putChain(batch, chain, flush, sync)
}

func UpdateGenesisTag(namespace string, genesis uint64, batch db.Batch, flush bool, sync bool) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ChainNotExistErr
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.Genesis = genesis
		return putChain(batch, &chain.data, flush, sync)
	}
}

func GetGenesisTag(namespace string) (error, uint64) {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ChainNotExistErr, 0
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return nil, chain.data.Genesis
	}
}

func RemoveChain(batch db.Batch, flush bool, sync bool) error {
	if err := batch.Delete(ChainKey); err != nil {
		return err
	}
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// WaitUtilHeightChan get chain from channel. if channel is empty, the func will be blocked
func GetChainUntil(namespace string) *types.Chain {
	chain := chains.GetChain(namespace)
	data := <-chain.cpChan
	return &data
}

// WriteChain to channel, if the channel is not read, will be blocked
func WriteChainChan(namespace string) {
	chain := chains.GetChain(namespace)
	chain.cpChan <- chain.data
}

// putChain put chain database
func putChain(batch db.Batch, t *types.Chain, flush bool, sync bool) error {
	// assign version tag
	data, err := proto.Marshal(t)
	if err != nil {
		return err
	}
	if err := batch.Put(ChainKey, data); err != nil {
		return err
	}
	if flush {
		if sync {
			batch.Write()
		} else {
			go batch.Write()
		}
	}
	return nil
}

// getChain - get chain from database.
func getChain(namespace string) (*types.Chain, error) {
	db, err := hyperdb.GetDBDatabaseByNamespace(namespace)
	if err != nil {
		return nil, err
	}
	return getChainFn(db)
}

func getChainFn(db db.Database) (*types.Chain, error) {
	if db == nil {
		// short circuit if db is empty
		return nil, errors.New("empty db")
	}
	var chain types.Chain
	data, err := db.Get(ChainKey)
	if err != nil {
		return nil, err
	}
	err = proto.Unmarshal(data, &chain)
	if err != nil {
		return nil, err
	}
	return &chain, nil
}

func GetChain(namespace string) (*types.Chain, error) {
	return getChain(namespace)
}
