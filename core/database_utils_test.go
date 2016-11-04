package core

import (
	"fmt"
	"github.com/syndtr/goleveldb/leveldb"
	"hyperchain/core/types"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"testing"
	"time"
	"hyperchain/accounts"
	"hyperchain/common"
	"github.com/golang/protobuf/proto"
	"os"
)

var transactionCases = []*types.Transaction{
	&types.Transaction{
		From:      []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
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

func TestSignTime(t *testing.T) {
	tx := transactionCases[0]
	kec256Hash := crypto.NewKeccak256Hash("keccak256")

	keydir := "../build/config/keystore"
	//
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := accounts.NewAccountManager(keydir,encryption)
	ac:=accounts.Account{
		Address:common.HexToAddress(string(tx.From)),
		File:am.KeyStore.JoinPath(string(tx.From)),
	}

	am.Unlock(ac,"123")

	tx.From = common.HexToAddress(string(tx.From)).Bytes()
	hash := tx.SighHash(kec256Hash)
	var set = []int{1,100}

	for _,j:=range set{
		start := time.Now()
		for i:=0;i<j;i++{
			tx.Signature,_ = am.SignWithPassphrase(common.BytesToAddress(tx.From),hash[:],"123")
			//fmt.Println(tx.Signature)
			//tx.ValidateSign(encryption,kec256Hash)
		}
		fmt.Printf("signtx test %dtxs: %s",j,time.Since(start))
		fmt.Println()

	}
	//tx.Signature,_ = am.SignWithPassphrase(common.BytesToAddress(tx.From),hash[:],"123")
	for _,j:=range set{
		start := time.Now()
		for i:=0;i<j;i++{
			//am.SignWithPassphrase(common.HexToAddress(string(tx.From)),hash[:],"123")
			if !tx.ValidateSign(encryption,kec256Hash){
				t.Error("unsign error")
			}
		}
		fmt.Printf("unsigntx test %dtxs: %s",j,time.Since(start))
		fmt.Println()

	}
	transactionCases[0].Signature = []byte("signature1")
}

// TestInitDB tests for InitDB
func TestInitDB(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	hyperdb.GetLDBDatabase()
}

// TestPutTransaction tests for PutTransaction
func TestPutTransaction(t *testing.T) {
	log.Info("test =============> > > TestPutTransaction")
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	for _, trans := range transactionCases {
		key := trans.Hash(commonHash).Bytes()
		err = PutTransaction(db, key, trans)
		if err != nil {
			log.Fatal(err)
		}
	}
}

// TestGetTransaction tests for GetTransaction
func TestGetTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetTransaction")
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	PutTransactions(db, commonHash, transactionCases)
	for _, trans := range transactionCases {
		key := trans.Hash(commonHash).Bytes()
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PutBlock(db, blockUtilsCase.BlockHash, &blockUtilsCase)
	commonHash := crypto.NewKeccak256Hash("keccak256")
	PutTransactions(db, commonHash, transactionCases)
	if err != nil {
		log.Fatal(err)
	}
	block, err := GetBlockByNumber(db, 1)
	if err != nil {
		log.Fatal(err)
	}
	if len(block.Transactions)>0{
		fmt.Println("tx hash", block.Transactions[0].BuildHash())
		tx := block.Transactions[0]
		bn, i := GetTxWithBlock(db, tx.BuildHash().Bytes())
		fmt.Println("block num :", bn, "tx index:", i)
	}

}

// TestGetAllTransaction tests for GetAllTransaction
func TestGetAllTransaction(t *testing.T) {
	log.Info("test =============> > > TestGetAllTransaction")
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	commonHash := crypto.NewKeccak256Hash("keccak256")
	PutTransactions(db, commonHash, transactionCases)
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	for _, trans := range transactionCases {
		commonHash := crypto.NewKeccak256Hash("keccak256")
		key := trans.Hash(commonHash).Bytes()
		PutTransaction(db,key,trans)
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
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
	log.Info("test =============> > > TestPutBlock")
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatal(err)
	}
	err = PutBlock(db, blockUtilsCase.BlockHash, &blockUtilsCase)
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
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
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	replicas := make([]uint64, 10)
	for i := 0; i < 10; i += 1 {
		replicas[i] = uint64(i)
	}
	SetReplicas(replicas)
	t.Log(GetReplicas())
}

func TestGetId(t *testing.T) {
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	SetId(uint64(100))
	t.Log(GetId())
}

func TestGetInvaildTx(t *testing.T) {
	InitDB("/tmp",8001)
	defer os.RemoveAll("/tmp/hyperchain/cache8001")
	tx := transactionCases[0]
	record := &types.InvalidTransactionRecord{
		Tx:      tx,
		ErrType: types.InvalidTransactionRecord_OUTOFBALANCE,
	}
	data,_ := proto.Marshal(record)
	// save to db
	db, _ := hyperdb.GetLDBDatabase()
	db.Put(append(InvalidTransactionPrefix, tx.TransactionHash...), data)

	result,_ := GetInvaildTxErrType(db,tx.TransactionHash)
	fmt.Println(result)

}

