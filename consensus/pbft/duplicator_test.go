//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

import (
	"testing"
	"time"
	"container/list"

	"hyperchain/core/types"
)

func TestTransactionStore_Len(t *testing.T) {
	txs := &transactionStore{}
	txs.empty()

	r1 := createTx(1, "r1")
	r2 := createTx(2, "r2")
	if txs.has(txs.wrapTransaction(r1).key) {
		t.Error("should not have tx")
	}
	txs.add(r1)
	if !txs.has(txs.wrapTransaction(r1).key) {
		t.Error("should have tx")
	}
	if txs.has(txs.wrapTransaction(r2).key) {
		t.Error("should not have tx")
	}
	if txs.remove(r2) {
		t.Error("should not have removed tx")
	}
	if !txs.remove(r1) {
		t.Error("should have removed tx")
	}
	if txs.remove(r1) {
		t.Error("should not have removed tx")
	}
	if txs.order.Len() != 0 || len(txs.presence) != 0 {
		t.Error("should have 0 len")
	}
}

func createTx(replica uint64, payload string) (tx *types.Transaction) {

	tx = &types.Transaction{
		Timestamp:	time.Now().UnixNano(),
		Id:		replica,
		TransactionHash: []byte(payload),
	}

	return
}

func BenchmarkTransactionStore(b *testing.B) {
	txs := &transactionStore{}
	txs.empty()

	Ntx := 1000

	ts := make(map[string]*types.Transaction)
	for i := 0; i < Ntx; i++ {
		rc := txs.wrapTransaction(createTx(uint64(i), "test"))
		ts[rc.key] = rc.tx
	}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		for _, r := range ts {
			txs.add(r)
		}

		for k := range ts {
			_ = txs.has(k)
		}

		for _, r := range ts {
			txs.remove(r)
		}
	}
}

func TestTransactionStoreLen(t *testing.T)  {
	oq := &transactionStore{presence:make(map[string]*list.Element), order:list.List{}}
	if oq.Len() != 0 {
		t.Error("error orderedRequests len() error!")
	}
	//oq = nil
	t1 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("100"), TransactionHash:[]byte("t1")}
	t2 := &types.Transaction{From:[]byte("from addr"), To:[]byte("to addr"), Value:[]byte("200"), TransactionHash:[]byte("t2")}
	oq.add(t1)
	oq.add(t2)
	if oq.Len() != 2 {
		t.Errorf("error Len() = %d, expected: %d", oq.Len(), 2)
	}
}

func TestNewTransactionStore(t *testing.T)  {
	rs := newTransactionStore()
	if rs.order.Len() != 0 || len(rs.presence) != 0 {
		t.Error("error newTransactionStore not worked!")
	}
}