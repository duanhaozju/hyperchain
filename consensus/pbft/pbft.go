//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package pbft implement the pbft algorithm
//The PBFT key features:
//	1. atomic sequence transactions guarantee
//      2. leader selection by viewchange
//      3. dynamic add or delete new node
//      4. support self recovery
package pbft

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

// New return a instance of pbftImpl
func New(namespace string, conf *common.Config, h helper.Stack, n int) (*pbftImpl, error) {
	return newPBFT(namespace, conf, h, n)
}

// RecvMsg receives messages from other validating peers
func (pbft *pbftImpl) RecvMsg(e []byte) error {
	msg := &protos.Message{}
	err := proto.Unmarshal(e, msg)
	if err != nil {
		pbft.logger.Errorf("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message %v", err)
		return err
	}
	switch msg.Type {
	case protos.Message_CONSENSUS:
		return pbft.enqueueConsensusMsg(msg)
	case protos.Message_NULL_REQUEST:
		return pbft.processNullRequest(msg)

	default:
		pbft.logger.Errorf("Unsupport message type: %v", msg.Type)
		return nil
	}
}

// RecvLocal receives messages form local other modules
func (pbft *pbftImpl) RecvLocal(msg interface{}) error {
	if negoView, ok := msg.(*protos.Message); ok {
		if negoView.Type == protos.Message_NEGOTIATE_VIEW {
			return pbft.initNegoView()
		}
	} else if tx, ok := msg.(*types.Transaction); ok {
		// if we receive transaction from local module, we will broadcast it to others
		payload, err := proto.Marshal(tx)
		if err != nil {
			pbft.logger.Errorf("ConsensusMessage_TRANSACTION Marshal Error", err)
			return nil
		}
		consensusMsg := &ConsensusMessage{
			Type:    ConsensusMessage_TRANSACTION,
			Payload: payload,
		}
		pbMsg := cMsgToPbMsg(consensusMsg, pbft.id)
		pbft.helper.InnerBroadcast(pbMsg)
	}

	go pbft.pbftEventQueue.Push(msg)

	return nil
}

// Start initializes and starts the consensus service
func (pbft *pbftImpl) Start() {
	pbft.logger.Noticef("--------PBFT starting, nodeID: %d--------", pbft.id)
	pbft.timerMgr = newTimerMgr(pbft)
	pbft.initTimers()
	pbft.initStatus()

	atomic.StoreUint32(&pbft.activeView, 1)
	pbft.pbftManager.Start()
	pbft.pbftEventQueue = events.GetQueue(pbft.pbftManager.Queue())

	pbft.restoreState()
	pbft.vcMgr.viewChangeSeqNo = ^uint64(0)
	pbft.vcMgr.updateViewChangeSeqNo(pbft.seqNo, pbft.K, pbft.id)
	pbft.batchMgr.start(pbft.pbftEventQueue)
	pbft.timerMgr.makeRequestTimeoutLegal()

	pbft.logger.Noticef("======== PBFT finish start, nodeID: %d", pbft.id)
}

// Close closes the consensus service
func (pbft *pbftImpl) Close() {
	pbft.logger.Notice("PBFT stop event process service")
	pbft.timerMgr.Stop()
	pbft.batchMgr.stop()
	pbft.pbftManager.Stop()

	pbft.logger.Notice("PBFT clear some resources")
	pbft.vcMgr = newVcManager(pbft)
	pbft.storeMgr = newStoreMgr()
	pbft.nodeMgr = newNodeMgr()

	pbft.batchMgr = newBatchManager(pbft)
	pbft.batchVdr = newBatchValidator()
	pbft.recoveryMgr = newRecoveryMgr()

	pbft.logger.Noticef("PBFT stopped!")
}

// GetStatus returns the current consensus status:
// 1. normal: true means not in viewchange, negotiate or state transfer
// 2. full: true means txPool is full
func (pbft *pbftImpl) GetStatus() (normal bool, full bool) {
	normal = false
	full = false

	if atomic.LoadUint32(&pbft.normal) == 1 {
		normal = true
	}
	if atomic.LoadUint32(&pbft.poolFull) == 1 {
		full = true
	}
	return
}
