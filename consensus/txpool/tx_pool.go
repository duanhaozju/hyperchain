//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import (
	"crypto/md5"
	"encoding/hex"

	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/manager/event"

	"github.com/op/go-logging"
)

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They exit the pool when they are included in the blockchain.
//
// The pool separates processable transactions (which can be applied to the
// current state) and future transactions. Transactions move between those
// two states over time as they are received and processed.
type TxPool interface {
	GenerateTxBatch() error
	AddNewTx(tx *types.Transaction, isPrimary bool, checkPool bool) (bool, error)
	RemoveBatchedTxs(hashList []string) error
	RemoveOneBatchedTxs(hash string) error
	IsPoolFull() bool
	HasTxInPool() bool
	GetTxsBack(hashList []string) error
	GetOneTxsBack(hash string) error
	GetTxsByHashList(id string, hashList []string) (txs []*types.Transaction, missingTxsHash []string, err error)
	ReturnFetchTxs(id string, missingHashList []string) (txs []*types.Transaction, err error)
	GotMissingTxs(id string, txs []*types.Transaction) error
}

// TxHashBatch contains transactions that batched by primary.
type TxHashBatch struct {
	BatchHash  string
	TxHashList []string
	TxList     []*types.Transaction
}

type MissingTxHashList struct {
	TxHashList []string
	BatchHash  string
}

type ReturnFetchTxs struct {
	BatchHash        string
	ReturnedFetchTxs []*types.Transaction
}

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

// txPoolImpl implement the txpool
type txPoolImpl struct {
	txPool         map[string]*types.Transaction // store all non-batched txs
	txPoolHash     []string                      // store all non-batched txs' hash by order
	batchStore     []*TxHashBatch                // store all batches created by current primary in order, removed in
	                                             // viewchange as new primary may create batches in other order
	batchedTxs     map[string]bool               // store batched txs' hash corresponding to batchStore
	missingTxs     map[string][]string           // store missing txs' hash using missing batch's id as key
	poolSize       int                           // upper limit of txPool
	queue          *event.TypeMux                // when we generate a batch, we would post it to this channel
	batchSize      int                           // a batch contains how many transactions
	logger         *logging.Logger
}

// NewTxPool creates a new transaction pool
func NewTxPool(namespace string, poolsize int, queue *event.TypeMux, batchsize int) (TxPool, error) {
	return newTxPoolImpl(namespace, poolsize, queue, batchsize)
}

// AddNewTx adds a transaction to txPool, and when current node is primary, isPrimary should be true. checkPool is used
// to guarantee cached txs' size in tx pool, when we receive txs from local(which is sent from client), we need to check
// pool size before add txs into tx pool, else, when we receive txs from other nodes(which has been added to other's
// tx pool), we need not to check pool size as this tx has go through hpc layer, must be committed to blockchain.
func (pool *txPoolImpl) AddNewTx(tx *types.Transaction, isPrimary bool, checkPool bool) (bool, error) {
	if isPrimary {
		return pool.primaryAddNewTx(tx, checkPool)
	} else {
		return pool.replicaAddNewTx(tx, checkPool)
	}
}

// addTxs attempts to queue a batch of transactions into the pool without check pool size after receiving missing txs
// from primary.
func (pool *txPoolImpl) addTxs(txs []*types.Transaction) error {
	tmpPool := pool.txPool
	tmpPoolHash := pool.txPoolHash
	for _, tx := range txs {
		isDuplicate := false
		txHash := tx.GetHash().Hex()

		// find tx in batchStore(batched txs)
		if _, ok := pool.batchedTxs[txHash]; ok {
			pool.logger.Warningf("Duplicate transaction with hash : %s which has been batched", txHash)
			return ErrDuplicateTx
		}

		// find tx in txPool(non-batched txs)
		if pool.txPool[txHash] != nil {
			pool.logger.Debugf("Duplicate transaction with hash : %s, may be caused by receiving certain tx" +
				"during fetching missing tx from primary", txHash)
			isDuplicate = true
		}

		if isDuplicate == false {
			tmpPool[txHash] = tx
			tmpPoolHash = append(tmpPoolHash, txHash)
		}
	}
	pool.txPool = tmpPool
	pool.txPoolHash = tmpPoolHash
	pool.logger.Debugf("Replica add transactions, and there are %d transactions in txpool",
		len(pool.txPool))
	return nil
}

// When sth like view change happens, consensus module should post this event to stop a running batch timer
func (pool *txPoolImpl) GenerateTxBatch() error {
	return pool.generateTxBatch()
}

// GetTxsByHashList returns the transaction list found by given hash list.
// When replicas receive hashList from primary, they need to generate batches the same as primary.
// If they miss some transactions, they need to fetch these transactions from primary.
func (pool *txPoolImpl) GetTxsByHashList(id string, hashList []string) (txs []*types.Transaction, missingTxsHash []string, err error) {
	var hasMissing bool
	if batch, e := pool.getBatchById(id); e == nil { // If replica already has this batch, return
		pool.logger.Noticef("Replica already has this batch, id: %s", id)
		txs = batch.TxList
		missingTxsHash = nil
		return
	}
	// If we have check this batch and find we miss some transactions,
	// just return the same missingTxsHash as before
	if miss, ok := pool.missingTxs[id]; ok {
		txs = nil
		missingTxsHash = miss
		return
	}
	for _, hash := range hashList {
		_, ok := pool.batchedTxs[hash] // If this transaction has been batched
		if ok {
			pool.logger.Warningf("Duplicate transaction with hash : %s", hash)
			err = ErrDuplicateTx
			return
		}
		if pool.txPool[hash] != nil { // If this node has this transaction
			if !hasMissing {
				txs = append(txs, pool.txPool[hash])
			}
		} else {
			pool.logger.Infof("Can't find tx by hash: %s from txpool", hash)
			hasMissing = true
			missingTxsHash = append(missingTxsHash, hash)
		}
	}

	// fetch missing txs if found missing txs from txpool
	if hasMissing {
		txs = nil
		pool.missingTxs[id] = missingTxsHash
		return
	} else {
		batch := &TxHashBatch{
			BatchHash:  id,
			TxHashList: hashList,
			TxList:     txs,
		}
		pool.removeTxPoolTxs(hashList)
		pool.batchStore = append(pool.batchStore, batch)
		for _, hash := range batch.TxHashList {
			pool.batchedTxs[hash] = true
		}
		pool.logger.Debugf("Replica generate a transaction batch by hash list, which digest is %s, and now there are %d "+
			"pending transactions and %d batches in txPool", id, len(pool.txPool), len(pool.batchStore))
		return
	}
}

// When replica miss some transactions and ask for these transactions, primary could use ReturnFetchTxs to
// fetch these transactions.
func (pool *txPoolImpl) ReturnFetchTxs(id string, missingHashList []string) (txs []*types.Transaction, err error) {
	// if this node doesn't have this batch, there is an error.
	if batch, e := pool.getBatchById(id); e != nil {
		err = e
		return
	} else {
		i := 0
		// If all transactions in missingHashList are in this batch, they should keep the order
		// So we can find them all by scanning this batch in order
		batchLen := len(batch.TxHashList)
		for _, hash := range missingHashList {
			for i < batchLen {
				if batch.TxHashList[i] == hash {
					txs = append(txs, batch.TxList[i])
					i++
					break
				}
				i++
			}

			if i == batchLen && len(txs) != len(missingHashList) {
				err = ErrMismatch
				return
			}
		}
		return
	}
}

// GotMissingTxs receives txs fetched from primary and add txs to txpool
func (pool *txPoolImpl) GotMissingTxs(id string, txs []*types.Transaction) error {
	if _, ok := pool.missingTxs[id]; !ok {
		pool.logger.Errorf("Received missing txs, but can't find corresponding batch hash: %s", id)
		return ErrNoBatch
	}
	if len(txs) != len(pool.missingTxs[id]) {
		return ErrMismatch
	}
	for i, tx := range txs {
		txHash := tx.GetHash().Hex()
		if txHash != pool.missingTxs[id][i] {
			pool.logger.Warningf("Received missing txs, but find an unmatch tx hash: %s", txHash)
			return ErrMismatch
		}
	}

	if err := pool.addTxs(txs); err != nil {
		return err
	}
	delete(pool.missingTxs, id)
	return nil
}

// removeTxBatch removes several batches by given digests of transaction batches from the pool(batchedTxs).
func (pool *txPoolImpl) RemoveBatchedTxs(hashList []string) error {
	hashMap := make(map[string]bool)
	for _, hash := range hashList { // store hash of batch which needs to be removed
		hashMap[hash] = true
	}
	// if a batch doesn't need to be removed, store it in newBatchedTxs.
	// And let newBatchedTxs be batchStore finally
	var newBatchedTxs []*TxHashBatch
	for _, batch := range pool.batchStore {
		if _, ok := hashMap[batch.BatchHash]; !ok {
			newBatchedTxs = append(newBatchedTxs, batch)
		} else {
			for _, hash := range batch.TxHashList {
				delete(pool.batchedTxs, hash)
			}
		}
	}
	pool.batchStore = newBatchedTxs
	pool.logger.Debugf("Replica removes some batches in txPool, and now there are"+
		" %d batches in txPool", len(pool.batchStore))
	return nil
}

// RemoveOneTxBatch removes one batch by given digest of transaction batch from the pool(batchedTxs).
func (pool *txPoolImpl) RemoveOneBatchedTxs(hash string) error {
	find := false
	index := 0
	for _, batch := range pool.batchStore {
		if batch.BatchHash == hash {
			find = true
			break
		}
		index++
	}
	if find {
		batch := pool.batchStore[index]
		for _, hash := range batch.TxHashList {
			delete(pool.batchedTxs, hash)
		}
		pool.batchStore = append(pool.batchStore[:index], pool.batchStore[index+1:]...)
		pool.logger.Debugf("Replica removes one transaction batch, which hash is %s, and now there are "+
			"%d batches in txPool", hash, len(pool.batchStore))
	} else {
		return ErrNoTxHash
	}
	return nil
}

// GetTxsBack move some batch in batchStore to txpool
func (pool *txPoolImpl) GetTxsBack(hashList []string) error {
	var batches []*TxHashBatch
	for _, hash := range hashList {
		if batch, e := pool.getBatchById(hash); e != nil {
			return e
		} else {
			batches = append(batches, batch)
		}
	}
	pool.RemoveBatchedTxs(hashList)
	var newTxPoolHash []string
	for _, batch := range batches {
		newTxPoolHash = append(newTxPoolHash, batch.TxHashList...)
		for _, tx := range batch.TxList {
			pool.txPool[tx.GetHash().Hex()] = tx
		}
	}
	pool.txPoolHash = append(newTxPoolHash, pool.txPoolHash...)
	return nil
}

// GetOneTxsBack move one batch in batchStore to txpool
func (pool *txPoolImpl) GetOneTxsBack(hash string) error {
	batch, err := pool.getBatchById(hash)
	if err != nil {
		return err
	}
	pool.RemoveOneBatchedTxs(hash)
	pool.txPoolHash = append(batch.TxHashList, pool.txPoolHash...)
	for _, tx := range batch.TxList {
		pool.txPool[tx.GetHash().Hex()] = tx
	}
	return nil
}

// newTxPoolImpl creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func newTxPoolImpl(namespace string, poolsize int, queue *event.TypeMux, batchsize int) (*txPoolImpl, error) {
	txpool := &txPoolImpl{
		poolSize:  poolsize,
		queue:     queue,
		batchSize: batchsize,
	}
	txpool.txPool = make(map[string]*types.Transaction)
	txpool.txPoolHash = nil
	txpool.batchStore = nil
	txpool.batchedTxs = make(map[string]bool)
	txpool.missingTxs = make(map[string][]string)
	txpool.logger = common.GetLogger(namespace, "consensus")
	return txpool, nil
}

// primaryAddNewTx enqueues a single new transaction into the pool with check pool size and batch size.
func (pool *txPoolImpl) primaryAddNewTx(tx *types.Transaction, checkPool bool) (bool, error) {
	if checkPool {
		if pool.IsPoolFull() {
			pool.logger.Warningf("Reach the upper limit of txpool")
			return false, ErrPoolFull
		}
	}
	txHash := tx.GetHash().Hex()
	if _, ok := pool.batchedTxs[txHash]; ok {
		pool.logger.Warningf("Duplicate transaction with hash : %s in batched txs", txHash)
		return false, ErrDuplicateTx
	}

	if pool.txPool[txHash] != nil {
		pool.logger.Warningf("Duplicate transaction with hash : %s in tx pool", txHash)
		return false, ErrDuplicateTx
	}
	pool.txPool[txHash] = tx
	pool.txPoolHash = append(pool.txPoolHash, txHash)

	isGenerated := false
	for len(pool.txPool) >= pool.batchSize {
		pool.logger.Debugf("Reach batch size, generate a batch")
		err := pool.generateTxBatch()
		if err == nil {
			isGenerated = true
		}
	}

	return isGenerated, nil
}

// replicaAddNewTx enqueues a single new transaction into the pool with check pool size, but don't generate a new batch
func (pool *txPoolImpl) replicaAddNewTx(tx *types.Transaction, checkPool bool) (bool, error) {
	if checkPool {
		if pool.IsPoolFull() {
			pool.logger.Warningf("Reach the upper limit of txpool")
			return false, ErrPoolFull
		}
	}

	txHash := tx.GetHash().Hex()
	if _, ok := pool.batchedTxs[txHash]; ok {
		pool.logger.Warningf("Duplicate transaction with hash : %s in batched txs", txHash)
		return false, ErrDuplicateTx
	}

	if pool.txPool[txHash] != nil {
		pool.logger.Warningf("Duplicate transaction with hash : %s in tx pool", txHash)
		return false, ErrDuplicateTx
	}
	pool.txPool[txHash] = tx
	pool.txPoolHash = append(pool.txPoolHash, txHash)
	return false, nil
}

// IsPoolFull check is txPool is full(including batched txs)
func (pool *txPoolImpl) IsPoolFull() bool {
	length := len(pool.txPool) + len(pool.batchedTxs)
	return length >= pool.poolSize
}

// HasTxInPool checks if there exists tx(s) in tx pool
func (pool *txPoolImpl) HasTxInPool() bool {
	length := len(pool.txPool)
	return length > 0
}

// generateTxBatch generates a transaction batch by batch limit (timeout or size).
func (pool *txPoolImpl) generateTxBatch() error {
	poolLen := len(pool.txPool)
	if poolLen == 0 {
		return ErrEmptyFull
	} else {
		batch := pool.newTxBatch()
		if batch != nil {
			pool.postTxBatch(*batch)
		}
	}
	return nil
}

// postTxBash post a batch to chan which should be listened by consensus module
func (pool *txPoolImpl) postTxBatch(msg TxHashBatch) error {
	pool.queue.Post(msg)
	return nil
}

// newTxBatch creates a new transaction batch to store the transactions.
func (pool *txPoolImpl) newTxBatch() *TxHashBatch {
	var hashList []string
	var txList []*types.Transaction
	// if txPool stores transactions more than batchSize,
	// use the first batchSize transactions to generate a batch
	if poolLen := len(pool.txPool); poolLen > pool.batchSize {
		hashList = pool.txPoolHash[:pool.batchSize]
		pool.txPoolHash = pool.txPoolHash[pool.batchSize:]
	} else {
		hashList = pool.txPoolHash
		pool.txPoolHash = nil
	}
	for _, hash := range hashList {
		if tx, ok := pool.txPool[hash]; !ok {
			pool.logger.Errorf("Can't find transaction by hash %s in txPool", hash)
			return nil
		} else {
			txList = append(txList, tx)
			delete(pool.txPool, hash)
			pool.batchedTxs[hash] = true
		}
	}

	txbatch := &TxHashBatch{
		TxHashList: hashList,
		TxList:     txList,
	}
	batchHash := hash(txbatch)
	txbatch.BatchHash = batchHash
	pool.batchStore = append(pool.batchStore, txbatch)
	pool.logger.Debugf("Primary generate a transaction batch with %d txs, which hash is %s, and now there are %d "+
		"pending transactions and %d batches in txPool", len(hashList), batchHash, len(pool.txPool), len(pool.batchStore))
	return txbatch
}

// removeTxPoolTxs removes all hash in hashList from the pool(txPoolHash)
func (pool *txPoolImpl) removeTxPoolTxs(hashList []string) error {
	hashMap := make(map[string]bool)
	for _, hash := range hashList {
		hashMap[hash] = true
		delete(pool.txPool, hash)
	}
	var newPoolHash []string
	for _, poolHash := range pool.txPoolHash {
		if _, ok := hashMap[poolHash]; !ok {
			newPoolHash = append(newPoolHash, poolHash)
		}
	}
	pool.txPoolHash = newPoolHash
	return nil
}

// getBatchById find batch by id
func (pool *txPoolImpl) getBatchById(id string) (*TxHashBatch, error) {
	for _, batch := range pool.batchStore {
		if hash(batch) == id {
			return batch, nil
		}
	}
	return nil, ErrNoBatch
}


func hash(batch *TxHashBatch) string {
	h := md5.New()
	for _, hash := range batch.TxHashList {
		h.Write([]byte(hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}