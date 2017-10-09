//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"time"
	"testing"
	"sync/atomic"

	"github.com/stretchr/testify/assert"
	"hyperchain/consensus/txpool"
	"hyperchain/core/types"
)

func TestVid(t *testing.T) {
	bv := newBatchValidator()
	ast := assert.New(t)
	lastVid := uint64(10)
	currentVid := uint64(11)
	bv.setCurrentVid(&currentVid)
	bv.setLastVid(lastVid)

	ast.Equal(lastVid, bv.lastVid, "set lastVid failed")
	ast.Equal(currentVid, *bv.currentVid, "set currentVid failed")

	bv.updateLCVid()
	ast.Equal(currentVid, bv.lastVid, "updateLCVid failed")
	ast.Nil(bv.currentVid, "updateLCVid failed")
}

func TestCVB(t *testing.T) {
	bv := newBatchValidator()
	ast := assert.New(t)
	cb1 := &cacheBatch{
		seqNo: uint64(1),
		resultHash: "cb1",
	}
	bv.saveToCVB(cb1.resultHash, cb1)
	ast.Equal(true, bv.containsInCVB(cb1.resultHash), "saveToCVB failed")
	ast.Equal(false, bv.containsInCVB("not exist"), "containsInCVB failed")
	CVB := make(map[string]*cacheBatch)
	CVB[cb1.resultHash] = cb1
	ast.Equal(CVB, bv.getCVB(), "getCVB failed")
	ast.Equal(cb1, bv.getCacheBatchFromCVB(cb1.resultHash), "getCacheBatchFromCVB failed")
	bv.deleteCacheFromCVB(cb1.resultHash)
	ast.Equal(0, len(bv.getCVB()), "deleteCacheFromCVB failed")
}

func TestBatchTimer(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	ast.Equal(false, rbft.batchMgr.isBatchTimerActive(), "batchTimer initialize failed")
	rbft.Start()

	rbft.batchMgr.eventMux.Post(txpool.TxHashBatch{
		BatchHash: "test",
	})
	rbft.startBatchTimer()
	ast.Equal(true, rbft.batchMgr.isBatchTimerActive(), "batchTimer start failed")
	rbft.restartBatchTimer()
	ast.Equal(true, rbft.batchMgr.isBatchTimerActive(), "batchTimer restart failed")
	rbft.stopBatchTimer()
	ast.Equal(false, rbft.batchMgr.isBatchTimerActive(), "batchTimer restart failed")
	rbft.batchMgr.stop()
}

func TestPrimaryValidateBatch(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	ast.Equal(false, rbft.batchMgr.isBatchTimerActive(), "batchTimer initialize failed")
	rbft.Start()

	tx1 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("testTx1"),
		Id:              uint64(1),
		TransactionHash: []byte("txHash1"),
	}
	txBatch := &TransactionBatch{
		TxList: 	[]*types.Transaction{tx1},
		HashList: 	[]string{"txHash1"},
		Timestamp: 	time.Now().UnixNano(),
	}
	rbft.primaryValidateBatch("test1", txBatch, 0)
	ast.Equal(uint64(1), rbft.seqNo, "seqNo increase failed")
	rbft.primaryValidateBatch("test2", txBatch, 100)
	ast.Equal(uint64(100), rbft.seqNo, "seqNo increase failed")
}

func TestFindNextValidateBatch(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)

	cert1 := rbft.storeMgr.getCert(0, 1, "1")
	idx1 := vidx{view: 0, seqNo: 1}
	rbft.batchVdr.preparedCert[idx1] = "1"

	cert2 := rbft.storeMgr.getCert(0, 2, "2")
	idx2 := vidx{view: 0, seqNo: 2}
	rbft.batchVdr.preparedCert[idx2] = "2"
	pp2 := &PrePrepare{
		View:             uint64(0),
		SequenceNumber:   uint64(2),
		BatchDigest:      "2",
		ResultHash:	  "2",
		HashBatch:	  &HashBatch{
			List:	  []string{"2"},
			Timestamp:time.Now().UnixNano(),
		},
		ReplicaId:        uint64(0),
	}
	cert2.prePrepare = pp2
	rbft.findNextValidateBatch()
	ast.NotEqual(rbft.batchVdr.currentVid, idx1.seqNo)


	pp1 := &PrePrepare{
		View:             uint64(0),
		SequenceNumber:   uint64(1),
		BatchDigest:      "1",
		ResultHash:	  "1",
		HashBatch:	  &HashBatch{
			List:	  []string{},
			Timestamp:time.Now().UnixNano(),
		},
		ReplicaId:        uint64(0),
	}
	cert1.prePrepare = pp1
	find, _, _, _ := rbft.findNextValidateBatch()
	ast.Equal(true, find, "findNextValidateBatch failed")
	ast.Equal(*rbft.batchVdr.currentVid, idx1.seqNo, "findNextValidateBatch failed")
	rbft.batchVdr.updateLCVid()

	find, _, _, _ = rbft.findNextValidateBatch()
	ast.Equal(false, find, "findNextValidateBatch failed")
	ast.Nil(rbft.batchVdr.currentVid, "findNextValidateBatch failed")


	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("testTx"),
		Id:              uint64(1),
		TransactionHash: []byte("3"),
	}
	cert3 := rbft.storeMgr.getCert(0, 3, "3")
	idx3 := vidx{view: 0, seqNo: 3}
	rbft.batchVdr.preparedCert[idx3] = "3"
	pp3 := &PrePrepare{
		View:             uint64(0),
		SequenceNumber:   uint64(3),
		BatchDigest:      "3",
		ResultHash:	  "3",
		HashBatch:	  &HashBatch{
			List:	  []string{tx.GetHash().Hex()},
			Timestamp:time.Now().UnixNano(),
		},
		ReplicaId:        uint64(0),
	}
	cert3.prePrepare = pp3
	currentVid := idx2.seqNo
	rbft.batchVdr.setCurrentVid(&currentVid)
	rbft.batchVdr.updateLCVid()

	rbft.batchMgr.txPool.AddNewTx(tx, true, true)
	rbft.batchMgr.txPool.GenerateTxBatch()
	find, _, _, _ = rbft.findNextValidateBatch()
	ast.Equal(false, find, "findNextValidateBatch failed")
	ast.Equal(uint32(0), atomic.LoadUint32(&rbft.activeView), "sendViewChange failed")

}

func TestValidatePending(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 0, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)

	cert1 := rbft.storeMgr.getCert(0, 1, "1")
	idx1 := vidx{view: 0, seqNo: 1}
	rbft.batchVdr.preparedCert[idx1] = "1"
	pp1 := &PrePrepare{
		View:             uint64(0),
		SequenceNumber:   uint64(1),
		BatchDigest:      "1",
		ResultHash:	  "1",
		HashBatch:	  &HashBatch{
			List:	  []string{},
			Timestamp:time.Now().UnixNano(),
		},
		ReplicaId:        uint64(0),
	}
	cert1.prePrepare = pp1

	currentVid := idx1.seqNo
	rbft.batchVdr.setCurrentVid(&currentVid)
	rbft.validatePending()

	rbft.batchVdr.setCurrentVid(nil)
	rbft.validatePending()
	ast.Equal(uint64(1), rbft.batchVdr.lastVid)
}

func TestHandleTransactionsAfterAbnormal(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	rbft.status.inActiveState(&rbft.status.inNegoView)
	rbft.status.inActiveState(&rbft.status.inRecovery)
	rbft.handleTransactionsAfterAbnormal()

	rbft.id = 1
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("testTx"),
		Id:              uint64(1),
		TransactionHash: []byte("3"),
	}
	rbft.processTransaction(txRequest{
		tx:	tx,
		new:	true,
	})
	ast.Equal(true, rbft.batchMgr.txPool.HasTxInPool(), "add transaction failed")
	rbft.handleTransactionsAfterAbnormal()
}




















