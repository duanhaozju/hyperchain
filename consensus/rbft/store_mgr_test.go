//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"testing"

	"hyperchain/common"
	"hyperchain/consensus/helper/persist"
	mdb "hyperchain/hyperdb/mdb"

	"github.com/stretchr/testify/assert"
)

func TestNewStoreMgr(t *testing.T) {
	storeMgr := newStoreMgr()
	structName, nilElems, err := checkNilElems(storeMgr)
	if err != nil {
		t.Error(err.Error())
	}
	if nilElems != nil {
		t.Errorf("There exists some nil elements: %v in struct: %s", nilElems, structName)
	}
}

func TestMoveWatermarks(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	rbft.storeMgr.chkpts[uint64(9)] = "9"
	rbft.storeMgr.chkpts[uint64(10)] = "10"
	rbft.storeMgr.chkpts[uint64(11)] = "11"

	rbft.vcMgr.qlist[qidx{n: uint64(9)}] = nil
	rbft.vcMgr.qlist[qidx{n: uint64(10)}] = nil
	rbft.vcMgr.qlist[qidx{n: uint64(11)}] = nil

	rbft.vcMgr.plist[uint64(9)] = nil
	rbft.vcMgr.plist[uint64(10)] = nil
	rbft.vcMgr.plist[uint64(11)] = nil

	rbft.storeMgr.moveWatermarks(rbft, uint64(10))
	ast.Equal(2, len(rbft.storeMgr.chkpts), "moveWatermarks delete checkpoints failed")
	ast.Equal(1, len(rbft.vcMgr.qlist), "moveWatermarks delete qset failed")
	ast.Equal(1, len(rbft.vcMgr.plist), "moveWatermarks delete pset failed")
}

func TestGetCert(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	v := uint64(10)
	n := uint64(100)
	d := "digest"
	cert1 := rbft.storeMgr.getCert(v, n, d)
	cert2, ok := rbft.storeMgr.certStore[msgID{v, n, d}]
	ast.Equal(true, ok, "getCert failed")
	ast.Equal(cert1, cert2, "getCert failed")

	v = uint64(11)
	cert := &msgCert{
		resultHash: "cert",
	}
	rbft.storeMgr.certStore[msgID{v, n, d}] = cert
	cert3 := rbft.storeMgr.getCert(v, n, d)
	ast.Equal(cert, cert3, "getCert failed")
}

func TestGetChkptCert(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	n := uint64(100)
	id := "digest"
	cert1 := rbft.storeMgr.getChkptCert(n, id)
	cert2, ok := rbft.storeMgr.chkptCertStore[chkptID{n, id}]
	ast.Equal(true, ok, "getChkptCert failed")
	ast.Equal(cert1, cert2, "getChkptCert failed")

	n = uint64(11)
	cert := &chkptCert{
		chkptCount: 10,
	}
	rbft.storeMgr.chkptCertStore[chkptID{n, id}] = cert
	cert3 := rbft.storeMgr.getChkptCert(n, id)
	ast.Equal(cert, cert3, "getChkptCert failed")
}

func TestExistedDigest(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

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
		ResultHash:     "resultHash",
	}
	cert1.resultHash = "resultHash"
	cert1.prePrepare = prePrepare1

	ast.Equal(true, rbft.storeMgr.existedDigest(uint64(9), v1, digest1))
	ast.Equal(false, rbft.storeMgr.existedDigest(uint64(9), uint64(10), digest1))
}
