//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package rbft implement the rbft algorithm
//The PBFT key features:
//	1. atomic sequence transactions guarantee
//      2. leader selection by viewchange
//      3. dynamic add or delete new node
//      4. support self recovery
package rbft

import (
	"sync/atomic"

	"hyperchain/common"
	"hyperchain/consensus/events"
	"hyperchain/consensus/helper"
	"hyperchain/core/types"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
)

/**
This file implement the API of consensus
which can be invoked by outer services.
*/

// New return a instance of rbftImpl
func New(namespace string, conf *common.Config, h helper.Stack, n int) (*rbftImpl, error) {
	return newPBFT(namespace, conf, h, n)
}

// RecvMsg receives messages from other validating peers
func (rbft *rbftImpl) RecvMsg(e []byte) error {
	msg := &protos.Message{}
	err := proto.Unmarshal(e, msg)
	if err != nil {
		rbft.logger.Errorf("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message %v", err)
		return err
	}
	switch msg.Type {
	case protos.Message_CONSENSUS:
		return rbft.enqueueConsensusMsg(msg)
	case protos.Message_NULL_REQUEST:
		return rbft.processNullRequest(msg)

	default:
		rbft.logger.Errorf("Unsupport message type: %v", msg.Type)
		return nil
	}
}

// RecvLocal receives messages form local other modules
func (rbft *rbftImpl) RecvLocal(msg interface{}) error {
	if negoView, ok := msg.(*protos.Message); ok {
		if negoView.Type == protos.Message_NEGOTIATE_VIEW {
			return rbft.initNegoView()
		}
	} else if tx, ok := msg.(*types.Transaction); ok {
		// if we receive transaction from local module, we will broadcast it to others
		payload, err := proto.Marshal(tx)
		if err != nil {
			rbft.logger.Errorf("ConsensusMessage_TRANSACTION Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_TRANSACTION,
			Payload: payload,
		}
		pbMsg := cMsgToPbMsg(consensusMsg, rbft.id)
		rbft.helper.InnerBroadcast(pbMsg)
	}

	go rbft.rbftEventQueue.Push(msg)

	return nil
}

// Start initializes and starts the consensus service
func (rbft *rbftImpl) Start() {
	rbft.logger.Noticef("--------PBFT starting, nodeID: %d--------", rbft.id)
	rbft.timerMgr = newTimerMgr(rbft)
	rbft.initTimers()
	rbft.initStatus()

	atomic.StoreUint32(&rbft.activeView, 1)
	rbft.rbftManager.Start()
	rbft.rbftEventQueue = events.GetQueue(rbft.rbftManager.Queue())

	rbft.restoreState()
	rbft.vcMgr.viewChangeSeqNo = ^uint64(0)
	rbft.vcMgr.updateViewChangeSeqNo(rbft.seqNo, rbft.K, rbft.id)
	rbft.batchMgr.start(rbft.rbftEventQueue)
	rbft.timerMgr.makeRequestTimeoutLegal()

	rbft.logger.Noticef("======== PBFT finish start, nodeID: %d", rbft.id)
}

// Close closes the consensus service
func (rbft *rbftImpl) Close() {
	rbft.logger.Notice("PBFT stop event process service")
	rbft.timerMgr.Stop()
	rbft.batchMgr.stop()
	rbft.rbftManager.Stop()

	rbft.logger.Notice("PBFT clear some resources")
	rbft.vcMgr = newVcManager(rbft)
	rbft.storeMgr = newStoreMgr()
	rbft.nodeMgr = newNodeMgr()

	rbft.batchMgr = newBatchManager(rbft)
	rbft.batchVdr = newBatchValidator()
	rbft.recoveryMgr = newRecoveryMgr()

	rbft.logger.Noticef("PBFT stopped!")
}

// GetStatus returns the current consensus status:
// 1. normal: true means not in viewchange, negotiate or state transfer
// 2. full: true means txPool is full
func (rbft *rbftImpl) GetStatus() (normal bool, full bool) {
	normal = false
	full = false

	if atomic.LoadUint32(&rbft.normal) == 1 {
		normal = true
	}
	if atomic.LoadUint32(&rbft.poolFull) == 1 {
		full = true
	}
	return
}