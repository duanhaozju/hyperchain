//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package core

import (
	"testing"
)

func TestGenesis(t *testing.T) {
	InitDB("/tmp", 123)
	CreateInitBlock("genesis.json")
}
