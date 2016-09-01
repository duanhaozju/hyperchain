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

type batch struct {
	batchTimer       events.Timer
	batchTimerActive bool
	batchTimeout     time.Duration
	batchSize        int
	batchStore       []*Request 	//ordered message batch
	helperImpl       helper.Stack
	manager          events.Manager
	pbft             *pbftCore
	localID           uint64

	reqStore	*requestStore	//received messages
	deduplicator	*deduplicator
	//test_c                chan int8 //ToDo for test
	mux		sync.Mutex
}

type testEvent struct {} //ToDo for test

// batchMessageEvent is sent when a consensus message is received that is then to be sent to pbft
type batchMessageEvent batchMessage

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

type batchMessage struct {
	msg    *pb.Message
	sender uint64
}

func newBatch(id uint64, config *viper.Viper, h helper.Stack) *batch{
	var err error
	batchObj:=&batch{
		localID:	id,
		helperImpl:	h,
	}

	batchObj.manager=events.NewManagerImpl()
	batchObj.manager.SetReceiver(batchObj)
	etf := events.NewTimerFactoryImpl(batchObj.manager)
	batchObj.pbft = newPbftCore(id, config, batchObj, etf)
	batchObj.manager.Start()

	batchObj.batchTimer = etf.CreateTimer()
	batchObj.batchSize = config.GetInt("general.batchsize")
	batchObj.batchStore = nil
	batchObj.batchTimeout, err = time.ParseDuration(config.GetString("timeout.batch"))

	if err != nil {
		panic(fmt.Errorf("Cannot parse batch timeout: %s", err))
	}
	if batchObj.batchTimeout >= batchObj.pbft.requestTimeout {
		batchObj.pbft.requestTimeout = 3 * batchObj.batchTimeout / 2
		logger.Warningf("Configured request timeout must be greater than batch timeout, setting to %v", batchObj.pbft.requestTimeout)
	}
	if batchObj.pbft.requestTimeout >= batchObj.pbft.nullRequestTimeout && batchObj.pbft.nullRequestTimeout != 0 {
		batchObj.pbft.nullRequestTimeout = 3 * batchObj.pbft.requestTimeout / 2
		logger.Warningf("Configured null request timeout must be greater than request timeout, setting to %v", batchObj.pbft.nullRequestTimeout)
	}
	logger.Infof("PBFT Batch size = %d", batchObj.batchSize)
	logger.Infof("PBFT Batch timeout = %v", batchObj.batchTimeout)

	batchObj.reqStore = newRequestStore()

	return batchObj
}

func (op *batch) getHelper() helper.Stack {
	return op.helperImpl
}

func (op *batch) ProcessEvent(e events.Event) events.Event{
	logger.Debugf("Replica %d batch main thread looping", op.pbft.id)
	switch event:=e.(type) {
	//case *testEvent:
	//	fmt.Println("lalalla")
	//	b.test_c <- 1//ToDo for test
	case batchMessageEvent:
		return op.processMessage(event.msg,  event.sender)
	case batchTimerEvent:
		logger.Infof("Replica %d batch timer expired", op.pbft.id)
		if  (len(op.batchStore) > 0) {
			return op.sendBatch()
		}
	default:
		logger.Info("batch processEvent, default: ")
		return op.pbft.ProcessEvent(event)
	}
	return nil
}

func (op *batch) processMessage(msg *pb.Message, id uint64) events.Event {
	if msg.Type == pb.Message_TRANSACTION {
		req := op.txToReq(msg)
		return op.submitToLeader(req)
	}
	if msg.Type != pb.Message_CONSENSUS {
		return nil
	}
	batchMsg := &BatchMessage{}
	err := proto.Unmarshal(msg.Payload, batchMsg)
	if err != nil {
		logger.Errorf("Error unmarshaling message: %s", err)
		return nil
	}

	if req := batchMsg.GetRequest(); req != nil {
		op.reqStore.storeOutstanding(req)
		if (op.pbft.primary(op.pbft.view) == op.pbft.id) {
			fmt.Println("3")
			return op.leaderProcReq(req)
		}
		fmt.Println("4")
		op.startTimerIfOutstandingRequests()
		return nil
	} else if pbftMsg := batchMsg.GetPbftMessage(); pbftMsg != nil {
		senderID :=  id
		if err != nil {
			panic("Cannot map sender's PeerID to a valid replica ID")
		}
		msg := &Message{}
		err = proto.Unmarshal(pbftMsg, msg)
		if err != nil {
			logger.Errorf("Error unpacking payload from message: %s", err)
			return nil
		}
		return pbftMessageEvent{
			msg:    msg,
			sender: senderID,
		}
	}

	logger.Errorf("Unknown request: %+v", batchMsg)

	return nil
}

func (op *batch) txToReq(tx *pb.Message) *Request {
	req := &Request{
		Timestamp: 	tx.Timestamp,
		Payload:   	tx.Payload,
		ReplicaId: 	op.pbft.id,
	}
	// XXX sign req
	return req
}

func (op *batch) leaderProcReq(req *Request) events.Event {
	//digest := hash(req)
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

func (op *batch) startTimerIfOutstandingRequests() {
	op.pbft.softStartTimer(op.pbft.requestTimeout, "Batch outstanding requests")
}

func (op *batch) startBatchTimer() {
	op.batchTimer.Reset(op.batchTimeout, batchTimerEvent{})
	logger.Debugf("Replica %d started the batch timer", op.pbft.id)
	op.batchTimerActive = true
}
func (op *batch) RecvMsg(e []byte) error {

	tempMsg := &pb.Message{}
	err := proto.Unmarshal(e,tempMsg)
	if err!=nil {
		logger.Info("**********> Unmarshal error:",err)
		return err
	}

	event := batchMessageEvent{
		msg: 	tempMsg,
		sender:	tempMsg.Id,
	}


	go op.postEvent(event)

        return nil
}

func (op *batch) postEvent(event batchMessageEvent) {
	op.mux.Lock()
	defer op.mux.Unlock()
	op.manager.Queue() <- event
}

func (op *batch) stopBatchTimer() {
	op.batchTimer.Stop()
	logger.Debugf("Replica %d stopped the batch timer", op.pbft.id)
	op.batchTimerActive = false
}

func (op *batch) submitToLeader(req *Request) events.Event {
	// Broadcast the request to the network, in case we're in the wrong view

	pbMsg := batchMsgHelper(&BatchMessage{Payload: &BatchMessage_Request{Request:req}}, op.pbft.id)
	op.helperImpl.InnerBroadcast(pbMsg)
	op.reqStore.storeOutstanding(req)
	op.startTimerIfOutstandingRequests()
	if op.pbft.primary(op.pbft.view) == op.pbft.id {
		return op.leaderProcReq(req)
	}
	return nil
}

// Close tells us to release resources we are holding
func (op *batch) Close() {
	op.batchTimer.Halt()
	op.pbft.close()
}


