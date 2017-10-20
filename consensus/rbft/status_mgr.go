//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import "sync/atomic"

// specify the position of corresponding status name in rbft.status
const (
	inNegotiateView = iota
	inRecovery
	inViewChange
	byzantine
	skipInProgress
	stateTransferring
	valid
	timerActive
	inUpdatingN
	isNewNode
	inAddingNode
	inDeletingNode
	inVcReset
	vcHandled
	newNodeReady
	updateHandled
	vcToRecovery
)

type statusManager struct {
	status   uint64 // consensus status
	normal   uint32 // system is normal or not
	poolFull uint32 // txPool is full or not
}

func newStatusMgr() *statusManager {
	return &statusManager{}
}

// reset only resets consensus status to 0.
func (st *statusManager) reset() {
	st.status = 0
}

// setBit sets the bit at position in integer n.
func (st *statusManager) setBit(position uint64) {
	st.status |= 1 << position
}

// clearBit clears the bit at position in integer n.
func (st *statusManager) clearBit(position uint64) {
	st.status &= ^(1 << position)
}

// hasBit checks whether a bit position is set.
func (st *statusManager) hasBit(position uint64) bool {
	val := st.status & (1 << position)
	return val > 0
}

// on sets the status of specified positions
func (rbft *rbftImpl) on(statusPos ...uint64) {
	for _, pos := range statusPos {
		rbft.status.setBit(pos)
	}
}

// off resets the status of specified positions
func (rbft *rbftImpl) off(statusPos ...uint64) {
	for _, pos := range statusPos {
		rbft.status.clearBit(pos)
	}
}

// in returns the status of specified position
func (rbft *rbftImpl) in(pos uint64) bool {
	return rbft.status.hasBit(pos)
}

// inAll checks the result of several status computed with each other using '&&'
func (rbft *rbftImpl) inAll(poss ...uint64) bool {
	var rs bool = true
	for _, pos := range poss {
		rs = rs && rbft.in(pos)
	}
	return rs
}

// inOne checks the result of several status computed with each other using '||'
func (rbft *rbftImpl) inOne(poss ...uint64) bool {
	var rs bool = false
	for _, pos := range poss {
		rs = rs || rbft.in(pos)
	}
	return rs
}

func (rbft *rbftImpl) setNormal() {
	atomic.StoreUint32(&rbft.status.normal, 1)
}

func (rbft *rbftImpl) setAbNormal() {
	atomic.StoreUint32(&rbft.status.normal, 0)
}

func (rbft *rbftImpl) setFull() {
	atomic.StoreUint32(&rbft.status.poolFull, 1)
}

func (rbft *rbftImpl) setNotFull() {
	atomic.StoreUint32(&rbft.status.poolFull, 0)
}

func (rbft *rbftImpl) isNormal() bool {
	return atomic.LoadUint32(&rbft.status.normal) == 1
}

func (rbft *rbftImpl) isPoolFull() bool {
	return atomic.LoadUint32(&rbft.status.poolFull) == 1
}
