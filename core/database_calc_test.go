package core

import (
	"testing"
	"hyperchain/crypto"
)

func TestCalcResponseCount(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8089)
	blockUtilsCase.Number = GetHeightOfChain() + 1
	commonHash := crypto.NewKeccak256Hash("keccak256")
	WriteBlock(&blockUtilsCase, commonHash, 122)
	count, _ := CalcResponseCount(GetHeightOfChain(), 1000)
	if count != 2 {
		t.Errorf("%d not equal 2, TestCalcResponseCount fail", count)
	}
	//InitDB(8081)
	//db, _ := hyperdb.GetLDBDatabase()
	//height := GetHeightOfChain()
	//fmt.Println(height)
	//block,_ := GetBlockByNumber(db,height)
	////var ch crypto.CommonHash
	//kec256Hash := crypto.NewKeccak256Hash("keccak256")
	//fmt.Println(block.Hash(kec256Hash))
	//fmt.Println(block.HashBlock(kec256Hash))
}

/*func TestCalcResponseCount(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8084)
	//blockUtilsCase.Number = GetHeightOfChain() + 1
	//commonHash := crypto.NewKeccak256Hash("keccak256")
	//WriteBlock(blockUtilsCase, commonHash)
	fmt.Println(GetHeightOfChain())
	count := CalcResponseCount(5, int64(300))
	*//*if count != 2 {
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

