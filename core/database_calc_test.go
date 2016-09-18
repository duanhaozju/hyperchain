package core

import (
	"testing"
	"fmt"
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
}

/*func TestCalcResponseCount(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8084)
	//blockUtilsCase.Number = GetHeightOfChain() + 1
	//commonHash := crypto.NewKeccak256Hash("keccak256")
	//WriteBlock(blockUtilsCase, commonHash)
	fmt.Println(GetHeightOfChain())
	for i := uint64(0); i <= GetHeightOfChain(); i += 1 {
		count,_ := CalcResponseCount(i, int64(300))

		fmt.Println(count)
	}

}*/
func TestCalcCommitBatchAVGTime(t *testing.T) {
	InitDB(8084)
	fmt.Println(CalcCommitBatchAVGTime(uint64(10),uint64(20)))
}

/*func TestGetBlockHash(t *testing.T) {
	log.Info("test =============> > > TestInitDB")
	InitDB(8082)
	db, _ := hyperdb.GetLDBDatabase()
	hash, _ := GetBlockHash(db, 1)
	fmt.Println("1: ", hash)

	block, _ := GetBlock(db, hash)
	fmt.Printf("%#v", block)
}*/

