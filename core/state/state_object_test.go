//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"bytes"
	"fmt"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"hyperchain/trie"
	"math/big"
	"testing"
)

func TestStorageTree(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	obj := NewStateObject(common.HexToAddress("01234567890"), db)
	obj.SetState(common.StringToHash("key1"), common.StringToHash("value1"))
	obj.SetState(common.StringToHash("key2"), common.StringToHash("value2"))
	obj.Update()

	val1:= obj.GetState(common.StringToHash("key1"))
	val2:= obj.GetState(common.StringToHash("key2"))
	if val1!=common.StringToHash("value1")||val2!=common.StringToHash("value2"){
		t.Error("setstate fail")
	}

	obj.SetState(common.StringToHash("key1"), common.StringToHash(""))
	obj.Update()
	val1= obj.GetState(common.StringToHash("key1"))
	if val1!=common.StringToHash(""){
		t.Error("setstate fail")
	}
}

func TestCopy(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	obj := NewStateObject(common.HexToAddress("01234567890"), db)
	obj.SetState(common.StringToHash("key1"), common.StringToHash("value1"))
	obj.SetState(common.StringToHash("key2"), common.StringToHash("value2"))
	obj.SetCode([]byte("code"))
	obj.SetABI([]byte("abi"))
	obj.SetBalance(big.NewInt(123))
	obj.SetNonce(123)
	newObj := obj.Copy()
	if !SOCompare(obj, newObj) {
		t.Error("Copy failed")
	}
}

func TestEncode(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	obj := NewStateObject(common.HexToAddress("01234567890"), db)
	obj.SetState(common.StringToHash("key1"), common.StringToHash("value1"))
	obj.SetState(common.StringToHash("key2"), common.StringToHash("value2"))
	obj.Update()
	obj.SetCode([]byte("code"))
	obj.SetABI([]byte("abi"))
	obj.SetBalance(big.NewInt(123))
	obj.SubBalance(big.NewInt(111))
	obj.SetNonce(123)
	obj.trie.Commit()
	db.Put(obj.codeHash, obj.code)
	buf, _ := obj.EncodeObject()
	obj2, err := DecodeObject(common.HexToAddress("01234567890"), db, buf)
	if err != nil {
		fmt.Println(err)
	}
	if !SOCompare(obj, obj2) {
		t.Error("Decode failed")
	}
}
func TestGetNull(t *testing.T) {
	db, _ := hyperdb.NewMemDatabase()
	obj := NewStateObject(common.HexToAddress("01234567890"), db)
	if (obj.GetState(common.StringToHash("key1"))!=common.Hash{}){
		t.Error("expected null,but got something return")
	}

	if(obj.getAddr(common.StringToHash("key1"))!=common.Hash{}){
		t.Error("expected null,but got somthing return")
	}
}
func SOCompare(so1 *StateObject, so2 *StateObject) bool {
	if bytes.Compare(so1.address.Bytes(), so2.address.Bytes()) != 0 {
		fmt.Println("address mismatch")
		return false
	}
	if so1.BalanceData.Cmp(so2.BalanceData) != 0 {
		fmt.Println("balance mismatch")
		return false
	}
	if so1.nonce != so2.nonce {
		fmt.Println("nonce mismatch")
		return false
	}
	if bytes.Compare(so1.Root(), so2.Root()) != 0 {
		fmt.Println("root mismatch")
		return false
	}
	if bytes.Compare(so1.abi, so2.abi) != 0 {
		fmt.Println("abi mismatch")
		return false
	}
	if bytes.Compare(so1.code, so2.code) != 0 {
		fmt.Println("code mismatch")
		return false
	}
	if bytes.Compare(so1.codeHash, so2.codeHash) != 0 {
		fmt.Println("codehash mismatch")
		return false
	}
	if !StCompare(so1.trie, so2.trie) {
		fmt.Println("storage mismatch")
		return false
	}
	return true
}
func StCompare(trie1 *trie.SecureTrie, trie2 *trie.SecureTrie) bool {
	it := trie1.Iterator()
	for it.Next() {
		key := trie1.GetKey(it.Key)
		val2 := trie2.Get(key)
		if bytes.Compare(it.Value, val2) != 0 {
			return false
		}
	}
	return true
}
