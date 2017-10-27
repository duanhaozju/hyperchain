//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"fmt"
	"testing"

	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/consensus/helper/persist"
	hdb "github.com/hyperchain/hyperchain/hyperdb/db"
	mdb "github.com/hyperchain/hyperchain/hyperdb/mdb"

	"github.com/stretchr/testify/assert"
)

func TestPersistAndRestoreAndDelPSet(t *testing.T) {

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
	cert1.prepare = make(map[Prepare]bool)
	cert1.prepare[Prepare{ResultHash: "c1p0"}] = true
	cert1.prepare[Prepare{ResultHash: "c1p1"}] = true
	cert1.prepare[Prepare{ResultHash: "c1p2"}] = true
	rbft.persistPSet(v1, seqNo1, digest1)

	v2 := uint64(2)
	seqNo2 := uint64(101)
	digest2 := "d2"
	cert2 := &msgCert{}
	idx2 := msgID{v2, seqNo2, digest2}
	rbft.storeMgr.certStore[idx2] = cert2
	cert2.prepare = make(map[Prepare]bool)
	cert2.prepare[Prepare{ResultHash: "c2p0"}] = true
	cert2.prepare[Prepare{ResultHash: "c2p1"}] = true
	cert2.prepare[Prepare{ResultHash: "c2p2"}] = true
	rbft.persistPSet(v2, seqNo2, digest2)

	PSet, err := rbft.restorePSet()
	ast.Nil(err, "restorePSet fail")
	prepares1, ok := PSet[msgID{v1, seqNo1, digest1}]
	ast.Equal(true, ok, "persistPSet fail")
	prepares2, ok := PSet[msgID{v2, seqNo2, digest2}]
	ast.Equal(true, ok, "persistPSet fail")

	expectedResultHash1 := make(map[string]bool)
	expectedResultHash2 := make(map[string]bool)
	receivedResultHash1 := make(map[string]bool)
	receivedResultHash2 := make(map[string]bool)
	for p := range cert1.prepare {
		expectedResultHash1[p.ResultHash] = true
	}
	for p := range cert2.prepare {
		expectedResultHash2[p.ResultHash] = true
	}
	for _, p := range prepares1.Set {
		receivedResultHash1[p.ResultHash] = true
	}
	for _, p := range prepares2.Set {
		receivedResultHash2[p.ResultHash] = true
	}
	ast.Equal(expectedResultHash1, receivedResultHash1, "restorePSet fail")
	ast.Equal(expectedResultHash2, receivedResultHash2, "restorePSet fail")

	rbft.persistDelPSet(v1, seqNo1, digest1)
	PSet, err = rbft.restorePSet()
	ast.Nil(err, "restorePSet fail")
	prepares1, ok = PSet[msgID{v1, seqNo1, digest1}]
	ast.Equal(false, ok, "persistDelPSet fail")
}

func TestPersistAndRestoreAndDelQSet(t *testing.T) {
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
	prePrepare1 := &PrePrepare{
		View:           v1,
		SequenceNumber: seqNo1,
		BatchDigest:    digest1,
	}
	rbft.persistQSet(prePrepare1)

	v2 := uint64(2)
	seqNo2 := uint64(101)
	digest2 := "d2"
	prePrepare2 := &PrePrepare{
		View:           v2,
		SequenceNumber: seqNo2,
		BatchDigest:    digest2,
	}
	rbft.persistQSet(prePrepare2)

	QSet, err := rbft.restoreQSet()
	prePre1, ok := QSet[msgID{v1, seqNo1, digest1}]
	ast.Equal(true, ok, "persistQSet fail")
	prePre2, ok := QSet[msgID{v2, seqNo2, digest2}]
	ast.Equal(true, ok, "persistQSet fail")

	ast.EqualValues(*prePrepare1, *prePre1)
	ast.EqualValues(*prePrepare2, *prePre2)

	rbft.persistDelQSet(v1, seqNo1, digest1)
	QSet, err = rbft.restoreQSet()
	prePre1, ok = QSet[msgID{v1, seqNo1, digest1}]
	ast.Equal(false, ok, "persistDelQSet fail")
}

func TestPersistAndRestoreAndDelCSet(t *testing.T) {

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
	cert1.commit = make(map[Commit]bool)
	cert1.commit[Commit{ResultHash: "c1c0"}] = true
	cert1.commit[Commit{ResultHash: "c1c1"}] = true
	cert1.commit[Commit{ResultHash: "c1c2"}] = true
	rbft.persistCSet(v1, seqNo1, digest1)

	v2 := uint64(2)
	seqNo2 := uint64(101)
	digest2 := "d2"
	cert2 := &msgCert{}
	idx2 := msgID{v2, seqNo2, digest2}
	rbft.storeMgr.certStore[idx2] = cert2
	cert2.commit = make(map[Commit]bool)
	cert2.commit[Commit{ResultHash: "c2c0"}] = true
	cert2.commit[Commit{ResultHash: "c2c1"}] = true
	cert2.commit[Commit{ResultHash: "c2c2"}] = true
	rbft.persistCSet(v2, seqNo2, digest2)

	CSet, err := rbft.restoreCSet()
	ast.Nil(err, "restoreCSet fail")
	commits1, ok := CSet[msgID{v1, seqNo1, digest1}]
	ast.Equal(true, ok, "persistCSet fail")
	commits2, ok := CSet[msgID{v2, seqNo2, digest2}]
	ast.Equal(true, ok, "persistCSet fail")

	expectedResultHash1 := make(map[string]bool)
	expectedResultHash2 := make(map[string]bool)
	receivedResultHash1 := make(map[string]bool)
	receivedResultHash2 := make(map[string]bool)
	for c := range cert1.commit {
		expectedResultHash1[c.ResultHash] = true
	}
	for c := range cert2.commit {
		expectedResultHash2[c.ResultHash] = true
	}
	for _, c := range commits1.Set {
		receivedResultHash1[c.ResultHash] = true
	}
	for _, c := range commits2.Set {
		receivedResultHash2[c.ResultHash] = true
	}
	ast.Equal(expectedResultHash1, receivedResultHash1, "restoreCSet failed")
	ast.Equal(expectedResultHash2, receivedResultHash2, "restoreCSet failed")

	rbft.persistDelCSet(v1, seqNo1, digest1)
	CSet, err = rbft.restoreCSet()
	ast.Nil(err, "restoreCSet fail")
	commits1, ok = CSet[msgID{v1, seqNo1, digest1}]
	ast.Equal(false, ok, "persistDelCSet fail")
}

func TestRestoreCert(t *testing.T) {
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
	rbft.persistQSet(prePrepare1)
	cert1.prepare = make(map[Prepare]bool)
	cert1.prepare[Prepare{ResultHash: "c1p0"}] = true
	cert1.prepare[Prepare{ResultHash: "c1p1"}] = true
	cert1.prepare[Prepare{ResultHash: "c1p2"}] = true
	rbft.persistPSet(v1, seqNo1, digest1)
	cert1.commit = make(map[Commit]bool)
	cert1.commit[Commit{ResultHash: "c1c0"}] = true
	cert1.commit[Commit{ResultHash: "c1c1"}] = true
	cert1.commit[Commit{ResultHash: "c1c2"}] = true
	rbft.persistCSet(v1, seqNo1, digest1)

	delete(rbft.storeMgr.certStore, idx1)
	_, ok := rbft.storeMgr.certStore[idx1]
	ast.Equal(false, ok, "delete failed")

	rbft.restoreCert()
	_, ok = rbft.storeMgr.certStore[idx1]
	ast.Equal(false, ok, "restoreCert check lastExec failed")

	rbft.storeMgr.certStore[idx1] = cert1
	rbft.persistQSet(prePrepare1)
	rbft.persistPSet(v1, seqNo1, digest1)
	rbft.persistCSet(v1, seqNo1, digest1)
	rbft.exec.setLastExec(uint64(200))
	delete(rbft.storeMgr.certStore, idx1)
	_, ok = rbft.storeMgr.certStore[idx1]
	ast.Equal(false, ok, "delete failed")
	rbft.restoreCert()
	cert2, ok := rbft.storeMgr.certStore[idx1]
	ast.Equal(true, ok, "restoreCert failed")
	ast.Equal(cert1.resultHash, cert2.resultHash, "restoreCert failed")

	expectedResultHash1 := make(map[string]bool)
	receivedResultHash1 := make(map[string]bool)
	for p := range cert1.prepare {
		expectedResultHash1[p.ResultHash] = true
	}
	for p := range cert2.prepare {
		receivedResultHash1[p.ResultHash] = true
	}
	ast.Equal(expectedResultHash1, receivedResultHash1, "restoreCert fail")

	delete(rbft.storeMgr.certStore, idx1)
	_, ok = rbft.storeMgr.certStore[idx1]
	ast.Equal(false, ok, "delete failed")
	rbft.persistDelQPCSet(v1, seqNo1, digest1)
	rbft.restoreCert()
	_, ok = rbft.storeMgr.certStore[idx1]
	ast.Equal(false, ok, "persistDelQPCSet failed")
}

func TestParseSpecifyCertStore(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)
	rbft.view = uint64(10)

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

	v2 := uint64(3)
	seqNo2 := uint64(100)
	digest2 := "d1"
	cert2 := &msgCert{}
	idx2 := msgID{v2, seqNo2, digest2}
	rbft.storeMgr.certStore[idx2] = cert2
	prePrepare2 := &PrePrepare{
		View:           v2,
		SequenceNumber: seqNo2,
		BatchDigest:    digest2,
		ResultHash:     "resultHash2",
	}
	cert2.resultHash = "resultHash2"
	cert2.prePrepare = prePrepare2
	cert2.prepare = make(map[Prepare]bool)
	cert2.prepare[Prepare{ResultHash: "c2p0"}] = true
	cert2.prepare[Prepare{ResultHash: "c2p1"}] = true
	cert2.prepare[Prepare{ResultHash: "c2p2"}] = true
	cert2.commit = make(map[Commit]bool)
	cert2.commit[Commit{ResultHash: "c2c0"}] = true
	cert2.commit[Commit{ResultHash: "c2c1"}] = true
	cert2.commit[Commit{ResultHash: "c2c2"}] = true

	rbft.parseSpecifyCertStore()
	ast.Equal(1, len(rbft.storeMgr.certStore), "parseSpecifyCertStore failed")
	ast.Equal(uint64(10), rbft.storeMgr.certStore[msgID{uint64(10), seqNo1, digest1}].prePrepare.View, "parseSpecifyCertStore failed")
}

func TestTxBatchPersist(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	txBatch1 := &TransactionBatch{
		SeqNo:      uint64(10),
		ResultHash: "10",
	}
	txBatch2 := &TransactionBatch{
		SeqNo:      uint64(11),
		ResultHash: "11",
	}
	rbft.storeMgr.txBatchStore[txBatch1.ResultHash] = txBatch1
	rbft.storeMgr.txBatchStore[txBatch2.ResultHash] = txBatch2

	rbft.persistTxBatch(txBatch1.ResultHash)
	rbft.persistTxBatch(txBatch2.ResultHash)

	delete(rbft.storeMgr.txBatchStore, txBatch1.ResultHash)
	delete(rbft.storeMgr.txBatchStore, txBatch2.ResultHash)
	ast.Equal(0, len(rbft.storeMgr.txBatchStore), "delete txBatchStore failed")

	rbft.restoreTxBatchStore()
	ast.Equal(2, len(rbft.storeMgr.txBatchStore), "restoreTxBatchStore failed")

	delete(rbft.storeMgr.txBatchStore, txBatch1.ResultHash)
	delete(rbft.storeMgr.txBatchStore, txBatch2.ResultHash)
	ast.Equal(0, len(rbft.storeMgr.txBatchStore), "delete txBatchStore failed")
	rbft.persistDelTxBatch(txBatch1.ResultHash)
	rbft.restoreTxBatchStore()
	ast.Equal(1, len(rbft.storeMgr.txBatchStore), "persistDelTxBatch failed")

	delete(rbft.storeMgr.txBatchStore, txBatch2.ResultHash)
	ast.Equal(0, len(rbft.storeMgr.txBatchStore), "delete txBatchStore failed")
	rbft.persistDelAllTxBatches()
	rbft.restoreTxBatchStore()
	ast.Equal(0, len(rbft.storeMgr.txBatchStore), "persistDelTxBatch failed")
}

func TestLocalKeyPersist(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	hash := []byte("localkey00000000001")
	rbft.persistLocalKey(hash)

	key := fmt.Sprint("localkey")
	rs, err := rbft.persister.ReadState(key)
	ast.Equal(hash, rs, fmt.Sprintf("error persistLocalKey(%v) not success", hash))

	rbft.persistDellLocalKey()
	rs, err = rbft.persister.ReadState(key)
	ast.Equal(hdb.DB_NOT_FOUND, err, "error persistDellLocalKey() not success")
}

func TestViewPersist(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)
	rbft.view = uint64(10)

	view := uint64(128)
	rbft.persistView(view)
	rbft.restoreView()
	ast.Equal(view, rbft.view)

	rbft.persistDelView()
	_, err = rbft.persister.ReadState(fmt.Sprint("view"))
	ast.Equal(hdb.DB_NOT_FOUND, err, "error persistDelView() not success")
}

func TestCheckpointPersist(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	seqNo := uint64(1)
	id := []byte("checkpoint00000000001")
	rbft.persistCheckpoint(seqNo, id)

	key := fmt.Sprintf("chkpt.%d", seqNo)
	rs, err := rbft.persister.ReadState(key)
	ast.Equal(id, rs, fmt.Sprintf("error persistCheckpoint(%v, %v) not success", seqNo, id))

	rbft.persistDelCheckpoint(seqNo)
	rs, err = rbft.persister.ReadState(key)
	ast.Equal(hdb.DB_NOT_FOUND, err, fmt.Sprintf("error persistDelCheckpoint(%q) not success", key))
}

func TestRestoreState(t *testing.T) {
	ast := assert.New(t)
	rbft, _, err := TNewRbft("./Testdatabase/", "../../configuration/namespaces/", "global", 2, t)
	defer CleanData(rbft.namespace)
	ast.Equal(nil, err, err)
	rbft.Start()
	db, err := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	rbft.persister = persist.New(db)

	rbft.K = 1
	rbft.N = 3

	rbft.persistView(uint64(10))
	rbft.restoreState()
	ast.Equal(uint64(10), rbft.view, "restoreState: restore view failed")

	checkPointSeq := uint64(5)
	checkPointId := []byte("checkpoint00000000001")
	rbft.persistCheckpoint(checkPointSeq, checkPointId)
	rbft.restoreState()
	_, ok := rbft.storeMgr.chkpts[checkPointSeq]
	ast.Equal(true, ok, "restoreState: restore checkpoint failed")

	rbft.persistN(10)
	rbft.restoreState()
	ast.Equal(10, rbft.N, "restoreState: restore nodes failed")

	rbft.persistNewNode(1)
	rbft.restoreState()
	ast.Equal(true, rbft.in(isNewNode), "restoreState: restore newNode failed")

	localKeyHash := []byte("localkey00000000001")
	localKey := string(localKeyHash)
	rbft.persistLocalKey(localKeyHash)
	rbft.restoreState()
	ast.Equal(localKey, rbft.nodeMgr.localKey, "restoreState: restore localKey failed")

}
