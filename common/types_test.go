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

package common

import (
	"bytes"
	"math/big"
	"testing"
)

func TestBytesConversion(t *testing.T) {
	bytes := []byte{5}
	hash := BytesToHash(bytes)

	var exp Hash
	exp[31] = 5

	if hash != exp {
		t.Errorf("expected %x got %x", exp, hash)
	}
}

func TestHashJsonValidation(t *testing.T) {
	var h Hash
	var tests = []struct {
		Prefix string
		Size   int
		Error  error
	}{
		{"", 2, hashJsonLengthErr},
		{"", 62, hashJsonLengthErr},
		{"", 66, hashJsonLengthErr},
		{"", 65, hashJsonLengthErr},
		{"0X", 64, nil},
		{"0x", 64, nil},
		{"0x", 62, hashJsonLengthErr},
	}
	for i, test := range tests {
		if err := h.UnmarshalJSON(append([]byte(test.Prefix), make([]byte, test.Size)...)); err != test.Error {
			t.Errorf("test #%d: error mismatch: have %v, want %v", i, err, test.Error)
		}
	}
}

func TestAddressUnmarshalJSON(t *testing.T) {
	var a Address
	var tests = []struct {
		Input     string
		ShouldErr bool
		Output    *big.Int
	}{
		{"", true, nil},
		{`""`, true, nil},
		{`"0x"`, true, nil},
		{`"0x00"`, true, nil},
		{`"0xG000000000000000000000000000000000000000"`, true, nil},
		{`"0x0000000000000000000000000000000000000000"`, false, big.NewInt(0)},
		{`"0x0000000000000000000000000000000000000010"`, false, big.NewInt(16)},
	}
	for i, test := range tests {
		err := a.UnmarshalJSON([]byte(test.Input))
		if err != nil && !test.ShouldErr {
			t.Errorf("test #%d: unexpected error: %v", i, err)
		}
		if err == nil {
			if test.ShouldErr {
				t.Errorf("test #%d: expected error, got none", i)
			}
			if a.Big().Cmp(test.Output) != 0 {
				t.Errorf("test #%d: address mismatch: have %v, want %v", i, a.Big(), test.Output)
			}
		}
	}
}

func TestStringAndHex(t *testing.T) {
	in := "99800aacd"
	out := StringToHex(in)
	exp := "0x99800aacd"
	in1 := "0x99800aacd"
	if exp != out {
		t.Errorf("expected %s got %s", exp, out)
	}
	out1 := StringToHex(in1)
	if exp != out1 {
		t.Errorf("expected %s got %s", exp, out1)
	}
	out2 := HexToString(out)
	if out2 != in {
		t.Errorf("expected %s got %s", in, out2)
	}
	out3 := HexToString(in)
	if out3 != in {
		t.Errorf("expected %s got %s", in, out3)
	}
}
func TestHash_Set(t *testing.T) {
	b := []byte{1, 22, 3}
	a := []byte{1, 2, 33}
	ahash := BytesToHash(a)
	bhash := BytesToHash(b)

	bhash.Set(ahash)
	if !hashCmp(ahash, bhash) {
		t.Error("hash.set error")
	}
}
func TestHash(t *testing.T) {
	//b:= []byte{0xa,0xa,3,0,9,8,7}
	hex := "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
	hash := HexToHash(hex)
	byte := []byte{0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 15, 26, 122, 8, 204, 196, 142, 93, 48, 248, 8, 80, 207, 28, 242, 131, 170, 58, 189}
	big := Big("336817623557988624971520752579810150033537725")

	exphex := "0x000000000000000000000000000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
	if hash.Hex() != exphex {
		t.Errorf("expected %s got %s", exphex, hash.Hex())
	}
	if bytes.Compare(hash.Bytes(), byte) != 0 {
		t.Errorf("expected %x got %x", byte, hash.Bytes())
	}
	if big.Cmp(hash.Big()) != 0 {
		t.Errorf("expected %d got %d", big, hash.Big())
	}

	str := hash.Str()

	if !hashCmp(StringToHash(str), hash) {
		t.Error("string to hash error")
	}
	if !hashCmp(BigToHash(big), hash) {
		t.Error("big to hash error")
	}
	if !hashCmp(HexToHash(hex), hash) {
		t.Error("hex to hash error")
	}
	if !hashCmp(BytesToHash(byte), hash) {
		t.Error("byte to hash error")
	}

}
func TestIsHexAddress(t *testing.T) {
	from := "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
	out := IsHexAddress(from)
	if !out {
		t.Error("it is an address")
	}

	from1 := "12m,dg"

	if out1 := IsHexAddress(from1); out1 {
		t.Error("it is not an address representation")
	}

	from2 := "000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
	if out2 := IsHexAddress(from2); !out2 {
		t.Error("it is an address")
	}

}
func TestAddress(t *testing.T) {
	from := "0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"
	address := HexToAddress(from)

	hex := address.Hex()
	str := address.Str()
	big := address.Big()
	byte := address.Bytes()

	if !addrCmp(address, HexToAddress(hex)) {
		t.Error("hex to address error")
	}
	if !addrCmp(address, StringToAddress(str)) {
		t.Error("str to address error")
	}
	if !addrCmp(address, BigToAddress(big)) {
		t.Error("big to address error")
	}
	if !addrCmp(address, BytesToAddress(byte)) {
		t.Error("byte to address error")
	}

}
func addrCmp(a1, a2 Address) bool {
	if bytes.Compare(a1[:], a2[:]) == 0 {
		return true
	}
	return false
}

func hashCmp(h1, h2 Hash) bool {
	if bytes.Compare(h1[:], h2[:]) == 0 {
		return true
	}
	return false
}
