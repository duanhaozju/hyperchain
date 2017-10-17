//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	"hyperchain/common"
	"hyperchain/consensus/consensusMocks"
	"hyperchain/consensus/helper/persist"
	mdb "hyperchain/hyperdb/mdb"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func TestNewRecoveryMgr(t *testing.T) {
	recoveryMgr := newRecoveryMgr()
	structName, nilElems, err := checkNilElems(recoveryMgr)
	if err != nil {
		t.Error(err.Error())
	}
	if nilElems != nil {
		t.Errorf("There exists some nil elements: %v in struct: %s", nilElems, structName)
	}
}

func TestInitNegoView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		negoView := &NegotiateView{}
		err = proto.Unmarshal(consensus.Payload, negoView)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(rbft.id, negoView.ReplicaId, "initNegoView failed")
	}()

	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.initNegoView()

	rbft.status.activeState(&rbft.status.inNegoView)
	rbft.initNegoView()

	time.Sleep(1 * time.Nanosecond)
}

func TestRecvNegoView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerUnicast").Return(nil)
	from := uint64(9)
	go func() {
		_ = <-eChan
		to := <-eChan
		ast.Equal(from, to, "recvNegoView failed")
	}()

	atomic.StoreUint32(&rbft.activeView, 1)
	negoViewMsg := &NegotiateView{
		ReplicaId: from,
	}
	payload, err := proto.Marshal(negoViewMsg)
	ast.Nil(err, err)
	consensusMsg := &ConsensusMessage{
		Type:    ConsensusMessage_NEGOTIATE_VIEW,
		Payload: payload,
	}
	msg := cMsgToPbMsg(consensusMsg, rbft.id)
	tmpMsg, err := proto.Marshal(msg)
	ast.Nil(err, err)
	rbft.RecvMsg(tmpMsg)

	time.Sleep(1 * time.Nanosecond)
}

func TestRecvNegoViewRsp(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	var ret consensusEvent
	rbft.status.inActiveState(&rbft.status.inNegoView)
	nvr := &NegotiateViewResponse{}
	ret = rbft.recvNegoViewRsp(nvr)
	ast.Nil(ret, "not in inNegoView, expect ret nil")

	rbft.id = uint64(2)
	rbft.status.activeState(&rbft.status.inNegoView)
	atomic.StoreUint32(&rbft.activeView, 1)
	rbft.seqNo = uint64(0)
	rbft.recoveryMgr.negoViewRspStore = make(map[uint64]*NegotiateViewResponse)
	// duplicate negoViewRsp
	nvr1 := &NegotiateViewResponse{
		ReplicaId: 1,
		View:      0,
	}
	rbft.recoveryMgr.negoViewRspStore[nvr1.ReplicaId] = nvr1
	ret = rbft.recvNegoViewRsp(nvr1)
	ast.Nil(ret, "recvNegoViewRsp duplicate, expect ret nil")

	// recv negoViewRsp doesn't above N-f
	nvr2 := &NegotiateViewResponse{
		ReplicaId: 2,
		View:      0,
	}
	ret = rbft.recvNegoViewRsp(nvr2)
	ast.Nil(ret, "recvNegoViewRsp not reach N-f+1, expect ret nil")

	// recv negoViewRsp above N-f but cannot find quorum
	nvr3 := &NegotiateViewResponse{
		ReplicaId: 3,
		View:      1,
	}
	ret = rbft.recvNegoViewRsp(nvr3)
	ast.Nil(ret, "recvNegoViewRsp above N-f, but cannot find quorum, expect ret nil")

	// recv negoViewRsp above N-f and find quorum
	nvr4 := &NegotiateViewResponse{
		ReplicaId: 4,
		View:      0,
	}
	ret = rbft.recvNegoViewRsp(nvr4)
	e, ok := ret.(*LocalEvent)
	ast.Equal(true, ok, "recvNegoViewRsp achieve quorum, expect LocalEvent")
	if ok {
		ast.Equal(RECOVERY_NEGO_VIEW_DONE_EVENT, e.EventType, "recvNegoViewRsp achieve quorum, expect RECOVERY_NEGO_VIEW_DONE_EVENT")
	}
}

func TestInitRecovery(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(ConsensusMessage_RECOVERY_INIT, consensus.Type, "initRecovery failed: expect ConsensusMessage_RECOVERY_INIT")
		reco := &RecoveryInit{}
		err = proto.Unmarshal(consensus.Payload, reco)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(rbft.id, reco.ReplicaId, "initRecovery failed")
	}()

	seqNo := uint64(10)
	rbft.recoveryMgr.recoveryToSeqNo = &seqNo
	rbft.initRecovery()
	ast.Nil(rbft.recoveryMgr.recoveryToSeqNo, "initRecovery, expect nil")

	time.Sleep(1 * time.Nanosecond)
}

func TestRestartRecovery(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	nvr1 := &NegotiateViewResponse{N: uint64(1)}
	nvr2 := &NegotiateViewResponse{N: uint64(2)}
	rbft.recoveryMgr.negoViewRspStore[uint64(1)] = nvr1
	rbft.recoveryMgr.negoViewRspStore[uint64(2)] = nvr2

	rr1 := &RecoveryResponse{BlockHeight: uint64(1)}
	rr2 := &RecoveryResponse{BlockHeight: uint64(2)}
	rbft.recoveryMgr.rcRspStore[uint64(1)] = rr1
	rbft.recoveryMgr.rcRspStore[uint64(2)] = rr2

	rbft.status.inActiveState(&rbft.status.inNegoView, &rbft.status.inRecovery)

	ast.Equal(2, len(rbft.recoveryMgr.negoViewRspStore), "set failed")
	ast.Equal(2, len(rbft.recoveryMgr.rcRspStore), "set failed")
	ast.Equal(false, rbft.status.getState(&rbft.status.inNegoView), "inActiveState failed")
	ast.Equal(false, rbft.status.getState(&rbft.status.inRecovery), "inActiveState failed")

	rbft.restartRecovery()

	ast.Equal(0, len(rbft.recoveryMgr.negoViewRspStore), "restartRecovery failed")
	ast.Equal(0, len(rbft.recoveryMgr.rcRspStore), "restartRecovery failed")
	ast.Equal(true, rbft.status.getState(&rbft.status.inNegoView), "restartRecovery failed")
	ast.Equal(true, rbft.status.getState(&rbft.status.inRecovery), "restartRecovery failed")
}

func TestRecvRecovery(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerUnicast").Return(nil)
	from := uint64(9)
	go func() {
		_ = <-eChan
		to := <-eChan
		ast.Equal(from, to, "recvRecovery failed")
	}()

	rbft.status.activeState(&rbft.status.skipInProgress)
	recoveryMsg := &RecoveryInit{
		ReplicaId: from,
	}
	event := rbft.recvRecovery(recoveryMsg)
	ast.Nil(event, "skipInProgress, expect nil")

	rbft.status.inActiveState(&rbft.status.skipInProgress)
	rbft.recvRecovery(recoveryMsg)

	time.Sleep(1 * time.Nanosecond)
}

func TestRecvRecoveryRsp(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	rr1 := &RecoveryResponse{
		ReplicaId:   uint64(1),
		Chkpts:      make(map[uint64]string),
		BlockHeight: uint64(100),
		Genesis:     uint64(100),
	}
	encode60 := byteToString([]byte("60"))
	encode70 := byteToString([]byte("70"))
	encode80 := byteToString([]byte("80"))

	rr1.Chkpts[uint64(60)] = encode60
	rr1.Chkpts[uint64(70)] = encode70
	rr1.Chkpts[uint64(80)] = encode80

	rr2 := &RecoveryResponse{
		ReplicaId:   uint64(2),
		Chkpts:      make(map[uint64]string),
		BlockHeight: uint64(100),
		Genesis:     uint64(100),
	}
	rr2.Chkpts[uint64(60)] = "60"

	rr3 := &RecoveryResponse{
		ReplicaId:   uint64(3),
		Chkpts:      make(map[uint64]string),
		BlockHeight: uint64(100),
		Genesis:     uint64(100),
	}
	rr3.Chkpts[uint64(60)] = encode60

	// pbft not in recovery, should just return
	rbft.status.inActiveState(&rbft.status.inRecovery)
	rbft.recvRecoveryRsp(rr1)
	_, ok := rbft.recoveryMgr.rcRspStore[uint64(1)]
	ast.Equal(false, ok, "Pbft not in recovery, expect false")

	rbft.status.activeState(&rbft.status.inRecovery)
	rbft.recvRecoveryRsp(rr1)
	_, ok = rbft.recoveryMgr.rcRspStore[uint64(1)]
	ast.Equal(true, ok, "Pbft in recovery, expect true")

	recoveryToSeqNo := uint64(10)
	rbft.recoveryMgr.recoveryToSeqNo = &recoveryToSeqNo
	rbft.recvRecoveryRsp(rr2)
	rbft.recvRecoveryRsp(rr3)
	ast.NotEqual(uint64(60), *rbft.recoveryMgr.recoveryToSeqNo, "recoveryToSeqNo is not nil, expect not equal")

	rbft.recoveryMgr.recoveryToSeqNo = nil
	rbft.recvRecoveryRsp(rr3)
	ast.Equal(uint64(60), *rbft.recoveryMgr.recoveryToSeqNo, "recoveryToSeqNo is not nil, expect not equal")
}

func TestFindHighestChkptQuorum(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rr1 := &RecoveryResponse{
		Chkpts:      make(map[uint64]string),
		BlockHeight: uint64(100),
		Genesis:     uint64(100),
	}
	rr1.Chkpts[uint64(60)] = "60"
	rr1.Chkpts[uint64(70)] = "70"
	rr1.Chkpts[uint64(80)] = "80"

	rr2 := &RecoveryResponse{
		Chkpts:      make(map[uint64]string),
		BlockHeight: uint64(100),
		Genesis:     uint64(100),
	}
	rr2.Chkpts[uint64(60)] = "60"

	rbft.recoveryMgr.rcRspStore[uint64(1)] = rr1
	rbft.recoveryMgr.rcRspStore[uint64(2)] = rr2

	seqNo, id, _, find, behind := rbft.findHighestChkptQuorum()
	ast.Equal(uint64(60), seqNo, "findHighestChkptQuorum failed")
	ast.Equal("60", id, "findHighestChkptQuorum failed")
	ast.Equal(true, find, "findHighestChkptQuorum failed")
	ast.Equal(true, behind, "findHighestChkptQuorum failed")

	rbft.recoveryMgr.rcRspStore[uint64(2)].Chkpts[uint64(80)] = "80"
	seqNo, id, _, find, behind = rbft.findHighestChkptQuorum()
	ast.Equal(uint64(80), seqNo, "findHighestChkptQuorum failed")
	ast.Equal("80", id, "findHighestChkptQuorum failed")
	ast.Equal(true, find, "findHighestChkptQuorum failed")
	ast.Equal(true, behind, "findHighestChkptQuorum failed")
}

func TestReturnRecoveryPQC(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerUnicast").Return(nil)
	from := uint64(9)
	go func() {
		_ = <-eChan
		to := <-eChan
		ast.Equal(from, to, "recvRecovery failed")
	}()

	recoveryFetchPQC := &RecoveryFetchPQC{
		ReplicaId: from,
		H:         uint64(40),
	}

	event := rbft.returnRecoveryPQC(recoveryFetchPQC)
	ast.Nil(event, "h is too big, expect nil")

	recoveryFetchPQC.H = uint64(20)
	rbft.view = uint64(2)

	v1 := uint64(2)
	seqNo1 := uint64(100)
	digest1 := "d1"
	cert1 := &msgCert{}
	idx1 := msgID{v1, seqNo1, digest1}
	rbft.storeMgr.certStore[idx1] = cert1
	prePrepare1 := &PrePrepare{
		View:           v1,
		SequenceNumber: seqNo1,
		BatchDigest:    digest1,
		ResultHash:     "resultHash1",
	}
	cert1.resultHash = "resultHash1"
	cert1.prePrepare = prePrepare1
	cert1.prepare = make(map[Prepare]bool)
	cert1.prepare[Prepare{ResultHash: "c1p0"}] = true
	cert1.prepare[Prepare{ResultHash: "c1p1"}] = true
	cert1.prepare[Prepare{ResultHash: "c1p2"}] = true
	cert1.commit = make(map[Commit]bool)
	cert1.commit[Commit{ResultHash: "c1c0"}] = true
	cert1.commit[Commit{ResultHash: "c1c1"}] = true
	cert1.commit[Commit{ResultHash: "c1c2"}] = true

	v2 := uint64(2)
	seqNo2 := uint64(100)
	digest2 := "d2"
	cert2 := &msgCert{}
	idx2 := msgID{v2, seqNo2, digest2}
	rbft.storeMgr.certStore[idx2] = cert2
	cert2.resultHash = "resultHash2"
	cert2.prepare = make(map[Prepare]bool)
	cert2.prepare[Prepare{ResultHash: "c2p0"}] = true
	cert2.prepare[Prepare{ResultHash: "c2p1"}] = true
	cert2.prepare[Prepare{ResultHash: "c2p2"}] = true
	cert2.commit = make(map[Commit]bool)
	cert2.commit[Commit{ResultHash: "c2c0"}] = true
	cert2.commit[Commit{ResultHash: "c2c1"}] = true
	cert2.commit[Commit{ResultHash: "c2c2"}] = true

	rbft.returnRecoveryPQC(recoveryFetchPQC)
}
