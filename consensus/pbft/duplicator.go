//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"container/list"

	"encoding/hex"
	"github.com/pkg/errors"
	"hyperchain/core/types"
)

type transactionContainer struct {
	key string
	tx  *types.Transaction
}

type transactionStore struct {
	order    list.List
	presence map[string]*list.Element
}

func (a *transactionStore) Len() int {
	return a.order.Len()
}

func (a *transactionStore) wrapTransaction(tx *types.Transaction) transactionContainer {
	return transactionContainer{
		key: string(tx.TransactionHash),
		tx:  tx,
	}
}

func (ts *transactionStore) has(key string) bool {
	_, ok := ts.presence[key]
	return ok
}

func (ts *transactionStore) add(tx *types.Transaction) {
	rc := ts.wrapTransaction(tx)
	if !ts.has(rc.key) {
		e := ts.order.PushBack(rc)
		ts.presence[rc.key] = e
	}
}

func (ts *transactionStore) remove(tx *types.Transaction) bool {
	rc := ts.wrapTransaction(tx)
	e, ok := ts.presence[rc.key]
	if !ok {
		return false
	}
	ts.order.Remove(e)
	delete(ts.presence, rc.key)
	return true
}

func (ts *transactionStore) empty() {
	ts.order.Init()
	ts.presence = make(map[string]*list.Element)
}

// newRequestStore creates a new requestStore.
func newTransactionStore() *transactionStore {
	rs := &transactionStore{}
	// initialize data structures
	rs.empty()

	return rs
}

// =============================================================================
// helper functions for duplicator
// =============================================================================
// check if a tx is duplicate in a block
func (pbft *pbftImpl) checkDuplicateInBlock(tx *types.Transaction, txStore *transactionStore) bool {
	key := hex.EncodeToString(tx.TransactionHash)
	return txStore.has(key)
}

// check if a tx is duplicate in cache
func (pbft *pbftImpl) checkDuplicateInCache(tx *types.Transaction) (exist bool) {
	exist = false
	for _, txStore := range pbft.duplicator {
		if txStore != nil && pbft.checkDuplicateInBlock(tx, txStore) {
			exist = true
			break
		}
	}
	return
}

// backup put the packaged batch to transactionStore and check if primary's batch result is right
func (pbft *pbftImpl) checkDuplicate(txBatch *TransactionBatch) (txStore *transactionStore, err error) {
	txStore = newTransactionStore()
	err = nil
	for _, tx := range txBatch.Batch {
		key := hex.EncodeToString(tx.TransactionHash)
		if txStore.has(key) || pbft.checkDuplicateInCache(tx) {
			err = errors.New("Find duplicate transaction in the batch sent by primary")
			break
		} else {
			txStore.add(tx)
		}
	}
	return
}

// primary remove duplicate transaction for packaged batch
func (pbft *pbftImpl) removeDuplicate(txBatch *TransactionBatch) (newBatch *TransactionBatch, txStore *transactionStore) {
	newBatch = &TransactionBatch{Timestamp: txBatch.Timestamp}
	txStore = newTransactionStore()
	for _, tx := range txBatch.Batch {
		key := hex.EncodeToString(tx.TransactionHash)
		if txStore.has(key) || pbft.checkDuplicateInCache(tx) {
			pbft.logger.Warningf("Primary %d received duplicate transaction %v", pbft.id, tx)
		} else {
			txStore.add(tx)
			newBatch.Batch = append(newBatch.Batch, tx)
		}
	}
	return
}

// previous primary rebuild the duplicator after view change
func (pbft *pbftImpl) rebuildDuplicator() {
	temp := make(map[uint64]*transactionStore)
	dv := pbft.batchVdr.vid - pbft.h
	for i, txStore := range pbft.duplicator {
		temp[i-dv] = txStore
	}
	pbft.duplicator = temp
	pbft.clearDuplicator()
}

// replica clear the duplicator after view change
func (pbft *pbftImpl) clearDuplicator() {
	h := pbft.h
	for i := range pbft.duplicator {
		if i > h {
			delete(pbft.duplicator, i)
		}
	}
}
