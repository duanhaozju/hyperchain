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
	"math/big"
	"testing"
)

var (
	num1 = big.NewInt(2)
	num2 = big.NewInt(3)
	num3 = big.NewInt(4)
)

func TestStack(t *testing.T) {
	stack := newstack()
	stack.push(num1)
	d := stack.len()
	if d != 1 {
		t.Error("wrong stack push")
	}
	stack.pushN(num2, num3)
	if d = stack.len(); d != 3 {
		t.Error("wrong stack pushN")
	}
	ret := stack.pop()
	if ret != num3 {
		t.Error("wrong stack pop")
	}
	data := stack.Data()
	if len(data) != 2 {
		t.Error("wrong stack Data method")
	}
	peek := stack.peek()
	if peek != num2 {
		t.Error("wrong stack peek")
	}
	if err := stack.require(3); err == nil {
		t.Error("stack require error")
	}
	if err := stack.require(1); err != nil {
		t.Error("stack requier error")
	}
	stack.pushN(num1, num2, num3)
	//now the data in stack:[2,3,2,3,4]
	stack.dup(1)
	if stack.len() != 6 {
		t.Error("stack dup error")
		tmp := stack.peek()
		if tmp.Cmp(num3) != 0 {
			t.Error("stack dup error")
		}
	}
	stack.swap(3)
	stack.Print()
	//now stack is [2 3 2 4 4 3]
	len := stack.len()
	swp1 := stack.data[len-1]
	swp2 := stack.data[len-3]
	if swp1.Cmp(num2) != 0 || swp2.Cmp(num3) != 0 {
		t.Error("stack swap error")
	}
	empty := newstack()
	empty.Print()

}
