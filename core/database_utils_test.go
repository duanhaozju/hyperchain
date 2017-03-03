package core

import (
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"reflect"
	"testing"
	"time"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		Version:   []byte("1.2"),
		From:      []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano() - int64(time.Second),
		Signature: []byte("signature1"),
		Id:        1,
	},
	&types.Transaction{
		Version:   []byte("1.2"),
		From:      []byte("0000000000000000000000000000000000000001"),
		To:        []byte("0000000000000000000000000000000000000002"),
		Value:     []byte("100"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature2"),
		Id:        2,
	},
	&types.Transaction{
		Version:   []byte("1.2"),
		From:      []byte("0000000000000000000000000000000000000002"),
		To:        []byte("0000000000000000000000000000000000000003"),
		Value:     []byte("700"),
		Timestamp: time.Now().UnixNano(),
		Signature: []byte("signature3"),
		Id:        3,
	},
}

func TestSignTime(t *testing.T) {
	tx := transactionCases[0]
	kec256Hash := crypto.NewKeccak256Hash("keccak256")

	keydir := "../build/config/keystore"
	//
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := accounts.NewAccountManager(keydir, encryption)
	ac := accounts.Account{
		Address: common.HexToAddress(string(tx.From)),
		File:    am.KeyStore.JoinPath(string(tx.From)),
	}

	am.Unlock(ac, "123")

	tx.From = common.HexToAddress(string(tx.From)).Bytes()
	hash := tx.SighHash(kec256Hash)
	var set = []int{1, 100}

	for _, j := range set {
		start := time.Now()
		for i := 0; i < j; i++ {
			tx.Signature, _ = am.SignWithPassphrase(common.BytesToAddress(tx.From), hash[:], "123")
			//fmt.Println(tx.Signature)
			//tx.ValidateSign(encryption,kec256Hash)
		}
		fmt.Printf("signtx test %dtxs: %s", j, time.Since(start))
		fmt.Println()

	}
	//tx.Signature,_ = am.SignWithPassphrase(common.BytesToAddress(tx.From),hash[:],"123")
	for _, j := range set {
		start := time.Now()
		for i := 0; i < j; i++ {
			//am.SignWithPassphrase(common.HexToAddress(string(tx.From)),hash[:],"123")
			if !tx.ValidateSign(encryption, kec256Hash) {
				t.Error("unsign error")
			}
		}
		fmt.Printf("unsigntx test %dtxs: %s", j, time.Since(start))
		fmt.Println()

	}
	transactionCases[0].Signature = []byte("signature1")
}

// TestPutTransaction tests for PutTransaction
func TestPutTransaction(t *testing.T) {
	log.Info("test =============> > > TestPutTransaction")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	for _, trans := range transactionCases {
		key := trans.Hash(commonHash).Bytes()
		trans.TransactionHash = key
		err, _ = PersistTransaction(db.NewBatch(), trans, "1.2", false, false)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetTransaction")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	key := transactionCases[0].Hash(commonHash).Bytes()
	transactionCases[0].TransactionHash = key
	err, _ = PersistTransaction(db.NewBatch(), transactionCases[0], "1.2", true, true)
	if err != nil {
		log.Fatal(err)
	}
	for _, trans := range transactionCases[:1] {
		key := trans.GetTransactionHash().Bytes()
		tr, err := GetTransaction(db, key)
		if err != nil {
			log.Fatal(err)
		}
		if string(tr.Signature) != string(trans.Signature) {
			t.Errorf("%s not equal %s, TestGetTransaction fail", string(tr.Signature), string(trans.Signature))
		}
	}
}

var transactionMeta = types.TransactionMeta{
	BlockIndex: 1,
	Index:      1,
}

func TestGetTransactionBLk(t *testing.T) {
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err, _ = PersistBlock(db.NewBatch(), &blockUtilsCase, "1.2", true, true)
	err, _ = PersistTransaction(db.NewBatch(), transactionCases[0], "1.2", true, true)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlockByNumber(db, 1)
	if err != nil {
		log.Fatal(err)
	}
	PersistTransactionMeta(db.NewBatch(), &transactionMeta, transactionCases[0].GetTransactionHash(), true, true)
	if len(block.Transactions) > 0 {
		fmt.Println("tx hash", block.Transactions[0].GetTransactionHash().Bytes())
		tx := block.Transactions[0]
		bn, i := GetTxWithBlock(db, tx.GetTransactionHash().Bytes())
		fmt.Println("block num :", bn, "tx index:", i)
	}

}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetAllTransaction")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	key := transactionCases[0].Hash(commonHash).Bytes()
	transactionCases[0].TransactionHash = key
	err, _ = PersistTransaction(db.NewBatch(), transactionCases[0], "1.2", true, true)
	if err != nil {
		log.Fatal(err)
	}
	trs, err := GetAllTransaction(db)
	var zero = types.Transaction{}
	for _, trans := range trs {
		if !reflect.DeepEqual(*trans, zero) {
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
}

// TestDeleteTransaction tests for DeleteTransaction
func TestDeleteTransaction(t *testing.T) {
	log.Info("test =============> > > TestDeleteTransaction")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, trans := range transactionCases[:1] {
		err, _ = PersistTransaction(db.NewBatch(), trans, "1.2", true, true)
		if err != nil {
			log.Fatal(err)
		}
		batch := db.NewBatch()
		DeleteTransaction(batch, trans.GetTransactionHash().Bytes())
		batch.Write()
		_, err := GetTransaction(db, trans.GetTransactionHash().Bytes())
		if err.Error() != "not found" {
			t.Errorf("the transaction key [%s] delete fail, TestDeleteTransaction fail", trans.GetTransactionHash().Bytes())
		}
	}
}

// TestPutTransactions tests for PutTransactions
func TestPutTransactions(t *testing.T) {
	log.Info("test =============> > > TestPutTransactions")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PersistTransactions(db.NewBatch(), transactionCases, "1.2", true, true)
	if err != nil {
		log.Fatal(err)
	}
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
	log.Info("test =============> > > TestPutBlock")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err, _ = PersistBlock(db.NewBatch(), &blockUtilsCase, "1.2", true, true)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlock(db, blockUtilsCase.BlockHash)
	fmt.Println(block.Number)
}

// TestGetBlock tests for GetBlock
func TestGetBlock(t *testing.T) {
	log.Info("test =============> > > TestGetBlock")
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err, _ = PersistBlock(db.NewBatch(), &blockUtilsCase, "1.2", true, true)
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
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err, _ = PersistBlock(db.NewBatch(), &blockUtilsCase, "1.2", true, true)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlock(db, blockUtilsCase.BlockHash)
	fmt.Println(block.Number)
	err = DeleteBlock(db, blockUtilsCase.BlockHash)
	block, err = GetBlock(db, blockUtilsCase.BlockHash)
	if err.Error() != "not found" {
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
	db, err := hyperdb.NewMemDatabase()
	if err != nil {
		log.Fatal(err)
	}
	UpdateChain(db.NewBatch(), &blockUtilsCase, true, true, true)
	lasthash := GetLatestBlockHash()
	parentHash := GetParentBlockHash()
	if string(lasthash) != string(blockUtilsCase.BlockHash) {
		t.Errorf("TestUpdateChain fail")
	}
	if string(parentHash) != string(blockUtilsCase.ParentHash) {
		t.Errorf("TestUpdateChain fail")
	}
}

func TestGetInvaildTx(t *testing.T) {
	tx := transactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	data, _ := proto.Marshal(record)
	// save to db
	db, _ := hyperdb.NewMemDatabase()
	db.Put(append(InvalidTransactionPrefix, tx.TransactionHash...), data)

	result, _ := GetInvaildTxErrType(db, tx.TransactionHash)
	fmt.Println(result)

}
