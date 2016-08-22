package core

import (
	"time"
	"hyperchain-alpha/core/types"
	"testing"
	"strconv"
	"fmt"
)

var transactionCases = []types.Transaction{
	{"jack", "hash1", "snail", 100, time.Now(), "signedhash"},
	{"jack", "hash1", "snail", 45, time.Now(), "signedhash"},
	{"jack", "dfghg", "snail", 23, time.Now(), "signedhash"},
}

func TestInitDB(t *testing.T) {
	InitDB(2001)
	fmt.Println("InitDB is successful")
}

func TestPutTransactionToLDB(t *testing.T) {
	for i, ca := range transactionCases {
		err := PutTransactionToLDB("key"+strconv.Itoa(i), ca)
		if err == nil{
			fmt.Println("PutTransactionToLDB is successful")
		} else {
			t.Error("PutTransactionToLDB is fail")
		}
	}
}

func TestGetTransactionFromLDB(t *testing.T) {
	for i, ca := range transactionCases {
		trans, _ := GetTransactionFromLDB("key"+strconv.Itoa(i))
		if trans.TimeStamp != ca.TimeStamp{
			t.Error("GetTransactionFromLDB is fail")
		}else {
			fmt.Println("GetTransactionFromLDB is successful")
		}
	}
}

func TestGetAllTransactionFromLDB(t *testing.T) {
	cas , err := GetAllTransactionFromLDB()
	if err != nil && cas == nil{
		t.Error("GetAllTransactionFromLDB is fail")
	}else {
		fmt.Println("GetAllTransactionFromLDB is successful")
	}
}

func TestDeleteTransactionFromLDB(t *testing.T) {
	for i, _ := range transactionCases {
		DeleteTransactionFromLDB("key"+strconv.Itoa(i))
		_ , err :=  GetTransactionFromLDB("key"+strconv.Itoa(i))
		if (err == nil){
			t.Error("DeleteTransactionFromLDB is fail")
		}else {
			fmt.Println("DeleteTransactionFromLDB is successful")
		}
	}
}