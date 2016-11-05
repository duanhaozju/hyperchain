package hpc

import (
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"hyperchain/p2p"
	"testing"
)

var peermanager = &p2p.GrpcPeerManager{
	NodeID: 1,
}
var db, _ = hyperdb.GetLDBDatabase()
var pm = &manager.ProtocolManager{
	Peermanager:    peermanager,
	AccountManager: am,
}
var keydir = "../config/keystore/"

var am = accounts.NewAccountManager(keydir, encryption)
var from = common.HexToAddress("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
var to = common.HexToAddress("0x0000000000000000000000000000000000000003")
var args = SendTxArgs{
	From:     from,
	To:       &to,
	Gas:      NewInt64ToNumber(1000),
	GasPrice: NewInt64ToNumber(1000),
	Value:    NewInt64ToNumber(1000),
	Payload:  "",
}
var newTx common.Hash
var publicTransactionAPI = NewPublicTransactionAPI(nil, pm, db)

func Test_GetBlockTransactionCountByHash(t *testing.T) {

	hash := [32]byte{'1', '2'}
	_, err := publicTransactionAPI.GetBlockTransactionCountByHash(hash)
	if err == nil {
		t.Errorf("GetBlockTransactionCountByHash wrong")
	}
}

func Test_GetDiscardTransactions(t *testing.T) {
	_, err := publicTransactionAPI.GetDiscardTransactions()
	if err != nil {
		t.Errorf("GetDiscardTransactions wrong")
	}
}

func Test_GetSighHash(t *testing.T) {
	var from = common.HexToAddress("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	var to = common.HexToAddress("0x0000000000000000000000000000000000000003")
	var args = SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
		PrivKey:  "123456",
	}
	ref, err := publicTransactionAPI.GetSighHash(args)
	if err != nil {
		fmt.Println(ref)
		t.Errorf(err.Error())
	}
}

func Test_GetTransactionByBlockHashAndIndex(t *testing.T) {
	core.InitDB("C:/hyperchain", 8023)

	var peermanager = &p2p.GrpcPeerManager{
		NodeID: 1,
	}
	var keydir = "../config/keystore/"
	var am = accounts.NewAccountManager(keydir, encryption)
	var db, _ = hyperdb.GetLDBDatabase()
	var pm = &manager.ProtocolManager{
		Peermanager:    peermanager,
		AccountManager: am,
	}
	var publicTransactionAPI = NewPublicTransactionAPI(nil, pm, db)

	_, err := publicTransactionAPI.GetTransactionByBlockHashAndIndex([32]byte{42, 136, 73, 59, 200, 83, 12, 108, 225, 221, 83, 243, 113, 218, 60, 174, 47, 163, 144, 111, 225, 233, 176, 255, 193, 180, 10, 131, 111, 227, 45, 239}, Number(1))
	if err == nil {
		t.Errorf("GetTransactionByBlockHashAndIndex wrong ")
	}
}

func Test_GetTransactionByBlockNumberAndIndex(t *testing.T) {
	_, err := publicTransactionAPI.GetTransactionByBlockNumberAndIndex(Number(-1), Number(1))
	if err == nil {
		t.Errorf("GetTransactionByBlockNumberAndIndex wrong ")
	}
}

func Test_GetTransactionByBlockNumberAndIndex2(t *testing.T) {
	core.InitDB("C:/hyperchain", 8023)
	var peermanager = &p2p.GrpcPeerManager{
		NodeID: 1,
	}
	var keydir = "../config/keystore/"
	var am = accounts.NewAccountManager(keydir, encryption)
	var db, _ = hyperdb.GetLDBDatabase()
	var pm = &manager.ProtocolManager{
		Peermanager:    peermanager,
		AccountManager: am,
	}
	var publicTransactionAPI = NewPublicTransactionAPI(nil, pm, db)

	ref, err := publicTransactionAPI.GetTransactionByBlockNumberAndIndex(Number(1), Number(1))
	fmt.Println(ref)
	fmt.Println(err)

}

func Test_GetTransactionReceipt(t *testing.T) {
	ref, err := publicTransactionAPI.GetTransactionReceipt([32]byte{42, 136, 73, 59, 200, 83, 12, 108, 225, 221, 83, 243, 113, 218, 60, 174, 47, 163, 144, 111, 225, 233, 176, 255, 193, 180, 10, 131, 111, 227, 45, 239})
	fmt.Println("GetTransactionReceipt:", ref)
	fmt.Println("GetTransactionReceipt:", err)

}

func Test_GetTransactions(t *testing.T) {
	core.InitDB("C:/hyperchain", 8023)
	var peermanager = &p2p.GrpcPeerManager{
		NodeID: 1,
	}
	var keydir = "../config/keystore/"
	var am = accounts.NewAccountManager(keydir, encryption)
	var db, _ = hyperdb.GetLDBDatabase()
	var pm = &manager.ProtocolManager{
		Peermanager:    peermanager,
		AccountManager: am,
	}
	var publicTransactionAPI = NewPublicTransactionAPI(nil, pm, db)

	//	var publicTransactionAPI2 = NewPublicTransactionAPI(nil, pm, nil)

	num1 := Number(0000000000000000000000000000000000000001)
	num2 := Number(0000000000000000000000000000000000000003)
	Inargs := IntervalArgs{
		From: &num1,
		To:   &num2,
	}
	ref, err := publicTransactionAPI.GetTransactions(Inargs)
	if err == nil {
		fmt.Println(ref)
		t.Errorf("GetTransactions wrong")
	}

	//	ref1, err1 := publicTransactionAPI2.GetTransactions(Inargs)
	//	fmt.Println(ref1)
	//	fmt.Println(err1)

}

func Test_GetTransactionsCount(t *testing.T) {

	ref, err := publicTransactionAPI.GetTransactionsCount()
	fmt.Println(*ref)
	fmt.Println(err)
}

func Test_SendTransaction(t *testing.T) {
	ref, err := publicTransactionAPI.SendTransaction(args)
	if err != nil {
		t.Errorf(ref.Hex())
	}
}
