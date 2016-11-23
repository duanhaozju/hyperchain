// author: Zhenlong Zhao
// email: zhenlongzhao@hyperchain.cn
// date: 16/11/15
// last modified: 16/11/15
// last Modified Author: zhenlongzhao
// change log:


package pbft

import (
	"time"
	"testing"
	"hyperchain/protos"

	"github.com/golang/protobuf/proto"

	"hyperchain/core"
	"hyperchain/event"
	"hyperchain/core/types"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/events"
)

func getPbftConfigPath() string {
	return "/Users/zarczhao/Documents/GoWorkspace/src/hyperchain/config/pbft.yaml"
}

func TestRecvMsgMaliciousEvent(t *testing.T) {

	pbft := new(pbftProtocal)

	// test malicious input
	maliciousEv := []byte("testbytes")
	err := pbft.RecvMsg(maliciousEv)
	if err == nil {
		t.Errorf("Recv receive malicious bytes, expect err")
	}
}

func TestRecvMsgProcessTransaction(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	// first generate a Transaction
	tx := &types.Transaction{
		From: 		[]byte{1},
		To:   		[]byte{2},
		Value:		[]byte{1},
		Timestamp:	time.Now().UnixNano(),
		Signature: 	[]byte("test"),
		Id:		uint64(1),
		TransactionHash:[]byte("hash"),
	}
	txPayload, _ := proto.Marshal(tx)

	message := &protos.Message{
		Type: protos.Message_TRANSACTION,
		Timestamp: time.Now().UnixNano(),
		Payload: txPayload,
		Id: uint64(1),
	}

	msg, err := proto.Marshal(message)

	err = pbft.RecvMsg(msg)
	if err != nil {
		t.Errorf("RecvMsg error not nil, expect nil")
	}
}

func TestRecvMsgProcessConsensus(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

 	//  Messsage(Message_CONSENSUS) contains
	// ConsensusMessage(ConsenssusMessage_TRANSACTION)
	tx := &types.Transaction{
		From: 		[]byte{1},
		To:   		[]byte{2},
		Value:		[]byte{1},
		Timestamp:	time.Now().UnixNano(),
		Signature: 	[]byte("test"),
		Id:		uint64(1),
		TransactionHash:[]byte("hash"),
	}
	txPayload, _ := proto.Marshal(tx)

	cs := &ConsensusMessage {
		Type:		ConsensusMessage_TRANSACTION,
		Payload:	txPayload,
	}
	csPayload, err := proto.Marshal(cs)
	if err != nil {
		t.Errorf("TestProcessConsensus Marshal error")
	}

	message := &protos.Message{
		Type: protos.Message_CONSENSUS,
		Timestamp: time.Now().UnixNano(),
		Payload: csPayload,
		Id: uint64(1),
	}
	msg, err := proto.Marshal(message)
	if err != nil {
		t.Errorf("TestProcessConsensus Marshal error")
	}

	err = pbft.RecvMsg(msg)
	if err != nil {
		t.Errorf("ProcessConsensus not nil, expect nil")
	}

	//Message(Message_CONSENSUS) contains
	//non-ConsensusMessage_TRANSACTION, eg ConsensusMessage_TRANSATION_BATCH
	preprep := &PrePrepare{
		View:             0,
		SequenceNumber:   1,
		BatchDigest:      "digest",
		TransactionBatch: nil,
		ReplicaId:        1,
	}
	preprePayload, err := proto.Marshal(preprep)
	nonTxConsensus := &ConsensusMessage {
		Type:		ConsensusMessage_PRE_PREPARE,
		Payload:	preprePayload,
	}
	nonTxConsensusPayload, _ := proto.Marshal(nonTxConsensus)

	nonTxCsWrapper := &protos.Message {
		Type:		 protos.Message_CONSENSUS,
		Timestamp:	 time.Now().UnixNano(),
		Payload:	 nonTxConsensusPayload,
		Id: 		 uint64(1),
	}
	msg, _ = proto.Marshal(nonTxCsWrapper)

	err = pbft.RecvMsg(msg)
	if err != nil {
		t.Errorf("ProcessConsensus not nil, expect nil")
	}

	//  Messsage(Message_CONSENSUS) contains
	// ConsensusMessage(malicious)

	maliciousPayload := []byte("maliciousPayload")
	maliciousCsWrapper := &protos.Message {
		Type:		 protos.Message_CONSENSUS,
		Timestamp:	 time.Now().UnixNano(),
		Payload:	 maliciousPayload,
		Id: 		 uint64(1),
	}
	msg, _ = proto.Marshal(maliciousCsWrapper)
	err = pbft.RecvMsg(msg)
	if err == nil {
		t.Error("Process consensus malicious payload, error nil, expect non-nil")
	}
}

func TestProcessStateUpdated(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	defer pbft.Close()

	// malicious proto.Message
	payload := &protos.StateUpdatedMessage{
		SeqNo: uint64(1),
	}
	msg, _ := proto.Marshal(payload)
	maliciousMsg := &protos.Message{
		Type:		protos.Message_STATE_UPDATED,
		Timestamp: 	time.Now().UnixNano(),
		Payload:	msg,
		Id:		uint64(1),
	}
	msg, _ = proto.Marshal(maliciousMsg)

	err := pbft.RecvMsg(msg)
	if err != nil {
		t.Error("RecvMsg error not nil, expect nil")
	}
}

func TestProcessNullRequest(t *testing.T) {

	pbft := new(pbftProtocal)
	pbft.replicaCount = 4 // otherwise integer divide by zero in nullReqTimerReset()
	pbft.id = uint64(1)
	pbft.inNegoView = false
	pbft.pbftManager = events.NewManagerImpl()
	pbft.pbftManager.SetReceiver(pbft)
	pbftTimerFactory := events.NewTimerFactoryImpl(pbft.pbftManager)
	pbft.nullRequestTimer = pbftTimerFactory.CreateTimer()

	message := &protos.Message{
		Type:		protos.Message_NULL_REQUEST,
		Timestamp:	time.Now().UnixNano(),
		Payload:	nil,
		Id:		uint64(1),
	}
	msg, _ := proto.Marshal(message)

	err := pbft.RecvMsg(msg)
	if err != nil {
		t.Errorf("RecvMsg error not nil, expect nil")
	}

	pbft.inNegoView = true
	err = pbft.RecvMsg(msg)
	if err != nil {
		t.Errorf("RecvMsg error not nil, expect nil")
	}
}

func TestRecvProcessNegotiateView(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	message := &protos.Message{
		Type: 		protos.Message_NEGOTIATE_VIEW,
		Timestamp:	time.Now().UnixNano(),
		Payload:	nil,
		Id:		uint64(1),
	}
	msg, _ := proto.Marshal(message)

	pbft.inNegoView = true
	err := pbft.RecvMsg(msg)
	if err != nil {
		t.Errorf("recv error not nil, expect nil")
	}
}

func TestProcessTxEvent(t *testing.T) {

	tx := &types.Transaction{}

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()
	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	pbft.activeView = false
	err := pbft.processTxEvent(tx)
	if err != nil {
		t.Errorf("processTxEvent error, expect nil")
	}

	pbft.activeView = true
	pbft.inNegoView = false
	pbft.inRecovery = false
	pbft.id = uint64(2)

	err = pbft.processTxEvent(tx)
	if err != nil {
		t.Errorf("processTxEvent error, expect nil")
	}

	pbft.id = uint64(1)
	err = pbft.processTxEvent(tx)
	if err != nil {
		t.Error("processTxEvent error, expect nil")
	}
}

func TestProcessCachedTxs(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	tx := &types.Transaction{}
	pbft.reqStore.storeOutstanding(tx)
	if pbft.reqStore.outstandingRequests.Len() != 1 {
		num := pbft.reqStore.outstandingRequests.Len()
		t.Errorf("reqStore outstanding requests number %d, expect %d", num, 1)
	}
	pbft.processCachedTransactions()
	if pbft.reqStore.outstandingRequests.Len() != 0 {
		num := pbft.reqStore.outstandingRequests.Len()
		t.Errorf("reqStore outstanding requests number %d, expect %d", num, 0)
	}
}

func TestLeaderProcReq(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	tx := &types.Transaction{}
	pbft.leaderProcReq(tx)

	if len(pbft.batchStore) != 1 {
		t.Errorf("leaderProcReq batch, batch store not 1 tx, expect 1")
	}

	if pbft.batchTimerActive == false {
		t.Errorf("leaderProcReq batch timer active flase, expect true")
	}
}

func TestSendBatch(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	tx := &types.Transaction{}
	pbft.leaderProcReq(tx)

	err := pbft.sendBatch()
	if err != nil {
		t.Errorf("sendBatch error, expect err nil")
	}
	if pbft.batchTimerActive {
		t.Errorf("after sendbatch, batchtimer still open, expect stop")
	}
}

func TestNullRequestHandler(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)


	pbft.id = 2
	pbft.nullRequestHandler()
	// pbft in negotiate view should not do anything
	if pbft.activeView == false {
		t.Errorf("test null request handler, pbft in nego view, should not send view change")
	}

	pbft.inNegoView = false
	pbft.activeView = true
	pbft.nullRequestHandler()
	if pbft.activeView == true {
		t.Errorf("test null request handler, pbft not primary, should send view change")
	}
}

func TestRecvStateUpdatedEvent(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	pbft.stateTransferring = true

	// pbft in nego view do nothing
	pbft.recvStateUpdatedEvent(nil)
	if pbft.stateTransferring == false {
		t.Errorf("recvStateUpdatedEvent, pbft in nego view, should do nothing")
	}

	// normal
	pbft.inNegoView = false
	// et.seqNo < pbft.h  hightStateTarget == nil
	event := &stateUpdatedEvent{
		seqNo:	uint64(40),
	}
	pbft.h = uint64(50)
	pbft.highStateTarget = nil

	err := pbft.recvStateUpdatedEvent(event)
	if err != nil {
		t.Errorf("recvStateupdatedEvent, expect error nil")
	}

	// et.seqNo < pbft.h  et.seqNo<pbft.highStateTarget.seqNo
	if pbft.stateTransferring == true {
		t.Errorf("recvStateUpdateEvent, pbft should end state transfer")
	}
	pbft.highStateTarget = &stateUpdateTarget{
		checkpointMessage:	checkpointMessage{
			seqNo:		80,
			id:		[]byte("checkpointMessage"),
		},
		replicas:		[]uint64{1, 2},
	}
	pbft.recvStateUpdatedEvent(event)
	if pbft.stateTransferring == false {
		t.Errorf("recvStateUpdatedEvent, pbft should in statetransferring")
	}
	pbft.stateTransferring = false

	// et.seqNo >= pbft.h
	event = &stateUpdatedEvent{
		seqNo: 	uint64(80),
	}
	err = pbft.recvStateUpdatedEvent(event)
	if pbft.lastExec != event.seqNo {
		t.Errorf("recvStateUpdatedEvent, pbft should update lastExec")
	}
	if pbft.h != event.seqNo {
		t.Errorf("recvStateUpdatedEvent, pbft should move water mark")
	}

	// pbft in recovery
	event = &stateUpdatedEvent{
		seqNo: 	uint64(90),
	}
	err = pbft.recvStateUpdatedEvent(event)
	if err != nil {
		t.Errorf("recvStateUpdatedEvent, pbft in recovery, expect error nil")
	}



}

func TestRecvRequestBatch(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)


	reqBatch := &TransactionBatch{}
	err := pbft.recvRequestBatch(reqBatch)
	if err != nil {
		t.Errorf("RecvReqBatch, pbft in nego view, expect error nil")
	}

	pbft.inNegoView = false
	pbft.recvRequestBatch(reqBatch)
	if pbft.vid != uint64(1) {
		t.Errorf("RecvReqBatch vid should add 1")
	}
}

func TestValidateBatch(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	// primary
	batch := &TransactionBatch{}
	pbft.validateBatch(batch, 0, 0)
	if pbft.vid != uint64(1) {
		t.Errorf("RecvReqBatch vid should add 1")
	}

	// not primary, not inWV
	pbft.id = 2
	pbft.validateBatch(batch, 1, 0)

	// not primary, inWV
	pbft.validateBatch(batch, 0, 0)
}

func TestCallSendPrePrepare(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	// cache nil
	ret := pbft.callSendPrePrepare("")
	if ret == true {
		t.Errorf("cache nil, expect return true")
	}

	// cache.vid != pbft.lastVid + 1
	batch := &cacheBatch{
		batch:	nil,
		vid:	uint64(2),
	}
	pbft.cacheValidatedBatch["a"] = batch
	ret = pbft.callSendPrePrepare("a")
	if ret == true {
		t.Error("cache.vid != pbft.lastVid+1, expect return false")
	}

	// len(cache.batch.Batch) == 0
	uint2 := uint64(2)
	pbft.currentVid = &uint2
	pbft.lastVid = uint64(1)
	batch2 := &cacheBatch{
		batch:	&TransactionBatch{
			Batch:	[]*types.Transaction{},
			Timestamp: int64(1),
		},
		vid:	uint64(2),
	}
	pbft.cacheValidatedBatch["a"] = batch2
	ret = pbft.callSendPrePrepare("a")
	if ret == false || pbft.currentVid != nil {
		t.Errorf("len of cache.batch.Batch is 0, expect return true")
	}

	// sendPreprepare
	//pbft.currentVid = &uint2
	//fmt.Println("cur", pbft.currentVid)
	//pbft.lastVid = uint64(1)
	//batch3 := &cacheBatch{
	//	batch:	&TransactionBatch{
	//		Batch:	[]*types.Transaction{&types.Transaction{}, &types.Transaction{}},
	//		Timestamp: int64(1),
	//	},
	//	vid:	uint64(2),
	//}
	//
	//pbft.cacheValidatedBatch["a"] = batch3
	//ret = pbft.callSendPrePrepare("a")
	//fmt.Println("ret", ret, pbft.currentVid)
	//if ret == false || *(pbft.currentVid) != batch3.vid {
	//	t.Errorf("callSendPrePrepare should return true")
	//}


}

func TestSendPrePrepare(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)

	preprep2 := &PrePrepare{
		View:			uint64(0),
		SequenceNumber:		uint64(2),
		BatchDigest:		"digest",
		TransactionBatch: 	nil,
		ReplicaId:		pbft.id,
	}

	// same digest different seqNo

	pbft.seqNo = uint64(0)
	cert := pbft.getCert(uint64(0), uint64(0))
	cert.prePrepare = preprep2
	cert.digest = "digest"

	pbft.sendPrePrepare(nil, "digest")
	if pbft.seqNo == 1 {
		t.Errorf("sendPrePrepare should not handle this preprepare")
	}

	// not inWV
	H := pbft.K * pbft.L
	pbft.seqNo = H
	pbft.sendPrePrepare(nil, "")
	if pbft.seqNo == 1 {
		t.Errorf("sendPrePrepare should not handle this preprepare")
	}

	// normal
	pbft.seqNo = uint64(0)
	pbft.view = uint64(0)
	curVid := uint64(0)
	pbft.currentVid = &curVid

	pbft.sendPrePrepare(nil, "normal")

	cert = pbft.getCert(uint64(0), uint64(1))
	if cert.digest != "normal" {
		t.Errorf("should be 'normal'")
	}
	if pbft.currentVid != nil {
		t.Errorf("should be nil")
	}
}

func TestRecvPrePrepare(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false

	// normal self replica 1, recv from replica 0
	txBatch := &TransactionBatch{
		Batch:		[]*types.Transaction{&types.Transaction{}},
		Timestamp:	int64(1),
	}
	pp := &PrePrepare{
		View:		uint64(0),
		SequenceNumber: uint64(1),
		BatchDigest:    "normal",
		ReplicaId:	uint64(1),
		TransactionBatch: txBatch,
	}
	pbft.recvPrePrepare(pp)

	cert := pbft.getCert(pp.View, pp.SequenceNumber)
	if cert.sentPrepare == false {
		t.Errorf("recv preprepare, should send prepare")
	}

	// pbft in negotiate view
	pbft.inNegoView = true
	err := pbft.recvPrePrepare(pp)
	if err != nil {
		t.Errorf("recv preprepare, in nego view, should not handle")
	}
	pbft.inNegoView = false

	// pbft in view change
	pbft.activeView = false
	err = pbft.recvPrePrepare(pp)
	if err != nil {
		t.Errorf("recv preprepare, in view change, should not handle")
	}
	pbft.activeView = true

	// pp not from primary
	pp2 := &PrePrepare{
		View:		uint64(0),
		SequenceNumber: uint64(1),
		BatchDigest:    "normal",
		ReplicaId:	uint64(2),
		TransactionBatch: txBatch,
	}
	err = pbft.recvPrePrepare(pp2)
	if err != nil {
		t.Errorf("recv preprepare, replicaId is not primary, should not handle")
	}

	// pp not in WV
	pbft.view = uint64(0)
	pp3 := &PrePrepare{
		View:		uint64(1),
		SequenceNumber: uint64(1),
		BatchDigest:    "normall",
		ReplicaId:	uint64(1),
		TransactionBatch: txBatch,
	}
	pbft.recvPrePrepare(pp3)
	cert = pbft.getCert(pp3.View, pp3.SequenceNumber)
	if cert.digest == pp3.BatchDigest {
		t.Errorf("recv preprepare, not in WV, expect not receipt")
	}


}

func TestRecvPrepare(t *testing.T) {

	// normal
	// test in maybeSendCommit

	// in negotiate view
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false

	pbft.inNegoView = true
	err := pbft.recvPrepare(&Prepare{})
	if err != nil {
		t.Errorf("recvPrepare while in negotiate view")
	}
	pbft.inNegoView = false

	// recv prepare from primary
	pbft.inRecovery = false
	prep := &Prepare{
		View:		uint64(0),
		SequenceNumber: uint64(1),
		BatchDigest:	"digest",
		ReplicaId:	uint64(1),
	}
	err = pbft.recvPrepare(prep)
	if err != nil {
		t.Errorf("recvPrepare from primary")
	}

	// recv prepare not in WV
	outofH := pbft.K * pbft.L + 1
	prep2 := &Prepare{
		View:		uint64(0),
		SequenceNumber: outofH,
		BatchDigest:	"digest",
		ReplicaId:	uint64(2),
	}
	err = pbft.recvPrepare(prep2)
	if err != nil {
		t.Errorf("recvPrepare not in WV")
	}



}

func TestMaybeSendCommit(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false

	// not preprared
	pbft.maybeSendCommit("", uint64(0), uint64(1))

	// prepared but in state transfer
	pbft.skipInProgress = true
	txBatch := &TransactionBatch{}
	pbft.validatedBatchStore["a"] = txBatch
	cert := pbft.getCert(uint64(0), uint64(1))
	pp := &PrePrepare{
		View:		uint64(0),
		SequenceNumber: uint64(1),
		BatchDigest:    "a",
		ReplicaId:	uint64(1),
		TransactionBatch: txBatch,
	}
	cert.prePrepare = pp // now preprepared
	cert.prepareCount = pbft.preparedReplicasQuorum() // now prepared

	err := pbft.maybeSendCommit(pp.BatchDigest, pp.View, pp.SequenceNumber)
	if err != nil {
		t.Errorf("should end in not prepared")
	}

	// normal replica
	pbft.skipInProgress = false
	txBatch2 := &TransactionBatch{}
	pp2 := &PrePrepare{
		View:		uint64(0),
		SequenceNumber: uint64(1),
		BatchDigest:    "a",
		ReplicaId:	uint64(1),
		TransactionBatch: txBatch2,
	}
	pbft.maybeSendCommit(pp2.BatchDigest, pp2.View, pp2.SequenceNumber)
	if cert.sentValidate == false {
		t.Errorf("replica is expected to sentValidate")
	}

	// normal primary
	// test in sendCommit
}

func TestSendCommit(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false

	// normal
	d := "digest"
	v := uint64(0)
	n := uint64(1)

	err := pbft.sendCommit(d, v, n)
	if err != nil {
		t.Errorf("sendCommit err")
	}
	cert := pbft.getCert(v, n)
	if cert.sentCommit == false {
		t.Errorf("cert sent commit false, expect true")
	}
}

func TestRecvCommit(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false

	// normal, committed, execute
	v := uint64(0)
	n := uint64(1)
	d := "digest"
	replicaId := uint64(1)
	cmt := &Commit{
		View:		v,
		SequenceNumber:	n,
		BatchDigest:	d,
		ReplicaId:	replicaId,
	}

	pbft.validatedBatchStore[d] = &TransactionBatch{}
	txBatch := &TransactionBatch{
		Timestamp:	int64(1),
	}
	preprep := &PrePrepare{
		View: 		v,
		SequenceNumber:	n,
		BatchDigest:	d,
		TransactionBatch: txBatch,
	}
	cert := pbft.getCert(v, n)
	cert.digest = d
	cert.prePrepare = preprep // now preprepared
	cert.prepareCount = pbft.preparedReplicasQuorum() // now prepared
	cert.commitCount = pbft.committedReplicasQuorum() // now committed
	cert.sentExecute = false
	cert.validated = true
	err := pbft.recvCommit(cmt)
	if err != nil {
		t.Errorf("recvCommit normal, expect excute")
	}
	certIdx := msgID{v: v, n: n}
	cert = pbft.certStore[certIdx]
	if cert == nil {
		t.Errorf("recvCommit normal, expect cert not nil")
	}
	if cert.commit[*cmt] != true {
		t.Errorf("recvCommit normal, expect commit exist")
	}

	if cert.sentExecute != true {
		t.Errorf("recvCommit should send execute")
	}

	// normal, not committed
	v2 := uint64(0)
	n2 := uint64(2)
	d2 := "digest2"
	replicaId = uint64(1)
	cmt2 := &Commit{
		View:		v2,
		SequenceNumber:	n2,
		BatchDigest:	d2,
		ReplicaId:	replicaId,
	}

	pbft.validatedBatchStore[d] = &TransactionBatch{}
	txBatch2 := &TransactionBatch{
		Timestamp:	int64(1),
	}
	preprep2 := &PrePrepare{
		View: 		v2,
		SequenceNumber:	n2,
		BatchDigest:	d2,
		TransactionBatch: txBatch2,
	}
	cert = pbft.getCert(v2, n2)
	cert.digest = d
	cert.prePrepare = preprep2// now preprepared
	cert.prepareCount = pbft.preparedReplicasQuorum() // now prepared

	pbft.recvCommit(cmt2)
	if cert.commitCount >= pbft.committedReplicasQuorum() {
		t.Errorf("recv commit, expect not commited")
	}
}

func TestExecuteAfterStateUpdate(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(5)


	// certs in certstore
	// cert1: idx.n <= pbft.seqNo
	v1, n1, _ := uint64(0), uint64(1), "d1"
	cert1 := pbft.getCert(v1, n1)
	pbft.executeAfterStateUpdate()
	if cert1.sentValidate == true {
		t.Errorf("executeAfterStateUpdate n < seqNo not handle this")
	}

	// cert2: idx.n > pbft.seqNo, not prepared
	v2, n2, d2 := uint64(0), uint64(6), "d2"
	cert2 := pbft.getCert(v2, n2)
	txBatch := &TransactionBatch{}
	pbft.validatedBatchStore[d2] = txBatch

	pp := &PrePrepare{
		View:		v2,
		SequenceNumber: n2,
		BatchDigest:    d2,
		ReplicaId:	uint64(1),
		TransactionBatch: txBatch,
	}
	cert2.prePrepare = pp // now preprepared
	cert2.digest = d2
	cert2.prepareCount = pbft.preparedReplicasQuorum() - 1
	pbft.executeAfterStateUpdate()
	if cert2.sentValidate == true {
		t.Errorf("executeAfterStateUpdate n > seqNo, but not prepared, not handle this")
	}

	// cert3: idx.n > pbft.seqNo, prepared, already validated
	v3, n3, d3 := uint64(0), uint64(7), "d3"
	cert3 := pbft.getCert(v3, n3)
	pbft.validatedBatchStore[d3] = txBatch
	pp3 := &PrePrepare{
		View:		v3,
		SequenceNumber:	n3,
		BatchDigest: 	d3,
		ReplicaId: 	uint64(1),
		TransactionBatch: txBatch,
	}
	cert3.prePrepare = pp3
	cert3.digest = d3
	cert3.prepareCount = pbft.preparedReplicasQuorum()
	cert3.validated = true
	pbft.executeAfterStateUpdate()
	if cert3.sentValidate == true {
		t.Errorf("executeAfterStateUpdate n > seqNo, prepared, already validated")
	}
	//cert4: idx.n > pbft.seqNo, prepared, not validated yet
	v4, n4, d4 := uint64(0), uint64(8), "d4"
	cert4 := pbft.getCert(v4, n4)
	pbft.validatedBatchStore[d4] = txBatch
	pp4 := &PrePrepare{
		View: 		v4,
		SequenceNumber: n4,
		BatchDigest:    d4,
		ReplicaId: 	uint64(1),
		TransactionBatch: txBatch,
	}
	cert4.prePrepare = pp4
	cert4.digest = d4
	cert4.prepareCount = pbft.preparedReplicasQuorum()
	cert4.validated = false
	pbft.executeAfterStateUpdate()
	if cert4.sentValidate == false {
		t.Errorf("executeAfterStateUpdate, expect sentValidate")
	}

}

func TestExecuteOne(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(0)

	// the cert is nil
	idx1 := msgID{0, 1}
	ret := pbft.executeOne(idx1)
	if ret == true {
		t.Errorf("cert is nil , expect false")
	}

	// the cert preprepare is nil
	idx2 := msgID{0, 1}
	cert2 := pbft.getCert(0, 1)
	cert2.prePrepare = nil
	ret = pbft.executeOne(idx2)
	if ret == true {
		t.Errorf("cert.prePrepare is nil, expect false")
	}

	// cert already sentExecute
	idx3 := msgID{0, 1}
	cert3 := pbft.getCert(0, 1)
	pp3 := &PrePrepare{
		View: 	0,
		SequenceNumber:	1,
		BatchDigest:    "digest",
		TransactionBatch: nil,
		ReplicaId:	1,
	}
	cert3.prePrepare = pp3
	cert3.sentExecute = true
	ret = pbft.executeOne(idx3)
	if ret == true {
		t.Errorf("cert sentExecute, expect return false")
	}

	// not the one to execute
	idx4 := msgID{0, 2}
	cert4 := pbft.getCert(0, 2)
	pp4 := &PrePrepare{
		View:		0,
		SequenceNumber:	2,
		BatchDigest:	"digest",
		TransactionBatch:	nil,
		ReplicaId:	1,
	}
	cert4.prePrepare = pp4
	cert4.sentExecute = false
	pbft.lastExec = uint64(0)
	ret = pbft.executeOne(idx4)
	if ret == true {
		t.Errorf("not the one to execute, expect return false")
	}

	// in skipInProgress
	idx5 := msgID{0, 3}
	cert5 := pbft.getCert(0, 3)
	pp5 := &PrePrepare{
		View:		0,
		SequenceNumber:	3,
		BatchDigest:	"digest",
		TransactionBatch: nil,
		ReplicaId:	1,
	}
	cert5.prePrepare = pp5
	cert5.sentExecute = false
	pbft.lastExec = uint64(2)
	pbft.skipInProgress = true
	ret = pbft.executeOne(idx5)
	if ret == true {
		t.Errorf("in the checkpoint, expect false")
	}

	// not committed
	idx6 := msgID{0, 4}
	cert6 := pbft.getCert(0, 4)
	pp6 := &PrePrepare{
		View:		0,
		SequenceNumber: 3,
		BatchDigest:	"digest",
		TransactionBatch: nil,
		ReplicaId:	1,
	}
	cert6.prePrepare = pp6
	cert6.sentExecute = false
	pbft.lastExec = uint64(3)
	pbft.skipInProgress = false
	ret = pbft.executeOne(idx6)
	if ret == true {
		t.Errorf("not committed, expect false")
	}


	// normal
	idx7 := msgID{0, 5}
	cert7 := pbft.getCert(0, 5)
	txBatch := &TransactionBatch{
		Timestamp:	1,
	}
	pp7 := &PrePrepare{
		View:		0,
		SequenceNumber: 5,
		BatchDigest:	"digest",
		TransactionBatch: txBatch,
		ReplicaId:	1,
	}

	cert7.prePrepare = pp7
	pbft.validatedBatchStore["digest"] = txBatch

	cert7.sentExecute = false
	pbft.lastExec = uint64(4)
	pbft.skipInProgress = false
	cert7.digest = "digest"
	cert7.prepareCount = pbft.preparedReplicasQuorum()
	cert7.commitCount = pbft.committedReplicasQuorum()
	ret = pbft.executeOne(idx7)
	if ret == false {
		t.Errorf("expect true")
	}
}

func TestExecDoneSync(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(0)

	pbft.currentExec = nil
	pbft.skipInProgress = false
	pbft.execDoneSync(msgID{0, 1})
	if pbft.skipInProgress == false {
		t.Errorf("currentExec is nil, expect true")
	}
}

func TestCheckpoint(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:		10,
		PreviousBlockHash:	[]byte("previous"),
		CurrentBlockHash:	[]byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(10)
	pbft.checkpoint(seqNo, bcInfo)
	idString, _ := pbft.chkpts[seqNo]
	if idString != idAsString {
		t.Errorf("checkpoint error")
	}
}

func TestRecvCheckpoint(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:		10,
		PreviousBlockHash:	[]byte("previous"),
		CurrentBlockHash:	[]byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(10)
	chkpt := &Checkpoint{
		SequenceNumber:		seqNo,
		ReplicaId:		uint64(3),
		Id:			idAsString,
	}
	cert := pbft.getChkptCert(seqNo, chkpt.Id)
	cert.chkptCount = pbft.intersectionQuorum()-1
	pbft.chkpts[seqNo] = idAsString
	pbft.recvCheckpoint(chkpt)
	if pbft.h != 10 {
		t.Errorf("should movewatermarks")
	}
}

func TestWeakCheckpointSetOutOfRange(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:		50,
		PreviousBlockHash:	[]byte("previous"),
		CurrentBlockHash:	[]byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(50)
	chkpt := &Checkpoint{
		SequenceNumber:		seqNo,
		ReplicaId:		uint64(3),
		Id:			idAsString,
	}
	pbft.hChkpts[1] = uint64(50)

	pbft.weakCheckpointSetOutOfRange(chkpt)

	if pbft.h == 0 {
		t.Errorf("should move water marks")
	}
}

func TestWitnessCheckpointWeakCert(t *testing.T) {
	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:		10,
		PreviousBlockHash:	[]byte("previous"),
		CurrentBlockHash:	[]byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(10)
	chkpt := &Checkpoint{
		SequenceNumber:		seqNo,
		ReplicaId:		uint64(3),
		Id:			idAsString,
	}
	pbft.checkpointStore[*chkpt] = true
	chkpt2 := &Checkpoint{
		SequenceNumber:  	seqNo,
		ReplicaId:		uint64(2),
		Id:			idAsString,
	}
	pbft.checkpointStore[*chkpt2] = true

	pbft.highStateTarget = nil

	pbft.witnessCheckpointWeakCert(chkpt)

	if pbft.highStateTarget == nil {
		t.Errorf("should update highStateTarget")
	}
}

func TestFetchRequestBatches(t *testing.T) {
	// TODO
}

func TestRecvFetchRequestBatch(t *testing.T) {
	// TODO
}

func TestRecvReturnRequestBatch(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.activeView = true
	pbft.seqNo = uint64(0)

	txBatch := &TransactionBatch{
		Batch:		nil,
		Timestamp:	1,
	}

	digest := hash(txBatch)
	pbft.missingReqBatches[digest] = true
	pbft.recvReturnRequestBatch(txBatch)
	_, ok := pbft.missingReqBatches[digest]
	if ok {
		t.Errorf("missing request batch should be 0")
	}
}

func TestMoveWatermarks(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.activeView = true
	pbft.seqNo = uint64(0)
	// set current h to 50
	pbft.h = uint64(50)
	// fill some record into store
	cert := pbft.getCert(0, 1)
	cert.digest = "v0n1digest"
	// validatedBatchStore
	pbft.validatedBatchStore[cert.digest] = &TransactionBatch{}
	// outstanding req batch
	pbft.outstandingReqBatches[cert.digest] = &TransactionBatch{}

	testChkpt := &Checkpoint{
		SequenceNumber:		10,
	}
	// checkpointstore
	pbft.checkpointStore[*testChkpt] = true
	// chkptCertStore
	chkpts := make(map[Checkpoint]bool)
	chkpts[*testChkpt] = true
	pbft.getChkptCert(10, "chkptId")

	// pset
	vcP := &ViewChange_PQ{}
	pbft.pset[10] = vcP

	// qset
	qIdx := qidx{d: "chkpt", n: 10}
	vcQ := &ViewChange_PQ{}
	pbft.qset[qIdx] = vcQ

	// chkpts
	pbft.chkpts[10] = "chkpt"

	// this seqno should not trigger movewatermark
	pbft.moveWatermarks(40)

	if _, ok := pbft.chkpts[10]; ok != true {
		t.Errorf("should not move watermark, expect true")
	}
	// test if all delete
	pbft.moveWatermarks(60)
	// validatedBatchStore
	if _, ok := pbft.validatedBatchStore[cert.digest]; ok {
		t.Errorf("should not move watermark, expect vb record not exist")
	}
	// outstanding req batch
	if _, ok := pbft.outstandingReqBatches[cert.digest]; ok {
		t.Errorf("should not move watermark, expect ob record not exist")
	}
	// checkpoint store
	if _, ok := pbft.checkpointStore[*testChkpt]; ok {
		t.Errorf("should not move watermark, expect checkpoint store record not exist")
	}
	// checkpoint cert store
	if _, ok := pbft.chkptCertStore[chkptID{n:10, id:"chkptId"}]; ok {
		t.Errorf("should not move watermark, expect chkptCertStore record not exist")
	}
	if _, ok := pbft.pset[10]; ok {
		t.Errorf("should not move watermark, expect pset record not exist")
	}
	if _, ok := pbft.qset[qidx{d: "chkpt", n: 10}]; ok {
		t.Errorf("should not move watermark, expect qset record not exist")
	}
	if _, ok := pbft.chkpts[10]; ok {
		t.Errorf("should not move watermark, expect chkpts record not exist")
	}
	if _, ok := pbft.chkpts[10]; ok != false {
		t.Errorf("should not move watermark, expect false")
	}
}

func TestUpdateHighStateTarget(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.activeView = true
	pbft.seqNo = uint64(0)

	checkpoint := checkpointMessage{
		seqNo:		20,
		id:		[]byte("checkpoint"),
	}
	peers := []uint64{1, 2}
	curTarget := &stateUpdateTarget {
		checkpointMessage:	checkpoint,
		replicas:		peers,
	}
	pbft.highStateTarget = curTarget

	// new target seqNo <= cur target seqNo
	newCheckpoint1 := checkpointMessage{
		seqNo:		10,
		id:		[]byte("checkpoint"),
	}
	newTargetSmaller := &stateUpdateTarget{
		checkpointMessage:	newCheckpoint1,
		replicas:		peers,
	}
	pbft.updateHighStateTarget(newTargetSmaller)
	if pbft.highStateTarget.checkpointMessage.seqNo > curTarget.checkpointMessage.seqNo {
		t.Errorf("should not update high state target, expect not change")
	}

	// new target seqNo > cur target seqNo
	newCheckpoint2 := checkpointMessage{
		seqNo:		30,
		id:		[]byte("checkpoint"),
	}
	newTargetLarger := &stateUpdateTarget {
		checkpointMessage:	newCheckpoint2,
		replicas:		peers,
	}
	pbft.updateHighStateTarget(newTargetLarger)
	if pbft.highStateTarget.checkpointMessage.seqNo <= curTarget.checkpointMessage.seqNo {
		t.Errorf("shoul update high state target, expect change")
	}
}

func TestStateTransfer(t *testing.T) {
	// TODO
}

func TestRetryStateTransfer(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = false
	pbft.activeView = true
	pbft.seqNo = uint64(0)

	pbft.stateTransferring = false
	newCheckpoint := checkpointMessage{
		seqNo:		10,
		id:		[]byte("checkpoint"),
	}
	peers := []uint64{1,2}
	target := &stateUpdateTarget{
		checkpointMessage:	newCheckpoint,
		replicas:		peers,
	}
	pbft.retryStateTransfer(target)
	if pbft.stateTransferring == false {
		t.Errorf("should in statetransferring")
	}

	pbft.stateTransferring = false
	pbft.highStateTarget = target
	pbft.retryStateTransfer(nil)
	if pbft.stateTransferring == false {
		t.Errorf("should in statetransferring")
	}
}

func TestResubmitRequestBatches(t *testing.T) {
	// TODO
}

func TestSkipTo(t *testing.T) {
	// TODO
}

func TestUpdateState(t *testing.T) {
	// TODO
}

func TestUpdateViewChangeSeqNo(t *testing.T) {
	// TODO
}

func TestProcessNegotiateView(t *testing.T) {

}

func TestRecvNegoView(t *testing.T) {

}

func TestRecvNegoViewRsp(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.id = uint64(2)
	pbft.inNegoView = true
	pbft.activeView = true
	pbft.seqNo = uint64(0)

	var ret events.Event
	pbft.negoViewRspStore = make(map[uint64]uint64)

	// duplicate negoViewRsp
	nvr1 := &NegotiateViewResponse{
		ReplicaId:	1,
		View:		0,
	}
	pbft.negoViewRspStore[nvr1.ReplicaId] = nvr1.View
	ret = pbft.recvNegoViewRsp(nvr1)
	if ret != nil {
		t.Errorf("recvNegoViewRsp duplicate, expect ret nil")
	}

	// recv negoViewRsp doesn't above N-f
	nvr2 := &NegotiateViewResponse{
		ReplicaId:	2,
		View:		0,
	}
	ret = pbft.recvNegoViewRsp(nvr2)
	if ret != nil {
		t.Errorf("recvNegoViewRsp not reach N-f+1, expect ret nil")
	}

	// recv negoViewRsp above N-f but cannot find quorum
	nvr3 := &NegotiateViewResponse{
		ReplicaId:	3,
		View:		1,
	}
	ret = pbft.recvNegoViewRsp(nvr3)
	if ret != nil {
		t.Errorf("recvNegoViewRsp above N-f, but cannot find quorum, expect ret nil")
	}

	// recv negoViewRsp above N-f and find quorum
	nvr4 := &NegotiateViewResponse{
		ReplicaId:	4,
		View:		0,
	}
	ret = pbft.recvNegoViewRsp(nvr4)
	if _, ok := ret.(negoViewDoneEvent); !ok {
		t.Errorf("recvNegoViewRsp achieve quorum, expect ret negoViewDoneEvent")
	}
}

func TestRecvValidateResult(t *testing.T) {

	core.InitDB("/temp/leveldb", 8088)
	defer clearDB()

	id := 1
	pbftConfigPath := getPbftConfigPath()
	config := loadConfig(pbftConfigPath)
	eventMux := new(event.TypeMux)
	h := helper.NewHelper(eventMux)
	pbft := newPbft(uint64(id), config, h)
	pbft.inNegoView = false
	pbft.activeView = true
	pbft.seqNo = uint64(0)

	// primary recv ValidateResult
	pbft.id = uint64(1)
	pbft.view = uint64(0)

	txes := make([]*types.Transaction, 1)
	vali := event.ValidatedTxs{
		Transactions: 	txes,
		Hash:		"hash",
		SeqNo:		1,
		View: 		0,
		Timestamp:	1,
	}
	pbft.recvValidatedResult(vali)
	if _, ok := pbft.validatedBatchStore["hash"]; !ok {
		t.Errorf("primary recv ValidatedResult, expect exist in validatedBatchStore")
	}

	if _, ok := pbft.outstandingReqBatches["hash"]; !ok {
		t.Errorf("primary recv Validatedresult, expect exist in outstandingReqBatches")
	}

	// replica recvValidatedResult
	pbft.id = uint64(2)
	pbft.recvValidatedResult(vali)
	cert, ok := pbft.certStore[msgID{v: 0, n: 1}]
	if ok != true {
		t.Errorf("replica recv validated result, expect exist in cert store")
	}
	if cert.validated != true {
		t.Errorf("replica recv validated result, expect validated")
	}

}