//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package pbft implement the pbft algorithm
//The PBFT key features:
//	1. atomic sequence transactions guarantee
//      2. leader selection by viewchange
//      3. dynamic delete or add new node
//      4. support state recovery.
package pbft

import (
	"fmt"

	"github.com/golang/protobuf/proto"

	"hyperchain/common"
	"hyperchain/consensus"
	"hyperchain/consensus/events"
	"hyperchain/consensus/helper"
	"hyperchain/core/types"
	"hyperchain/manager/protos"
	"sync/atomic"
)

/**
This file implement the API of consensus
which can be invoked by outer services.
*/

// New return a instance of pbftProtocal  TODO: rename helper.Stack ??
func New(namespace string, conf *common.Config, h helper.Stack) (*pbftImpl, error) {
	var err error
	pcPath := conf.GetString(consensus.CONSENSUS_ALGO_CONFIG_PATH)
	if pcPath == "" {
		err = fmt.Errorf("Invalid consensus algorithm configuration path, %s: %s",
			consensus.CONSENSUS_ALGO_CONFIG_PATH, pcPath)
		return nil, err
	}
	conf, err = conf.MergeConfig(pcPath)
	if err != nil {
		err = fmt.Errorf("Load pbft config error: %v", err)
		return nil, err
	}
	return newPBFT(namespace, conf, h)
}

// RecvMsg receive messages from outer services.
func (pbft *pbftImpl) RecvMsg(e []byte) error {

	msg := &protos.Message{}
	err := proto.Unmarshal(e, msg)
	if err != nil {
		pbft.logger.Errorf("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message %v", err)
		return err
	}
	switch msg.Type {
	case protos.Message_CONSENSUS:
		return pbft.enqueueConsensusMsg(msg) //msgs from other peers
	case protos.Message_NULL_REQUEST:
		return pbft.processNullRequest(msg)

	default:
		pbft.logger.Errorf("Unsupport message type: %v", msg.Type)
		return nil //TODO: define PBFT error type
	}
}

//RecvMsg receive messages form local services
func (pbft *pbftImpl) RecvLocal(msg interface{}) error {
	if negoView, ok := msg.(*protos.Message); ok {
		if negoView.Type == protos.Message_NEGOTIATE_VIEW {
			return pbft.initNegoView()
		}
		return nil
	} else {
		switch msg.(type) {
		case protos.RemoveCache:
			if atomic.LoadUint32(&pbft.activeView) == 1 && pbft.primary(pbft.view) == pbft.id {
				go pbft.reqEventQueue.Push(msg)
			} else {
				go pbft.pbftEventQueue.Push(msg)
			}
		case *types.Transaction: //local transaction
			go pbft.reqEventQueue.Push(msg)

		default:
			go pbft.pbftEventQueue.Push(msg)
		}
		return nil
	}
}

//Start start the consensus service
func (pbft *pbftImpl) Start() {
	pbft.logger.Noticef("--------PBFT starting, nodeID: %d--------", pbft.id)
	pbft.pbftTimerMgr = newTimerMgr(pbft)
	pbft.initTimers()
	pbft.initStatus()

	atomic.StoreUint32(&pbft.activeView, 1)
	pbft.pbftManager.Start()
	pbft.pbftEventQueue = events.GetQueue(pbft.pbftManager.Queue()) // init pbftEventQueue
	pbft.reqEventQueue = events.GetQueue(pbft.batchMgr.batchEventsManager.Queue())

	//1.restore state.
	pbft.restoreState()

	pbft.vcMgr.viewChangeSeqNo = ^uint64(0) // infinity
	pbft.vcMgr.updateViewChangeSeqNo(pbft.seqNo, pbft.K, pbft.id)
	pbft.batchMgr.start()

	pbft.pbftTimerMgr.makeRequestTimeoutLegal()

	pbft.logger.Noticef("======== PBFT finish start, nodeID: %d", pbft.id)
}

//Close close the consenter service
func (pbft *pbftImpl) Close() {
	pbft.logger.Notice("PBFT stop event process service")
	pbft.pbftTimerMgr.Stop()
	pbft.batchMgr.stop()
	pbft.pbftManager.Stop()

	pbft.logger.Notice("PBFT clear some resources")

	pbft.vcMgr = newVcManager(pbft.pbftTimerMgr, pbft, pbft.config)
	pbft.storeMgr = newStoreMgr()
	pbft.nodeMgr = newNodeMgr()

	pbft.duplicator = make(map[uint64]*transactionStore)
	pbft.batchMgr = newBatchManager(pbft.config, pbft) // init after pbftEventQueue
	// new batch manager
	pbft.batchVdr = newBatchValidator(pbft)
	pbft.reqStore = newRequestStore()
	pbft.recoveryMgr = newRecoveryMgr()

	pbft.logger.Noticef("PBFT stopped!")

}
