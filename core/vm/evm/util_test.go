// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package evm

import (
	"hyperchain/common"
	"math/big"
	"reflect"
	"testing"
)

func TestDisassemble(t *testing.T) {
	code := []byte{byte(PUSH2), 0x10, 0x20, byte(PUSH1), 0x1}
	result := []string{"PUSH2", "0x1020", "PUSH1", "0x01"}
	ret := Disassemble(code)

	if !reflect.DeepEqual(result, ret) {
		t.Error("Disassemble error")
	}

	code1 := []byte{byte(PUSH6), 0x10, byte(PUSH1), 0x1}
	ret1 := Disassemble(code1)
	if ret1 != nil {
		t.Error("Disassemble error")
	}

	code2 := []byte{byte(PUSH6)}
	ret2 := Disassemble(code2)
	if ret2 != nil {
		t.Error("Disassemble error")
	}
}

type matchTest struct {
	input   []OpCode
	match   []OpCode
	matches int
}

func TestMatchFn(t *testing.T) {
	tests := []matchTest{
		matchTest{
			[]OpCode{PUSH1, PUSH1, MSTORE, JUMP},
			[]OpCode{PUSH1, MSTORE},
			1,
		},
		matchTest{
			[]OpCode{PUSH1, PUSH1, MSTORE, JUMP},
			[]OpCode{PUSH1, MSTORE, PUSH1},
			0,
		},
		matchTest{
			[]OpCode{},
			[]OpCode{PUSH1},
			0,
		},
	}

	for i, test := range tests {
		var matchCount int
		MatchFn(test.input, test.match, func(i int) bool {
			matchCount++
			return true
		})
		if matchCount != test.matches {
			t.Errorf("match count failed on test[%d]: expected %d matches, got %d", i, test.matches, matchCount)
		}
	}
}

type parseTest struct {
	base   OpCode
	size   int
	output OpCode
}

func TestParser(t *testing.T) {
	tests := []parseTest{
		parseTest{PUSH1, 32, PUSH},
		parseTest{DUP1, 16, DUP},
		parseTest{SWAP1, 16, SWAP},
		parseTest{MSTORE, 1, MSTORE},
	}

	for _, test := range tests {
		for i := 0; i < test.size; i++ {
			code := append([]byte{byte(byte(test.base) + byte(i))}, make([]byte, i+1)...)
			output := Parse(code)
			if len(output) == 0 {
				t.Fatal("empty output")
			}
			if output[0] != test.output {
				t.Errorf("%v failed: expected %v but got %v", test.base+OpCode(i), test.output, output[0])
			}
		}
	}
}
func TestPrecompiledContracts(t *testing.T) {
	res := PrecompiledContracts()

	ecrecoverkey := string(common.LeftPadBytes([]byte{1}, 20))
	sha256key := string(common.LeftPadBytes([]byte{2}, 20))
	ripemendkey := string(common.LeftPadBytes([]byte{3}, 20))
	lastkey := string(common.LeftPadBytes([]byte{4}, 20))

	if res[ecrecoverkey].Gas(1).Cmp(big.NewInt(3000)) != 0 {
		t.Error("cal gas error")
	}
	if res[sha256key].Gas(1).Cmp(big.NewInt(72)) != 0 {
		t.Error("cal gas error")
	}
	if res[ripemendkey].Gas(1).Cmp(big.NewInt(720)) != 0 {
		t.Error("cal gas error")
	}
	if res[lastkey].Gas(1).Cmp(big.NewInt(18)) != 0 {
		t.Error("cal gas error")
	}

}

func TestCrypto(t *testing.T) {
	msg := []byte{1, 2, 3}
	shaExpect := []byte{3, 144, 88, 198, 242, 192, 203, 73, 44, 83, 59, 10, 77, 20, 239, 119, 204, 15, 120, 171, 204, 206, 213, 40, 125, 132, 161, 162, 1, 28, 251, 129}
	ripExpect := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 121, 249, 1, 218, 38, 9, 240, 32, 173, 173, 191, 46, 95, 104, 161, 108, 140, 63, 125, 87}
	sha256 := sha256Func(msg)
	if !reflect.DeepEqual(sha256, shaExpect) {
		t.Error("sha256 error")
	}
	rip := ripemd160Func(msg)
	if !reflect.DeepEqual(rip, ripExpect) {
		t.Error("ripemd160 error")
	}
}

func TestEcrecoverFunc(t *testing.T) {
	msg := common.FromHex("c4e6d704fc8e4cc1ef3138286202e14edf15186c4a32d1efe81fe853f2bfe10854d4a906baf0d9c93119e1206e7d18caccac36754a2aefdfff47be5565606e9401")
	msg[0] = 27
	in := common.LeftPadBytes(msg, 128)
	out := ecrecoverFunc(in)

	if reflect.DeepEqual(in, out) {
		t.Error("ecrecover error")
	}

	msg[32] = 27
	out = ecrecoverFunc(msg)
	if out != nil {
		t.Error("ecrecover error")
	}

}
