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
	"testing"
)

var (
	offset uint64
	size   uint64
)

func TestMemory(t *testing.T) {
	memory := NewMemory()
	offset = 34
	size = 2
	value := []byte{1, 2}
	//memory.Print()
	memory.Resize(uint64(64))
	if memory.Len() != 64 {
		t.Error("memory resize error")
	}
	memory.Set(offset, size, value)
	for i := uint64(0); i < size; i++ {
		if memory.store[offset+i] != value[i] {
			t.Error("memory set error")
		}
	}
	getvalue := memory.Get(int64(offset), int64(size))
	for i := uint64(0); i < size; i++ {
		if memory.store[offset+i] != getvalue[i] {
			t.Error("memory get error")
		}
	}
	getEmptyValue := memory.Get(int64(offset), 0)
	if len(getEmptyValue) != 0 {
		t.Error("memory get error")
	}

	getExceedValue := memory.Get(int64(90), 2)
	if getExceedValue != nil {
		t.Error("memory get error")
	}

	getprtvalue := memory.GetPtr(int64(offset), int64(size))
	for i := uint64(0); i < size; i++ {
		if memory.Data()[offset+i] != getprtvalue[i] {
			t.Error("memory set error")
		}
	}
	getPtrEmpty := memory.GetPtr(int64(offset), 0)
	if len(getPtrEmpty) != 0 {
		t.Error("memory get error")
	}
	getPtrExceed := memory.GetPtr(int64(90), 2)
	if getPtrExceed != nil {
		t.Error("memory get error")
	}
	//memory.Print()
}
