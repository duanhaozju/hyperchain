package core

import (
	"testing"
	"hyperchain/crypto"
	"time"
	"log"
)

func TestCalcResponseCount(t *testing.T) {
	log.Println("test =============> > > TestInitDB")
	InitDB(100000)
	blockUtilsCase.Number = GetHeightOfChain() + 1
	commonHash := crypto.NewKeccak256Hash("keccak256")
	WriteBlock(blockUtilsCase, commonHash)
	count := CalcResponseCount(GetHeightOfChain(), int64(time.Second))
	if count != 2 {
		t.Errorf("%d not equal 2, TestCalcResponseCount fail", count)
	}
}
