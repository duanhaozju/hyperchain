package hpc

import (
	"testing"

	"hyperchain/core"
	"fmt"
)

func TestPublicBlockAPI_GetBlocks(t *testing.T) {

	//init db
	core.InitDB(8083)

	//init genesis
	core.CreateInitBlock("../core/genesis.json")
	//
	blockAPI := NewPublicBlockAPI()
	//
	a := blockAPI.GetBlocks()
	//
	fmt.Println(a)

	cTime, bTime := blockAPI.QueryCommitAndBatchTime(SendQueryArgs{From:"1",To:"2",})

	fmt.Println(cTime,bTime)

	num := blockAPI.QueryTransactionSum()
	fmt.Println(num)
}