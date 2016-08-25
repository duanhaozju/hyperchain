package pbft

import (
	"time"
	"fmt"
	"hyperchain-alpha/consensus"
	"hyperchain-alpha/consensus/events"
	pb "github.com/hyperledger/fabric/protos"
	"github.com/golang/protobuf/proto"
	"github.com/spf13/viper"
)

type batch struct {
	batchTimer       events.Timer
	batchTimerActive bool
	batchTimeout     time.Duration
	batchSize        int
	batchStore       []*Request


	manager          events.Manager
	pbft        *pbftCore

	c                chan int8 //ToDo for test
}

type testEvent struct {} //ToDo for test

// batchMessageEvent is sent when a consensus message is received that is then to be sent to pbft
type batchMessageEvent batchMessage

// batchTimerEvent is sent when the batch timer expires
type batchTimerEvent struct{}

type batchMessage struct {
	msg    *pb.Message
	sender *pb.PeerID
}

const configPrefix = "CORE_PBFT"


func (b *batch) ProcessEvent(e events.Event) events.Event{
	logger.Debugf("Replica %d batch main thread looping", b.pbft.id)
	switch et:=e.(type) {
	case *testEvent:
		fmt.Println("lalalla")
		b.c <- 1
	case batchMessageEvent:
		ocMsg := et
		return b.processMessage(ocMsg.msg,  ocMsg.sender)
	case batchTimerEvent:
		logger.Infof("Replica %d batch timer expired", b.pbft.id)
		if  (len(b.batchStore) > 0) {
			return b.sendBatch()
		}
	default:
		fmt.Println("default")
	}
	return nil
}


func (op *batch) processMessage(ocMsg *pb.Message, senderHandle *pb.PeerID) events.Event {

	batchMsg := &BatchMessage{}
	err := proto.Unmarshal(ocMsg.Payload, batchMsg)
	if err != nil {
		logger.Errorf("Error unmarshaling message: %s", err)
		return nil
	}

	if req := batchMsg.GetRequest(); req != nil {
		if (op.pbft.primary(op.pbft.view) == op.pbft.id) {
			return op.leaderProcReq(req)
		}
		op.startTimerIfOutstandingRequests()
		return nil
	} else if pbftMsg := batchMsg.GetPbftMessage(); pbftMsg != nil {
		senderID, err := getValidatorID(senderHandle) // who sent this?
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

func (op *batch) leaderProcReq(req *Request) events.Event {
	digest := hash(req)
	logger.Debugf("Batch primary %d queueing new request %s", op.pbft.id, digest)
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
func (b *batch) RecvMsg(e events.Event){
	fmt.Println("RecvMsg")
	b.manager.Queue()<- e
}
func (op *batch) stopBatchTimer() {
	op.batchTimer.Stop()
	logger.Debugf("Replica %d stopped the batch timer", op.pbft.id)
	op.batchTimerActive = false
}



func newBatch(id uint64,config *viper.Viper,stack consensus.Stack) *batch{
	var err error
	fmt.Println("new batch")
	batchObj:=&batch{
		manager:events.NewManagerImpl(),
		c:c,//TODO   for test
	}


	batchObj.manager.SetReceiver(batchObj)
	batchObj.manager.Start()

	etf := events.NewTimerFactoryImpl(batchObj.manager)
	batchObj.pbft = newPbftCore(id, config, batchObj, etf)

	batchObj.batchTimer = etf.CreateTimer()
	batchObj.batchSize = config.GetInt("general.batchsize")
	batchObj.batchStore = nil
	batchObj.batchTimeout, err = time.ParseDuration(config.GetString("general.timeout.batch"))

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



	return batchObj
}




