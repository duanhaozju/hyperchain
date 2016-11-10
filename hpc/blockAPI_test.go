//author:zsx
//data:2016-11-3
package hpc

//import (
//	"fmt"
//	"hyperchain/common"
//	"hyperchain/core"
//	"hyperchain/hyperdb"
//	"testing"
//)

//func Test_blockAPI(t *testing.T) {
//	fmt.Println("-------------Test BlockAPI start--------------------------------")
//	core.InitDB("./build/keystore", 8023)
//	db1, _ := hyperdb.GetLDBDatabase()
//	publicBlockAPI2 := NewPublicBlockAPI(db1)
//	if publicBlockAPI2.db != db1 {
//		t.Errorf("NewPublicBlockAPI wrong")
//	}

//	a := Number(123)
//	b := Number(124)
//	args := &IntervalArgs{
//		From: &a,
//		To:   &b,
//	}
//	_, err := publicBlockAPI2.GetBlocks(*args)
//	if err.Error() != "leveldb: not found" {
//		t.Errorf("getBlock1 wrong")
//	}

//	var a1, b1 *Number
//	args1 := &IntervalArgs{
//		From: a1,
//		To:   b1,
//	}
//	_, err1 := publicBlockAPI2.GetBlocks(*args1)
//	if err1.Error() != "leveldb: not found" {
//		t.Errorf("getBlock2 wrong")
//	}

//	_, err2 := publicBlockAPI2.LatestBlock()
//	if err2.Error() != "leveldb: not found" {
//		t.Errorf("LatestBlock wrong")
//	}

//	var hash common.Hash
//	ref1, _ := publicBlockAPI2.GetBlockByHash(hash)
//	if ref1 != nil {
//		t.Errorf("GetBlockByHash wrong")
//	}

//	ref2, _ := publicBlockAPI2.GetBlockByNumber(Number(0))
//	if ref2 != nil {
//		t.Errorf("GetBlockByNumber wrong")
//	}

//}
