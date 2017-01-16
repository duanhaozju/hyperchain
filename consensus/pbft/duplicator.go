//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"container/list"

	"hyperchain/core/types"
)

type transactionContainer struct {
	key string
	tx *types.Transaction
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
		tx: tx,
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

