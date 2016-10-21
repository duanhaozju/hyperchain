// author: Lizhong kuang
// date: 16-8-26
// last modified: 16-8-26 13:08

package core

import (
	"testing"
	"fmt"

	"encoding/hex"
)


func TestBlockPool(t *testing.T){
	InitDB("/tmp",8082)
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
	fmt.Println("Block number",currentChain.Height)
	fmt.Println("Block hash",hex.EncodeToString(currentChain.LatestBlockHash))





}
