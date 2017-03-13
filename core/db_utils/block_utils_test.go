package db_utils

import (
	"time"
	"testing"
	"hyperchain/core/types"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"strconv"
	"fmt"
	"os"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		Version:   []byte(TransactionVersion),
		From:      []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
		Id:        1,
		TransactionHash: []byte("transactionHash1"),
		Nonce: 1,
	},
	&types.Transaction{
		Version:   []byte(TransactionVersion),
		From:      []byte("0000000000000000000000000000000000000001"),
		To:        []byte("0000000000000000000000000000000000000002"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature2"),
		Id:        2,
		TransactionHash: []byte("transactionHash2"),
		Nonce: 2,
	},
	&types.Transaction{
		Version:   []byte(TransactionVersion),
		From:      []byte("0000000000000000000000000000000000000002"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("700"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature3"),
		Id:        3,
		TransactionHash: []byte("transactionHash3"),
		Nonce: 3,
	},
}

var blockCases = types.Block{
	//ParentHash:   []byte("parenthash"),
	//BlockHash:    []byte("blockhash"),
	//Transactions: transactionCases,
	//Timestamp:    time.Now().UnixNano(),
	//MerkleRoot:   []byte("merkeleroot"),
	//Number:       1,
	//WriteTime:    time.Now().UnixNano() + int64(time.Second)/2,
	Version:      []byte(BlockVersion),
	ParentHash:   []byte("parentHash"),
	BlockHash:    []byte("blockHash"),
	Transactions: transactionCases,
	Timestamp:    1489387222,
	MerkleRoot:   []byte("merkleRoot"),
	TxRoot:       []byte("txRoot"),
	ReceiptRoot:  []byte("receiptRoot"),
	Number:       1,
	WriteTime:    1489387223,
	CommitTime:   1489387224,
	EvmTime:      1489387225,
}

var transactionMeta = types.TransactionMeta{
	BlockIndex: 1,
	Index:      1,
}

var receipt = types.Receipt{
	//Version : []byte("1.2"),
	//TxHash : []byte("12345678901234567890123456789012"),
	Version:           []byte(ReceiptVersion),
	PostState:         []byte("postState"),
	CumulativeGasUsed: 1,
	TxHash:           []byte("12345678901234567890123456789012"),
	ContractAddress:   []byte("contractAddress"),
	GasUsed:           1,
	Ret:               []byte("ret"),
	Logs:              []byte("logs"),
	Status:            types.Receipt_SUCCESS,
	Message:           []byte("message"),
}

func InitDataBase() {
	conf := common.NewConfig("../../config/global.yaml")
	conf.MergeConfig("../../config/db.yaml")
	hyperdb.SetDBConfig("../../config/db.yaml", strconv.Itoa(conf.GetInt(common.C_NODE_ID)))
	hyperdb.InitDatabase(conf, hyperdb.DefautNameSpace)
	InitializeChain(hyperdb.DefautNameSpace)
}

// TestPutBlock tests for PutBlock
func TestPersistBlock(t *testing.T) {
	logger.Info("test =============> > > TestPersistBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err,_ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace, blockCases.BlockHash)
	if block.Number != 1 {
		t.Error("TestPersistBlock fail")
	}
	deleteTestData()
}

// TestGetBlock tests for GetBlock
func TestGetBlockHash(t *testing.T) {
	logger.Info("test =============> > > TestGetBlockHash")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	blockHash, err := GetBlockHash(hyperdb.DefautNameSpace, 1)
	if err != nil {
		logger.Fatal(err)
	}
	if string(blockHash) != string(blockCases.BlockHash) {
		t.Errorf("both blockhash is not equal, %s not equal %s, TestGetBlock fail", string(blockHash), string(blockCases.BlockHash))
	}
	deleteTestData()
}

// TestDeleteBlock tests for DeleteBlock
func TestDeleteBlock(t *testing.T) {
	logger.Info("test =============> > > TestDeleteBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace, blockCases.BlockHash)
	if block.Number != 1 {
		t.Error("GetBlock fail")
	}
	err = DeleteBlock(hyperdb.DefautNameSpace, db.NewBatch(), blockCases.BlockHash, true, true)
	block, err = GetBlock(hyperdb.DefautNameSpace, blockCases.BlockHash)
	if err.Error() != "leveldb: not found" {
		t.Errorf("block delete fail, TestDeleteBlock fail")
	}
	deleteTestData()
}

func TestGetMarshalBlock(t *testing.T) {
	logger.Info("test =============> > > TestGetMarshalBlock")
	InitDataBase()
	db, _ := hyperdb.GetDBDatabaseByNamespace(hyperdb.DefautNameSpace)
	err, _ := PersistBlock(db.NewBatch(), &blockCases, true, true)
	if err != nil {
		logger.Fatal(err)
	}
	block, err := GetBlock(hyperdb.DefautNameSpace, blockCases.BlockHash)
	if block.Number != 1 {
		t.Error("GetBlock fail")
	}
	err, _ = GetMarshalBlock(block)
	if err != nil {
		t.Errorf("TestGetMarshalBlock fail")
	}
	deleteTestData()
}

func deleteTestData() {
	childPath := "/build"
	current_dir, _ := os.Getwd()
	path := current_dir + childPath
	err := os.RemoveAll(path)
	if err != nil {
		fmt.Println("remove test data file fail!", err)
	} else {
		fmt.Println("remove test data file success!")
	}
}