//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package txpool

import (
	"hyperchain/core/types"
	"hyperchain/manager/event"
	"testing"
	"time"
)

func batchLoop(batchSub event.Subscription) {

	for obj := range batchSub.Chan() {
		switch ev := obj.Data.(type) {
		case TxHashBatch:
			logger.Info(ev)
		}
	}

}

func TestNewTxPoolImpl(t *testing.T) {
	eventMux := new(event.TypeMux)
	//batchSub := eventMux.Subscribe(TxHashBatch{})
	//recvFromChan(batchSub.Chan())
	if txPool, err := newTxPoolImpl(400, eventMux, 10*time.Second, 40); err != nil {
		t.Error("new txPool fail")
	} else {
		if txPool.batchSize != 40 || txPool.batchTimerActive != false {
			t.Error("new txPool wrong")
		}
	}

}

func TestPrimaryAddNewTx(t *testing.T) {
	eventMux := new(event.TypeMux)
	batchSub := eventMux.Subscribe(TxHashBatch{})
	txPool, err := newTxPoolImpl(400, eventMux, 10*time.Second, 40)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx1 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test1"),
		Id:              uint64(1),
		TransactionHash: []byte("hash1"),
	}
	err = txPool.primaryAddNewTx(tx1, false)
	if err != nil {
		t.Error("primary add transaction fail")
	}
	if len(txPool.txPoolHash) != 1 || len(txPool.txPool) != 1 {
		t.Error("Only add one transaction to txPool")
	}
	if !txPool.batchTimerActive {
		t.Error("We active the batch timer, expect true")
	}
	txPool.batchSize = 2
	go batchLoop(batchSub)
	err = txPool.primaryAddNewTx(tx1, false)
	if err == nil {
		t.Error("should return a duplicate transaction error")
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	err = txPool.primaryAddNewTx(tx2, false)
	if err != nil {
		t.Error("primary add transaction fail")
	}
	if len(txPool.txPoolHash) != 0 || len(txPool.txPool) != 0 || len(txPool.batchStore) != 1 {
		t.Error("should generate a batch, but fail")
	}
	if txPool.batchTimerActive == true || txPool.batchTimer != nil {
		t.Error("timer should be stopped")
	}
	time.Sleep(time.Second)
}

func TestReplicaAddNewTx(t *testing.T) {
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(1, eventMux, 10*time.Second, 40)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	err = txPool.replicaAddNewTx(tx, true)
	if err != nil {
		t.Error("replica add transaction fail")
	}
	if len(txPool.txPoolHash) != 1 || len(txPool.txPool) != 1 {
		t.Error("Only add one transaction to txPool")
	}
	err = txPool.replicaAddNewTx(tx, true)
	if err != ErrPoolFull {
		t.Error("pool should be full and return a full error")
	}
	txPool.poolSize = 2
	err = txPool.replicaAddNewTx(tx, true)
	if err != ErrDuplicateTx {
		t.Error("pool get a existed transaction")
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	err = txPool.replicaAddNewTx(tx2, true)
	if err != nil {
		t.Error("should add a new transaction, but fail now")
	}
	if len(txPool.txPool) != 2 || len(txPool.txPoolHash) != 2 {
		t.Error("should store 2 transaction, expect 2")
	}
}

func TestAddNewTx(t *testing.T) {
	eventMux := new(event.TypeMux)
	txPool, err := newTxPoolImpl(10, eventMux, 10*time.Second, 40)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	err = txPool.AddNewTx(tx, false, true)
	if err != nil {
		t.Error("add transaction false")
	}
	if txPool.batchTimerActive == true {
		t.Error("repilca can't active the batch timer")
	}
	if len(txPool.txPool) != 1 || len(txPool.txPoolHash) != 1 {
		t.Error("should store 1 transaction")
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	err = txPool.AddNewTx(tx2, true, false)
	if err != nil {
		t.Error("add transaction false")
	}
	if txPool.batchTimerActive == false {
		t.Error("primary should active the batch timer")
	}
	if len(txPool.txPool) != 2 || len(txPool.txPoolHash) != 2 {
		t.Error("should store 2 transaction")
	}
}

func TestAddTxs(t *testing.T) {
	txPool, err := newTxPoolImpl(10, new(event.TypeMux), 10*time.Second, 40)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	txs := []*types.Transaction{tx, tx}
	err = txPool.addTxs(txs)
	if err != nil {
		t.Error("add transaction false")
	}
	if len(txPool.txPoolHash) != 1 || len(txPool.txPool) != 1 {
		t.Error("add dupilcate transactions, expect 1")
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	txs = []*types.Transaction{tx, tx2}
	err = txPool.addTxs(txs)
	if err != nil {
		t.Error("add transaction false")
	}
	if len(txPool.txPoolHash) != 2 || len(txPool.txPool) != 2 {
		t.Error("add dupilcate transactions, expect 1")
	}

}

func TestBatchTimer(t *testing.T) {
	txPool, err := newTxPoolImpl(10, new(event.TypeMux), 2*time.Second, 3)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	if err := txPool.primaryAddNewTx(tx, false); err != nil {
		t.Error("add transaction false")
	}
	if txPool.batchTimerActive != true {
		t.Error("the batch timer should be active, expect true")
	}

	txPool.StopBatch()
	time.Sleep(3 * time.Second)

}

func TestReturnFetchTxs(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}})
	_, err = txPool.ReturnFetchTxs("1", []string{"1"})
	if err == nil {
		t.Error("should not find this batch")
	}
	txBatch := &TxHashBatch{
		TxHashList: []string{"1", "2", "3", "4"},
		TxList: []*types.Transaction{
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test1"),
				Id:              uint64(1),
				TransactionHash: []byte("hash1"),
			},
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test2"),
				Id:              uint64(1),
				TransactionHash: []byte("hash2"),
			},
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test3"),
				Id:              uint64(1),
				TransactionHash: []byte("hash3"),
			},
			&types.Transaction{
				From:            []byte{1},
				To:              []byte{2},
				Value:           []byte{1},
				Timestamp:       time.Now().UnixNano(),
				Signature:       []byte("test4"),
				Id:              uint64(1),
				TransactionHash: []byte("hash4"),
			},
		},
	}
	txBatch.BatchHash = hash(txBatch)
	txPool.batchStore = append(txPool.batchStore, txBatch)
	txBatches, err := txPool.ReturnFetchTxs(txBatch.BatchHash, []string{"5"})
	if err == nil {
		t.Error("should not find this transanction")
	}
	txBatches, err = txPool.ReturnFetchTxs(txBatch.BatchHash, []string{"2", "4"})
	if err != nil {
		t.Error("should find these transanctions")
	}
	if len(txBatches) != 2 {
		t.Error("should get two transanctions")
	}
	if string(txBatches[0].TransactionHash) != "hash2" || string(txBatches[1].TransactionHash) != "hash4" {
		t.Error("should find transanctions whose hash values are 'hash2' and 'hash4'")
	}
}

func TestGotMissingTxs(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}

	txs := []*types.Transaction{
		{
			From:            []byte{1},
			To:              []byte{2},
			Value:           []byte{1},
			Timestamp:       time.Now().UnixNano(),
			Signature:       []byte("test2"),
			Id:              uint64(1),
			TransactionHash: []byte("hash2"),
		},
		{
			From:            []byte{1},
			To:              []byte{2},
			Value:           []byte{1},
			Timestamp:       time.Now().UnixNano(),
			Signature:       []byte("test4"),
			Id:              uint64(1),
			TransactionHash: []byte("hash4"),
		},
	}
	txPool.missingTxs["2"] = []string{txs[0].GetHash().Hex(), txs[1].GetHash().Hex()}
	if _, err = txPool.GotMissingTxs("2", txs); err != ErrNoCachedBatch {
		t.Error("should not find this id in cachedHashList")
	}
	txPool.cachedHashList["2"] = txPool.missingTxs["2"]
	hashList, err := txPool.GotMissingTxs("2", txs)
	if err != nil {
		t.Error(err)
	}
	if hashList[0] != txs[0].GetHash().Hex() {
		t.Error("should get the same hash")
	}
}

func TestGetTxsByHashList(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}})
	id := hash(txPool.batchStore[0])
	txs, missingHash, err := txPool.GetTxsByHashList(id, []string{})
	if err != ErrDuplicateBatch {
		t.Error("This batch already exists")
	}
	txPool.txPool["1"] = &types.Transaction{}
	txPool.txPool["2"] = &types.Transaction{}
	txPool.txPool["3"] = &types.Transaction{}
	txPool.txPool["4"] = &types.Transaction{}
	txPool.txPoolHash = []string{"1", "2", "3", "4"}
	txs, missingHash, err = txPool.GetTxsByHashList("2", []string{"2", "5", "6"})
	if err != ErrMissing || txs != nil || len(missingHash) != 2 {
		t.Error("should return missing error and miss two transactions in total")
	}
	txs, missingHash, err = txPool.GetTxsByHashList("2", []string{"2", "3", "4"})
	if err != nil {
		t.Error("should generate a new batch")
	}
	if txPool.batchStore[1].BatchHash != "2" || len(txPool.batchStore[1].TxList) != 3 {
		t.Error("The new batch hash should be '2' and it's length is 3")
	}
	if txPool.batchedTxs["2"] != true || txPool.batchedTxs["3"] != true || txPool.batchedTxs["4"] != true {
		t.Error("should store batched transactions in batchedTxs")
	}
}

func TestRemoveBatchedTxs(t *testing.T) {
	txPool, err := newTxPoolImpl(1, new(event.TypeMux), 10*time.Second, 40)
	if err != nil {
		t.Error("new txPool fail")
	}
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "2"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "3"})
	hashList := []string{"1", "2"}
	err = txPool.RemoveBatchedTxs(hashList)
	if err != nil {
		t.Error("remove batched transactions failed")
	}
	if len(txPool.batchStore) != 1 {
		t.Error("remove two batch, expect 1")
	}
	hashList = []string{"2", "3"}
	err = txPool.RemoveBatchedTxs(hashList)
	if err != nil {
		t.Error("remove batched transactions failed")
	}
	if len(txPool.batchStore) != 0 {
		t.Error("remove one batch, expect 1")
	}
}

func TestRemoveOneBatchedTxs(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 40)
	if err != nil {
		t.Error("new txPool fail")
	}
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "2"})
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "3"})
	if err = txPool.RemoveOneBatchedTxs("4"); err != ErrNoTxHash {
		t.Error("should not find the batch whose hash is '4'")
	}
	err = txPool.RemoveOneBatchedTxs("2")
	if err != nil || txPool.batchStore[1].BatchHash != "3" {
		t.Error("the batch should be removed")
	}

}

func TestGetTxsBack(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx1 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test1"),
		Id:              uint64(1),
		TransactionHash: []byte("hash1"),
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	batch1 := &TxHashBatch{
		TxList:     []*types.Transaction{tx1},
		TxHashList: []string{tx1.GetHash().Hex()},
	}
	batch2 := &TxHashBatch{
		TxList:     []*types.Transaction{tx2},
		TxHashList: []string{tx2.GetHash().Hex()},
	}
	batch1.BatchHash = hash(batch1)
	batch2.BatchHash = hash(batch2)

	txPool.batchStore = append(txPool.batchStore, batch1)
	txPool.batchedTxs[tx1.GetHash().Hex()] = true
	txPool.batchStore = append(txPool.batchStore, batch2)
	txPool.batchedTxs[tx2.GetHash().Hex()] = true
	err = txPool.GetTxsBack([]string{batch1.BatchHash, "1"})
	if err == nil || len(txPool.batchStore) != 2 {
		t.Error("should not move because there is no batch whose hash is '1'")
	}
	err = txPool.GetTxsBack([]string{batch1.BatchHash, batch2.BatchHash})
	if err != nil || txPool.batchStore != nil {
		t.Error("should move all batched batches")
	}
	if txPool.txPoolHash[0] != tx1.GetHash().Hex() {
		t.Error("The first transaction in batch should be the first transaction in txpool")
	}
}

func TestGetAllTxsLength(t *testing.T) {
	eventMux := new(event.TypeMux)
	batchSub := eventMux.Subscribe(TxHashBatch{})
	txPool, err := newTxPoolImpl(3, eventMux, 10*time.Second, 2)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	txPool.AddNewTx(tx, true, false)
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{
		BatchHash:  "1",
		TxList:     []*types.Transaction{tx},
		TxHashList: []string{tx.GetHash().Hex()},
	})
	if txPool.IsPoolFull() != 2 {
		t.Error("store two transactions in total, expect 2")
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	go batchLoop(batchSub)
	txPool.AddNewTx(tx2, true, false)
	if txPool.IsPoolFull() != 3 {
		t.Error("store three transactions in total, expect 3")
	}
}

func TestGenerateTxBatch(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}
	err = txPool.generateTxBatch()
	if err != ErrEmptyFull {
		t.Error("pool is empty, expect ErrEmptyFull")
	}
}

func TestNewTxBatch(t *testing.T) {
	eventMux := new(event.TypeMux)
	batchSub := eventMux.Subscribe(TxHashBatch{})
	txPool, err := newTxPoolImpl(3, eventMux, 10*time.Second, 1)
	if err != nil {
		t.Error("new txPool fail")
	}
	tx := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test"),
		Id:              uint64(1),
		TransactionHash: []byte("hash"),
	}
	tx2 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test2"),
		Id:              uint64(1),
		TransactionHash: []byte("hash2"),
	}
	tx3 := &types.Transaction{
		From:            []byte{1},
		To:              []byte{2},
		Value:           []byte{1},
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("test3"),
		Id:              uint64(1),
		TransactionHash: []byte("hash3"),
	}
	go batchLoop(batchSub)
	txPool.AddNewTx(tx, true, false)
	if len(txPool.txPool) != 0 || len(txPool.txPoolHash) != 0 || len(txPool.batchStore) != 1 {
		t.Error("should generate one batch")
	}
	txPool.addTxs([]*types.Transaction{tx2, tx3})
	txPool.newTxBatch()
	if len(txPool.txPool) != 1 || len(txPool.txPoolHash) != 1 || len(txPool.batchStore) != 2 {
		t.Error("should generate one batch again")
	}
}

func TestRemoveTxPoolTxs(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}
	txPool.txPool["1"] = &types.Transaction{}
	txPool.txPool["2"] = &types.Transaction{}
	txPool.txPool["3"] = &types.Transaction{}
	txPool.txPool["4"] = &types.Transaction{}
	txPool.txPoolHash = []string{"1", "2", "3", "4"}
	txPool.removeTxPoolTxs([]string{"2", "4"})
	if len(txPool.txPool) != 2 || len(txPool.txPoolHash) != 2 {
		t.Error("should remove two transactions")
	}
	if txPool.txPoolHash[0] != "1" || txPool.txPoolHash[1] != "3" {
		t.Error("should be ordered")
	}
}

func TestGetBatchById(t *testing.T) {
	txPool, err := newTxPoolImpl(100, new(event.TypeMux), 10*time.Second, 10)
	if err != nil {
		t.Error("new txPool fail")
	}
	txPool.batchStore = append(txPool.batchStore, &TxHashBatch{BatchHash: "1", TxHashList: []string{"1", "2"}})
	h := hash(txPool.batchStore[0])
	batch, err := txPool.getBatchById(h)
	if err != nil {
		t.Error(err)
	}
	if batch.BatchHash != "1" {
		t.Error("Get a wrong batch")
	}
	batch, err = txPool.getBatchById("a")
	if err == nil {
		t.Error("should not get a batch")
	}
}
