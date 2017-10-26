//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"encoding/base64"
	"fmt"
	"reflect"
	"testing"
	"time"

	"hyperchain/consensus/consensusMocks"
	"hyperchain/core/types"
	"hyperchain/manager/protos"

	"github.com/golang/protobuf/proto"
	"github.com/stretchr/testify/assert"
)

func getTestViewChange() (vc *ViewChange) {
	vc = &ViewChange{
		Basis: &VcBasis{
			View: uint64(1),
			H:    uint64(50),
			Cset: []*Vc_C{
				{
					SequenceNumber: 60,
					Id:             "chkpoint---digest60",
				},
				{
					SequenceNumber: 70,
					Id:             "chkpoint---digest70",
				},
			},
			Pset: []*Vc_PQ{
				{
					SequenceNumber: 51,
					BatchDigest:    "batch---P1",
					View:           0,
				},
				{
					SequenceNumber: 52,
					BatchDigest:    "batch---P2",
					View:           0,
				},
				{
					SequenceNumber: 53,
					BatchDigest:    "batch---P3",
					View:           0,
				},
			},
			Qset: []*Vc_PQ{
				{
					SequenceNumber: 54,
					BatchDigest:    "batch---Q4",
					View:           0,
				},
				{
					SequenceNumber: 55,
					BatchDigest:    "batch---Q5",
					View:           0,
				},
				{
					SequenceNumber: 56,
					BatchDigest:    "batch---Q6",
					View:           0,
				},
			},
			ReplicaId: 2,
		},
		Timestamp: time.Now().UnixNano(),
	}

	return
}

func TestNewVcManager(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	structName, nilElems, err := checkNilElems(rbft.vcMgr)
	if err != nil {
		t.Error(err.Error())
	}
	if nilElems != nil {
		t.Errorf("There exists some nil elements: %v in struct: %s", nilElems, structName)
	}
}

func TestNewViewTimer(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.off(timerActive)
	ast.Equal(false, rbft.in(timerActive), "be set to inActive, expect false")

	rbft.startNewViewTimer(time.Second, "nothing")
	ast.Equal(true, rbft.in(timerActive), "startNewViewTimer set this to active, expect true")

	rbft.stopNewViewTimer()
	ast.Equal(false, rbft.in(timerActive), "stopNewViewTimer set it to inActive, expect false")

	rbft.softStartNewViewTimer(time.Second, "nothing")
	ast.Equal(true, rbft.in(timerActive), "softStartNewViewTimer set this to active, expect true")
}

func TestCalcPSet(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	calc_pset := rbft.calcPSet()
	ast.Equal(0, len(calc_pset), fmt.Sprintf("pset should be nil, but it is :%v", calc_pset))

	//plist map[uint64]*Vc_PQ
	rbft.vcMgr.plist[uint64(10)] = &Vc_PQ{SequenceNumber: uint64(10), View: uint64(3)}
	v := uint64(1)
	n := uint64(10)
	d := "1"
	rbft.storeMgr.getCert(v, n, d)
	calc_pset = rbft.calcPSet()
	ast.Equal(1, len(calc_pset), fmt.Sprintf("pset should contain one prepare, but it is :%v", calc_pset))

	rbft.getCertPreprepared(d, v, n)
	calc_pset = rbft.calcPSet()
	ast.Equal(1, len(calc_pset), fmt.Sprintf("pset should contain one prepare, but it is :%v", calc_pset))

	rbft.getCertPrepared(d, v, n)
	calc_pset = rbft.calcPSet()
	ast.Equal(1, len(calc_pset), fmt.Sprintf("pset should contain one prepare, but it is :%v", calc_pset))

	rbft.getCertPrepared(d, v, uint64(11))
	calc_pset = rbft.calcPSet()
	ast.Equal(2, len(calc_pset), fmt.Sprintf("pset should contain two prepare, but it is :%v", calc_pset))
}

func TestCalcQSet(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	calc_qset := rbft.calcQSet()
	ast.Equal(0, len(calc_qset), fmt.Sprintf("qset should be nil, but it is :%v", calc_qset))

	rbft.vcMgr.qlist[qidx{d: "10", n: uint64(10)}] = &Vc_PQ{SequenceNumber: uint64(10), View: uint64(3)}
	v := uint64(1)
	n := uint64(10)
	d := "10"
	rbft.storeMgr.getCert(v, n, d)
	calc_qset = rbft.calcQSet()
	ast.Equal(1, len(calc_qset), fmt.Sprintf("qset should contain one prePrepare, but it is :%v", calc_qset))

	rbft.getCertPreprepared(d, v, n)
	calc_qset = rbft.calcQSet()
	ast.Equal(1, len(calc_qset), fmt.Sprintf("qset should contain one prePrepare, but it is :%v", calc_qset))

	rbft.getCertPreprepared(d, v, uint64(11))
	calc_qset = rbft.calcQSet()
	ast.Equal(2, len(calc_qset), fmt.Sprintf("qset should contain two prePrepare, but it is :%v", calc_qset))
}

func TestBeforeSendVC(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.on(inNegotiateView)
	err = rbft.beforeSendVC()
	ast.NotNil(err, "inNegotiateView, expect err not nil")

	rbft.off(inNegotiateView)
	rbft.on(inRecovery)
	err = rbft.beforeSendVC()
	ast.NotNil(err, "inRecovery, expect err not nil")

	rbft.off(inRecovery)
	rbft.on(timerActive)
	rbft.vcMgr.viewChangeStore[vcidx{uint64(0), uint64(1)}] = nil
	err = rbft.beforeSendVC()
	ast.Nil(err, "go on, expect nil")
	ast.Equal(false, rbft.in(timerActive), "stopNewViewTimer in beforeSendVC failed")
	ast.Equal(uint64(1), rbft.view, "beforeSendVC failed")
	ast.Equal(0, len(rbft.vcMgr.viewChangeStore), "delete old vc failed")
}

func TestCorrectViewChange(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	vc := getTestViewChange()

	ast.Equal(true, rbft.correctViewChange(vc), "Should be a correct view change, expect true")

	vc.Basis.View = uint64(0)
	ast.Equal(false, rbft.correctViewChange(vc), "Should not be a correct view change: vc.View is lower than view in PQset.")

	vc.Basis.View = 1
	vc.Basis.Cset[0].SequenceNumber = uint64(40)
	ast.Equal(false, rbft.correctViewChange(vc), "Should not be a correct view change: vc.H is too high.")
}

func TestGatherPQC(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.vcMgr.plist[uint64(20)] = &Vc_PQ{SequenceNumber: uint64(20), View: uint64(3)}
	rbft.vcMgr.qlist[qidx{d: "20", n: uint64(20)}] = &Vc_PQ{SequenceNumber: uint64(20), View: uint64(3)}
	rbft.storeMgr.chkpts[uint64(20)] = "20"
	cset, pset, qset := rbft.gatherPQC()
	ast.Equal(2, len(cset), "should contain two checkpoint")
	ast.Equal(1, len(pset), "should contain one prepare")
	ast.Equal(1, len(qset), "should contain one prePrepare")

	rbft.h = uint64(10)
	rbft.vcMgr.plist[uint64(5)] = &Vc_PQ{SequenceNumber: uint64(5), View: uint64(3)}
	rbft.vcMgr.qlist[qidx{d: "5", n: uint64(5)}] = &Vc_PQ{SequenceNumber: uint64(5), View: uint64(3)}
	cset, pset, qset = rbft.gatherPQC()
	ast.Equal(2, len(cset), "should contain two checkpoint")
	ast.Equal(1, len(pset), "should contain one prepare")
	ast.Equal(1, len(qset), "should contain one prePrepare")
}

func TestRecvAndSendViewChange(t *testing.T) { // test some pre-check procedure and resend problem
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.off(inNegotiateView)
	rbft.on(inRecovery)
	if err := rbft.sendViewChange(); err != nil {
		t.Error("should end in inRecovery, expect nil")
	}

	vc := getTestViewChange()

	rbft.on(inNegotiateView)
	ast.Nil(rbft.recvViewChange(vc), fmt.Sprintf("Replica %d is in negoview, so it should not receive this view-change message", rbft.id))

	rbft.off(inNegotiateView)
	rbft.on(inRecovery)
	ast.Nil(rbft.recvViewChange(vc), fmt.Sprintf("Replica %d is in recovery, so it should not receive this view-change message", rbft.id))

	rbft.off(inRecovery)

	rbft.view = uint64(3)
	ast.Nil(rbft.recvViewChange(vc), fmt.Sprintf("Replica %d found view-change message for old view, so it should not receive this view-change message", rbft.id))

	rbft.view = uint64(0)

	vc.Basis.Pset[0].View = 5
	ast.Nil(rbft.recvViewChange(vc), fmt.Sprintf("Replica %d found view-change message incorrect, so it should not receive this view-change message", rbft.id))

	vc.Basis.Pset[0].View = 0
	rbft.recvViewChange(vc)
	ast.Equal(1, rbft.vcMgr.vcResendCount, "recv view change from itself for the first time, expect 1")
	ast.Equal(vc, rbft.vcMgr.viewChangeStore[vcidx{vc.Basis.View, vc.Basis.ReplicaId}], "should store this vc, expect true")

	rbft.recvViewChange(vc)
	ast.Equal(2, rbft.vcMgr.vcResendCount, "recv this view change before, expect 1")

	rbft.vcMgr.vcResendLimit = 3
	vc.Basis.View = uint64(2)
	rbft.recvViewChange(vc)
	ast.Equal(0, rbft.vcMgr.vcResendCount, "should recover, expect 0")
}

func TestRecvAndSendViewChange2(t *testing.T) { // test normal case
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	vc := getTestViewChange()
	vc2 := getTestViewChange()
	vc3 := getTestViewChange()
	vc.Basis.ReplicaId = 4
	vc2.Basis.ReplicaId = 3
	vc3.Basis.ReplicaId = 2

	rbft.on(inViewChange)
	rbft.off(timerActive)
	rbft.recvViewChange(vc)
	rbft.recvViewChange(vc2)
	ast.Equal(uint64(1), rbft.view, "should send view change, expect 1")
	ast.Equal(true, rbft.in(timerActive), "should startNewViewTimer, expect true")

	rbft.off(inViewChange)
	rbft.recvViewChange(vc3)
	ast.Equal(uint64(2), rbft.view, "should send view change again, expect 2")
}

func TestSendAndRecvNewView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft2, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft2.namespace)
	ast.Equal(nil, err, err)
	rbft2.Start()

	chkpt50 := &protos.BlockchainInfo{
		Height:            50,
		CurrentBlockHash:  []byte("current block is 50."),
		PreviousBlockHash: []byte("previous block is 49."),
	}
	chkpt60 := &protos.BlockchainInfo{
		Height:            60,
		CurrentBlockHash:  []byte("current block is 60."),
		PreviousBlockHash: []byte("previous block is 59."),
	}
	chkpt70 := &protos.BlockchainInfo{
		Height:            70,
		CurrentBlockHash:  []byte("current block is 70."),
		PreviousBlockHash: []byte("previous block is 69."),
	}

	digest50, _ := proto.Marshal(chkpt50)
	digest60, _ := proto.Marshal(chkpt60)
	digest70, _ := proto.Marshal(chkpt70)

	vcC1 := []*Vc_C{
		{
			SequenceNumber: 50,
			Id:             base64.StdEncoding.EncodeToString(digest50),
		},
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
	}
	vcP1 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcQ1 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}

	vcC2 := []*Vc_C{
		{
			SequenceNumber: 50,
			Id:             base64.StdEncoding.EncodeToString(digest50),
		},
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
	}
	vcP2 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcQ2 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 67,
			BatchDigest:    "batch---67",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcC3 := []*Vc_C{
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
		{
			SequenceNumber: 70,
			Id:             base64.StdEncoding.EncodeToString(digest70),
		},
	}
	vcP3 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
		{
			SequenceNumber: 80,
			BatchDigest:    "batch---80",
			View:           0,
		},
	}
	vcQ3 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 71,
			BatchDigest:    "batch---71",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcC4 := []*Vc_C{
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
	}
	vcP4 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
		{
			SequenceNumber: 87,
			BatchDigest:    "batch---87",
			View:           0,
		},
	}
	vcQ4 := []*Vc_PQ{
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
		{
			SequenceNumber: 79,
			BatchDigest:    "batch---79",
			View:           0,
		},
	}

	vcBasis1 := &VcBasis{
		View:      1,
		H:         40,
		ReplicaId: 1,
	}
	vc1 := &ViewChange{Basis: vcBasis1}

	vcBasis2 := &VcBasis{
		View:      2,
		H:         40,
		ReplicaId: 2,
	}
	vc2 := &ViewChange{Basis: vcBasis2}

	vcBasis3 := &VcBasis{
		View:      2,
		H:         50,
		ReplicaId: 3,
	}
	vc3 := &ViewChange{Basis: vcBasis3}

	vcBasis4 := &VcBasis{
		View:      2,
		H:         50,
		ReplicaId: 4,
	}
	vc4 := &ViewChange{Basis: vcBasis4}

	vcidx1 := vcidx{v: vc1.Basis.View, id: vc1.Basis.ReplicaId}
	vcidx2 := vcidx{v: vc2.Basis.View, id: vc2.Basis.ReplicaId}
	vcidx3 := vcidx{v: vc3.Basis.View, id: vc3.Basis.ReplicaId}
	vcidx4 := vcidx{v: vc4.Basis.View, id: vc4.Basis.ReplicaId}
	rbft.vcMgr.viewChangeStore = map[vcidx]*ViewChange{
		vcidx1: vc1,
		vcidx2: vc2,
		vcidx3: vc3,
		vcidx4: vc4,
	}

	rbft.on(inNegotiateView)
	ast.Nil(rbft.sendNewView(), fmt.Sprintf("Replica %d try to sendNewView, but it's in nego-view", rbft.id))

	rbft.off(inNegotiateView)
	rbft.vcMgr.newViewStore[rbft.view] = nil
	ast.Nil(rbft.sendNewView(), fmt.Sprintf("Replica %d try to sendNewView, but there is the same view in newViewStore", rbft.id))

	delete(rbft.vcMgr.newViewStore, rbft.view)
	ast.Nil(rbft.sendNewView(), fmt.Sprintf("Replica %d try to sendNewView, but there is no cSet in vc", rbft.id))

	rbft.vcMgr.viewChangeStore[vcidx1].Basis.Cset = vcC1
	rbft.vcMgr.viewChangeStore[vcidx2].Basis.Cset = vcC2
	rbft.vcMgr.viewChangeStore[vcidx3].Basis.Cset = vcC3
	rbft.vcMgr.viewChangeStore[vcidx4].Basis.Cset = vcC4
	ast.Nil(rbft.sendNewView(), fmt.Sprintf("Replica %d try to sendNewView, but there is no pSet and qSet", rbft.id))

	rbft.vcMgr.viewChangeStore[vcidx1].Basis.Pset = vcP1
	rbft.vcMgr.viewChangeStore[vcidx2].Basis.Pset = vcP2
	rbft.vcMgr.viewChangeStore[vcidx3].Basis.Pset = vcP3
	rbft.vcMgr.viewChangeStore[vcidx4].Basis.Pset = vcP4
	rbft.vcMgr.viewChangeStore[vcidx1].Basis.Qset = vcQ1
	rbft.vcMgr.viewChangeStore[vcidx2].Basis.Qset = vcQ2
	rbft.vcMgr.viewChangeStore[vcidx3].Basis.Qset = vcQ3
	rbft.vcMgr.viewChangeStore[vcidx4].Basis.Qset = vcQ4

	rbft.vcMgr.newViewStore = make(map[uint64]*NewView)
	rbft.sendNewView()

	vset := rbft.getViewChanges()

	cp, ok, _ := rbft.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, Vc_C{SequenceNumber: 60, Id: base64.StdEncoding.EncodeToString(digest60)})) {
		t.Error("should get initial checkpoint!")
	}

	msglist := rbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	ast.Equal(vc1.Basis.Pset[0].BatchDigest, msglist[61], "sequence number 61 should be assigned but not assigned actually")

	nv := &NewView{
		View:      rbft.view,
		Xset:      msglist,
		ReplicaId: rbft.id,
	}
	ast.Equal(nv, rbft.vcMgr.newViewStore[rbft.view], "should be the same, expect equal")

	rbft2.view = uint64(2)
	rbft2.on(inNegotiateView, inViewChange)
	ast.Nil(rbft2.recvNewView(nv), "in nego-view, so it should not receive this message")

	rbft2.off(inNegotiateView)
	rbft2.on(inRecovery)
	ast.Nil(rbft2.recvNewView(nv), "in recovery, so it should not receive this message")
	ast.Equal(true, rbft2.recoveryMgr.recvNewViewInRecovery, "receive newView message in recovery, expect true")

	rbft2.off(inRecovery)
	rbft2.recoveryMgr.recvNewViewInRecovery = false
	ast.Nil(rbft2.recvNewView(nv), "nv's view is smaller than rbft2.view, so it should not receive this message")

	nv.View = uint64(3)
	nv.ReplicaId = uint64(4)
	rbft2.recvNewView(nv)
	nv2, ok := rbft2.vcMgr.newViewStore[nv.View]
	ast.Equal(true, ok, "should be stored, expect true")
	if ok {
		ast.Equal(nv, nv2, "should be stored, expect equal")
	}

	rbft2.vcMgr.viewChangeStore[vcidx1] = vc1
	rbft2.vcMgr.viewChangeStore[vcidx2] = vc2
	rbft2.vcMgr.viewChangeStore[vcidx3] = vc3
	delete(rbft2.vcMgr.newViewStore, nv.View)
	rbft2.recvNewView(nv)
}

func TestFeedMissingReqBatchIfNeeded(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.h = uint64(50)
	xset := make(map[uint64]string)
	xset[uint64(40)] = "40"
	xset[uint64(51)] = ""
	xset[uint64(52)] = "52"
	miss := rbft.feedMissingReqBatchIfNeeded(xset)
	ast.Equal(true, miss, fmt.Sprintf("miss the batch whose digest is %v", xset[uint64(52)]))
	_, ok := rbft.storeMgr.missingReqBatches[xset[uint64(52)]]
	ast.Equal(true, ok, fmt.Sprintf("miss the batch whose digest is %v", xset[uint64(52)]))
}

func TestPrimaryCheckNewView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.h = uint64(10)
	initialCp := Vc_C{
		SequenceNumber: uint64(10),
		Id:             "123",
	}
	replicas := []replicaInfo{}
	ast.Nil(rbft.primaryCheckNewView(initialCp, replicas, nil), "hash could not be decoded, expect nil")

	xset := make(map[uint64]string)
	xset[uint64(51)] = ""
	nv := &NewView{Xset: xset}
	rbft.off(vcHandled)
	rbft.storeMgr.outstandingReqBatches["1"] = nil
	ast.Equal(1, len(rbft.storeMgr.outstandingReqBatches), "store one, expect 1")
	rbft.exec.lastExec = uint64(100)
	rbft.primaryCheckNewView(initialCp, replicas, nv)
	ast.Equal(0, len(rbft.storeMgr.outstandingReqBatches), "should reset all states, expect 0")
}

func TestReplicaCheckNewView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	ast.Nil(rbft.replicaCheckNewView(), "no new view message stored, expect nil")

	chkpt50 := &protos.BlockchainInfo{
		Height:            50,
		CurrentBlockHash:  []byte("current block is 50."),
		PreviousBlockHash: []byte("previous block is 49."),
	}
	chkpt60 := &protos.BlockchainInfo{
		Height:            60,
		CurrentBlockHash:  []byte("current block is 60."),
		PreviousBlockHash: []byte("previous block is 59."),
	}
	chkpt70 := &protos.BlockchainInfo{
		Height:            70,
		CurrentBlockHash:  []byte("current block is 70."),
		PreviousBlockHash: []byte("previous block is 69."),
	}

	digest50, _ := proto.Marshal(chkpt50)
	digest60, _ := proto.Marshal(chkpt60)
	digest70, _ := proto.Marshal(chkpt70)

	vcC1 := []*Vc_C{
		{
			SequenceNumber: 50,
			Id:             base64.StdEncoding.EncodeToString(digest50),
		},
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
	}
	vcP1 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcQ1 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}

	vcC2 := []*Vc_C{
		{
			SequenceNumber: 50,
			Id:             base64.StdEncoding.EncodeToString(digest50),
		},
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
	}
	vcP2 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcQ2 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 67,
			BatchDigest:    "batch---67",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcC3 := []*Vc_C{
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
		{
			SequenceNumber: 70,
			Id:             base64.StdEncoding.EncodeToString(digest70),
		},
	}
	vcP3 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
		{
			SequenceNumber: 80,
			BatchDigest:    "batch---80",
			View:           0,
		},
	}
	vcQ3 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 71,
			BatchDigest:    "batch---71",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
	}
	vcC4 := []*Vc_C{
		{
			SequenceNumber: 60,
			Id:             base64.StdEncoding.EncodeToString(digest60),
		},
	}
	vcP4 := []*Vc_PQ{
		{
			SequenceNumber: 66,
			BatchDigest:    "batch---66",
			View:           0,
		},
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
		{
			SequenceNumber: 87,
			BatchDigest:    "batch---87",
			View:           0,
		},
	}
	vcQ4 := []*Vc_PQ{
		{
			SequenceNumber: 75,
			BatchDigest:    "batch---75",
			View:           0,
		},
		{
			SequenceNumber: 79,
			BatchDigest:    "batch---79",
			View:           0,
		},
	}

	vcBasis1 := &VcBasis{
		View:      1,
		H:         40,
		ReplicaId: 1,
		Cset:      vcC1,
		Pset:      vcP1,
		Qset:      vcQ1,
	}
	vc1 := &ViewChange{Basis: vcBasis1}

	vcBasis2 := &VcBasis{
		View:      2,
		H:         40,
		ReplicaId: 2,
		Cset:      vcC2,
		Pset:      vcP2,
		Qset:      vcQ2,
	}
	vc2 := &ViewChange{Basis: vcBasis2}

	vcBasis3 := &VcBasis{
		View:      2,
		H:         50,
		ReplicaId: 3,
		Cset:      vcC3,
		Pset:      vcP3,
		Qset:      vcQ3,
	}
	vc3 := &ViewChange{Basis: vcBasis3}

	vcBasis4 := &VcBasis{
		View:      2,
		H:         50,
		ReplicaId: 4,
		Cset:      vcC4,
		Pset:      vcP4,
		Qset:      vcQ4,
	}
	vc4 := &ViewChange{Basis: vcBasis4}

	vcidx1 := vcidx{v: vc1.Basis.View, id: vc1.Basis.ReplicaId}
	vcidx2 := vcidx{v: vc2.Basis.View, id: vc2.Basis.ReplicaId}
	vcidx3 := vcidx{v: vc3.Basis.View, id: vc3.Basis.ReplicaId}
	vcidx4 := vcidx{v: vc4.Basis.View, id: vc4.Basis.ReplicaId}

	vset := []*VcBasis{vcBasis1, vcBasis2, vcBasis3, vcBasis4}
	cp, ok, _ := rbft.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, Vc_C{SequenceNumber: 60, Id: base64.StdEncoding.EncodeToString(digest60)})) {
		t.Error("should get initial checkpoint!")
	}

	msglist := rbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	ast.Equal(vc1.Basis.Pset[0].BatchDigest, msglist[61], "sequence number 61 should be assigned but not assigned actually")

	nv := &NewView{
		View:      rbft.view,
		Xset:      msglist,
		ReplicaId: rbft.id,
	}
	rbft.vcMgr.newViewStore[rbft.view] = nv
	rbft.off(inViewChange)
	ast.Nil(rbft.replicaCheckNewView(), "activeView is 1, expect nil")

	rbft.on(inViewChange)
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.replicaCheckNewView()
	ast.Equal(uint64(1), rbft.view, "selectInitialCheckpoint failed and sendViewChange, expect 1")

	rbft.on(inViewChange)
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.vcMgr.viewChangeStore = map[vcidx]*ViewChange{
		vcidx1: vc1,
		vcidx2: vc2,
		vcidx3: vc3,
	}
	rbft.vcMgr.newViewStore[rbft.view] = nv
	rbft.replicaCheckNewView()
	ast.Equal(uint64(2), rbft.view, "assignSequenceNumbers failed and sendViewChange, expect 1")

	rbft.on(inViewChange)
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.vcMgr.viewChangeStore = map[vcidx]*ViewChange{
		vcidx1: vc1,
		vcidx2: vc2,
		vcidx3: vc3,
		vcidx4: vc4,
	}
	rbft.vcMgr.newViewStore[rbft.view] = &NewView{
		View:      uint64(100),
		Xset:      make(map[uint64]string),
		ReplicaId: uint64(3),
	}
	rbft.replicaCheckNewView()
	ast.Equal(uint64(3), rbft.view, "nv's vset is different and sendViewChange, expect 1")

	rbft.on(inViewChange)
	rbft.off(inNegotiateView)
	rbft.off(inRecovery)
	rbft.vcMgr.viewChangeStore = map[vcidx]*ViewChange{
		vcidx1: vc1,
		vcidx2: vc2,
		vcidx3: vc3,
		vcidx4: vc4,
	}
	rbft.vcMgr.newViewStore[rbft.view] = nv
	ast.Nil(rbft.replicaCheckNewView(), "hash could not be decoded, expect nil")
}

func TestResetStateForNewView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.on(vcHandled)
	ast.Nil(rbft.resetStateForNewView(), "vcHandled, expect true")

	rbft.off(vcHandled)
	rbft.resetStateForNewView()
	ast.Equal(true, rbft.in(vcHandled), "should be actived, expect true")
}

func TestRecvFinishVcReset(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	finish := &FinishVcReset{
		ReplicaId: rbft.id,
		View:      uint64(100),
		LowH:      uint64(100),
	}

	rbft.off(inViewChange)
	rbft.recvFinishVcReset(finish)
	ok := rbft.vcMgr.vcResetStore[*finish]
	ast.Equal(false, ok, "activeView is 1, should not received, expect false")

	rbft.on(inViewChange)
	rbft.recvFinishVcReset(finish)
	ok = rbft.vcMgr.vcResetStore[*finish]
	ast.Equal(false, ok, "view is not the same as rbft, should not received, expect false")

	finish.View = rbft.view
	rbft.vcMgr.vcResetStore[*finish] = true
	ast.Nil(rbft.recvFinishVcReset(finish), "already stored it, expect nil")

	rbft.vcMgr.vcResetStore[*finish] = false
	rbft.recvFinishVcReset(finish)
	ok = rbft.vcMgr.vcResetStore[*finish]
	ast.Equal(true, ok, "should store it, expect true")
}

func TestProcessReqInNewView(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.view = 1

	finish1 := FinishVcReset{
		ReplicaId: 1,
		View:      1,
		LowH:      10,
	}
	finish2 := FinishVcReset{
		ReplicaId: 2,
		View:      1,
		LowH:      10,
	}
	finish3 := FinishVcReset{
		ReplicaId: 3,
		View:      1,
		LowH:      10,
	}

	rbft.vcMgr.vcResetStore[finish2] = true
	rbft.vcMgr.vcResetStore[finish3] = true
	ast.Nil(rbft.processReqInNewView(), "less than quorum, expect nil")

	rbft.vcMgr.vcResetStore[finish1] = true
	rbft.on(inVcReset)
	rbft.off(skipInProgress)
	ast.Nil(rbft.processReqInNewView(), "has not done with vcReset and not in stateUpdate, expect nil")

	rbft.off(inVcReset)
	ast.Nil(rbft.processReqInNewView(), "have not stored new view, expect nil")

	rbft.h = uint64(10)
	nv := &NewView{
		Xset: map[uint64]string{
			11: "chkponit---digest60",
			12: "chkponit---digest70",
		},
	}
	rbft.vcMgr.newViewStore[rbft.view] = nv
	event := rbft.processReqInNewView()
	viewChangedEvent := &LocalEvent{
		Service:   VIEW_CHANGE_SERVICE,
		EventType: VIEW_CHANGED_EVENT,
	}
	ast.Equal(viewChangedEvent, event, "should return a VIEW_CHANGED_EVENT")
}

func TestGetViewChanges(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	vc1 := getTestViewChange()
	vc1.Basis.ReplicaId = uint64(1)

	vc2 := getTestViewChange()
	vc2.Basis.ReplicaId = uint64(2)

	vc3 := getTestViewChange()
	vc3.Basis.ReplicaId = uint64(3)

	vc4 := getTestViewChange()
	vc4.Basis.ReplicaId = uint64(4)

	rbft.vcMgr.viewChangeStore = map[vcidx]*ViewChange{
		vcidx{v: vc1.Basis.View, id: vc1.Basis.ReplicaId}: vc1,
		vcidx{v: vc2.Basis.View, id: vc2.Basis.ReplicaId}: vc2,
		vcidx{v: vc3.Basis.View, id: vc3.Basis.ReplicaId}: vc3,
		vcidx{v: vc4.Basis.View, id: vc4.Basis.ReplicaId}: vc4,
	}

	vset := rbft.getViewChanges()
	count := 0
	for _, value := range vset {
		if value.ReplicaId == uint64(1) {
			if reflect.DeepEqual(value, vc1.Basis) {
				count = count + 1
			} else {
				t.Errorf("Error while storing View-change1 to VSet.\nActual: %v\nExpect: %v", value, vc1)
			}
		} else if value.ReplicaId == uint64(2) {
			if reflect.DeepEqual(value, vc2.Basis) {
				count = count + 1
			} else {
				t.Errorf("Error while storing View-change2 to VSet.\nActual: %v\nExpect: %v", value, vc2)
			}
		} else if value.ReplicaId == uint64(3) {
			if reflect.DeepEqual(value, vc3.Basis) {
				count = count + 1
			} else {
				t.Errorf("Error while storing View-change3 to VSet.\nActual: %v\nExpect: %v", value, vc3)
			}
		} else if value.ReplicaId == uint64(4) {
			if reflect.DeepEqual(value, vc4.Basis) {
				count = count + 1
			} else {
				t.Errorf("Error while storing View-change4 to VSet.\nActual: %v\nExpect: %v", value, vc4)
			}
		} else {
			t.Error("Error while storing View-change to VSet.")
		}
	}

	if !(count == 4) {
		t.Error("Error while storing View-change to VSet: lack of View-change in VSet")
	}
}

func TestSelectInitialCheckpoint(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	vc1 := getTestViewChange()
	vc1.Basis.ReplicaId = uint64(1)

	vc2 := getTestViewChange()
	vc2.Basis.ReplicaId = uint64(2)

	vc3 := getTestViewChange()
	vc3.Basis.ReplicaId = uint64(3)

	vc4 := getTestViewChange()
	vc4.Basis.ReplicaId = uint64(4)

	vset := []*VcBasis{}

	_, ok, _ := rbft.selectInitialCheckpoint(vset)
	ast.Equal(false, ok, "VSet is nil. so we should not be able to select initial checkpoint.")

	vc1.Basis.H = 20
	vc1.Basis.Cset = []*Vc_C{
		{
			SequenceNumber: 30,
			Id:             "chkpoint---digest30",
		},
		{
			SequenceNumber: 30,
			Id:             "chkpoint---digest30",
		},
		{
			SequenceNumber: 40,
			Id:             "chkpoint---digest40",
		},
	}

	vc2.Basis.H = 30
	vc2.Basis.Cset = []*Vc_C{
		{
			SequenceNumber: 40,
			Id:             "chkpoint---digest40",
		},

		{
			SequenceNumber: 50,
			Id:             "chkpoint---digest50",
		},
	}

	vc3.Basis.H = 50
	vc3.Basis.Cset = []*Vc_C{
		{
			SequenceNumber: 60,
			Id:             "chkpoint---digest60",
		},
		{
			SequenceNumber: 70,
			Id:             "chkpoint---digest70",
		},
	}

	vc4.Basis.H = 50
	vc4.Basis.Cset = []*Vc_C{
		{
			SequenceNumber: 60,
			Id:             "chkpoint---digest60",
		},
		{
			SequenceNumber: 70,
			Id:             "chkpoint---digest70",
		},
	}

	vset = []*VcBasis{vc1.Basis, vc2.Basis, vc3.Basis, vc4.Basis}

	cp, ok, _ := rbft.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, Vc_C{SequenceNumber: 70, Id: "chkpoint---digest70"})) {
		t.Error("should get initial checkpoint!")
	}

}

func TestAssignSequenceNumbers(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	vc1 := &VcBasis{
		View:      1,
		H:         40,
		ReplicaId: 1,

		Cset: []*Vc_C{
			{
				SequenceNumber: 50,
				Id:             "chkpoint---digest50",
			},
			{
				SequenceNumber: 60,
				Id:             "chkpoint---digest60",
			},
		},

		Pset: []*Vc_PQ{
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
			{
				SequenceNumber: 77,
				BatchDigest:    "batch---77",
				View:           0,
			},
		},

		Qset: []*Vc_PQ{
			{
				SequenceNumber: 61,
				BatchDigest:    "batch---61",
				View:           0,
			},
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
		},
	}

	vc2 := &VcBasis{
		View:      2,
		H:         40,
		ReplicaId: 2,

		Cset: []*Vc_C{
			{
				SequenceNumber: 60,
				Id:             "chkpoint---digest60",
			},
		},

		Pset: []*Vc_PQ{
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
			{
				SequenceNumber: 77,
				BatchDigest:    "batch---77",
				View:           0,
			},
			{
				SequenceNumber: 79,
				BatchDigest:    "batch---79",
				View:           0,
			},
		},

		Qset: []*Vc_PQ{
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 67,
				BatchDigest:    "batch---67",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
		},
	}

	vc3 := &VcBasis{
		View:      2,
		H:         50,
		ReplicaId: 3,

		Cset: []*Vc_C{
			{
				SequenceNumber: 60,
				Id:             "chkpoint---digest60",
			},
			{
				SequenceNumber: 70,
				Id:             "chkpoint---digest70",
			},
		},

		Pset: []*Vc_PQ{
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
			{
				SequenceNumber: 80,
				BatchDigest:    "batch---80",
				View:           0,
			},
		},

		Qset: []*Vc_PQ{
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 71,
				BatchDigest:    "batch---71",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
			{
				SequenceNumber: 77,
				BatchDigest:    "batch---77",
				View:           0,
			},
		},
	}

	vc4 := &VcBasis{
		View:      2,
		H:         50,
		ReplicaId: 4,

		Cset: []*Vc_C{
			{
				SequenceNumber: 60,
				Id:             "chkpoint---digest60",
			},
		},

		Pset: []*Vc_PQ{
			{
				SequenceNumber: 66,
				BatchDigest:    "batch---66",
				View:           0,
			},
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
			{
				SequenceNumber: 85,
				BatchDigest:    "batch---85",
				View:           0,
			},
			{
				SequenceNumber: 87,
				BatchDigest:    "batch---87",
				View:           0,
			},
		},

		Qset: []*Vc_PQ{
			{
				SequenceNumber: 75,
				BatchDigest:    "batch---75",
				View:           0,
			},
			{
				SequenceNumber: 77,
				BatchDigest:    "batch---77",
				View:           0,
			},
			{
				SequenceNumber: 79,
				BatchDigest:    "batch---79",
				View:           0,
			},
		},
	}

	vset := []*VcBasis{vc1, vc2, vc3, vc4}

	cp, ok, _ := rbft.selectInitialCheckpoint(vset)
	if !(ok == true && reflect.DeepEqual(cp, Vc_C{SequenceNumber: 60, Id: "chkpoint---digest60"})) {
		t.Error("should get initial checkpoint!")
	}

	msglist := rbft.assignSequenceNumbers(vset, cp.SequenceNumber)
	ast.Equal(vc1.Pset[0].BatchDigest, msglist[61], "sequence number 66 should be assigned but not assigned actually")
	ast.Equal(vc1.Pset[2].BatchDigest, msglist[63], "sequence number 77 should be assigned but not assigned actually")
}

func TestRecvFetchRequestBatch(t *testing.T) {
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

	frb := &FetchRequestBatch{
		BatchDigest: "a",
		ReplicaId:   rbft.id,
	}

	rbft.on(inNegotiateView)
	ast.Nil(rbft.recvFetchRequestBatch(frb), "replica in negotiate view, expect nil")

	rbft.off(inNegotiateView)
	rbft.on(inRecovery)
	ast.Nil(rbft.recvFetchRequestBatch(frb), "replica in recovery, expect nil")

	rbft.off(inRecovery)
	ast.Nil(rbft.recvFetchRequestBatch(frb), "didn't have this batch, expect nil")

	rbft.storeMgr.txBatchStore[frb.BatchDigest] = nil
	go func() {
		_ = <-eChan
		to := <-eChan
		ast.Equal(rbft.id, to, "recvNegoView failed")
	}()
	rbft.recvFetchRequestBatch(frb)
	time.Sleep(time.Nanosecond)
}

func TestRecvReturnRequestBatch(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.on(inNegotiateView)
	ast.Nil(rbft.recvReturnRequestBatch(nil), "in negotiate view, expect nil")

	rbft.off(inNegotiateView)
	rbft.on(inRecovery)
	ast.Nil(rbft.recvReturnRequestBatch(nil), "in recovery, expect nil")

	batch := &ReturnRequestBatch{
		Batch:       nil,
		BatchDigest: "batch",
		ReplicaId:   rbft.id,
	}
	rbft.off(inRecovery)
	ast.Nil(rbft.recvReturnRequestBatch(batch), "didn't store this in missingReqBatches, expect nil")

	rbft.storeMgr.missingReqBatches[batch.BatchDigest] = true
	rbft.on(inViewChange)
	ast.Nil(rbft.recvReturnRequestBatch(batch), "didn't store newView in newViewStore, expect nil")
	_, ok := rbft.storeMgr.missingReqBatches[batch.BatchDigest]
	ast.Equal(false, ok, "should be deleted, expect false")

	rbft.storeMgr.missingReqBatches[batch.BatchDigest] = true
	rbft.on(inViewChange)
	rbft.vcMgr.newViewStore[rbft.view] = nil
	rbft.off(vcHandled)
	rbft.recvReturnRequestBatch(batch)
	ast.Equal(true, rbft.in(vcHandled), "resetStateForNewView, expect true")

	rbft.storeMgr.missingReqBatches[batch.BatchDigest] = true
	rbft.off(inViewChange)
	rbft.off(inUpdatingN)
	ast.Nil(rbft.recvReturnRequestBatch(batch), "no missing batch and activeView and not in updatingN, expect nil")

	rbft.storeMgr.missingReqBatches[batch.BatchDigest] = true
	rbft.on(inUpdatingN)
	ast.Nil(rbft.recvReturnRequestBatch(batch), "no stored UpdateN, expect nil")
}

func TestPrimaryResendBatch(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	xset := make(map[uint64]string)
	xset[uint64(1)] = ""
	xset[uint64(2)] = "2"
	xset[uint64(3)] = "3"
	xset[uint64(5)] = "5"

	rbft.storeMgr.txBatchStore["3"] = &TransactionBatch{HashList: []string{}}
	rbft.primaryResendBatch(xset)
	_, ok := rbft.storeMgr.outstandingReqBatches["3"]
	ast.Equal(true, ok, "should validate this batch, expect true")
}

func TestRebuildCertStore(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()

	rbft.h = uint64(50)
	rbft.exec.lastExec = uint64(60)
	xset := make(map[uint64]string)
	xset[uint64(40)] = "40"
	xset[uint64(51)] = "51"
	xset[uint64(52)] = "52"
	rbft.storeMgr.txBatchStore["52"] = &TransactionBatch{
		TxList:     []*types.Transaction{},
		HashList:   []string{},
		Timestamp:  time.Now().UnixNano(),
		SeqNo:      uint64(52),
		ResultHash: "52",
	}
	rbft.rebuildCertStore(xset)
	cert := rbft.storeMgr.getCert(rbft.view, uint64(52), "52")
	ast.Equal(true, cert.validated, "should send prePrepare, expect true")
	ast.Equal(true, cert.sentCommit, "should send prePrepare, expect true")

	rbft.id = uint64(2)
	xset[uint64(53)] = "53"
	rbft.storeMgr.txBatchStore["53"] = &TransactionBatch{
		TxList:     []*types.Transaction{},
		HashList:   []string{},
		Timestamp:  time.Now().UnixNano(),
		SeqNo:      uint64(53),
		ResultHash: "53",
	}
	rbft.rebuildCertStore(xset)
	cert = rbft.storeMgr.getCert(rbft.view, uint64(53), "53")
	ast.Equal(true, cert.sentPrepare, "should send prepare, expect true")
}
