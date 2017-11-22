//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus/helper"
	"github.com/hyperchain/hyperchain/consensus/helper/persist"
	"github.com/hyperchain/hyperchain/consensus/txpool"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/hyperdb"
	hcom "github.com/hyperchain/hyperchain/hyperdb/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"github.com/hyperchain/hyperchain/manager/protos"

	"fmt"
	"github.com/golang/protobuf/proto"
)

// Package rbft implement the rbft algorithm
// The RBFT key features:
// 1. atomic sequence transactions guarantee
// 2. leader selection by viewchange
// 3. dynamically add or delete new node
// 4. support self recovery
// This file implements the API of consensus which can be invoked by outer services.

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
func (rbft *rbftImpl) Start() error {
	rbft.logger.Noticef("--------RBFT starting, nodeID: %d--------", rbft.id)

	var err error

	metaDb, err := hyperdb.GetOrCreateDatabase(rbft.config, rbft.namespace, hcom.DBNAME_META)
	if err != nil {
		rbft.logger.Errorf("get meta db by namespace: %s failed.", rbft.namespace)
		return fmt.Errorf("get meta db by namespace: %s failed", rbft.namespace)
	}
	rbft.persister = persist.New(metaDb)

	// new timer manager
	rbft.timerMgr = newTimerMgr(rbft.logger)
	rbft.initTimers()

	// new status manager
	rbft.status = newStatusMgr()
	rbft.initStatus()

	// new executor
	opLog, err := hyperdb.GetOrCreateDatabase(rbft.config, rbft.namespace, hcom.DBNAME_OPLOG)
	if err != nil {
		rbft.logger.Errorf("get opLog db by namespace: %s failed.", rbft.namespace)
		return fmt.Errorf("get opLog db by namespace: %s failed", rbft.namespace)
	}
	rbft.exec = newExecutor(opLog, rbft.logger)

	// new store manager
	rbft.storeMgr = newStoreMgr(rbft.logger)

	// new batch manager
	batchMgr, err := newBatchManager(rbft.namespace, rbft.config, rbft.logger)
	if err != nil {
		return err
	}
	rbft.batchMgr = batchMgr

	// new batch validator
	rbft.batchVdr = newBatchValidator()

	// new recovery manager
	rbft.recoveryMgr = newRecoveryMgr()

	// new viewchange manager
	rbft.vcMgr = newVcManager(rbft.config, rbft.logger)

	// new node manager
	rbft.nodeMgr = newNodeMgr()

	// restore state from consensus database
	if err = rbft.restoreState(); err != nil {
		return err
	}
	// update viewchange seqNo after restore state which may update seqNo
	rbft.updateViewChangeSeqNo(rbft.seqNo, rbft.K, rbft.id)

	// start listen batch event from tx pool
	rbft.eventMux = new(event.TypeMux)
	rbft.batchSub = rbft.eventMux.Subscribe(txRequest{}, txpool.TxHashBatch{}, protos.RoutersMessage{}, &LocalEvent{}, &ConsensusMessage{})
	rbft.batchMgr.start(rbft.eventMux)

	// start listen consensus event
	rbft.close = make(chan bool)
	go rbft.listenEvent()

	rbft.logger.Infof("RBFT Max number of validating peers (N) = %v", rbft.N)
	rbft.logger.Infof("RBFT Max number of failing peers (f) = %v", rbft.f)
	rbft.logger.Infof("RBFT byzantine flag = %v", rbft.in(byzantine))
	rbft.logger.Infof("RBFT Checkpoint period (K) = %v", rbft.K)
	rbft.logger.Infof("RBFT Log multiplier = %v", rbft.logMultiplier)
	rbft.logger.Infof("RBFT log size (L) = %v", rbft.L)

	rbft.logger.Noticef("======== RBFT finished start, nodeID: %d", rbft.id)

	return nil
}

// Close closes the consensus service
func (rbft *rbftImpl) Stop() {
	rbft.logger.Notice("RBFT stopping...")

	// resource database
	// TODO how to close this shared database
	//hyperdb.CloseDatabase(rbft.namespace, hcom.DBNAME_META)
	hyperdb.CloseDatabase(rbft.namespace, hcom.DBNAME_OPLOG)

	// stop listen consensus event
	if rbft.close != nil {
		close(rbft.close)
		rbft.close = nil
	}

	// stop listen batch event
	rbft.batchMgr.stop()

	// stop all timer event
	rbft.timerMgr.Stop()

	rbft.logger.Notice("RBFT clear some resources...")
	rbft.resetComponents()

	rbft.logger.Noticef("======== RBFT stopped!")
}

// GetStatus returns the current consensus status:
// 1. normal: true means not in viewchange, negotiate or state transfer
// 2. full: true means txPool is full
func (rbft *rbftImpl) GetStatus() (normal bool, full bool) {

	normal = rbft.isNormal()
	full = rbft.isPoolFull()

	return
}
