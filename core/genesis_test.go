// author: Lizhong kuang
// date: 16-8-26
// last modified: 16-8-26 13:08

package core

import (
	"testing"


)


func TestGenesis(t *testing.T){
	InitDB(123)
	CreateInitBlock("genesis.json")



}
