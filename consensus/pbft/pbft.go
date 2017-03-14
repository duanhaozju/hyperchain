//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

//Package pbft implement the pbft algorithm
//The PBFT key features:
//		1. atomic sequence transactions guarantee
//      2. leader selection by viewchange
//      3. dynamic delete or add new node
//      4. support state recovery.
package pbft

import (
	"fmt"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"

	"hyperchain/protos"
	"hyperchain/common"
	"hyperchain/consensus/helper"
	"hyperchain/consensus"
	"sync/atomic"
)

/**
	This file implement the API of consensus
	which can be invoked by outer services.
 */

var logger *logging.Logger

func init() {
	logger = logging.MustGetLogger("consensus")
}

// New return a instance of pbftProtocal  TODO: rename helper.Stack ??
func New(namespace string, conf * common.Config, h helper.Stack) (*pbftImpl, error) {
	var err error
	pcPath := conf.GetString(consensus.CONSENSUS_ALGO_CONFIG_PATH)
	if pcPath == "" {
		err = fmt.Errorf("Invalid consensus algorithm configuration path, %s: %s",
			consensus.CONSENSUS_ALGO_CONFIG_PATH,  pcPath)
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
		logger.Errorf("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message %v", err)
		return err
	}
	switch msg.Type {
	case protos.Message_TRANSACTION:// tx send to current node and current node is primary
		return pbft.enqueueTx(msg)
	case protos.Message_CONSENSUS:
		return pbft.enqueueConsensusMsg(msg) //msgs from other peers
	case protos.Message_STATE_UPDATED:
		return pbft.enqueueStateUpdatedMsg(msg)
	case protos.Message_NULL_REQUEST:
		return pbft.processNullRequest(msg)
	case protos.Message_NEGOTIATE_VIEW:
		return pbft.processNegotiateView()
	default:
		logger.Errorf("Unsupport message type: %v", msg.Type)
		return nil//TODO: define PBFT error type
	}
}

//RecvMsg receive messages form local services
func (pbft *pbftImpl) RecvLocal(msg interface{}) error {

	switch msg.(type) {
	case protos.RemoveCache:
		if atomic.LoadUint32(&pbft.activeView) == 1 && pbft.primary(pbft.view) == pbft.id {
			go pbft.reqEventQueue.Push(msg)
		} else {
			go pbft.pbftEventQueue.Push(msg)
		}

	default:
		go pbft.pbftEventQueue.Push(msg)
	}
	return nil
}

//Start start the consensus service
func (pbft *pbftImpl) Start()  {
	logger.Noticef("--------PBFT starting, nodeID: %d--------", pbft.id)

	//1.restore state.
	pbft.restoreState()

	pbft.vcMgr.viewChangeSeqNo = ^uint64(0) // infinity
	pbft.vcMgr.updateViewChangeSeqNo(pbft.seqNo, pbft.K, pbft.id)
	pbft.batchMgr.start()

	pbft.pbftTimerMgr.makeRequestTimeoutLegal()

	logger.Noticef("======== PBFT finish start, nodeID: %d", pbft.id)
}

//Close close the consenter service
func (*pbftImpl) Close()  {
	//TODO: stop the PBFT service
}

