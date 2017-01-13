//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"container/list"

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

func (a *transactionStore) wrapRequest(tx *types.Transaction) transactionContainer {
	return transactionContainer{
		key: string(tx.TransactionHash),
		tx:  tx,
	}
}

func (a *transactionStore) has(key string) bool {
	_, ok := a.presence[key]
	return ok
}

func (a *transactionStore) add(tx *types.Transaction) {
	rc := a.wrapRequest(tx)
	if !a.has(rc.key) {
		e := a.order.PushBack(rc)
		a.presence[rc.key] = e
	}
}

func (a *transactionStore) remove(tx *types.Transaction) bool {
	rc := a.wrapRequest(tx)
	e, ok := a.presence[rc.key]
	if !ok {
		return false
	}
	a.order.Remove(e)
	delete(a.presence, rc.key)
	return true
}

func (a *transactionStore) empty() {
	a.order.Init()
	a.presence = make(map[string]*list.Element)
}

// newRequestStore creates a new requestStore.
func newTransactionStore() *transactionStore {
	rs := &transactionStore{}
	// initialize data structures
	rs.empty()

	return rs
}
