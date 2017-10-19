//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"strconv"
	"sync/atomic"
	"testing"
	"time"

	"hyperchain/common"
	"hyperchain/consensus/consensusMocks"
	"hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/manager/protos"

	"github.com/facebookgo/ensure"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

//for test, make a cert pre-prepared.
func (rbft *rbftImpl) getCertPreprepared(digest string, v uint64, n uint64) {

	prePrepare := &PrePrepare{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    digest,
	}
	cert := rbft.storeMgr.getCert(v, n, digest)
	cert.prePrepare = prePrepare
	cert.resultHash = digest
}

func (rbft *rbftImpl) getCertPrepared(digest string, v uint64, n uint64) {
	if !rbft.prePrepared(digest, v, n) {
		rbft.getCertPreprepared(digest, v, n)
	}
	cert := rbft.storeMgr.getCert(v, n, digest)
	cert.prepare[Prepare{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    digest,
		ReplicaId:      uint64(1),
	}] = true
	cert.prepare[Prepare{ //now prepared
		View:           v,
		SequenceNumber: n,
		BatchDigest:    digest,
		ReplicaId:      uint64(2),
	}] = true
}

func (rbft *rbftImpl) getCertCommitted(digest string, v uint64, n uint64) {
	if !rbft.prepared(digest, v, n) {
		rbft.getCertPrepared(digest, v, n)
	}
	cert := rbft.storeMgr.getCert(v, n, digest)
	cmt := &Commit{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    digest,
		ReplicaId:      uint64(0),
	}
	cmt2 := &Commit{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    digest,
		ReplicaId:      uint64(1),
	}
	cmt3 := &Commit{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    digest,
		ReplicaId:      uint64(2),
	}
	cert.sentExecute = false
	cert.validated = true
	cert.commit[*cmt] = true
	cert.commit[*cmt2] = true
	cert.commit[*cmt3] = true
}

func TestPbftImpl_NewPbft(t *testing.T) {

	//new PBFT
	rbft, conf, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ensure.Nil(t, err)

	ensure.DeepEqual(t, rbft.namespace, "global-"+strconv.Itoa(int(rbft.id)))
	ensure.DeepEqual(t, rbft.status.getState(&rbft.status.inViewChange), false)
	ensure.DeepEqual(t, rbft.f, (rbft.N-1)/3)
	ensure.DeepEqual(t, rbft.N, conf.GetInt("self.N"))
	ensure.DeepEqual(t, rbft.h, uint64(0))
	ensure.DeepEqual(t, rbft.id, uint64(conf.GetInt64(common.C_NODE_ID)))
	ensure.DeepEqual(t, rbft.K, uint64(10))
	ensure.DeepEqual(t, rbft.logMultiplier, uint64(4))
	ensure.DeepEqual(t, rbft.L, rbft.logMultiplier*rbft.K)
	ensure.DeepEqual(t, rbft.seqNo, uint64(0))
	ensure.DeepEqual(t, rbft.view, uint64(0))

	//Test Consenter interface
	rbft.Start()

	//pbft.RecvMsg()
	rbft.Close()

}

func TestProcessNullRequest(t *testing.T) {
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	ensure.Nil(t, err)
	pbMsg := &protos.Message{
		Type:      protos.Message_NULL_REQUEST,
		Payload:   nil,
		Timestamp: time.Now().UnixNano(),
		Id:        1,
	}
	event, err := proto.Marshal(pbMsg)
	rbft.RecvMsg(event)

	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.activeState(&rbft.status.inViewChange)
	rbft.RecvMsg(event)

	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.RecvMsg(event)

	err = CleanData(rbft.namespace)
	ensure.Nil(t, err)
}

func TestFindNextPrePrepareBatch(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	cb1 := &cacheBatch{
		seqNo:      uint64(1),
		resultHash: "1",
	}
	cb2 := &cacheBatch{
		seqNo:      uint64(2),
		resultHash: "2",
	}
	cb3 := &cacheBatch{
		seqNo:      uint64(3),
		resultHash: "3",
	}
	rbft.batchVdr.saveToCVB("notExist", nil)
	rbft.batchVdr.saveToCVB(cb2.resultHash, cb2)
	find, digest, resultHash := rbft.findNextPrePrepareBatch()
	ast.Equal(false, find, "cannot find the next prePrepare batch, expect false")

	rbft.batchVdr.saveToCVB(cb1.resultHash, cb1)
	find, digest, resultHash = rbft.findNextPrePrepareBatch()
	ast.Equal(true, find, "can find the next prePrepare batch, expect true")
	ast.Equal("1", digest, "can find the next prePrepare batch, expect '1'")
	ast.Equal("1", resultHash, "can find the next prePrepare batch, expect '1'")
	rbft.batchVdr.updateLCVid()

	rbft.getCertPreprepared("2", rbft.view, uint64(10))
	find, digest, resultHash = rbft.findNextPrePrepareBatch()
	ast.Equal(false, find, "cannot find the next prePrepare batch, expect false")
	ast.Equal(uint64(2), rbft.batchVdr.lastVid, "existedDigest, expect 2")

	rbft.storeMgr.certStore = make(map[msgID]*msgCert)
	rbft.batchVdr.saveToCVB(cb3.resultHash, cb3)
	rbft.h = uint64(10)
	find, digest, resultHash = rbft.findNextPrePrepareBatch()
	ast.Equal(false, find, "cannot find the next prePrepare batch, expect false")
	ast.Nil(rbft.batchVdr.currentVid, "not sendInWV, expect nil")
}

func TestSendPendingPrePrepares(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	cb1 := &cacheBatch{
		seqNo:      uint64(1),
		resultHash: "1",
	}
	cb2 := &cacheBatch{
		seqNo:      uint64(2),
		resultHash: "2",
	}
	rbft.batchVdr.saveToCVB(cb1.resultHash, cb1)
	rbft.batchVdr.saveToCVB(cb2.resultHash, cb2)
	rbft.storeMgr.outstandingReqBatches["1"] = &TransactionBatch{
		HashList:  []string{},
		Timestamp: time.Now().UnixNano(),
	}
	rbft.storeMgr.outstandingReqBatches["2"] = &TransactionBatch{
		HashList:  []string{},
		Timestamp: time.Now().UnixNano(),
	}
	rbft.sendPendingPrePrepares()
	ast.Equal(uint64(2), rbft.batchVdr.lastVid, "sendPendingPrePrepares failed")

	currentVid := uint64(10)
	rbft.batchVdr.setCurrentVid(&currentVid)
	rbft.sendPendingPrePrepares()
	ast.Equal(uint64(2), rbft.batchVdr.lastVid, "sendPendingPrePrepares failed")
}

func TestSendPrePrepare(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	seqNo := uint64(10)
	digest := "10"
	hash := "10"

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		prePre := &PrePrepare{}
		err = proto.Unmarshal(consensus.Payload, prePre)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(seqNo, prePre.SequenceNumber, "sendPrePrepare failed")
		ast.Equal(digest, prePre.BatchDigest, "sendPrePrepare failed")
	}()

	reqBatch := &TransactionBatch{
		HashList:  []string{},
		Timestamp: time.Now().UnixNano(),
	}
	rbft.batchVdr.setCurrentVid(&seqNo)
	rbft.sendPrePrepare(seqNo, digest, hash, reqBatch)
	ast.Nil(rbft.batchVdr.currentVid, "sendPrePrepare failed")
	ast.Equal(seqNo, rbft.batchVdr.lastVid, "sendPrePrepare failed")
}

func TestRecvPrePrepare(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	v := uint64(0)
	seqNo := uint64(1)
	digest := "d1"
	rbft.status.inActiveState(&rbft.status.inNegoView)
	pp := &PrePrepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    "normal",
		ResultHash:     digest,
		ReplicaId:      uint64(1),
	}
	rbft.recvPrePrepare(pp)

	cert := rbft.storeMgr.getCert(pp.View, pp.SequenceNumber, pp.BatchDigest)
	ast.Equal(true, cert.sentPrepare, "recv preprepare, should send prepare")
	ast.Equal(pp.ResultHash, cert.resultHash, "recv preprepare, should send prepare")
}

func TestRecvPrepare(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	rbft.status.inActiveState(&rbft.status.inNegoView)

	v := uint64(0)
	seqNo := uint64(10)
	digest := "d1"
	prep := &Prepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
		ResultHash:     digest,
		ReplicaId:      uint64(2),
	}
	err = rbft.recvPrepare(prep)
	ast.Nil(err, "should succeed, expect nil")
	err = rbft.recvPrepare(prep)
	ast.Nil(err, "duplicate prepare, expect nil")
	rbft.status.inActiveState(&rbft.status.inRecovery)
	err = rbft.recvPrepare(prep)
	ast.Nil(err, "duplicate prepare, expect nil")
}

func TestMaybeSendCommit(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	v := uint64(0)
	seqNo := uint64(10)
	digest := "d1"

	err = rbft.maybeSendCommit(digest, v, seqNo)
	ast.Nil(err, "cert is not prepared, expect nil")
	rbft.getCertPrepared(digest, v, seqNo)
	rbft.status.activeState(&rbft.status.skipInProgress)
	err = rbft.maybeSendCommit(digest, v, seqNo)
	ast.Nil(err, "skipInProgress, expect nil")

	rbft.status.inActiveState(&rbft.status.skipInProgress)
	err = rbft.maybeSendCommit(digest, v, seqNo)
	d, ok := rbft.batchVdr.preparedCert[vidx{view: v, seqNo: seqNo}]
	ast.Equal(true, ok, "maybeSendCommit failed")
	if ok {
		ast.Equal(digest, d, "maybeSendCommit failed")
	}

	rbft.id = uint64(1)
	err = rbft.maybeSendCommit(digest, v, seqNo)
	ast.Nil(err, "maybeSendCommit failed")
	cert := rbft.storeMgr.getCert(v, seqNo, digest)
	ast.Equal(true, cert.sentCommit, "maybeSendCommit failed")
}

func TestSendCommit(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)

	// normal
	d := "digest"
	v := uint64(0)
	n := uint64(1)
	rbft.getCertPrepared(d, v, n)

	err = rbft.sendCommit(d, v, n)
	ast.Nil(err, "sendCommit failed")

	cert := rbft.storeMgr.getCert(v, n, d)
	ast.Equal(true, cert.sentCommit, "send commit false, expect true")
}

func TestRecvCommit(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)

	// normal, committed, execute
	v := uint64(0)
	n := uint64(1)
	d := "digest"
	replicaId := uint64(1)
	cmt := &Commit{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    d,
		ReplicaId:      replicaId,
	}

	//inNegoView
	err = rbft.recvCommit(cmt)
	ast.Nil(err, "should end in inNegoView, expect nil")

	//not inWV
	rbft.status.inActiveState(&rbft.status.inNegoView)
	cmt.View = uint64(2)
	err = rbft.recvCommit(cmt)
	ast.Nil(err, "should end in inWV, expect nil")

	cmt.View = uint64(0)
	cert := rbft.storeMgr.getCert(cmt.View, cmt.SequenceNumber, cmt.BatchDigest)
	cert.commit[*cmt] = true
	err = rbft.recvCommit(cmt)
	ast.Nil(err, "already received, expect nil")

	cert.commit = make(map[Commit]bool)
	rbft.startNewViewTimer(5*time.Second, "nothing")
	err = rbft.recvCommit(cmt)
	ast.Nil(err, "not prepared, expect nil")
	ast.Equal(true, rbft.status.getState(&rbft.status.timerActive))

	cert.commit = make(map[Commit]bool)
	rbft.getCertPrepared(cmt.BatchDigest, cmt.View, cmt.SequenceNumber)
	hashBatch := &HashBatch{Timestamp: time.Now().UnixNano()}
	cert.prePrepare.HashBatch = hashBatch
	err = rbft.recvCommit(cmt)
	ast.Nil(err, "no enough commit message, expect nil")

	cmt2 := &Commit{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    d,
		ReplicaId:      uint64(2),
	}
	cmt3 := &Commit{
		View:           v,
		SequenceNumber: n,
		BatchDigest:    d,
		ReplicaId:      uint64(3),
	}
	ast.Equal(true, rbft.status.getState(&rbft.status.timerActive))
	rbft.recvCommit(cmt2)
	rbft.recvCommit(cmt3) //now committed
	ast.Equal(false, rbft.status.getState(&rbft.status.timerActive), "should stop new view timer, expect false")
	cert.validated = true
	delete(cert.commit, *cmt3)
	rbft.recvCommit(cmt3)
}

func TestFindNextCommitTx(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	v := uint64(0)
	n := uint64(1)
	d := "digest"
	rbft.storeMgr.committedCert[msgID{v, n, d}] = "first"
	find, _, _ := rbft.findNextCommitTx()
	ast.Equal(false, find, "cert is nil, expect false")

	rbft.getCertPreprepared(d, v, n)
	cert0 := rbft.storeMgr.getCert(v, n, d)
	cert0.sentExecute = true
	find, _, _ = rbft.findNextCommitTx()
	ast.Equal(false, find, "sentExecute is false, expect false")

	cert0.sentExecute = false
	rbft.exec.setLastExec(uint64(10))
	find, _, _ = rbft.findNextCommitTx()
	ast.Equal(false, find, "n is not equal to lastExec + 1, expect false")

	rbft.exec.setLastExec(uint64(0))
	rbft.status.activeState(&rbft.status.skipInProgress)
	find, _, _ = rbft.findNextCommitTx()
	ast.Equal(false, find, "skipInProgress, expect false")

	rbft.status.inActiveState(&rbft.status.skipInProgress)
	find, _, _ = rbft.findNextCommitTx()
	ast.Equal(false, find, "not committed, expect false")

	rbft.getCertCommitted(d, v, n)
	find, _, _ = rbft.findNextCommitTx()
	ast.Equal(true, find, "committed, expect true")
	ast.Equal(uint64(1), *rbft.exec.currentExec, "committed, expect 1")
}

func TestAfterCommitBlock(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	v := uint64(0)
	n := uint64(1)
	d := "digest"

	rbft.status.inActiveState(&rbft.status.skipInProgress)
	currentExec := uint64(10)
	rbft.exec.setCurrentExec(nil)
	rbft.afterCommitBlock(msgID{v, n, d})
	ast.Equal(true, rbft.status.getState(&rbft.status.skipInProgress), "currentExec is not nil, expect true")
	ast.Nil(rbft.exec.currentExec, "currentExec should be set to nil")

	go func() {
		chain.WriteChainChan(rbft.namespace)
	}()
	rbft.exec.setCurrentExec(&currentExec)
	rbft.afterCommitBlock(msgID{v, n, d})
	ast.Equal(uint64(10), rbft.exec.lastExec, "lastExec is set to currentExec, expect 10")
	ast.Nil(rbft.exec.currentExec, "currentExec should be set to nil")
}

func TestRecvFetchMissingTransaction(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerUnicast").Return(nil)

	v := rbft.view
	seqNo := uint64(10)
	batchDigest := "fetch"
	hashList1 := []string{"0", "1"}
	hashList2 := make(map[uint64]string)
	txList := []*types.Transaction{nil, nil}
	hashList2[uint64(0)] = "0"
	hashList2[uint64(1)] = "1"
	hashList2[uint64(2)] = "2"

	tranBatch := &TransactionBatch{
		HashList: hashList1,
		TxList:   txList,
	}
	rbft.storeMgr.txBatchStore[batchDigest] = tranBatch

	fetch := &FetchMissingTransaction{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    batchDigest,
		HashList:       hashList2,
		ReplicaId:      rbft.id,
	}
	err = rbft.recvFetchMissingTransaction(fetch)
	ast.Nil(err, "mismatch tx hash, expect nil")

	delete(fetch.HashList, uint64(2))
	go func() {
		_ = <-eChan
		to := <-eChan
		ast.Equal(rbft.id, to, "recvNegoView failed")
	}()
	rbft.recvFetchMissingTransaction(fetch)
	time.Sleep(time.Nanosecond)

	delete(rbft.storeMgr.txBatchStore, fetch.BatchDigest)
	err = rbft.recvFetchMissingTransaction(fetch)
	ast.Nil(err, "find in txPool and cannot find it, expect ni;")
}

func TestRecvReturnMissingTransaction(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	v := rbft.view
	seqNo := uint64(10)
	batchDigest := "return"
	hashList := make(map[uint64]string)
	txList := make(map[uint64]*types.Transaction)
	hashList[uint64(0)] = "0"
	hashList[uint64(1)] = "1"
	txList[uint64(0)] = nil

	re := &ReturnMissingTransaction{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    batchDigest,
		HashList:       hashList,
		TxList:         txList,
		ReplicaId:      rbft.id,
	}
	event := rbft.recvReturnMissingTransaction(re)
	ast.Nil(event, "length mismatch, expect nil")

	re.TxList[uint64(1)] = nil
	rbft.batchVdr.lastVid = uint64(10)
	event = rbft.recvReturnMissingTransaction(re)
	ast.Nil(event, "SequenceNumber <= lastVid, expect nil")

	rbft.batchVdr.lastVid = uint64(0)
	event = rbft.recvReturnMissingTransaction(re)
	ast.Nil(event, "no prePrepare, expect nil")

	rbft.getCertPreprepared(batchDigest, v, seqNo)
	event = rbft.recvReturnMissingTransaction(re)
	ast.Nil(event, "GotMissingTxs false, expect nil")
}

func TestProcessTransaction(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	// first generate a Transaction
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	txRequest := txRequest{
		tx:  tx,
		new: true,
	}
	rbft.processTransaction(txRequest)

	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.processTransaction(txRequest)
	ast.Equal(true, rbft.batchMgr.batchTimerActive, "should start batch timer, expect true")

	rbft.id = uint64(2)
	rbft.processTransaction(txRequest)
	ast.Equal(true, rbft.batchMgr.batchTimerActive, "should start batch timer, expect true")
}

func TestRecvStateUpdatedEvent(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	// normal
	// et.seqNo < pbft.h  hightStateTarget == nil
	e := protos.StateUpdatedMessage{
		SeqNo: uint64(40),
	}
	rbft.h = uint64(50)
	rbft.storeMgr.highStateTarget = nil

	err = rbft.recvStateUpdatedEvent(e)
	ast.Nil(err, "et.SeqNo < rbft.h and highStateTarget is nil, expect nil")

	// et.seqNo < pbft.h  et.seqNo<pbft.highStateTarget.seqNo
	rbft.storeMgr.highStateTarget = &stateUpdateTarget{
		checkpointMessage: checkpointMessage{
			seqNo: 80,
			id:    []byte("checkpointMessage"),
		},
		replicas: []replicaInfo{},
	}
	rbft.recvStateUpdatedEvent(e)
	ast.Equal(uint32(0), atomic.LoadUint32(&rbft.normal), "et.SeqNo < rbft.h and et.SeqNo < highStateTarget.seqNo, expect 0")

	rbft.storeMgr.highStateTarget.seqNo = 40
	err = rbft.recvStateUpdatedEvent(e)
	ast.Nil(err, "et.SeqNo < rbft.h, expect nil")

	// et.seqNo >= pbft.h
	e = protos.StateUpdatedMessage{
		SeqNo: uint64(80),
	}
	rbft.status.inActiveState(&rbft.status.inNegoView)
	err = rbft.recvStateUpdatedEvent(e)
	ast.Equal(e.SeqNo, rbft.exec.lastExec, "recvStateUpdatedEvent, rbft should update lastExec")
	ast.Equal(uint32(1), atomic.LoadUint32(&rbft.normal), "should be set to 1, expect 1")

	rbft.status.activeState(&rbft.status.inRecovery)
	err = rbft.recvStateUpdatedEvent(e)
	ast.Nil(err, "recoveryToSeqNo == nil, expect nil")
	recoveryToSeqNo := uint64(70)
	rbft.recoveryMgr.recoveryToSeqNo = &recoveryToSeqNo
	err = rbft.recvStateUpdatedEvent(e)
	ast.Nil(err, "lastExec >= recoveryToSeqNo, expect nil")
	time.Sleep(time.Second)

	recoveryToSeqNo = uint64(90)
	rbft.recoveryMgr.recoveryToSeqNo = &recoveryToSeqNo
	rbft.status.activeState(&rbft.status.inRecovery)
	err = rbft.recvStateUpdatedEvent(e)
	ast.Equal(true, rbft.status.getState(&rbft.status.inNegoView), "restartRecovery, expect true")
	ast.Equal(true, rbft.status.getState(&rbft.status.inRecovery), "restartRecovery, expect true")
}

func TestExecuteAfterStateUpdate(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	currentVid := uint64(10)
	rbft.batchVdr.setCurrentVid(&currentVid)
	rbft.getCertPrepared("123", uint64(0), uint64(10))
	rbft.executeAfterStateUpdate()
	d, ok := rbft.batchVdr.preparedCert[vidx{uint64(0), uint64(10)}]
	ast.Equal(true, ok, "should be stored, expect true")
	if ok {
		ast.Equal("123", d, "should be stored, expect '123'")
	}
}

func TestCheckpoint(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	eChan := make(chan interface{})
	mockHelper := &consensusMocks.MockHelper{
		EventChan: eChan,
	}
	rbft.helper = mockHelper
	mockHelper.On("InnerBroadcast").Return(nil)

	rbft.id = uint64(2)
	rbft.seqNo = uint64(0)
	bcInfo := &protos.BlockchainInfo{
		Height:            10,
		PreviousBlockHash: []byte("previous"),
		CurrentBlockHash:  []byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(10)

	go func() {
		event := <-eChan
		e, ok := event.(*protos.Message)
		ast.Equal(true, ok, "InnerBroadcast failed")

		consensus := &ConsensusMessage{}
		err := proto.Unmarshal(e.Payload, consensus)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		checkpoint := &Checkpoint{}
		err = proto.Unmarshal(consensus.Payload, checkpoint)
		ast.Nil(err, fmt.Sprint("processConsensus, unmarshal error: can not unmarshal ConsensusMessage", err))
		ast.Equal(seqNo, checkpoint.SequenceNumber, "checkpoint failed")
		ast.Equal(idAsString, checkpoint.Id, "checkpoint failed")
	}()

	rbft.checkpoint(seqNo, bcInfo)
	idString, _ := rbft.storeMgr.chkpts[seqNo]
	ast.Equal(idAsString, idString, "checkpoint failed")

	rbft.checkpoint(uint64(1), bcInfo)
	if _, ok := rbft.storeMgr.chkpts[uint64(1)]; ok {
		t.Error("n % pbft.K != 0")
	}
}

func TestRecvCheckpoint(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:            10,
		PreviousBlockHash: []byte("previous"),
		CurrentBlockHash:  []byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(10)
	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      uint64(3),
		Id:             idAsString,
	}
	cert := rbft.storeMgr.getChkptCert(seqNo, chkpt.Id)
	cert.chkptCount = rbft.commonCaseQuorum() - 1

	rbft.storeMgr.chkpts[seqNo] = idAsString
	rbft.recvCheckpoint(chkpt)
	ast.Equal(uint64(10), rbft.h, "should movewatermarks")

	delete(rbft.storeMgr.chkpts, seqNo)
	rbft.status.activeState(&rbft.status.skipInProgress, &rbft.status.inRecovery)
	rbft.recvCheckpoint(chkpt)
	ast.Equal(uint64(10), rbft.h, "should movewatermarks")

	chkpt.SequenceNumber = 30
	cert = rbft.storeMgr.getChkptCert(30, chkpt.Id)
	cert.chkptCount = rbft.commonCaseQuorum() - 1
	rbft.status.inActiveState(&rbft.status.inRecovery)
	rbft.recvCheckpoint(chkpt)
	ast.Equal(uint64(30), rbft.h, "should movewatermarks")
}

func TestWeakCheckpointSetOutOfRange(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	rbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:            50,
		PreviousBlockHash: []byte("previous"),
		CurrentBlockHash:  []byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(50)
	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      uint64(3),
		Id:             idAsString,
	}
	rbft.storeMgr.hChkpts[1] = uint64(50)
	rbft.storeMgr.hChkpts[uint64(3)] = uint64(50)
	chkpt.SequenceNumber = uint64(10)
	rbft.weakCheckpointSetOutOfRange(chkpt)
	if _, ok := rbft.storeMgr.hChkpts[uint64(3)]; ok {
		t.Error("should be deleted, expect false")
	}

	rbft.exec.lastExec = uint64(100)
	chkpt.SequenceNumber = seqNo
	ok := rbft.weakCheckpointSetOutOfRange(chkpt)
	if !ok || rbft.status.getState(&rbft.status.skipInProgress) {
		t.Error("Replica is ahead of others, expect true")
	}

	rbft.exec.lastExec = uint64(10)
	ok = rbft.weakCheckpointSetOutOfRange(chkpt)
	if _, ok := rbft.storeMgr.hChkpts[uint64(3)]; !ok {
		t.Error("should be deleted, expect true")
	}
	ast.NotEqual(0, rbft.h, "should movewatermarks")
	ast.Equal(true, rbft.status.getState(&rbft.status.skipInProgress), "should skip in progress")
}

func TestWitnessCheckpointWeakCert(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	rbft.seqNo = uint64(0)

	bcInfo := &protos.BlockchainInfo{
		Height:            10,
		PreviousBlockHash: []byte("previous"),
		CurrentBlockHash:  []byte("current"),
	}
	identity, _ := proto.Marshal(bcInfo)
	idAsString := byteToString(identity)
	seqNo := uint64(10)
	chkpt := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      uint64(3),
		Id:             idAsString,
	}
	rbft.storeMgr.checkpointStore[*chkpt] = true
	chkpt2 := &Checkpoint{
		SequenceNumber: seqNo,
		ReplicaId:      uint64(2),
		Id:             idAsString,
	}
	rbft.storeMgr.checkpointStore[*chkpt2] = true

	rbft.storeMgr.highStateTarget = nil

	rbft.witnessCheckpointWeakCert(chkpt)
	ast.NotNil(rbft.storeMgr.highStateTarget, "should update highStateTarget")
}

func TestMainMoveWatermarks(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.seqNo = uint64(0)
	// set current h to 50
	rbft.h = uint64(50)
	digest := "v0n1digest"
	// fill some record into store
	rbft.storeMgr.getCert(uint64(0), uint64(1), digest)
	rbft.storeMgr.txBatchStore["123"] = &TransactionBatch{SeqNo: uint64(30)}

	rbft.batchVdr.preparedCert[vidx{uint64(0), uint64(1)}] = "123"
	rbft.storeMgr.committedCert[msgID{uint64(0), uint64(1), digest}] = "123"

	testChkpt := &Checkpoint{
		SequenceNumber: uint64(10),
	}
	// checkpointstore
	rbft.storeMgr.checkpointStore[*testChkpt] = true

	// chkptCertStore
	rbft.storeMgr.getChkptCert(uint64(10), "chkptId")

	// pset
	vcP := &Vc_PQ{}
	rbft.vcMgr.plist[uint64(10)] = vcP

	// qset
	qIdx := qidx{d: "chkpt", n: 10}
	vcQ := &Vc_PQ{}
	rbft.vcMgr.qlist[qIdx] = vcQ

	// chkpts
	rbft.storeMgr.chkpts[uint64(10)] = "chkpt"

	// this seqno should not trigger movewatermark
	rbft.moveWatermarks(uint64(40))

	if _, ok := rbft.storeMgr.chkpts[uint64(10)]; ok != true {
		t.Error("should not move watermark, expect true")
	}
	// test if all delete
	rbft.moveWatermarks(uint64(60))

	// certStore
	if _, ok := rbft.storeMgr.certStore[msgID{uint64(0), uint64(1), digest}]; ok {
		t.Error("should move watermark, expect cert not exist")
	}
	// txBatchStore
	if _, ok := rbft.storeMgr.txBatchStore["123"]; ok {
		t.Error("should move watermark, expect cert not exist")
	}
	// preparedCert
	if _, ok := rbft.batchVdr.preparedCert[vidx{uint64(0), uint64(1)}]; ok {
		t.Error("should move watermark, expect cert not exist")
	}
	// committedCert
	if _, ok := rbft.storeMgr.committedCert[msgID{uint64(0), uint64(1), digest}]; ok {
		t.Error("should move watermark, expect cert not exist")
	}
	// checkpoint store
	if _, ok := rbft.storeMgr.checkpointStore[*testChkpt]; ok {
		t.Error("should not move watermark, expect checkpoint store record not exist")
	}
	// checkpoint cert store
	if _, ok := rbft.storeMgr.chkptCertStore[chkptID{n: 10, id: "chkptId"}]; ok {
		t.Error("should not move watermark, expect chkptCertStore record not exist")
	}
	if _, ok := rbft.vcMgr.plist[uint64(10)]; ok {
		t.Error("should not move watermark, expect pset record not exist")
	}
	if _, ok := rbft.vcMgr.qlist[qIdx]; ok {
		t.Error("should not move watermark, expect qset record not exist")
	}
}

func TestUpdateHighStateTarget(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.id = uint64(2)
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.seqNo = uint64(0)

	checkpoint := checkpointMessage{
		seqNo: uint64(20),
		id:    []byte("checkpoint"),
	}
	peers := []replicaInfo{}
	curTarget := &stateUpdateTarget{
		checkpointMessage: checkpoint,
		replicas:          peers,
	}
	rbft.storeMgr.highStateTarget = curTarget

	// new target seqNo <= cur target seqNo
	newCheckpoint1 := checkpointMessage{
		seqNo: uint64(10),
		id:    []byte("checkpoint"),
	}
	newTargetSmaller := &stateUpdateTarget{
		checkpointMessage: newCheckpoint1,
		replicas:          peers,
	}
	rbft.updateHighStateTarget(newTargetSmaller)
	ast.Equal(curTarget.checkpointMessage.seqNo, rbft.storeMgr.highStateTarget.seqNo, "should not update high state target, expect not change")

	// new target seqNo > cur target seqNo
	newCheckpoint2 := checkpointMessage{
		seqNo: uint64(30),
		id:    []byte("checkpoint"),
	}
	newTargetLarger := &stateUpdateTarget{
		checkpointMessage: newCheckpoint2,
		replicas:          peers,
	}
	rbft.updateHighStateTarget(newTargetLarger)
	ast.Equal(uint64(30), rbft.storeMgr.highStateTarget.seqNo, "shoul update high state target, expect 30")
}

func TestStateTransfer(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.status.inActiveState(&rbft.status.skipInProgress)
	rbft.stateTransfer(nil)
	ast.Equal(true, rbft.status.getState(&rbft.status.skipInProgress), "should be actived, expect true")

	newCheckpoint := checkpointMessage{
		seqNo: 10,
		id:    []byte("checkpoint"),
	}
	peers := []replicaInfo{}
	peers = append(peers, replicaInfo{
		id:      uint64(1),
		height:  uint64(10),
		genesis: uint64(10),
	})
	peers = append(peers, replicaInfo{
		id:      uint64(2),
		height:  uint64(10),
		genesis: uint64(10),
	})
	target := &stateUpdateTarget{
		checkpointMessage: newCheckpoint,
		replicas:          peers,
	}
	rbft.status.activeState(&rbft.status.stateTransferring)
	rbft.stateTransfer(nil)

	rbft.status.inActiveState(&rbft.status.stateTransferring)
	rbft.stateTransfer(nil)

	rbft.storeMgr.highStateTarget = target
	rbft.stateTransfer(nil)
	ast.Equal(true, rbft.status.getState(&rbft.status.stateTransferring), "shoud be actived, expect true")

	rbft.status.inActiveState(&rbft.status.stateTransferring)
	info := &protos.BlockchainInfo{Height: uint64(100)}
	id, _ := proto.Marshal(info)
	target.id = id
	rbft.stateTransfer(target)
}

func TestRecvValidatedResult(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	txes := make([]*types.Transaction, 1)
	vali := protos.ValidatedTxs{
		Transactions: txes,
		Hash:         "hash",
		SeqNo:        uint64(1),
		View:         uint64(0),
		Timestamp:    time.Now().UnixNano(),
		Digest:       "hash",
	}

	currentVid := uint64(10)
	rbft.batchVdr.setCurrentVid(&currentVid)
	rbft.status.activeState(&rbft.status.inViewChange)
	err = rbft.recvValidatedResult(vali)
	ast.Nil(err, "activeView is 0, expect nil")

	rbft.status.inActiveState(&rbft.status.inViewChange)
	rbft.status.activeState(&rbft.status.inUpdatingN)
	err = rbft.recvValidatedResult(vali)
	ast.Nil(err, "inUpdatingN is 1, expect nil")

	rbft.status.inActiveState(&rbft.status.inUpdatingN)
	rbft.seqNo = uint64(0)

	// primary recv ValidateResult
	rbft.id = uint64(1)
	rbft.view = uint64(0)
	vali.View = uint64(10)
	rbft.recvValidatedResult(vali)
	if _, ok := rbft.batchVdr.cacheValidatedBatch["hash"]; ok {
		t.Error("not inV, expect not exist in cacheValidatedBatch")
	}

	vali.View = uint64(0)
	rbft.recvValidatedResult(vali)
	if _, ok := rbft.batchVdr.cacheValidatedBatch["hash"]; !ok {
		t.Error("primary recv Validatedresult, expect exist in cacheValidatedBatch")
	}

	// replica recvValidatedResult
	rbft.id = uint64(2)
	vali.View = uint64(10)
	rbft.recvValidatedResult(vali)
	cert, ok := rbft.storeMgr.certStore[msgID{v: vali.View, n: vali.SeqNo, d: vali.Digest}]
	ast.Equal(false, ok, "not inV, expect not exist in cert store")

	vali.View = uint64(0)
	rbft.recvValidatedResult(vali)
	cert, ok = rbft.storeMgr.certStore[msgID{v: vali.View, n: vali.SeqNo, d: vali.Digest}]
	ast.Equal(true, ok, "replica recv validated result, expect exist in cert store")
	if ok {
		ast.Equal(false, cert.validated, "resultHash is nil, expect not validated")
	}

	cert = rbft.storeMgr.getCert(vali.View, vali.SeqNo, vali.Digest)
	cert.resultHash = vali.Hash
	rbft.recvValidatedResult(vali)
	ast.Equal(true, cert.validated, "replica recv validated result, expect validated")

	cert.resultHash = "notexist"
	rbft.status.inActiveState(&rbft.status.inNegoView, &rbft.status.inRecovery)
	rbft.recvValidatedResult(vali)
	ast.Equal(uint64(1), rbft.view, "view change, expect 1")
}
