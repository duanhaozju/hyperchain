package core

import (
	"testing"
	//"strconv"
	"fmt"
	"hyperchain-alpha/core/types"
	"math/big"
)

var ChainCases = []types.Chain{
	{"sdflksdahkfhsdakjfhask", *big.NewInt(324875943875)},
	{"hjfghbfgfsdgvsdaff", *big.NewInt(34957934857934579)},
	{"dfgdffdgdfgfdbvcbcvb", *big.NewInt(756743654645645)},
}

func TestInitDB2(t *testing.T) {
	InitDB(2001)
	fmt.Println("InitDB is successful")
}

/*
func TestPutChainToLDB(t *testing.T) {
	for i, ca := range ChainCases {
		err := PutChainToMEM("key"+strconv.Itoa(i), ca)
		if err == nil{
			fmt.Println("PutChainToLDB is successful")
		} else {
			t.Error("PutChainToLDB is fail")
		}
	}
}

func TestGetChainFromLDB(t *testing.T) {
	for i, ca := range ChainCases {
		cha, _ := GetChainFromMEM("key"+strconv.Itoa(i))
		if ca.LastestBlockHash != cha.LastestBlockHash{
			t.Error("GetChainFromLDB is fail")
		}else {
			fmt.Println("GetChainFromLDB is successful")
		}
	}
}

func TestGetAllChainFromLDB(t *testing.T) {
	cas , err := GetAllChainFromMEM()
	fmt.Println(cas)
	if err != nil && cas == nil{
		t.Error("GetChainFromLDB is fail")
	}else {
		fmt.Println("GetAllChainFromLDB is successful")
	}
}

func TestDeleteChainFromLDB(t *testing.T) {
	for i, _ := range ChainCases {
		DeleteChainFromMEM("key"+strconv.Itoa(i))
		_ , err :=  GetChainFromMEM("key"+strconv.Itoa(i))
		if (err == nil){
			t.Error("DeleteChainFromLDB is fail")
		}else {
			fmt.Println("DeleteChainFromLDB is successful")
		}
	}
}*/
