package hpc

import (
	"testing"

	"hyperchain/core"
)

func TestPublicBlockAPI_GetBlocks(t *testing.T) {

	//init db
	core.InitDB(8083)

	//init genesis
	core.CreateInitBlock("../core/genesis.json")
	//
	blockAPI := NewPublicBlockAPI()
	//
	a, err := blockAPI.GetBlocks()
	//
	if err != nil {
		t.Errorf("%v",err)
	}
	t.Log(a)

	cTime, bTime := blockAPI.QueryCommitAndBatchTime(SendQueryArgs{From:*NewInt64ToNumber(1),To:*NewInt64ToNumber(2),})

	t.Log(cTime,bTime)

	num := blockAPI.QueryTransactionSum()
	t.Log(num)
}