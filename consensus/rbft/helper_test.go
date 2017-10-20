//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSortableUint64SliceFunctions(t *testing.T) {
	slice := sortableUint64Slice{1, 2, 3, 4, 5}
	if slice.Len() != 5 {
		t.Error("error slice.len != 5")
	}
	if slice.Less(2, 3) != true {
		t.Error("error slice[2] >= slice[3]")
	}
	if slice.Swap(2, 3); !(slice[2] == 4 && slice[3] == 3) {
		t.Error("error exchange slice[2], slice[3]")
	}
}

func TestRbftStateFunctions(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.off(valid)
	ast.Equal(false, rbft.in(valid), "should be set to inActive")

	rbft.validateState()
	ast.Equal(true, rbft.in(valid), "should be set to active")

	rbft.invalidateState()
	ast.Equal(false, rbft.in(valid), "should be set to inActive")
}

func TestPrimary(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.N = 100
	x := rbft.primary(3)
	ast.Equal(uint64(4), x, fmt.Sprintf("primary(%d) == %d, actual: %d", 3, x, 4))

	r1 := rbft.primary(0)
	ast.Equal(uint64(1), rbft.primary(0), fmt.Sprintf("primary(%d) == %d, actual: %d", 0, r1, 1))

	rbft.view = 1
	ast.Equal(true, rbft.isPrimary(2), "2 should be the primary node")

	rbft.h = 0
	rbft.L = 100
	ast.Equal(true, rbft.inW(100), fmt.Sprintf("inw(%d) = false, actual true", 100))

	rbft.view = 9
	ast.Equal(true, rbft.inWV(9, 30), fmt.Sprintf("inWV(%d, %d) = false, actual true", 9, 30))
	ast.Equal(false, rbft.inWV(8, 30), fmt.Sprintf("inWV(%d, %d) = true, actual false", 9, 30))

	ast.Equal(true, rbft.sendInW(100), fmt.Sprintf("sendInWV(%d, %d) = false, actual true", 9, 100))
	ast.Equal(false, rbft.sendInW(101), fmt.Sprintf("sendInWV(%d, %d) = true, actual false", 9, 101))
}

func TestGetNodeCert(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	cert1 := rbft.getAddNodeCert("addHash")
	cert2 := rbft.getAddNodeCert("addHash")
	ast.Equal(cert1, cert2, "getAddNodeCert failed")

	cert3 := rbft.getDelNodeCert("delHash")
	cert4 := rbft.getDelNodeCert("delHash")
	ast.Equal(cert3, cert4, "getDelNodeCert failed")
}

func TestGetNV(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.N, rbft.view = 3, 1
	n, v := rbft.getAddNV()
	ast.Equal(int64(4), n, "getAddNv failed, expect 4")
	ast.Equal(uint64(5), v, "getAddNv failed, expect 5")

	rbft.view = 10
	n, v = rbft.getAddNV()
	ast.Equal(int64(4), n, "getAddNv failed, expect 4")
	ast.Equal(uint64(13), v, "getAddNv failed, expect 13")

	rbft.N, rbft.view = 4, 2
	n, v = rbft.getDelNV(uint64(1))
	ast.Equal(int64(3), n, "getDelNV failed, expect 3")
	ast.Equal(uint64(4), v, "getDelNV failed, expect 4")

	n, v = rbft.getDelNV(uint64(3))
	ast.Equal(int64(3), n, "getDelNV failed, expect 3")
	ast.Equal(uint64(5), v, "getDelNV failed, expect 5")
}

func TestCommonCaseQuorum(t *testing.T) {
	rbft := new(rbftImpl)
	rbft.f = 1
	rbft.N = 4
	k := rbft.commonCaseQuorum()
	assert.Equal(t, 3, k, fmt.Sprintf("error commonCaseQuorum() = %d, expected: 3", k))
}

func TestAllCorrectReplicasQuorum(t *testing.T) {
	rbft := new(rbftImpl)
	rbft.N = 100
	rbft.f = 33
	k := rbft.allCorrectReplicasQuorum()
	assert.Equal(t, 67, k, fmt.Sprintf("error allCorrectReplicasQuorum() = %d, expected: %d", k, 67))
}

func TestOneCorrectQuorum(t *testing.T) {
	rbft := new(rbftImpl)
	rbft.f = 1
	rbft.N = 4
	k := rbft.oneCorrectQuorum()
	assert.Equal(t, 2, k, fmt.Sprintf("error oneCorrectQuorum() = %d, expected: 2", k))
}

func TestIntersectionQuorum(t *testing.T) {
	rbft := new(rbftImpl)
	rbft.N = 1
	rbft.f = 1
	k := rbft.allCorrectQuorum()
	assert.Equal(t, 1, k, fmt.Sprintf("error allCorrectQuorum() = %d, expected: 1", k))
}

func TestPrePrepared(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.view = 2
	rbft.seqNo = 100
	digest := ""
	v := uint64(2)
	seqNo := uint64(100)
	res := rbft.prePrepared(digest, v, seqNo)
	ast.Equal(false, res, fmt.Sprintf("error prePrepared(%q, %d, %d) = %t, actual: %t", digest, v, seqNo, res, false))

	digest = "d1"
	prePrepare := &PrePrepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
	}
	cert := &msgCert{
		prePrepare: prePrepare,
	}
	idx := msgID{v, seqNo, digest}
	rbft.storeMgr.certStore[idx] = cert

	res = rbft.prePrepared(digest, v, seqNo)
	ast.Equal(true, res, fmt.Sprintf("error prePrepared(%q, %d, %d) = %t, actual: %t", digest, v, seqNo, res, true))
}

func TestPrepared(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.view = 2
	rbft.seqNo = 100

	digest := ""
	v := uint64(2)
	seqNo := uint64(100)

	res := rbft.prepared("xxx", 12, 12)
	ast.Equal(false, res, fmt.Sprintf("error prepared(%q, %d, %d) = %t, expected: %t", digest, v, seqNo, res, false))

	digest = "d1"
	prePrepare := &PrePrepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
	}
	cert := &msgCert{
		prePrepare: prePrepare,
	}
	idx := msgID{v, seqNo, digest}
	rbft.storeMgr.certStore[idx] = cert
	res = rbft.prepared(digest, v, seqNo)
	ast.Equal(false, res, fmt.Sprintf("error prepared(%q, %d, %d) = %t, expected: %t", digest, v, seqNo, res, false))
	cert.prepare = make(map[Prepare]bool)
	cert.prepare[Prepare{ReplicaId: uint64(0)}] = true
	cert.prepare[Prepare{ReplicaId: uint64(1)}] = true
	cert.prepare[Prepare{ReplicaId: uint64(2)}] = true

	res = rbft.prepared(digest, v, seqNo)
	ast.Equal(true, res, fmt.Sprintf("error prepared(%q, %d, %d) = %t, expected: %t", digest, v, seqNo, res, true))
}

func TestCommitted(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.view = 2
	rbft.seqNo = 100

	digest := ""
	v := uint64(2)
	seqNo := uint64(100)
	digest = "d1"
	prePrepare := &PrePrepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
	}
	cert := &msgCert{
		prePrepare: prePrepare,
	}
	idx := msgID{v, seqNo, digest}
	rbft.storeMgr.certStore[idx] = cert
	cert.prepare = make(map[Prepare]bool)
	cert.prepare[Prepare{ReplicaId: uint64(0)}] = true
	cert.prepare[Prepare{ReplicaId: uint64(1)}] = true
	cert.prepare[Prepare{ReplicaId: uint64(2)}] = true

	res := rbft.committed(digest, v, seqNo)
	ast.Equal(false, res, fmt.Sprintf("error committed(%q, %d, %d) = %t, expected: %t", digest, v, seqNo, res, false))

	cert.commit = make(map[Commit]bool)
	cert.commit[Commit{ReplicaId: uint64(0)}] = true
	cert.commit[Commit{ReplicaId: uint64(1)}] = true
	cert.commit[Commit{ReplicaId: uint64(2)}] = true
	res = rbft.committed(digest, v, seqNo)
	ast.Equal(true, res, fmt.Sprintf("error committed(%q, %d, %d) = %t, expected: %t", digest, v, seqNo, res, true))
}

func TestStartTimerIfOutstandingRequests(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.on(skipInProgress)
	rbft.startTimerIfOutstandingRequests()
	ast.Equal(false, rbft.in(timerActive), "should not start newView timer")

	rbft.off(skipInProgress)
	rbft.exec.setCurrentExec(nil)
	rbft.startTimerIfOutstandingRequests()
	ast.Equal(false, rbft.in(timerActive), "should not start newView timer")

	rbft.storeMgr.outstandingReqBatches["something"] = nil
	rbft.startTimerIfOutstandingRequests()
	ast.Equal(true, rbft.in(timerActive), "should start newView timer")
}

func TestConsensusMsgHelper(t *testing.T) {
	msg := &ConsensusMessage{
		Type: ConsensusMessage_CHECKPOINT,
	}
	rs := cMsgToPbMsg(msg, 1211)
	assert.Equal(t, uint64(1211), rs.Id, "error consensusMsgHelper failed!")
}

func TestNullRequestMsgHelper(t *testing.T) {
	msg := nullRequestMsgToPbMsg(1211)
	assert.Equal(t, uint64(1211), msg.Id, fmt.Sprintf("error nullRequestMsgHelper(%d) failed!", 1211))
}

func TestDeleteExistedTx(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	currentVid := uint64(100)
	rbft.batchVdr.setCurrentVid(&currentVid)
	digest := "something"
	rbft.batchVdr.saveToCVB(digest, nil)
	rbft.storeMgr.outstandingReqBatches[digest] = nil

	rbft.deleteExistedTx(digest)
	ast.Equal(uint64(100), rbft.batchVdr.lastVid, "deleteExistedTx failed")
	ast.Nil(rbft.batchVdr.currentVid, "deleteExistedTx failed")
	_, ok := rbft.batchVdr.cacheValidatedBatch[digest]
	ast.Equal(false, ok, "deleteExistedTx failed")
	_, ok = rbft.storeMgr.outstandingReqBatches[digest]
	ast.Equal(false, ok, "deleteExistedTx failed")
}

func TestIsPrePrepareLegal(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	v := uint64(1)
	seqNo := uint64(100)
	digest := "d"
	prePrepare := &PrePrepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
		ReplicaId:      2,
	}

	rbft.on(inNegotiateView)
	res := rbft.isPrePrepareLegal(prePrepare)
	ast.Equal(false, res, "isPrePrepareLegal failed")

	rbft.off(inNegotiateView)
	rbft.on(inViewChange)
	res = rbft.isPrePrepareLegal(prePrepare)
	ast.Equal(false, res, "isPrePrepareLegal failed")

	rbft.off(inViewChange)
	res = rbft.isPrePrepareLegal(prePrepare)
	ast.Equal(false, res, "isPrePrepareLegal failed")

	prePrepare.ReplicaId = 1
	res = rbft.isPrePrepareLegal(prePrepare)
	ast.Equal(false, res, "isPrePrepareLegal failed")

	prePrepare.View = 0
	res = rbft.isPrePrepareLegal(prePrepare)
	ast.Equal(true, res, "isPrePrepareLegal failed")
}

func TestIsPrepareLegal(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	v := uint64(2)
	seqNo := uint64(100)
	digest := "d"
	prepare := &Prepare{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
		ReplicaId:      1,
	}

	rbft.on(inNegotiateView)
	res := rbft.isPrepareLegal(prepare)
	ast.Equal(false, res, "isPrepareLegal failed")

	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	res = rbft.isPrepareLegal(prepare)
	ast.Equal(false, res, "isPrepareLegal failed")

	rbft.on(inRecovery)
	res = rbft.isPrepareLegal(prepare)
	ast.Equal(false, res, "isPrePrepareLegal failed")

	prepare.View = 0
	res = rbft.isPrepareLegal(prepare)
	ast.Equal(true, res, "isPrePrepareLegal failed")
}

func TestIsCommitLegal(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	v := uint64(2)
	seqNo := uint64(100)
	digest := "d"
	commit := &Commit{
		View:           v,
		SequenceNumber: seqNo,
		BatchDigest:    digest,
		ReplicaId:      1,
	}
	rbft.on(inNegotiateView)
	res := rbft.isCommitLegal(commit)
	ast.Equal(false, res, "isCommitLegal failed")

	rbft.off(inNegotiateView)
	res = rbft.isCommitLegal(commit)
	ast.Equal(false, res, "isCommitLegal failed")

	commit.View = 0
	res = rbft.isCommitLegal(commit)
	ast.Equal(true, res, "isCommitLegal failed")
}
