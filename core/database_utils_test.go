//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"github.com/syndtr/goleveldb/leveldb"
	"hyperchain/core/types"
	"hyperchain/hyperdb"
	"os"
	"testing"
	"time"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		From:      []byte("0000000000000000000000000000000000000001"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("100"),
		Timestamp:time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
	},
	&types.Transaction{
		From:  []byte("0000000000000000000000000000000000000001"),
		To:    []byte("0000000000000000000000000000000000000002"),
		Value: []byte("100"), Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature2"),
	},
	&types.Transaction{
		From:      []byte("0000000000000000000000000000000000000002"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("700"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature3"),
	},
}

func isDirExists(path string) bool {
	fi, err := os.Stat(path)

	if err != nil {
		return os.IsExist(err)
	} else {
		return fi.IsDir()
	}
}

// TestInitDB tests for InitDB
func TestInitDB(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB("/tmp",8001)
	hyperdb.GetLDBDatabase()
}

// TestPutTransaction tests for PutTransaction
func TestPutTransaction(t *testing.T) {
	log.Info("test =============> > > TestPutTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, trans := range transactionCases {
		err, _ = PersistTransaction(db.NewBatch(), trans, "1.0", true, true)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	PersistTransactions(db.NewBatch(), transactionCases, "1.0", true, true)
	for _, trans := range transactionCases {
		tr, err := GetTransaction(db, trans.GetTransactionHash().Bytes())
		if err != nil {
			log.Fatal(err)
		}
		if string(tr.Signature) != string(trans.Signature) {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
	}
}

func TestGetTransactionBLk(t *testing.T) {
	InitDB("/tmp",9999)
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	txMeta := types.TransactionMeta{
		BlockIndex: 1,
		Index:1,
	}
	for _, tx := range transactionCases {
		//PersistTransaction(db.NewBatch(), tx, "1.0", true, true)
		PersistTransactionMeta(db.NewBatch(), &txMeta, tx.GetTransactionHash(), true, true)
	}
	if err != nil {
		log.Fatal(err)
	}
	txs, _ := GetAllTransaction(db)
	for i := 0; i < len(txs); i += 1 {
		tx := txs[i]
		//t.Log(tx)
		bn, _ := GetTxWithBlock(db, tx.GetTransactionHash().Bytes())
		if bn != 1 {
			t.Error("get tx with block failed")
		}
	}
}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetAllTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	PersistTransactions(db.NewBatch(), transactionCases, "1.0", true, true)
	trs, err := GetAllTransaction(db)
	for _, trans := range trs {
		isPass := false
		if string(trans.Signature) == "signature1" ||
			string(trans.Signature) == "signature2" ||
			string(trans.Signature) == "signature3" {
			isPass = true
		}
		if !isPass {
			t.Errorf("%s not exist", string(trans.Signature))
		}
	}
}

// TestDeleteTransaction tests for DeleteTransaction
func TestDeleteTransaction(t *testing.T) {
	log.Info("test =============> > > TestDeleteTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, trans := range transactionCases {
		PersistTransaction(db.NewBatch(), trans, "1.0", true, true)
		DeleteTransaction(db, trans.GetTransactionHash().Bytes())
		_, err := GetTransaction(db, trans.GetTransactionHash().Bytes())
		if err != leveldb.ErrNotFound {
			t.Errorf("the transaction key [%s] delete fail, TestDeleteTransaction fail", trans.GetTransactionHash().Hex())
		}
	}
}

// TestPutTransactions tests for PutTransactions
func TestPutTransactions(t *testing.T) {
	log.Info("test =============> > > TestPutTransactions")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	PersistTransactions(db.NewBatch(), transactionCases, "1.0", true, true)
	trs, err := GetAllTransaction(db)
	if err != nil {
		log.Fatal(err)
	}
	if len(trs) < 3 {
		t.Error("TestPutTransactions fail")
	}
	for _, trans := range trs {
		isPass := false
		if string(trans.Signature) == "signature1" ||
			string(trans.Signature) == "signature2" ||
			string(trans.Signature) == "signature3" {
			isPass = true
		}
		if !isPass {
			t.Errorf("%s not exist", string(trans.Signature))
		}
	}
}

var blockUtilsCase = types.Block{
	ParentHash:   []byte("parenthash"),
	BlockHash:    []byte("blockhash"),
	Transactions: transactionCases,
	Timestamp:    time.Now().UnixNano(),
	MerkleRoot:   []byte("merkeleroot"),
	Number:       1,
	WriteTime:    time.Now().UnixNano() + int64(time.Second)/2,
}

// TestPutBlock tests for PutBlock
func TestPutBlock(t *testing.T) {
	//InitDB(8084)
	log.Info("test =============> > > TestPutBlock")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	PersistBlock(db.NewBatch(), &blockUtilsCase, "1.0", true, true)
	block, err := GetBlock(db, blockUtilsCase.BlockHash)
	if block.Number !=  1 {
		t.Error("test put block failed")
	}
}

// TestGetBlock tests for GetBlock
func TestGetBlock(t *testing.T) {
	log.Info("test =============> > > TestGetBlock")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err, _ = PersistBlock(db.NewBatch(), &blockUtilsCase, "1.0", true, true)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlock(db, blockUtilsCase.BlockHash)
	if err != nil {
		log.Fatal(err)
	}
	if string(block.BlockHash) != string(blockUtilsCase.BlockHash) {
		t.Errorf("both blockhash is not equal, %s not equal %s, TestGetBlock fail", string(block.BlockHash), string(blockUtilsCase.BlockHash))
	}
}

// TestDeleteBlock tests for DeleteBlock
func TestDeleteBlock(t *testing.T) {
	log.Info("test =============> > > TestDeleteBlock")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err, _ = PersistBlock(db.NewBatch(), &blockUtilsCase, "1.0", true, true)
	if err != nil {
		log.Fatal(err)
	}
	err = DeleteBlock(db, blockUtilsCase.BlockHash)
	_, err = GetBlock(db, blockUtilsCase.BlockHash)
	if err != leveldb.ErrNotFound {
		t.Error("block delete fail, TestDeleteBlock fail")
	}
}

var blockHashcases = [][]byte{
	[]byte("blockhash1"),
	[]byte("blockhash2"),
	[]byte("blockhash3"),
	[]byte("blockhash4"),
}

// TestUpdateChain tests for UpdateChain
func TestUpdateChain(t *testing.T) {
	log.Info("test =============> > > TestUpdateChain")
	InitDB("/tmp",8001)
	UpdateChain(&blockUtilsCase, false)
	lasthash := GetLatestBlockHash()
	parentHash := GetParentBlockHash()
	if string(lasthash) != string(blockUtilsCase.BlockHash) {
		t.Error("TestUpdateChain fail")
	}
	if string(parentHash) != string(blockUtilsCase.ParentHash) {
		t.Error("TestUpdateChain fail")
	}
}

func TestGetReplicas(t *testing.T) {
	InitDB("/tmp",8001)
	replicas := make([]uint64, 10)
	for i := 0; i < 10; i += 1 {
		replicas[i] = uint64(i)
	}
	SetReplicas(replicas)
	t.Log(GetReplicas())
}

func TestGetId(t *testing.T) {
	InitDB("/tmp",8001)
	SetId(uint64(100))
	t.Log(GetId())
}

/*func TestGetInvaildTx(t *testing.T) {
	tx := transactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	data,_ := proto.Marshal(record)
	// save to db
	db, _ := hyperdb.GetLDBDatabase()
	db.Put(append(invalidTransactionPrefix, tx.TransactionHash...), data)

	result,_ := GetInvaildTxErrType(db,tx.TransactionHash)
	fmt.Println(result)

}*/

