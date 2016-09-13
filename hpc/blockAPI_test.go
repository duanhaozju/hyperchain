package hpc

import (
	"testing"

	"hyperchain/core"
	"fmt"
)

func TestPublicBlockAPI_GetBlocks(t *testing.T) {

	//init db
	core.InitDB(8086)

	//init genesis
	core.CreateInitBlock("../core/genesis.json")
	//
	blockAPI := NewPublicBlockAPI()
	//
	a := blockAPI.GetBlocks()
	//
	fmt.Println(a)
}