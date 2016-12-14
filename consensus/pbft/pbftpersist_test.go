//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"testing"
	"hyperchain/consensus/helper/persist"
	"reflect"

	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"github.com/syndtr/goleveldb/leveldb"
	"fmt"
	"encoding/binary"
	//"github.com/pkg/errors"
	"hyperchain/core"

)

func TestPersistPSet(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)

	var pset []*ViewChange_PQ
	rsraw, _ := proto.Marshal(&PQset{pset})

	pbft.persistPSet()
	rs, err := persist.ReadState("pset")
	if err != nil || !reflect.DeepEqual(rsraw, rs) {
		t.Errorf("error persistPSet() = %v, %v, actual: null, null", rs, err)
	}

	pbft.pset = make(map[uint64]*ViewChange_PQ)
	vcpq := ViewChange_PQ{
		1,
		"digest",
		2,
	}
	pbft.pset[1234] = &vcpq

	pset = append(pset, &vcpq)

	pbft.persistPSet()
	rs, err = persist.ReadState("pset")
	rsraw, _ = proto.Marshal(&PQset{pset})
	if err != nil || !reflect.DeepEqual(rs, rsraw) {
		t.Errorf("error persistPSet() = %v, %v, actual: %v, %v", rs, err, pset, nil)
	}
}

func TestPersistQSet(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)
	var qset []*ViewChange_PQ

	rsraw, _ := proto.Marshal(&PQset{qset})

	pbft.persistQSet()
	rs, err := persist.ReadState("qset")
	if err != nil || !reflect.DeepEqual(rsraw, rs) {
		t.Errorf("error persistQSet() = %v, %v, actual: null, null", rs, err)
	}

	pbft.qset = make(map[qidx]*ViewChange_PQ)
	vcpq := ViewChange_PQ{
		1,
		"digest",
		2,
	}
	pbft.qset[qidx{"112", uint64(112)}] = &vcpq
	qset = append(qset, &vcpq)

	pbft.persistQSet()
	rs, err = persist.ReadState("qset")
	rsraw, _ = proto.Marshal(&PQset{qset})
	if err != nil || !reflect.DeepEqual(rs, rsraw) {
		t.Errorf("error persistQSet() = %v, %v, actual: %v, %v", rs, err, qset, nil)
	}

}

func TestRestorePQSet(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)
	var pset []*ViewChange_PQ
	var qset []*ViewChange_PQ
	pbft.pset = make(map[uint64]*ViewChange_PQ)
	p := ViewChange_PQ{
		1,
		"digest",
		2,
	}
	pbft.pset[1234] = &p
	pset = append(pset, &p)

	pbft.qset = make(map[qidx]*ViewChange_PQ)
	q := ViewChange_PQ{
		1,
		"digest",
		2,
	}
	pbft.qset[qidx{"112", uint64(112)}] = &q
	qset = append(qset, &q)

	rs := pbft.restorePQSet("pset")
	if reflect.DeepEqual(rs, pset) {
		t.Errorf("error restorePQSet(%q) = %v, actual: %v", "pset", rs, pset)
	}

	rs = pbft.restorePQSet("qset")

	if reflect.DeepEqual(rs, qset) {
		t.Errorf("error restorePQSet(%q) = %v, actual: %v", "qset", rs, qset)
	}

	pbft.persistPSet()
	pbft.persistQSet()

	rs = pbft.restorePQSet("pset")
	if !reflect.DeepEqual(rs, pset) {
		t.Errorf("error restorePQSet(%q) = %v, actual: %v", "pset", rs, pset)
	}

	rs = pbft.restorePQSet("qset")

	if !reflect.DeepEqual(rs, qset) {
		t.Errorf("error restorePQSet(%q) = %v, actual: %v", "qset", rs, qset)
	}
}

func TestBatchRelatedPersistFunctions(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)
	pbft.validatedBatchStore = make(map[string]*TransactionBatch)
	pbft.validatedBatchStore["t1"] = &TransactionBatch{Timestamp:111}
	pbft.validatedBatchStore["t2"] =
		&TransactionBatch{
			Timestamp:int64(222),
			Batch:[]*types.Transaction{
				{
					From:[]byte("A"),
					To:[]byte("B"),
					Value:[]byte("123"),
				},
				{
					From:[]byte("B}"),
					To:[]byte("A"),
					Value:[]byte("321"),
				},
			},
		}
	pbft.persistRequestBatch("t1")
	pbft.persistRequestBatch("t2")

	rs, err := persist.ReadState("reqBatch." + "t1")
	raw, _ := proto.Marshal(pbft.validatedBatchStore["t1"])
	if err != nil || !reflect.DeepEqual(rs, raw) {
		t.Errorf("persistRequestBatch(%q) = %v, actual: %v", "t1", rs, raw)
	}

	rs, err = persist.ReadState("reqBatch." + "t2")
	raw, _ = proto.Marshal(pbft.validatedBatchStore["t2"])
	if err != nil || !reflect.DeepEqual(rs, raw) {
		t.Errorf("persistRequestBatch(%q) = %v, actual: %v", "t2", rs, raw)
	}

	pbft.persistRequestBatch("t2XXXXXX")
	rs, err = persist.ReadState("reqBatch." + "t1")
	raw, _ = proto.Marshal(pbft.validatedBatchStore["t1"])
	if err != nil || !reflect.DeepEqual(rs, raw) {
		t.Errorf("persistRequestBatch(%q) = %v, actual: %v", "t1", rs, raw)
	}

	pbft.persistDelRequestBatch("t1")

	rs, err = persist.ReadState("reqBatch." + "t1")
	if err != leveldb.ErrNotFound {
		t.Errorf(`error persistDelRequestBatch(%q)`, "t1")
	}

	pbft.persistDelAllRequestBatches()
	rs, err = persist.ReadState("reqBatch." + "t2")
	if err != leveldb.ErrNotFound {
		t.Error(`error persistDelAllRequestBatches, not clear all baches`)
	}
}

func TestCheckpointPersist(t *testing.T)  {
	pbft := new(pbftProtocal)
	seqNo := uint64(1)
	id := []byte("checkpoint00000000001")
	pbft.persistCheckpoint(seqNo, id)

	key := fmt.Sprintf("chkpt.%d", seqNo)
	rs, err := persist.ReadState(key)
	if err != nil || !reflect.DeepEqual(rs, id) {
		t.Errorf(`error persistCheckpoint(%v, %v) not success`, seqNo, id)
	}

	pbft.persistDelCheckpoint(seqNo)
	rs, err = persist.ReadState(key)
	if err != leveldb.ErrNotFound {
		t.Errorf("error persistDelCheckpoint(%q) not success", key)
	}
}


func TestViewPersist(t *testing.T)  {
	pbft := new(pbftProtocal)
	view := uint64(128)
	key := fmt.Sprint("view")
	b := make([]byte, 8)
	binary.LittleEndian.PutUint64(b, view)
	pbft.persistView(view)
	rs, err := persist.ReadState(key)
	if err == leveldb.ErrNotFound || !reflect.DeepEqual(rs, b) {
		t.Errorf(`error persistView(%q) not success`, key)
	}

	pbft.persistDelView()
	rs, err = persist.ReadState(key)
	if err != leveldb.ErrNotFound {
		t.Error(`error persistDelView() not success`)
	}
}

func TestSeqnoFunctions(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)

	core.InitDB("/temp/leveldb", 8088)
	lseqno, error := pbft.getLastSeqNo()

	if lseqno != 0 || error == nil {
		t.Errorf(`error getLastSeqNo() = (%v, %v), actual: %v, %v`, lseqno, error, 0,  "Height of chain is 0")
	}

	clearDB()

	core.UpdateChain(&types.Block{
		Timestamp:12,
		Number:1222,
	}, false)

	lseqno, error = pbft.getLastSeqNo()
	if error != nil || lseqno != 1222 {
		t.Errorf(`error getLastSeqNo() = (%v, %v), actual: %v, %v`, lseqno, error, 1222,  nil)
	}
}

func TestRestoreLastSeqNo(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)
	core.InitDB("/temp/leveldb", 8088)
	pbft.restoreLastSeqNo()
}

func TestRestoreState(t *testing.T)  {
	defer clearDB()
	pbft := new(pbftProtocal)
	pbft.K = 1
	pbft.N = 3
	core.InitDB("./temp/leveldb", 8088)

	var pset []*ViewChange_PQ
	pbft.pset = make(map[uint64]*ViewChange_PQ)
	vcpq := ViewChange_PQ{
		1,
		"digest",
		2,
	}
	pbft.pset[1234] = &vcpq
	pset = append(pset, &vcpq)
	pbft.persistPSet()

	pbft.restoreState()
	if pbft.view != 2 || pbft.seqNo != 1 {
		t.Error("restoreState.updateSeqView error")
	}

	var qset []*ViewChange_PQ

	pbft.qset = make(map[qidx]*ViewChange_PQ)
	vcpq = ViewChange_PQ{
		4,
		"digest",
		6,
	}
	pbft.qset[qidx{"112", uint64(112)}] = &vcpq
	qset = append(qset, &vcpq)

	pbft.persistQSet()
	pbft.restoreState()
	if pbft.view != 6 || pbft.seqNo != 4 {
		t.Error("restoreState.updateSeqView error")
	}

	pbft.validatedBatchStore = make(map[string]*TransactionBatch)
	pbft.validatedBatchStore["t1"] = &TransactionBatch{Timestamp:111}
	pbft.validatedBatchStore["t2"] =
		&TransactionBatch{
			Timestamp:int64(222),
			Batch:[]*types.Transaction{
				{
					From:[]byte("A"),
					To:[]byte("B"),
					Value:[]byte("123"),
				},
				{
					From:[]byte("B}"),
					To:[]byte("A"),
					Value:[]byte("321"),
				},
			},
		}
	pbft.persistRequestBatch("t1")
	pbft.persistRequestBatch("t2")

	pbft.restoreState()
	rs := pbft.validatedBatchStore[hash(pbft.validatedBatchStore["t1"])]

	if !reflect.DeepEqual(pbft.validatedBatchStore["t1"], rs){
		t.Error("error restoreState reqBatchesPacked")
	}

	view := Int64ToBytes(1234)
	persist.StoreState("view", view)
	pbft.restoreState()

	if !(pbft.view == binary.LittleEndian.Uint64(view)){
		t.Error("error restoreState restore view")
	}

	seqNo := uint64(0)
	pbft.chkpts = make(map[uint64]string)
	id := []byte("checkpoint00000000001")
	pbft.persistCheckpoint(seqNo, id)
	pbft.restoreState()

	if seqNo != pbft.h {
		t.Error("error restoreState moveWatermarks")
	}

}

func Int64ToBytes(i int64) []byte {
	var buf = make([]byte, 8)
	binary.BigEndian.PutUint64(buf, uint64(i))
	return buf
}

func clearDB()  {
	persist.DelAllState()
}