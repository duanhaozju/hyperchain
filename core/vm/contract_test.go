//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/hyperdb"
	"testing"
	"reflect"
)

func NewContractTest() *Contract{
	db, _ := hyperdb.NewMemDatabase()
	statedb, _ := NewStateDB(common.Hash{},db)
	from := common.HexToAddress("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd")
	sender := statedb.CreateAccount(from)
	contractAddr := common.HexToAddress("0xbbe2b6412ccf633222374de8958f2acc76cda9c9")
	to := statedb.CreateAccount(contractAddr)
	value := big.NewInt(0)
	contractdeploy := NewContract(sender, to, value, gas, gasPrice)
	return contractdeploy
}
func TestContract_ReturnGas(t *testing.T) {
	contract := NewContractTest()
	contract.ReturnGas(gas,gasPrice)
}

func TestContract_Attribute(t *testing.T) {
	contract := NewContractTest()
	addr := contract.Caller()
	if addr!=common.HexToAddress("000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"){
		t.Error("contract get caller address error")
	}

	contractaddr := contract.Address()
	if common.ToHex(contractaddr[:])!="0xbbe2b6412ccf633222374de8958f2acc76cda9c9"{
		t.Error("contract get address error")
	}

	value := contract.Value()
	if value.Cmp(big.NewInt(0))!=0{
		t.Error("contract get value error")
	}
}

func TestContract_Set(t *testing.T) {
	contract := NewContractTest()
	code := []byte{1,2}
	abi := []byte{3,4}
	from :=common.StringToAddress("addr")

	contract.SetCode(code)
	if !reflect.DeepEqual(contract.Code,code){
		t.Error("contract setcode error")
	}

	contract.SetABI(abi)
	if !reflect.DeepEqual(contract.ABI,abi){
		t.Error("contract setabi error")
	}

	code = []byte{5,6}
	contract.SetCallCode(&from,code)
	if !reflect.DeepEqual(contract.Code,code){
		t.Error("contract setcallcode error")
	}
	if !reflect.DeepEqual(contract.CodeAddr,&from){
		t.Error("contract setcallcode error")
	}

	res:=contract.GetByte(uint64(1))
	if res!=code[1]{
		t.Error("contract getbyte error")
	}
	res0 := contract.GetByte(uint64(2))
	if res0!=0{
		t.Error("contract getbyte error")
	}
	op:=contract.GetOp(uint64(1))//MOD
	if op!=MOD{
		t.Error("contract getOp error")
	}
}