//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package rbft implement the rbft algorithm
//The RBFT key features:
//	1. atomic sequence transactions guarantee
//      2. leader selection by viewchange
//      3. dynamically add or delete new node
//      4. support self recovery
package rbft

import (
	"hyperchain/common"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/txpool"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
)

/**
This file implements the API of consensus
which can be invoked by outer services.
*/

// New return a instance of rbftImpl
func New(namespace string, conf *common.Config, h helper.Stack, n int) (*rbftImpl, error) {
	return newRBFT(namespace, conf, h, n)
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

// RecvLocal receives messages form other modules of local system
func (rbft *rbftImpl) RecvLocal(msg interface{}) error {
	if negoView, ok := msg.(*protos.Message); ok {
		if negoView.Type == protos.Message_NEGOTIATE_VIEW {
			return rbft.initNegoView()
		}
	} else if tx, ok := msg.(*types.Transaction); ok {
		// if we receive transaction from local module, we will broadcast it to others.
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

		req := txRequest{
			tx:  tx,
			new: true,
		}
		go rbft.eventMux.Post(req)

		return nil
	}

	go rbft.eventMux.Post(msg)

	return nil
}

// Start initializes and starts the consensus service
func (rbft *rbftImpl) Start() {
	rbft.logger.Noticef("--------RBFT starting, nodeID: %d--------", rbft.id)
	rbft.timerMgr = newTimerMgr(rbft)
	rbft.initTimers()
	rbft.initStatus()

	rbft.eventMux = new(event.TypeMux)
	rbft.batchSub = rbft.eventMux.Subscribe(txRequest{}, txpool.TxHashBatch{}, protos.RoutersMessage{}, &LocalEvent{}, &ConsensusMessage{})
	rbft.close = make(chan bool)

	rbft.restoreState()
	rbft.vcMgr.viewChangeSeqNo = ^uint64(0)
	rbft.vcMgr.updateViewChangeSeqNo(rbft.seqNo, rbft.K, rbft.id)
	rbft.batchMgr.start(rbft.eventMux)
	rbft.timerMgr.makeRequestTimeoutLegal()

	go rbft.listenEvent()
	rbft.logger.Noticef("======== RBFT finish start, nodeID: %d", rbft.id)
}

// Close closes the consensus service
func (rbft *rbftImpl) Close() {
	rbft.logger.Notice("RBFT stop event process service")
	rbft.timerMgr.Stop()
	rbft.batchMgr.stop()

	if rbft.close != nil {
		close(rbft.close)
		rbft.close = nil
	}

	rbft.logger.Notice("RBFT clear some resources")
	rbft.vcMgr = newVcManager(rbft)
	rbft.storeMgr = newStoreMgr()
	rbft.nodeMgr = newNodeMgr()

	rbft.batchMgr = newBatchManager(rbft)
	rbft.batchVdr = newBatchValidator()
	rbft.recoveryMgr = newRecoveryMgr()

	rbft.logger.Noticef("RBFT stopped!")
}

// GetStatus returns the current consensus status:
// 1. normal: true means not in viewchange, negotiate or state transfer
// 2. full: true means txPool is full
func (rbft *rbftImpl) GetStatus() (normal bool, full bool) {
	normal = rbft.isNormal()
	full = rbft.isPoolFull()

	return
}
