//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"

	"github.com/stretchr/testify/assert"
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

func TestBatchManager(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 1, t)
	ast.Equal(nil, err, err)
	ast.Equal(false, rbft.batchMgr.batchTimerActive, "batchTimer initialize failed")
}