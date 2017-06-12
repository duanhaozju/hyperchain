//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package pbft

import (
	"sync/atomic"
)

type PbftStatus struct {
	byzantine         int32
	skipInProgress    int32
	stateTransferring int32
	valid             int32
	timerActive       int32
	inRecovery        int32
	inNegoView        int32
	isNewNode         int32
	inAddingNode      int32
	inDeletingNode    int32
	inVcReset         int32 // track if replica itself in vcReset
	vcHandled         int32 // track if replica handled the vc after receive newview
	newNodeReady      int32
	updateHandled     int32
	vcToRecovery      int32
}

//activeState sets the states to true
func (status PbftStatus) activeState(stateName ...*int32) {
	for _, s := range stateName {
		atomic.StoreInt32(s, ON)
	}
}

//inActiveState sets the states to false
func (status PbftStatus) inActiveState(stateName ...*int32) {
	for _, s := range stateName {
		atomic.StoreInt32(s, OFF)
	}
}

func (status PbftStatus) negCurrentState(stateName *int32) bool {
	tmp := atomic.LoadInt32(stateName)
	return (ON - tmp) == ON
}

func (status PbftStatus) getState(stateName *int32) bool {
	return atomic.LoadInt32(stateName) == ON
}

func toBool(stateName *int32) bool {
	return atomic.LoadInt32(stateName) == ON
}

//checkStatesAnd checks the result of several status and each other
func (status PbftStatus) checkStatesAnd(stateName ...*int32) bool {
	var rs bool = true
	for _, s := range stateName {
		rs = rs && toBool(s)
	}
	return rs
}

//checkStatesAnd checks the result of several status or each other
func (status PbftStatus) checkStatesOr(stateName ...*int32) bool {
	var rs bool = false
	for _, s := range stateName {
		rs = rs || toBool(s)
	}
	return rs
}
