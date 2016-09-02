// author: Lizhong kuang
// date: 16-8-26
// last modified: 16-8-26 13:08

package core

import (
	"testing"



	"fmt"
	"hyperchain/logger"

)


func TestBlockPool(t *testing.T){
	myLogger.NewLogger(12)
	InitDB(8082)
	//CreateInitBlock("./genesis.json")
	//eventMux:=new(event.TypeMux)
	/*bx:=NewBlockPool(eventMux)
	block:=&types.Block{
		ParentHash: common.FromHex("0x0000000000000000000000000000000000000000000000000000000000000000"),
		BlockHash: common.FromHex("0x00012"),
		Number:19,
	}
	bx.AddBlock(block)*/
	currentChain := GetChainCopy()
	fmt.Println("final number",currentChain.Height)
	fmt.Println("final hash",currentChain.LatestBlockHash)





}