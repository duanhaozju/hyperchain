//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"fmt"
	"math/big"
	"testing"
)

func TestLogger(t *testing.T) {
	env := &Env{gasLimit: big.NewInt(10000), depth: 0}
	env.evm = New(env, Config{
		EnableJit: true,
		ForceJit:  false,
	})
	cfg := LogConfig{
		DisableMemory:  true,
		DisableStack:   true,
		DisableStorage: true,
		FullStorage:    true,
	}
	logger := NewLogger(cfg, env)

	fmt.Println(logger)

	//pc := uint64(2)
	//op := SSTORE
	//mem := NewMemory()
	//stack := newstack()
	//cost := big.NewInt(10)
	//storage :=make(Storage)
	//structlogstack := []*big.Int{big.NewInt(1),big.NewInt(2)}
	//
	//log := StructLog{pc, op, new(big.Int).Set(gas), cost, []byte("mem"), structlogstack, storage, 0, nil}
	//// Add the log to the collector
	//logger.cfg.Collector.AddStructLog(log)
	//
	//db, _ := hyperdb.NewMemDatabase()
	//statedb, _ := NewStateDB(common.Hash{},db)
	//from := common.HexToAddress("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	//sender := statedb.CreateAccount(from)
	//contractAddr := common.HexToAddress("0xbbe2b6412ccf633222374de8958f2acc76cda9c9")
	//to := statedb.CreateAccount(contractAddr)
	//value := big.NewInt(0)
	//
	//contract := NewContract(sender, to, value, gas, gasPrice)
	//logger.captureState(pc,op,big.NewInt(100),cost,mem,stack,contract,1,nil)

}
