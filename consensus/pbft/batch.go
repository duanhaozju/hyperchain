package pbft

import (
	"time"
	"fmt"
	"sync"

	"hyperchain/consensus/helper"
	"hyperchain/consensus/events"
	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
)

// batch is used to construct reqbatch, the middle layer between outer to pbft
type batch struct {
	batchTimer       events.Timer
	batchTimerActive bool
	batchTimeout     time.Duration
	batchSize        int
	batchStore       []*Request 	//ordered message batch
	helperImpl       helper.Stack
	batchManager     events.Manager
	pbftManager      events.Manager
	pbft             *pbftCore
	localID           uint64

	reqStore	*requestStore	//received messages
	mux		sync.Mutex
}

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

// newBatch initializes a batch
func newBatch(id uint64, config *viper.Viper, h helper.Stack) *batch {
	var err error
	op:=&batch{
		localID:	id,
		helperImpl:	h,
	}

	// pbftManager is used to solve pbft message
	op.pbftManager = events.NewManagerImpl()
	op.pbftManager.SetReceiver(op)
	pbftTimerFactory := events.NewTimerFactoryImpl(op.pbftManager)
	op.pbft = newPbftCore(id, config, op, pbftTimerFactory)
	op.pbftManager.Start()

	// batchManager is used to solve batch message, like *Request
	op.batchManager = events.NewManagerImpl()
	op.batchManager.SetReceiver(op)
	etf := events.NewTimerFactoryImpl(op.batchManager)
	op.batchManager.Start()

	// initialize the batchTimeout
	op.batchTimer = etf.CreateTimer()
	op.batchSize = config.GetInt("general.batchsize")
	op.batchStore = nil
	op.batchTimeout, err = time.ParseDuration(config.GetString("timeout.batch"))

	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}

	if op.batchTimeout >= op.pbft.requestTimeout {
		op.pbft.requestTimeout = 3 * op.batchTimeout / 2
		logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", op.pbft.requestTimeout)
	}

	if op.pbft.requestTimeout >= op.pbft.nullRequestTimeout && op.pbft.nullRequestTimeout != 0 {
		op.pbft.nullRequestTimeout = 3 * op.pbft.requestTimeout / 2
		logger.Warningf("Configured null request timeout must be greater than request timeout, setting to %v", op.pbft.nullRequestTimeout)
	}

	logger.Infof("PBFT Batch size = %d", op.batchSize)
	logger.Infof("PBFT Batch timeout = %v", op.batchTimeout)
	op.reqStore = newRequestStore()

	return op
}

// RecvMsg is used by outer to send message to consensus
func (op *batch) RecvMsg(e []byte) error {

	msg := &pb.Message{}
	err := proto.Unmarshal(e,msg)
	
	if err!=nil {
		logger.Errorf("Inner RecvMsg Unmarshal error: can not unmarshal pb.Message", err)
		return err
	}

	if msg.Type == pb.Message_TRANSACTION {
		return op.processTransaction(msg)
	} else if msg.Type == pb.Message_CONSENSUS {
		return op.processConsensus(msg)
	}

	logger.Errorf("Unknown recvMsg: %+v", msg)

	return nil
}

// process the trasaction message
func (op *batch) processTransaction(msg *pb.Message) error {

	// Parse the trasaction message to request
	req := op.txToReq(msg)

	// Broadcast request to other peer
	consensusMsg := &ConsensusMessage{Payload: &ConsensusMessage_Request{Request: req}}
	pbMsg := consensusMsgHelper(consensusMsg, op.pbft.id)
	op.helperImpl.InnerBroadcast(pbMsg)

	// Post a requestEvent
	go op.postRequestEvent(req)

	return nil
}

// process the consensus message
func (op *batch) processConsensus(msg *pb.Message) error {

	consensus := &ConsensusMessage{}
	err := proto.Unmarshal(msg.Payload, consensus)

	if err != nil {
		logger.Errorf("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err)
		return err
	}

	if req := consensus.GetRequest(); req != nil {
		go op.postRequestEvent(req)
		return nil
	} else if pbft := consensus.GetPbftMessage(); pbft != nil {
		event := pbftMessageEvent{
			msg:	pbft,
			sender:	msg.Id,
		}
		op.postPbftEvent(event)
		return nil
	}

	logger.Errorf("Unknown ConsensusMessage: %+v", msg)

	return nil
}

func (op *batch) ProcessEvent(e events.Event) events.Event{

	logger.Debugf("Replica %d start solve event", op.pbft.id)

	switch event:=e.(type) {

	case *Request:
		req := event
		return op.processRequest(req)
	case batchTimerEvent:
		logger.Debugf("Replica %d batch timer expired", op.pbft.id)
		if  (len(op.batchStore) > 0) {
			return op.sendBatch()
		}
	default:
		logger.Debugf("batch processEvent, default: ")
		return op.pbft.ProcessEvent(event)
	}
	return nil
}

func (op *batch) processRequest(req *Request) events.Event {

	op.reqStore.storeOutstanding(req)
	op.startTimerIfOutstandingRequests()
	if (op.pbft.primary(op.pbft.view) == op.pbft.id) {
		return op.leaderProcReq(req)
	}

	return nil
}

// covert the transaction to request
func (op *batch) txToReq(tx *pb.Message) *Request {

	req := &Request{
		Timestamp: 	tx.Timestamp,
		Payload:   	tx.Payload,
		ReplicaId: 	op.pbft.id,
	}

	return req
}

func (op *batch) postRequestEvent(event *Request) {

	op.mux.Lock()
	defer op.mux.Unlock()
	op.batchManager.Queue() <- event

}

func (op *batch) postPbftEvent(event pbftMessageEvent) {
	op.pbftManager.Queue() <- event
}

func (op *batch) leaderProcReq(req *Request) events.Event {
	
	logger.Debugf("Batch primary %d queueing new request", op.pbft.id)
	op.batchStore = append(op.batchStore, req)

	if !op.batchTimerActive {
		op.startBatchTimer()
	}

	if len(op.batchStore) >= op.batchSize {
		return op.sendBatch()
	}

	return nil
}

func (op *batch) sendBatch() events.Event {
	
	op.stopBatchTimer()
	
	if len(op.batchStore) == 0 {
		logger.Error("Told to send an empty batch store for ordering, ignoring")
		return nil
	}

	reqBatch := &RequestBatch{Batch: op.batchStore}
	op.batchStore = nil
	logger.Infof("Creating batch with %d requests", len(reqBatch.Batch))

	return reqBatch
}


func (op *batch) getHelper() helper.Stack {
	return op.helperImpl
}

func (op *batch) startTimerIfOutstandingRequests() {
	op.pbft.softStartTimer(op.pbft.requestTimeout, "Batch outstanding requests")
}

func (op *batch) startBatchTimer() {
	op.batchTimer.Reset(op.batchTimeout, batchTimerEvent{})
	logger.Debugf("Replica %d started the batch timer", op.pbft.id)
	op.batchTimerActive = true
}

func (op *batch) stopBatchTimer() {
	op.batchTimer.Stop()
	logger.Debugf("Replica %d stopped the batch timer", op.pbft.id)
	op.batchTimerActive = false
}

// Close tells us to release resources we are holding
func (op *batch) Close() {
	op.batchTimer.Halt()
	op.pbft.close()
}


