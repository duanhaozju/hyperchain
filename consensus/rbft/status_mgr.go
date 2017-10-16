//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"sync/atomic"
)

// PbftStatus is used to store all the status in rbft
type RbftStatus struct {
	byzantine         int32 // whether this node is intentionally acting as byzantine
	activeView        int32 // track if replica is in active view
	skipInProgress    int32 // set when we have detested a fall behind scenario until we pick a new starting point
	stateTransferring int32 // set when state transfer is executing
	valid             int32 //
	timerActive       int32
	inRecovery        int32
	inNegoView        int32
	inUpdatingN       int32
	isNewNode         int32
	inAddingNode      int32
	inDeletingNode    int32
	inVcReset         int32 // track if replica itself in vcReset
	vcHandled         int32 // track if replica handled the vc after receive newview
	newNodeReady      int32
	updateHandled     int32
	vcToRecovery      int32
}

// activeState sets the states to true
func (status RbftStatus) activeState(stateName ...*int32) {
	for _, s := range stateName {
		atomic.StoreInt32(s, ON)
	}
}

// inActiveState sets the states to false
func (status RbftStatus) inActiveState(stateName ...*int32) {
	for _, s := range stateName {
		atomic.StoreInt32(s, OFF)
	}
}

// getState returns the state of specified state name, true means ON, false means OFF
func (status RbftStatus) getState(stateName *int32) bool {
	return atomic.LoadInt32(stateName) == ON
}

// checkStatesAnd checks the result of several status computed with each other using '&&'
func (status RbftStatus) checkStatesAnd(stateName ...*int32) bool {
	var rs bool = true
	for _, s := range stateName {
		rs = rs && status.getState(s)
	}
	return rs
}

// checkStatesOr checks the result of several status computed with each other using '||'
func (status RbftStatus) checkStatesOr(stateName ...*int32) bool {
	var rs bool = false
	for _, s := range stateName {
		rs = rs || status.getState(s)
	}
	return rs
}
