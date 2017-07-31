//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"container/list"
	"hyperchain/core/types"
	"sync"
)

type requestContainer struct {
	key string
	req *types.Transaction
	tmux	sync.Mutex
}

type orderedRequests struct {
	order    list.List
	presence map[string]*list.Element
}

func (a *orderedRequests) Len() int {
	return a.order.Len()
}

func (a *orderedRequests) wrapRequest(req *types.Transaction) requestContainer {
	return requestContainer{
		key: hash(req),
		req: req,
	}
}

func (a *orderedRequests) has(key string) bool {
	_, ok := a.presence[key]
	return ok
}

func (a *orderedRequests) add(request *types.Transaction) {
	rc := a.wrapRequest(request)
	if !a.has(rc.key) {
		e := a.order.PushBack(rc)
		a.presence[rc.key] = e
	}
}

func (a *orderedRequests) adds(requests []*types.Transaction) {
	for _, req := range requests {
		a.add(req)
	}
}

func (a *orderedRequests) remove(request *types.Transaction) bool {
	rc := a.wrapRequest(request)
	e, ok := a.presence[rc.key]
	if !ok {
		return false
	}
	a.order.Remove(e)
	delete(a.presence, rc.key)
	return true
}

func (a *orderedRequests) removes(requests []*types.Transaction) bool {
	allSuccess := true
	for _, req := range requests {
		if !a.remove(req) {
			allSuccess = false
		}
	}

	return allSuccess
}

func (a *orderedRequests) empty() {
	a.order.Init()
	a.presence = make(map[string]*list.Element)
}

type requestStore struct {
	outstandingRequests *orderedRequests
	pendingRequests     *orderedRequests
}

// newRequestStore creates a new requestStore.
func newRequestStore() *requestStore {
	rs := &requestStore{
		outstandingRequests: &orderedRequests{},
		pendingRequests:     &orderedRequests{},
	}
	// initialize data structures
	rs.outstandingRequests.empty()
	rs.pendingRequests.empty()

	return rs
}

// storeOutstanding adds a request to the outstanding request list
func (rs *requestStore) storeOutstanding(request *types.Transaction) {
	rs.tmux.Lock()
	defer rs.tmux.Unlock()
	rs.outstandingRequests.add(request)
}

// storePending adds a request to the pending request list
func (rs *requestStore) storePending(request *types.Transaction) {
	rs.pendingRequests.add(request)
}

// storePending adds a slice of requests to the pending request list
func (rs *requestStore) storePendings(requests []*types.Transaction) {
	rs.pendingRequests.adds(requests)
}

// remove deletes the request from both the outstanding and pending lists, it returns whether it was found in each list respectively
func (rs *requestStore) remove(request *types.Transaction) (outstanding, pending bool) {
	rs.tmux.Lock()
	defer rs.tmux.Unlock()
	outstanding = rs.outstandingRequests.remove(request)
	pending = rs.pendingRequests.remove(request)
	return
}

// getNextNonPending returns up to the next n outstanding, but not pending requests
func (rs *requestStore) hasNonPending() bool {
	return rs.outstandingRequests.Len() > rs.pendingRequests.Len()
}

// getNextNonPending returns up to the next n outstanding, but not pending requests
func (rs *requestStore) getNextNonPending(n int) (result []*types.Transaction) {
	for oreqc := rs.outstandingRequests.order.Front(); oreqc != nil; oreqc = oreqc.Next() {
		oreq := oreqc.Value.(requestContainer)
		if rs.pendingRequests.has(oreq.key) {
			continue
		}
		result = append(result, oreq.req)
		if len(result) == n {
			break
		}
	}
	return result
}
