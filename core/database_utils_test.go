package core

import (
	"hyperchain/core/types"
	"time"
	"testing"
	"os"
	"hyperchain/hyperdb"
	"log"
	"strconv"
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		From: []byte("zhangsan"),
		To: []byte("wangwu"),
		Value: []byte("100"),
		TimeStamp: time.Now().Unix(),
		Signature: []byte("signature1"),
	},
	&types.Transaction{
		From: []byte("zhangsan"),
		To: []byte("lisi"),
		Value: []byte("100"),TimeStamp: time.Now().Unix(),
		Signature: []byte("signature2"),
	},
	&types.Transaction{
		From: []byte("lisi"),
		To: []byte("wangwu"),
		Value: []byte("700"),
		TimeStamp: time.Now().Unix(),
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
	fmt.Println("test =============> > > TestInitDB")
	InitDB(2048)
	hyperdb.GetLDBDatabase()
	if !isDirExists("/tmp/hyperchain/cache/2048") {
		t.Errorf("TestInitDB fail")
	}
}

// TestPutTransaction tests for PutTransaction
func TestPutTransaction(t *testing.T) {
	fmt.Println("test =============> > > TestPutTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for i, trans := range transactionCases {
		key := []byte("key" + strconv.Itoa(i))
		err = PutTransaction(db, key, *trans)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	fmt.Println("test =============> > > TestGetTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for i, trans := range transactionCases {
		key := []byte("key" + strconv.Itoa(i))
		tr, err := GetTransaction(db, key)
		if err != nil {
			log.Fatal(err)
		}
		if string(tr.Signature) != string(trans.Signature) {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
	}
}
// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransaction(t *testing.T) {
	fmt.Println("test =============> > > TestGetAllTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	trs, err := GetAllTransaction(db)
	if err != nil {
		log.Fatal(err)
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

// TestDeleteTransaction tests for DeleteTransaction
func TestDeleteTransaction(t *testing.T) {
	fmt.Println("test =============> > > TestDeleteTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for i, _ := range transactionCases {
		key := []byte("key" + strconv.Itoa(i))
		DeleteTransaction(db, key)
		_, err := GetTransaction(db, key)
		if err != leveldb.ErrNotFound {
			t.Errorf("the transaction key [%s] delete fail, TestDeleteTransaction fail", string(key))
		}
	}
}

var blockUtilsCase = types.Block{
	ParentHash: []byte("parenthash"),
	BlockHash: []byte("blockhash"),
	Transactions: transactionCases,
	Timestamp    : time.Now().Unix(),
	MerkleRoot  : []byte("merkeleroot"),
	Number       : 1,
}

// TestPutBlock tests for PutBlock
func TestPutBlock(t *testing.T) {
	fmt.Println("test =============> > > TestPutBlock")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PutBlock(db, blockUtilsCase.BlockHash, blockUtilsCase)
	if err != nil {
		log.Fatal(err)
	}
}

// TestGetBlock tests for GetBlock
func TestGetBlock(t *testing.T) {
	fmt.Println("test =============> > > TestGetBlock")
	db, err := hyperdb.GetLDBDatabase()
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
	fmt.Println("test =============> > > TestDeleteBlock")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = DeleteBlock(db, blockUtilsCase.BlockHash)
	_, err = GetBlock(db, blockUtilsCase.BlockHash)
	if err != leveldb.ErrNotFound {
		t.Errorf("block delete fail, TestDeleteBlock fail")
	}
}