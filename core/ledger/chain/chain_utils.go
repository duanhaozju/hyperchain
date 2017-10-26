package chain

import (
	"errors"
	"sync"

	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"hyperchain/hyperdb/db"
    "github.com/golang/protobuf/proto"
)

// ChainNotExistErr is thrown by UpdateGenesisTag or GetGenesisTag call if chain is nil
var ChainNotExistErr = errors.New("chain doesn't exist")

// memChain manages safe chain
type memChain struct {
	data    types.Chain      // Chain
	lock    sync.RWMutex     // Lock of chain
	cpChan  chan types.Chain // Will be written when data.Height reach check point
	txDelta uint64           // Transaction delta of a memChain
}

// memChains manages the map the memChain with corresponding namespace.
type memChains struct {
	chains map[string]*memChain
	lock   sync.RWMutex
}

// newMenChain constructs new memChain instance.
// It read from db firstly, if not exist, create a empty chain.
func NewMemChains() *memChains {
	return &memChains{
		chains: make(map[string]*memChain),
	}
}

// AddChain adds the chain to memChains with given namespace.
func (chains *memChains) AddChain(namespace string, chain *memChain) error {
	chains.lock.Lock()
	defer chains.lock.Unlock()
	chains.chains[namespace] = chain
	return nil
}



// GetChain returns the chain with given namespace.
func (chains *memChains) GetChain(namespace string) *memChain {
	chains.lock.RLock()
	defer chains.lock.RUnlock()
	return chains.chains[namespace]
}

var (
	chains         iChain // Singleton chains of memChains
	chainsInitOnce sync.Once  // Ensure chains will only be created once
)

func GetMemChain(namespace string) *types.MemChain {
    chain := chains.(*memChains).chains[namespace]
    cpChan := <-chain.cpChan
    memChain := &types.MemChain{
        Data: &chain.data,
        CpChan: &cpChan,
        TxDelta: chain.txDelta,
    }
    return memChain
}

// ==========================================================
// Public functions that invoked by outer service
// ==========================================================

// InitializeChain inits the chains instance at first time,
// and adds a memChain with given namespace.
func InitializeChain(namespace string) *memChain {
	// Construct the chains only at the first time
	chainsInitOnce.Do(func() {
		chains = NewMemChains()
	})

	chain, err := getChain(namespace)
	if err != nil {
		// Database not found, create an empty memChain, which can be used in test
		memChain := &memChain{
			data: types.Chain{
				Height: 0,
			},
			cpChan: make(chan types.Chain),
		}
		chains.AddChain(namespace, memChain)
		return memChain
	} else {
		memChain := &memChain{
			data:   *chain,
			cpChan: make(chan types.Chain),
		}
		chains.AddChain(namespace, memChain)
		return memChain
	}
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

// GetTxSumOfChain gets the sum of transactions in chain with given namespace.
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

// GetLatestBlockHash gets the latest blockHash with given namespace.
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

// GetParentBlockHash gets the latest block's parentHash with given namespace.
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

// GetTxDeltaOfMemChain gets the memChain's txDelta with given namespace.
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

// AddTxDeltaOfMemChain adds the memChain's txDelta with given namespace.
func AddTxDeltaOfMemChain(namespace string, txDelta uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.txDelta += txDelta
		return nil
	}
}

// GetChainCopy gets the copy of chain with given namespace.
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

// UpdateChain updates latest blockHash with given parameters,
// and adds 1 to the height of chain.
func UpdateChain(namespace string, batch db.Batch, block *types.Block, genesis bool, flush bool, sync bool) error {
	// Set the latest and parent block hash
	setLatestBlockHash(namespace, block.BlockHash)
	setParentBlockHash(namespace, block.ParentHash)
	if genesis {
		// For genesis block, set the height and txSum to 0
		setHeightOfChain(namespace, 0)
		setTxSumOfChain(namespace, 0)
	} else {
		if GetHeightOfChain(namespace) < block.Number {
			err := setHeightOfChain(namespace, block.Number)
			if err != nil {
				return err
			}
			err = increaseTxSumOfChain(namespace, uint64(len(block.Transactions)))
			if err != nil {
				return err
			}
		} else if GetHeightOfChain(namespace) > block.Number {
			// TODO: how could this happen?
			err := setHeightOfChain(namespace, block.Number)
			if err != nil {
				return err
			}
			err = decreaseTxSumOfChain(namespace)
			if err != nil {
				return err
			}
		}
	}
	chain := chains.GetChain(namespace)
	return putChain(batch, &chain.data, flush, sync)
}

// PutChain persists chain with given data.
func PutChain(batch db.Batch, chain *types.Chain, flush, sync bool) error {
	return putChain(batch, chain, flush, sync)
}

// UpdateChainByBlcokNum updates chain according block number, sets chain current height to the block number,
// returns error if correspondent block missing.
func UpdateChainByBlcokNum(namespace string, batch db.Batch, blockNumber uint64, flush bool, sync bool) error {
	block, err := GetBlockByNumber(namespace, blockNumber)
	if err != nil {
		logger(namespace).Warning("no required block number")
		return err
	}
	return UpdateChain(namespace, batch, block, block.Number == 0, flush, sync)
}

// UpdateGenesisTag updates the genesis tag with given parameters.
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

// GetGenesisTag gets the genesis tag of the chain with given namespace.
func GetGenesisTag(namespace string) (uint64, error) {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return 0, ChainNotExistErr
	} else {
		chain.lock.RLock()
		defer chain.lock.RUnlock()
		return chain.data.Genesis, nil
	}
}

// RemoveChain removes the chain and flushes it according to parameters.
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

// GetChainUntil gets the chain from channel, if channel is empty, the func will be blocked.
// It is used in the interaction between executor and consenter.
func GetChainUntil(namespace string) *types.Chain {
	chain := chains.GetChain(namespace)
	data := <-chain.cpChan
	return &data
}

// WriteChainChan writes the written chain to channel, if the channel is not read, will be blocked.
func WriteChainChan(namespace string) {
	chain := chains.GetChain(namespace)
	chain.cpChan <- chain.data
}

// GetChain gets the chain with given namespace.
func GetChain(namespace string) (*types.Chain, error) {
	return getChain(namespace)
}

// ==========================================================
// Private functions that invoked by inner service
// ==========================================================

// setHeightOfChain sets the height of chain with given namespace.
func setHeightOfChain(namespace string, height uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.Height = height
		return nil
	}
}

// setHeightOfChain sets the sum of transactions in chain with given namespace.
func setTxSumOfChain(namespace string, txNum uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.CurrentTxSum = txNum
		return nil
	}
}

// increaseTxSumOfChain increases the sum of transactions in chain with given namespace.
func increaseTxSumOfChain(namespace string, delta uint64) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.CurrentTxSum += delta
		return nil
	}
}

// decreaseTxSumOfChain decreases the sum of transactions in chain with given namespace.
func decreaseTxSumOfChain(namespace string) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.CurrentTxSum -= uint64(chain.txDelta)
		chain.txDelta = 0
		return nil
	}
}

// setLatestBlockHash sets latest blockHash in chain with given namespace.
func setLatestBlockHash(namespace string, hash []byte) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.LatestBlockHash = hash
		return nil
	}
}

// setParentBlockHash sets the latest block's parentHash in chain with given namespace.
func setParentBlockHash(namespace string, hash []byte) error {
	chain := chains.GetChain(namespace)
	if chain == nil {
		return ErrEmptyPointer
	} else {
		chain.lock.Lock()
		defer chain.lock.Unlock()
		chain.data.ParentBlockHash = hash
		return nil
	}
}

// putChain put chain into database
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
