package txpool

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/crypto"
	"sort"
	"sync"
)

// Validate checks validity of transactions in given batch.
func (pool *txPoolImpl) Validate(id string) (validTxs []*types.Transaction, invalidRecord []*types.InvalidTransactionRecord, invalidTxsHash string) {
	var (
		invalidTxsHashList []string
		wg                 sync.WaitGroup
		index              []uint64
		mu                 sync.Mutex
	)
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	encryption := crypto.NewEcdsaEncrypto("ecdsa")

	batch, _, err := pool.getBatchById(id)
	if err != nil {
		pool.logger.Errorf("Cannot find batch in batchStore with id: %s", id)
		// TODO return err?
		return
	}

	txs := batch.TxList
	// Parallel check the signature of each transaction
	for i := range txs {
		wg.Add(1)
		go func(i uint64) {
			tx := txs[i]
			if !tx.ValidateSign(encryption, kec256Hash) {
				pool.logger.Warningf("Found invalid signature, send from : %v", tx.Id)
				mu.Lock()
				invalidRecord = append(invalidRecord, &types.InvalidTransactionRecord{
					Tx:      tx,
					ErrType: types.InvalidTransactionRecord_SIGFAILED,
					ErrMsg:  []byte("Invalid signature"),
				})
				index = append(index, i)
				mu.Unlock()
			}
			wg.Done()
		}(uint64(i))
	}
	wg.Wait()

	// Remove invalid transaction from transaction list
	// Keep the txs sequentially as origin
	if len(index) > 0 {
		sort.Sort(common.SortableUint64Slice(index))
		count := uint64(0)
		for _, idx := range index {
			invalidTxsHashList = append(invalidTxsHashList, batch.TxHashList[idx])
			idx = idx - count
			txs = append(txs[:idx], txs[idx+1:]...)
			// TODO in case rollback?
			count++
		}
	}

	invalidTxsHash = hash(invalidTxsHashList)
	validTxs = txs
	return
}
