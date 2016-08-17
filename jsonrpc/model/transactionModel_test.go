package model

import (
	"testing"
	"fmt"
	"time"
	//"github.com/hashicorp/hcl/hcl/token"
	"strconv"
)

func TestInitCurrentKeyId(t *testing.T) {
	InitDB(0)
}

//测试模型建立方法


/************ 以下测试 ***************/
var transactionCases = []Transaction{
	{"hash1", "jack", "snail", 100, time.Now()},
	{"hash2", "jack", "snail", 34, time.Now()},
	{"hash3", "jack", "snail", 3, time.Now()},
}

func TestSaveTransction(t *testing.T){
	for _, tr := range transactionCases {
		SaveTransction(tr)
	}
}

func TestPutTransactionToDB(t *testing.T) {

	for i, tr := range transactionCases {
		err := PutTransactionToDB("key"+strconv.Itoa(i), tr)
		if err == nil{
			trans, _ := GetTransactionFromDB("key"+strconv.Itoa(i))
			if trans.Hash != tr.Hash {
				t.Error("PutTransactionToDB is fail")
			}else {
				fmt.Println("PutTransactionToDB is successful")
			}
		}
	}
}

func TestGetTransactionFromDB(t *testing.T) {
	for i, tr := range transactionCases {
		trans, _ := GetTransactionFromDB("key"+strconv.Itoa(i))
		if trans.Hash != tr.Hash {
			t.Error("GetTransactionFromDB is fail")
		}else {
			fmt.Println("GetTransactionFromDB is successful")
		}
	}
}

func TestGetAllTransaction(t *testing.T) {
	ts , err := GetAllTransaction()
	fmt.Println()
	if err != nil && ts != nil{
		t.Error("GetTransactionFromDB is fail")
	}else {
		fmt.Println("GetAllTransaction is successful")
	}
}

func TestDeleteTransactionFromDB(t *testing.T) {
	for i, _ := range transactionCases {
		DeleteTransactionFromDB("key"+strconv.Itoa(i))
		_ , err :=  GetTransactionFromDB("key"+strconv.Itoa(i))
		if (err == nil){
			t.Error("DeleteTransactionFromDB is fail")
		}else {
			fmt.Println("DeleteTransactionFromDB is successful")
		}
	}
}
/******************** end ***********************/
