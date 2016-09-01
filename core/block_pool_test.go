// author: Lizhong kuang
// date: 16-8-26
// last modified: 16-8-26 13:08

package core

import (
	"testing"

	"hyperchain/event"
	"hyperchain/core/types"

	"fmt"
	"hyperchain/logger"
	"hyperchain/common"
)


func TestBlockPool(t *testing.T){
	myLogger.NewLogger(12)
	InitDB(12)
	CreateInitBlock("./genesis.json")
	eventMux:=new(event.TypeMux)
	bx:=NewBlockPool(eventMux)
	block:=&types.Block{
		ParentHash: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		BlockHash: common.FromHex("0x00012"),
		Number:19,
	}
	bx.AddBlock(block)
	currentChain := GetChainCopy()
	fmt.Println("final number",currentChain.Height)





}
