//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

// RbftStatus is used to store all the status in rbft
type RbftStatus struct {
	byzantine         bool // whether this node is intentionally acting as byzantine
	skipInProgress    bool // set when we have detested a fall behind scenario until we pick a new starting point
	stateTransferring bool // set when state transfer is executing
	valid             bool
	timerActive       bool
	inRecovery        bool
	inNegoView        bool
	inUpdatingN       bool
	isNewNode         bool
	inAddingNode      bool
	inDeletingNode    bool
	inVcReset         bool // track if replica itself in vcReset
	vcHandled         bool // track if replica handled the vc after receive newview
	newNodeReady      bool
	updateHandled     bool
	vcToRecovery      bool
	inViewChange      bool
}

// activeState sets the states to true
func (status RbftStatus) activeState(stateName ...*bool) {
	for _, s := range stateName {
		*s = true
	}
}

// inActiveState sets the states to false
func (status RbftStatus) inActiveState(stateName ...*bool) {
	for _, s := range stateName {
		*s = false
	}
}

// getState returns the state of specified state name, true means ON, false means OFF
func (status RbftStatus) getState(stateName *bool) bool {
	return *stateName
}

// checkStatesAnd checks the result of several status computed with each other using '&&'
func (status RbftStatus) checkStatesAnd(stateName ...*bool) bool {
	var rs bool = true
	for _, s := range stateName {
		rs = rs && status.getState(s)
	}
	return rs
}

// checkStatesOr checks the result of several status computed with each other using '||'
func (status RbftStatus) checkStatesOr(stateName ...*bool) bool {
	var rs bool = false
	for _, s := range stateName {
		rs = rs || status.getState(s)
	}
	return rs
}
