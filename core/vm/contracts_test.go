//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
	"testing"
	"reflect"
	"hyperchain/common"
	"math/big"
)

func TestPrecompiledContracts(t *testing.T) {
	res:=PrecompiledContracts()

	ecrecoverkey :=string(common.LeftPadBytes([]byte{1}, 20))
	sha256key :=string(common.LeftPadBytes([]byte{2}, 20))
	ripemendkey :=string(common.LeftPadBytes([]byte{3}, 20))
	lastkey :=string(common.LeftPadBytes([]byte{4}, 20))

	if res[ecrecoverkey].Gas(1).Cmp(big.NewInt(3000))!=0{
		t.Error("cal gas error")
	}
	if res[sha256key].Gas(1).Cmp(big.NewInt(72))!=0{
		t.Error("cal gas error")
	}
	if res[ripemendkey].Gas(1).Cmp(big.NewInt(720))!=0{
		t.Error("cal gas error")
	}
	if res[lastkey].Gas(1).Cmp(big.NewInt(18))!=0{
		t.Error("cal gas error")
	}

}

func TestCrypto(t *testing.T) {
	msg :=[]byte{1,2,3}
	shaExpect:=[]byte{3,144,88,198,242,192,203,73,44,83,59,10,77,20,239,119,204,15,120,171,204,206,213,40,125,132,161,162,1,28,251,129}
	ripExpect :=[]byte{0,0,0,0,0,0,0,0,0,0,0,0,121,249,1,218,38,9,240,32,173,173,191,46,95,104,161,108,140,63,125,87}
	sha256 := sha256Func(msg)
	if !reflect.DeepEqual(sha256,shaExpect){
		t.Error("sha256 error")
	}
	rip:=ripemd160Func(msg)
	if !reflect.DeepEqual(rip,ripExpect){
		t.Error("ripemd160 error")
	}
}

func TestEcrecoverFunc(t *testing.T)  {
	msg := common.FromHex("c4e6d704fc8e4cc1ef3138286202e14edf15186c4a32d1efe81fe853f2bfe10854d4a906baf0d9c93119e1206e7d18caccac36754a2aefdfff47be5565606e9401")
	msg[0]=27
	in := common.LeftPadBytes(msg, 128)
	out :=ecrecoverFunc(in)

	if reflect.DeepEqual(in,out){
		t.Error("ecrecover error")
	}

	msg[32] = 27
	out = ecrecoverFunc(msg)
	if out!=nil{
		t.Error("ecrecover error")
	}

}
