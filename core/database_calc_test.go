package core

import (
	"testing"
	//"hyperchain/crypto"
	"fmt"

	"hyperchain/crypto"
)

func TestCalcResponseCount(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8089)
	blockUtilsCase.Number = GetHeightOfChain() + 1
	commonHash := crypto.NewKeccak256Hash("keccak256")
	WriteBlock(blockUtilsCase, commonHash)
	fmt.Println(GetHeightOfChain())
	count := CalcResponseCount(GetHeightOfChain(), 10000)
		/*if count != 2 {
			t.Errorf("%d not equal 2, TestCalcResponseCount fail", count)
		}*/
	fmt.Println(count)
}

/*func TestCalcResponseCount(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8084)
	//blockUtilsCase.Number = GetHeightOfChain() + 1
	//commonHash := crypto.NewKeccak256Hash("keccak256")
	//WriteBlock(blockUtilsCase, commonHash)
	fmt.Println(GetHeightOfChain())
	count := CalcResponseCount(5, 10000)
*//*	if count != 2 {
		t.Errorf("%d not equal 2, TestCalcResponseCount fail", count)
	}*//*
	fmt.Println(count)
}*/

/*func TestGetBlockHash(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8082)
	db, _ := hyperdb.GetLDBDatabase()
	hash, _ := GetBlockHash(db, 1)
	fmt.Println("1: ", hash)

	block, _ := GetBlock(db, hash)
	fmt.Printf("%#v", block)
}*/

