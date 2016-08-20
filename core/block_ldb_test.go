package core

import (
	"fmt"
	"testing"
	"time"
	"strconv"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/core/node"
)

var transactions = []types.Transaction{
	{"jack", "hash1", "snail", 100, time.Now(), "signedhash"},
	{"jack", "hash1", "snail", 45, time.Now(), "signedhash"},
	{"jack", "dfghg", "snail", 23, time.Now(), "signedhash"},
}

//var node1 = node.Node{"12.23.15.45", 8001, 8802, "coinbase"}
var node1 = node.NewNode("12.23.15.45", 8001, 8802)

var blockCases = []types.Block{
	{"ParentHash", "BlockHash", transactions, time.Now(), node1, "merkleroot"},
	{"gfd", "dsf", transactions, time.Now(), node1, "merkleroot"},
	{"dfgdfg", "BlockHasdfsdash", transactions, time.Now(), node1, "merkleroot"},
}

func TestInitDB1(t *testing.T) {
	InitDB(2001)
	fmt.Println("InitDB is successful")
}

func TestPutBlockToLDB(t *testing.T) {
	for i, ca := range blockCases {
		err := PutBlockToLDB("key"+strconv.Itoa(i), ca)
		if err == nil{
			fmt.Println("PutBlockToLDB is successful")
		} else {
			t.Error("PutBlockToLDB is fail")
		}
	}
}

func TestGetBlockFromLDB(t *testing.T) {
	for i, ca := range blockCases {
		bl, _ := GetBlockFromLDB("key"+strconv.Itoa(i))
		if bl.ParentHash != ca.ParentHash && bl.BlockHash != ca.BlockHash{
			t.Error("GetBlockFromLDB is fail")
		}else {
			fmt.Println("GetBlockFromLDB is successful")
		}
	}
}

func TestGetAllBlockFromLDB(t *testing.T) {
	cas , err := GetAllBlockFromLDB()
	//fmt.Println(cas)
	if err != nil && cas == nil{
		t.Error("GetBlockFromLDB is fail")
	}else {
		fmt.Println("GetBlockFromLDB is successful")
	}
}

func TestDeleteBlockFromLDB(t *testing.T) {
	for i, _ := range blockCases {
		DeleteBlockFromLDB("key"+strconv.Itoa(i))
		_ , err :=  GetBlockFromLDB("key"+strconv.Itoa(i))
		if (err == nil){
			t.Error("DeleteBlockFromLDB is fail")
		}else {
			fmt.Println("DeleteBlockFromLDB is successful")
		}
	}
}
