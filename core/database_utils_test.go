package core

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"os"
	"strconv"
	"testing"
	"time"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		From:      []byte("0000000000000000000000000000000000000001"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("100"),
		TimeStamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
	},
	&types.Transaction{
		From:  []byte("0000000000000000000000000000000000000001"),
		To:    []byte("0000000000000000000000000000000000000002"),
		Value: []byte("100"), TimeStamp: time.Now().UnixNano(),
		Signature: []byte("signature2"),
	},
	&types.Transaction{
		From:      []byte("0000000000000000000000000000000000000002"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("700"),
		TimeStamp: time.Now().UnixNano(),
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
	InitDB(2048)
	hyperdb.GetLDBDatabase()
}

// TestPutTransaction tests for PutTransaction
func TestPutTransaction(t *testing.T) {
	log.Info("test =============> > > TestPutTransaction")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for i, trans := range transactionCases {
		key := []byte("key" + strconv.Itoa(i))
		err = PutTransaction(db, key, trans)
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

func TestGetTransactionBLk(t *testing.T) {
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlockByNumber(db, 5)
	fmt.Println("tx hash", block.Transactions[2].BuildHash())
	tx := block.Transactions[2]
	bh, bn, i := GetTxWithBlock(db, tx.BuildHash().Bytes())
	fmt.Println("block hash", bh, "block num :", bn, "tx index:", i)
}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetAllTransaction")
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
	log.Info("test =============> > > TestDeleteTransaction")
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

// TestPutTransactions tests for PutTransactions
func TestPutTransactions(t *testing.T) {
	log.Info("test =============> > > TestPutTransactions")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	PutTransactions(db, commonHash, transactionCases)
	trs, err := GetAllTransaction(db)
	if err != nil {
		log.Fatal(err)
	}
	if len(trs) < 3 {
		t.Errorf("TestPutTransactions fail")
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
	err = PutBlock(db, blockUtilsCase.BlockHash, &blockUtilsCase)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlock(db, blockUtilsCase.BlockHash)
	fmt.Println(block.Number)
	//height := GetHeightOfChain()
	//for i:=uint64(1);i<=height;i++{
	//	block ,_:= GetBlockByNumber(db,i)
	//	PutBlock(db,block.BlockHash,block)
	//}

}

// TestGetBlock tests for GetBlock
func TestGetBlock(t *testing.T) {
	log.Info("test =============> > > TestGetBlock")
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
	log.Info("test =============> > > TestDeleteBlock")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PutBlock(db, blockUtilsCase.BlockHash, &blockUtilsCase)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlock(db, blockUtilsCase.BlockHash)
	fmt.Println(block.Number)
	err = DeleteBlock(db, blockUtilsCase.BlockHash)
	//err = DeleteBlockByNum(db, 1)
	block, err = GetBlock(db, blockUtilsCase.BlockHash)
	fmt.Println(block.Number)
	if err != leveldb.ErrNotFound {
		t.Errorf("block delete fail, TestDeleteBlock fail")
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
	UpdateChain(&blockUtilsCase, false)
	lasthash := GetLatestBlockHash()
	parentHash := GetParentBlockHash()
	if string(lasthash) != string(blockUtilsCase.BlockHash) {
		t.Errorf("TestUpdateChain fail")
	}
	if string(parentHash) != string(blockUtilsCase.ParentHash) {
		t.Errorf("TestUpdateChain fail")
	}
}

func TestGetReplicas(t *testing.T) {
	replicas := make([]uint64, 10)
	for i := 0; i < 10; i += 1 {
		replicas[i] = uint64(i)
	}
	SetReplicas(replicas)
	t.Log(GetReplicas())
}

func TestGetId(t *testing.T) {
	SetId(uint64(100))
	t.Log(GetId())
}

func TestGetBlockByNumber(t *testing.T) {
	db, _ := hyperdb.GetLDBDatabase()
	blk, err := GetBlockByNumber(db, 123)
	t.Log(blk, err)
}
