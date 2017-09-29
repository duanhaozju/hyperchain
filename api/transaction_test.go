//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"hyperchain/p2p"
	"strconv"
	"testing"
	"time"
)

func Test_SendTransaction(t *testing.T) {
	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据
	core.InitDB("./build/keystore", 8004)
	peermanager := &p2p.GRPCPeerManager{
		NodeID: 1,
	}
	db, _ := hyperdb.GetDBDatabase()
	keydir := "./build/keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GenerateNodeKey(strconv.Itoa(1), "./build/keynodes")
	am := accounts.NewAccountManager(keydir, encryption)
	eventMux1 := new(event.TypeMux)
	expiredTime := time.Time{}
	pm := manager.NewEventHub(nil, peermanager, eventMux1, nil, am, nil, 0, true, nil, expiredTime)
	publicTransactionAPI := NewPublicTransactionAPI(eventMux1, pm, db, false, 1, 10000)
	from := common.HexToAddress("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	to := common.HexToAddress("0x0000000000000000000000000000000000000003")
	args := SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
	}
	args1 := SendTxArgs{
		From:      from,
		To:        &to,
		Gas:       NewInt64ToNumber(1000),
		GasPrice:  NewInt64ToNumber(1000),
		Value:     NewInt64ToNumber(1000),
		Payload:   "",
		Signature: "123",
	}
	num := Number(3)
	args3 := SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
		Request:  &num,
	}

	ref, err := publicTransactionAPI.SendTransaction(args)
	if err.Error() != "account is locked" {
		t.Errorf("SendTransaction fail 期望的输出是account is locked")
	}

	am.UnlockAllAccount("./build/keystore")

	ref, err = publicTransactionAPI.SendTransaction(args)
	if err != nil {
		t.Errorf("SendTransaction fail 交易应该成功")
	}

	ref, err = publicTransactionAPI.SendTransaction(args1)
	if err.Error() != "invalid signature" {
		t.Errorf("SendTransaction fail 期望的输出是invalid signature")
	}

	ref, err = publicTransactionAPI.SendTransaction(args3)
	if err != nil {
		t.Errorf("SendTransaction fail 交易应该成功")
	}
	pm2 := manager.NewEventHub(nil, peermanager, nil, nil, am, nil, 0, true, nil, expiredTime)
	publicTransactionAPI2 := NewPublicTransactionAPI(eventMux1, pm2, db, true, 1, 10000)
	fmt.Println("11111111111111111111111111111111")
	ref, err = publicTransactionAPI2.SendTransaction(args)
	if err.Error() != "EventObject is nil" {
		fmt.Println(err)
		t.Errorf("SendTransaction fail 期望的输出是EventObject is nil")
	}

	ref, err = publicTransactionAPI2.SendTransaction(args1)
	if err.Error() != "System is too busy to response " {
		fmt.Println(err)
		t.Errorf("SendTransaction fail 期望的输出是System is too busy to response")
		fmt.Println(ref)
	}
}

func Test_GetTransactionReceipt(t *testing.T) {

	peermanager := &p2p.GRPCPeerManager{
		NodeID: 1,
	}
	db, _ := hyperdb.GetDBDatabase()
	keydir := "./build/keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	encryption.GenerateNodeKey(strconv.Itoa(1), "./build/keynodes")
	am := accounts.NewAccountManager(keydir, encryption)
	eventMux1 := new(event.TypeMux)
	expiredTime := time.Time{}
	pm := manager.NewEventHub(nil, peermanager, eventMux1, nil, am, nil, 0, true, nil, expiredTime)
	publicTransactionAPI := NewPublicTransactionAPI(eventMux1, pm, db, false, 1, 10000)

	from := common.HexToAddress("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	to := common.HexToAddress("0x0000000000000000000000000000000000000003")
	args := SendTxArgs{
		From:     from,
		To:       &to,
		Gas:      NewInt64ToNumber(1000),
		GasPrice: NewInt64ToNumber(1000),
		Value:    NewInt64ToNumber(1000),
		Payload:  "",
	}

	args2 := SendTxArgs{
		From:      from,
		Gas:       NewInt64ToNumber(1000),
		GasPrice:  NewInt64ToNumber(1000),
		Value:     NewInt64ToNumber(1000),
		Payload:   "",
		Timestamp: time.Now().UnixNano(),
	}

	args3 := SendTxArgs{
		From:      from,
		To:        &to,
		Gas:       NewInt64ToNumber(1000),
		GasPrice:  NewInt64ToNumber(1000),
		Value:     NewInt64ToNumber(1000),
		Payload:   "",
		Timestamp: time.Now().UnixNano(),
	}

	args8 := SendTxArgs{
		To:        &to,
		Gas:       NewInt64ToNumber(1000),
		GasPrice:  NewInt64ToNumber(1000),
		Value:     NewInt64ToNumber(1000),
		Payload:   "",
		Timestamp: time.Now().UnixNano(),
	}
	hash := common.HexToHash("")

	hash2 := common.HexToHash("0xca8c9e383b7c2737210cb68152630caf2486df92aad88a2bbb25fc0c4849930e")

	ref, err := publicTransactionAPI.GetTransactionReceipt(hash)
	fmt.Println(ref)
	if err == nil {
		t.Errorf("GetTransactionReceipt1 wrong")
	}

	ref, err = publicTransactionAPI.GetTransactionReceipt(hash2)
	if err != nil {
		t.Errorf("GetTransactionReceipt2 wrong")
	}

	num := BlockNumber(1)
	num1 := BlockNumber(2)

	args4 := IntervalArgs{
		From: &num,
		To:   &num1,
	}

	args5 := IntervalArgs{}

	args6 := IntervalArgs{
		From: &num,
	}

	args7 := IntervalArgs{
		To: &num1,
	}

	//	args5 := &IntervalArgs{
	//		From: Number(1),
	//		To:   Number(2),
	//	}

	ref1, err := publicTransactionAPI.GetTransactions(args4)
	fmt.Println(ref1)
	if err != nil {
		t.Errorf("publicTransactionAPI.GetTransactions(args4) wrong")
	}

	ref1, err = publicTransactionAPI.GetTransactions(args5)
	fmt.Println("publicTransactionAPI.GetTransactions(args5)")
	if err != nil {
		t.Errorf("publicTransactionAPI.GetTransactions(args5) wrong")
	}

	ref2, err := publicTransactionAPI.GetDiscardTransactions()
	if err != nil {
		t.Errorf("publicTransactionAPI.GetDiscardTransactions() fail")
	}

	hash3 := common.HexToHash("0x31f347530e693d16486107fbc9256b29609d69db100c9a8401a038e781a49a52")
	_, err = publicTransactionAPI.getDiscardTransactionByHash(hash3)
	if err != nil {
		t.Errorf("publicTransactionAPI.getDiscardTransactionByHash(hash3) fail")
	}

	hash4 := common.HexToHash("0x31f347530e693d16486107fbc9256b29609d69db100c9a8401a038e781a42222")
	ref4, err := publicTransactionAPI.getDiscardTransactionByHash(hash4)
	if err.Error() != "leveldb: not found" {
		fmt.Println(ref4)
		t.Errorf("publicTransactionAPI.getDiscardTransactionByHash(hash4) fail")
	}

	//	正确的区块哈希
	hash5 := common.HexToHash("0x1e8e6912eb9b88d0bdf7ed81651bbf8016341787ca3440f0c82906fb48bce7c5")
	_, err = publicTransactionAPI.GetTransactionByBlockHashAndIndex(hash5, Number(1))
	if err != nil {
		t.Errorf("publicTransactionAPI.GetTransactionByBlockHashAndIndex(hash5, Number(1)) fail")
	}

	_, err = publicTransactionAPI.GetTransactionByBlockHashAndIndex(hash5, Number(-1))
	if err == nil {
		t.Errorf("publicTransactionAPI.GetTransactionByBlockHashAndIndex(hash5, Number(-1)) fail")
	}

	_, err = publicTransactionAPI.GetTransactionByBlockHashAndIndex(hash3, Number(1))
	if err == nil {
		t.Errorf("publicTransactionAPI.GetTransactionByBlockHashAndIndex(hash3, Number(1)) fail")
	}

	//正确的hash
	ref6, err := publicTransactionAPI.GetBlockTransactionCountByHash(hash5)
	if err != nil {
		fmt.Println("ref6:")
		fmt.Println(ref6)
		t.Errorf("publicTransactionAPI.GetBlockTransactionCountByHash(hash5) fail ")
	}
	//错误的hash
	ref6, err = publicTransactionAPI.GetBlockTransactionCountByHash(hash3)
	if err.Error() != "leveldb: not found" {
		fmt.Println(ref6)
		fmt.Println(err)
		t.Errorf(" publicTransactionAPI.GetBlockTransactionCountByHash(hash3) fail")
	}

	//缺少timestamp
	ref7, err := publicTransactionAPI.GetSignHash(args)
	if err.Error() != "lack of param timestamp" {
		fmt.Println(ref7)
		fmt.Println(err)
		t.Errorf("publicTransactionAPI.GetSignHash(args) fail")
	}
	//from 为空
	ref7, err = publicTransactionAPI.GetSignHash(args8)
	fmt.Println(ref7)
	fmt.Println(err)

	//正确
	ref7, err = publicTransactionAPI.GetSignHash(args3)
	if err != nil {
		fmt.Println(ref7)
		fmt.Println(err)
		t.Errorf("publicTransactionAPI.GetSignHash(args) fail")
	}

	//没有to为部署合约
	ref7, err = publicTransactionAPI.GetSignHash(args2)
	if err != nil {
		fmt.Println(ref7)
		fmt.Println(err)
		t.Errorf("publicTransactionAPI.GetSignHash(args) fail")
	}

	ref8, err := publicTransactionAPI.GetTransactionsCount()
	if err != nil || ref8 == nil {
		fmt.Println("GetTransactionsCount()")
		fmt.Println(ref8)
		fmt.Println(err)
	}

	ref9, _ := publicTransactionAPI.GetTxAvgTimeByBlockNumber(args4)
	fmt.Println("arg4")
	fmt.Println(ref9)

	ref9, _ = publicTransactionAPI.GetTxAvgTimeByBlockNumber(args5)
	fmt.Println("arg5")
	fmt.Println(ref9)

	ref9, _ = publicTransactionAPI.GetTxAvgTimeByBlockNumber(args6)
	fmt.Println("arg6")
	fmt.Println(ref9)

	ref9, _ = publicTransactionAPI.GetTxAvgTimeByBlockNumber(args7)
	fmt.Println("arg7")
	fmt.Println(ref9)

	ref10, err := publicTransactionAPI.GetTransactionByHash(hash3)
	if err != nil {
		fmt.Println(ref10)
		fmt.Println(err)
		t.Errorf("GetTransactionByHash(hash5) fail")
	}

	ref10, err = publicTransactionAPI.GetTransactionByHash(hash4)
	if err.Error() != "leveldb: not found" {
		fmt.Println(ref10)
		fmt.Println(err)
		t.Errorf("GetTransactionByHash(hash4) fail")
	}

	//单例数据库状态设置为close
	hyperdb.Setclose()
	//初始化数据
	core.InitDB("./build/keystore1", 8004)
	db1, _ := hyperdb.GetDBDatabase()
	publicTransactionAPI2 := NewPublicTransactionAPI(eventMux1, pm, db1, false, 1, 10000)
	ref, err = publicTransactionAPI2.GetTransactionReceipt(hash2)
	if err == nil {
		t.Errorf("publicTransactionAPI2.GetTransactionReceipt(hash2)")
	}
	if err == nil {
		t.Errorf("GetTransactionReceipt3 wrong")
	}

	ref1, err = publicTransactionAPI2.GetTransactions(args5)
	if err != nil {
		t.Errorf("publicTransactionAPI2.GetTransactions(args5) wrong")
	}

	ref2, err = publicTransactionAPI2.GetDiscardTransactions()
	fmt.Println(ref2)
	if err != nil {
		t.Errorf("publicTransactionAPI2.GetDiscardTransactions()")
	}

	ref8, err = publicTransactionAPI2.GetTransactionsCount()
	if ref8 != nil {
		fmt.Println("GetTransactionsCount()")
		fmt.Println(ref8)
		fmt.Println(err)
	}

}
