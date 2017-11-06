//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import (
	"crypto/md5"
	"encoding/hex"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/op/go-logging"
)

// TxPool contains all currently known transactions. Transactions
// enter the pool when they are received from the network or submitted
// locally. They are deleted from the pool when they are included in
// the blockchain.
//
// The pool separates transactions being processed and transactions
// waiting to be processed. Transactions would be moved between those
// two states over time.
type TxPool interface {
	// GenerateTxBatch generates a transaction batch and post it
	// to outside if there are transactions in txPool.
	GenerateTxBatch() error

	// AddNewTx adds a transaction to txPool, and when current node
	// is primary, isPrimary should be true. checkPool is used to
	// guarantee cached txs' size in tx pool to be smaller than poolSize.
	// When we receive txs from local(which is sent from client), we need
	// to check pool size before add txs into tx pool. When we receive txs
	// from other nodes(which have been added to other's tx pool),
	// we need not to check pool size as these txs has go through hpc layer,
	// must be committed to blockchain.
	AddNewTx(tx *types.Transaction, isPrimary bool, checkPool bool) (bool, error)

	// RemoveTxBatch removes several batches by given digests of
	// transaction batches from the pool(batchedTxs).
	RemoveBatches(hashList []string) error

	// IsPoolFull check is txPool is full(including batched txs)
	IsPoolFull() bool

	// HasTxInPool checks if there is tx(s) in tx pool or not
	HasTxInPool() bool

	// GetBatchesBackExcept move some batch in batchStore to txPool
	GetBatchesBackExcept(hashList []string) error

	// GetOneBatchBack move one batch in batchStore to txPool
	GetOneBatchBack(hash string) error

	// PutBatchIntoPending puts one batch from batchStore into pendingBatches if
	// rbft reached validateCount.
	PutBatchIntoPending(hash string) error

	// GetTxsByHashList returns the transaction list found by given hash list.
	// When replicas receive hashList from primary, they need to generate
	// batches the same as primary.
	// 1. If this batch has been batched, just return its transactions without error.
	// 2. If we have check this batch and find we miss some transactions, just return
	//    the same missingTxsHash as before without error.
	// 3. If one transaction in hashList has been batched before in another batch,
	//    return ErrDuplicateTx
	// 4. If this node miss some transactions, it need to fetch these transactions
	//    from primary, and return missingTxsHash without error
	// 5. If this node get all transactions, generate a batch and return its
	//    transactions without error
	GetTxsByHashList(id string, hashList []string) (txs []*types.Transaction, missingTxsHash map[uint64]string, err error)

	// ReturnFetchTxs find this batch in this node it self, and return
	// transactions whose hash are in missingHashList. If there is no
	// such batch, return ErrNoBatch. If there is such a batch, but it
	// doesn't contain all transactions in missingHashList, return ErrMismatch.
	// When replica miss some transactions and ask for these transactions,
	// primary could use ReturnFetchTxs to return these transactions.
	ReturnFetchTxs(id string, missingHashList map[uint64]string) (txs map[uint64]*types.Transaction, err error)

	// GotMissingTxs receives txs fetched from primary and add txs to txPool
	GotMissingTxs(id string, txs map[uint64]*types.Transaction) error
}

// TxHashBatch contains transactions that batched by primary.
type TxHashBatch struct {
	BatchHash  string
	TxHashList []string
	TxList     []*types.Transaction
}

// txPoolImpl implements the TxPool interface
type txPoolImpl struct {
	// store all non-batched txs
	txPool map[string]*types.Transaction

	// store all non-batched txs' hash by order
	txPoolHash []string

	// store all batches created by current primary in order, removed in viewchange as
	// new primary may create batches in other order
	batchStore []*TxHashBatch

	// pending batches which have been batched by primary but not sent to validate
	pendingBatches []*TxHashBatch

	// store batched txs' hash corresponding to batchStore
	batchedTxs map[string]bool

	// store missing txs' hash using missing batch's id as key
	missingTxs map[string]map[uint64]string

	// upper limit of txPool
	poolSize int

	// when we generate a batch, we would post it to this channel
	queue *event.TypeMux

	// a batch contains how many transactions
	batchSize int

	logger *logging.Logger
}

// NewTxPool creates a new transaction pool
func NewTxPool(namespace string, poolsize int, queue *event.TypeMux, batchsize int) (TxPool, error) {

	return newTxPoolImpl(namespace, poolsize, queue, batchsize)
}

// AddNewTx adds a transaction to txPool, and when current node is primary,
// isPrimary should be true. checkPool is used to guarantee cached txs' size
// in tx pool to be smaller than poolSize. When we receive txs from local(which
// is sent from client), we need to check pool size before add txs into tx pool.
// When we receive txs from other nodes(which have been added to other's tx pool),
// we need not to check pool size as these txs has go through hpc layer, must be
// committed to blockchain.
func (pool *txPoolImpl) AddNewTx(tx *types.Transaction, isPrimary bool, checkPool bool) (bool, error) {

	if isPrimary {
		return pool.primaryAddNewTx(tx, checkPool)
	} else {
		return pool.replicaAddNewTx(tx, checkPool)
	}
}

// addTxs attempts to queue a batch of transactions into the pool
// without check pool size after receiving missing txs from primary.
func (pool *txPoolImpl) addTxs(txs []*types.Transaction) error {

	tmpPool := pool.txPool
	tmpPoolHash := pool.txPoolHash
	for _, tx := range txs {
		isDuplicate := false
		txHash := tx.GetHash().Hex()

		// find tx in batchStore(batched txs)
		if _, ok := pool.batchedTxs[txHash]; ok {
			pool.logger.Warningf("Duplicate transaction in addTxs with hash: %s which has been batched", txHash)
			return ErrDuplicateTx
		}

		// find tx in txPool(non-batched txs)
		if _, ok := pool.txPool[txHash]; ok {
			pool.logger.Debugf("Duplicate transaction in addTxs with hash: %s, may be caused by receiving certain tx"+
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
	pool.logger.Debugf("Replica add transactions, and there are %d transactions in txPool", len(pool.txPool))
	pool.logger.Debugf("There are %d txHash in txPool", len(pool.txPoolHash))
	return nil
}

// GenerateTxBatch generates a transaction batch and post it to outside
// if there are transactions in txPool.
func (pool *txPoolImpl) GenerateTxBatch() error {

	if pool.postTxBatchIfHasPending() {
		return nil
	}
	return pool.generateTxBatch()
}

// GetTxsByHashList returns the transaction list found by given hash list.
// When replicas receive hashList from primary, they need to generate
// batches the same as primary.
// 1. If this batch has been batched, just return its transactions without error.
// 2. If we have check this batch and find we miss some transactions, just return
//    the same missingTxsHash as before without error.
// 3. If one transaction in hashList has been batched before in another batch,
//    return ErrDuplicateTx
// 4. If this node miss some transactions, it need to fetch these transactions
//    from primary, and return missingTxsHash without error
// 5. If this node get all transactions, generate a batch and return its
//    transactions without error
func (pool *txPoolImpl) GetTxsByHashList(id string, hashList []string) (txs []*types.Transaction, missingTxsHash map[uint64]string, err error) {

	missingTxsHash = make(map[uint64]string)
	var hasMissing bool
	if batch, _, e := pool.getBatchById(id); e == nil { // If replica already has this batch, return
		pool.logger.Debugf("Replica already has this batch, id: %s", id)
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
	for i, hash := range hashList {
		_, ok := pool.batchedTxs[hash] // If this transaction has been batched
		if ok {
			pool.logger.Warningf("Duplicate transaction in getTxsByHashList with hash: %s, batch id: %s", hash, id)
			err = ErrDuplicateTx
			missingTxsHash = nil
			return
		}
		if _, ok := pool.txPool[hash]; ok { // If this node has this transaction
			if !hasMissing {
				txs = append(txs, pool.txPool[hash])
			}
		} else {
			pool.logger.Debugf("Can't find tx by hash: %s from txPool", hash)
			hasMissing = true
			missingTxsHash[uint64(i)] = hash
		}
	}

	// fetch missing txs if found missing txs from txPool
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
		pool.logger.Debugf("There are %d txHash in txPool", len(pool.txPoolHash))
		missingTxsHash = nil
		return
	}
}

// ReturnFetchTxs find this batch in this node it self, and return
// transactions whose hash are in missingHashList. If there is no
// such batch, return ErrNoBatch. If there is such a batch, but it
// doesn't contain all transactions in missingHashList, return ErrMismatch.
// When replica miss some transactions and ask for these transactions,
// primary could use ReturnFetchTxs to return these transactions.
// When replica miss some transactions and ask for these transactions, primary could use ReturnFetchTxs to
// fetch these transactions.
func (pool *txPoolImpl) ReturnFetchTxs(id string, missingHashList map[uint64]string) (txs map[uint64]*types.Transaction, err error) {

	txs = make(map[uint64]*types.Transaction)
	// if this node doesn't have this batch, there is an error.
	if batch, _, e := pool.getBatchById(id); e != nil {
		err = e
		txs = nil
		return
	} else {
		// If all transactions in missingHashList are in this batch, they should keep the order
		// So we can find them all by scanning this batch in order
		batchLen := uint64(len(batch.TxHashList))
		for i, hash := range missingHashList {
			if i >= batchLen || batch.TxHashList[i] != hash {
				txs = nil
				err = ErrMismatch
				return
			}
			txs[i] = batch.TxList[i]
		}
		return
	}
}

// GotMissingTxs receives txs fetched from primary and add txs to txpool
func (pool *txPoolImpl) GotMissingTxs(id string, txs map[uint64]*types.Transaction) error {

	pool.logger.Debugf("Replica received %d missingTxs, batch id : %s", len(txs), id)
	var validTxs []*types.Transaction
	if _, ok := pool.missingTxs[id]; !ok {
		pool.logger.Errorf("Receive missing txs, but can't find corresponding batch hash: %s", id)
		return ErrNoBatch
	}
	if len(txs) != len(pool.missingTxs[id]) {
		pool.logger.Errorf("Receive missing txs, but not match, expect length: %d, received length: %d", len(pool.missingTxs[id]), len(txs))
		return ErrMismatch
	}
	for i, tx := range txs {
		txHash := tx.GetHash().Hex()
		if txHash != pool.missingTxs[id][uint64(i)] {
			pool.logger.Warningf("Receive missing txs, but find a mismatch tx hash: %s", txHash)
			return ErrMismatch
		}
		validTxs = append(validTxs, tx)
	}

	if err := pool.addTxs(validTxs); err != nil {
		return err
	}
	delete(pool.missingTxs, id)
	return nil
}

// RemoveTxBatch removes several batches by given digests of transaction
// batches from the pool(batchedTxs).
func (pool *txPoolImpl) RemoveBatches(hashList []string) error {

	hashMap := make(map[string]bool)
	// store hash of batch which needs to be removed
	for _, hash := range hashList {
		hashMap[hash] = true
	}
	// if a batch doesn't need to be removed, store it in newBatches.
	// And let newBatches be batchStore finally
	var newBatches []*TxHashBatch
	for _, batch := range pool.batchStore {
		if _, ok := hashMap[batch.BatchHash]; !ok {
			newBatches = append(newBatches, batch)
		} else {
			for _, hash := range batch.TxHashList {
				delete(pool.batchedTxs, hash)
			}
		}
	}
	pool.batchStore = newBatches
	pool.logger.Debugf("Replica removes some batches in txPool, and now there are"+
		" %d batches in txPool", len(pool.batchStore))
	return nil
}

// GetBatchesBackExcept move some uncommitted batch in batchStore to txPool in viewchange and updatingN
func (pool *txPoolImpl) GetBatchesBackExcept(hashList []string) error {

	pool.logger.Debugf("Before put back txs except %d batches, there are %d batches in batchStore, " +
		"%d batches in pendingBatches", len(hashList), len(pool.batchStore), len(pool.pendingBatches))

	var (
		removeTxList      []string
		restoreBatches    []*TxHashBatch
		allBatches        []*TxHashBatch
		restoreBatchSet   map[string]bool = make(map[string]bool)
		restoreBatchedTxs map[string]bool = make(map[string]bool)
	)

	for _, hash := range hashList {
		restoreBatchSet[hash] = true
	}

	allBatches = append(pool.batchStore, pool.pendingBatches...)
	for _, batch := range allBatches {
		if _, ok := restoreBatchSet[batch.BatchHash]; ok {
			// Construct the restored batches
			restoreBatches = append(restoreBatches, batch)
			// Construct the batched txs
			for _, hash := range batch.TxHashList {
				restoreBatchedTxs[hash] = true
			}
		} else {
			// Construct the removed txHash list, which will be added again to pool
			removeTxList = append(removeTxList, batch.TxHashList...)
			// Add the txs again to pool
			for _, tx := range batch.TxList {
				pool.txPool[tx.GetHash().Hex()] = tx
			}
		}
	}

	// Reset the state
	pool.batchStore = restoreBatches
	pool.batchedTxs = restoreBatchedTxs
	pool.txPoolHash = append(removeTxList, pool.txPoolHash...)
	pool.pendingBatches = nil
	// clear missingTxs in viewchange or updatingN
	pool.missingTxs = make(map[string]map[uint64]string)

	pool.logger.Debugf("After put back txs, there are %d batches in batchStore, " +
		"%d batches in pendingBatches", len(hashList), len(pool.batchStore), len(pool.pendingBatches))

	return nil
}

// GetOneBatchBack move one batch in batchStore to txPool
func (pool *txPoolImpl) GetOneBatchBack(hash string) error {

	batch, index, err := pool.getBatchById(hash)
	if err != nil {
		pool.logger.Error(err)
		return err
	}

	// remove from batchedTxs and batchStore
	for _, hash := range batch.TxHashList {
		delete(pool.batchedTxs, hash)
	}
	pool.batchStore = append(pool.batchStore[:index], pool.batchStore[index+1:]...)

	pool.logger.Debugf("Replica removes one transaction batch, which hash is %s, and now there are "+
		"%d batches in txPool", hash, len(pool.batchStore))

	// put back into txPool
	pool.txPoolHash = append(batch.TxHashList, pool.txPoolHash...)
	for _, tx := range batch.TxList {
		pool.txPool[tx.GetHash().Hex()] = tx
	}
	return nil
}

func (pool *txPoolImpl) PutBatchIntoPending(hash string) error {

	pool.logger.Debugf("Put batch: %s from batchStore into pendingBatches", hash)
	batch, err := pool.removeBatchById(hash)
	if err != nil {
		pool.logger.Error(err)
		return err
	}
	pool.pendingBatches = append(pool.pendingBatches, batch)
	pool.logger.Debugf("After append, there are %d pending batches", len(pool.pendingBatches))

	return nil
}

// newTxPoolImpl creates a new transaction pool to gather, sort and filter inbound
// transactions from the network.
func newTxPoolImpl(namespace string, poolsize int, queue *event.TypeMux, batchsize int) (*txPoolImpl, error) {

	txPool := &txPoolImpl{
		poolSize:  poolsize,
		queue:     queue,
		batchSize: batchsize,
	}
	txPool.txPool = make(map[string]*types.Transaction)
	txPool.txPoolHash = nil
	txPool.batchStore = nil
	txPool.pendingBatches = nil
	txPool.batchedTxs = make(map[string]bool)
	txPool.missingTxs = make(map[string]map[uint64]string)
	txPool.logger = common.GetLogger(namespace, "consensus")
	return txPool, nil
}

// primaryAddNewTx enqueues a single new transaction into the pool with check pool size and batch size.
func (pool *txPoolImpl) primaryAddNewTx(tx *types.Transaction, checkPool bool) (bool, error) {

	if checkPool {
		if pool.IsPoolFull() {
			pool.logger.Warningf("Reach the upper limit of txPool")
			return false, ErrPoolFull
		}
	}
	txHash := tx.GetHash().Hex()
	if _, ok := pool.batchedTxs[txHash]; ok {
		pool.logger.Warningf("Duplicate transaction in primaryAddNewTx with hash: %s in batched txs", txHash)
		return false, ErrDuplicateTx
	}

	if _, ok := pool.txPool[txHash]; ok {
		pool.logger.Warningf("Duplicate transaction in primaryAddNewTx with hash: %s in tx pool", txHash)
		return false, ErrDuplicateTx
	}
	pool.txPool[txHash] = tx
	pool.txPoolHash = append(pool.txPoolHash, txHash)

	// if there exists pending batches, post those batches and not generate new tx batch.
	if pool.postTxBatchIfHasPending() {
		return false, nil
	}

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
			pool.logger.Warningf("Reach the upper limit of txPool")
			return false, ErrPoolFull
		}
	}

	txHash := tx.GetHash().Hex()
	if _, ok := pool.batchedTxs[txHash]; ok {
		pool.logger.Warningf("Duplicate transaction in replicaAddNewTx with hash : %s in batched txs", txHash)
		return false, ErrDuplicateTx
	}

	if _, ok := pool.txPool[txHash]; ok {
		pool.logger.Warningf("Duplicate transaction in replicaAddNewTx with hash : %s in tx pool", txHash)
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

// HasTxInPool checks if there is tx(s) in tx pool or not
func (pool *txPoolImpl) HasTxInPool() bool {

	length := len(pool.txPool)
	return length > 0
}

// generateTxBatch generates a transaction batch by batch limit (timeout or size).
func (pool *txPoolImpl) generateTxBatch() error {

	poolLen := len(pool.txPool)
	if poolLen == 0 {
		return ErrPoolEmpty
	} else {
		batch := pool.newTxBatch()
		if batch != nil {
			pool.postTxBatch(*batch)
		}
	}
	return nil
}

// postTxBash posts a batch to chan which should be listened by consensus module
// if there exists some batches in pendingBatches, first post those batches by
// order.
func (pool *txPoolImpl) postTxBatch(msg TxHashBatch) error {

	pool.queue.Post(msg)
	return nil
}

// postTxBatchIfHasPending returns true if there exists some batches in pendingBatches
// and post the first batch to rbft, else return false.
func (pool *txPoolImpl) postTxBatchIfHasPending() bool {

	if len(pool.pendingBatches) > 0 {
		batch := pool.pendingBatches[0]
		// remove batch from pendingBatches
		newPending := make([]*TxHashBatch, len(pool.pendingBatches)-1)
		copy(newPending, pool.pendingBatches[1:])
		pool.pendingBatches = newPending

		// add batch into batchStore
		pool.batchStore = append(pool.batchStore, batch)
		pool.queue.Post(*batch)
		pool.logger.Debugf("After post a batch in pending batches, there are %d pending " +
			"batches", len(pool.pendingBatches))
		return true
	}
	return false
}

// newTxBatch creates a new transaction batch to store the transactions.
func (pool *txPoolImpl) newTxBatch() *TxHashBatch {

	var hashList []string
	var txList []*types.Transaction
	// if txPool stores transactions more than batchSize,
	// use the first batchSize transactions to generate a batch
	if poolLen := len(pool.txPool); poolLen > pool.batchSize {
		// here, we must use copy() to get a new slice which points to an address
		// different from the origin pool.txPoolHash, if not, we may encounter a
		// changed hashList if we modify the origin pool.txPoolHash.
		hashList = make([]string, pool.batchSize)
		copy(hashList, pool.txPoolHash[:pool.batchSize])
		pool.txPoolHash = pool.txPoolHash[pool.batchSize:]
	} else {
		hashList = pool.txPoolHash
		pool.txPoolHash = nil
	}
	for _, hash := range hashList {
		if tx, ok := pool.txPool[hash]; !ok {
			pool.logger.Errorf("Can't find transaction by hash %s in txPool", hash)
			pool.txPoolHash = append(hashList, pool.txPoolHash...)
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

// getBatchById find batch together with its index in batchStore, returns error
// if not found id in batchStore.
func (pool *txPoolImpl) getBatchById(id string) (*TxHashBatch, int, error) {

	var batch *TxHashBatch
	index := 0
	find := false
	for _, batch = range pool.batchStore {
		if batch.BatchHash == id {
			find = true
			break
		}
		index++
	}

	if !find {
		return nil, 0, ErrNoBatch
	} else {
		return batch, index, nil
	}
}

// removeBatchById removes batch with given id from batchStore.
func (pool *txPoolImpl) removeBatchById(id string) (*TxHashBatch, error) {

	batch, index, err := pool.getBatchById(id)
	if err != nil {
		return nil, err
	}
	pool.batchStore = append(pool.batchStore[:index], pool.batchStore[index+1:]...)
	return batch, nil
}

func hash(batch *TxHashBatch) string {

	h := md5.New()
	for _, hash := range batch.TxHashList {
		h.Write([]byte(hash))
	}
	return hex.EncodeToString(h.Sum(nil))
}
