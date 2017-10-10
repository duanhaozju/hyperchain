// Copyright 2014 The go-ethereum Authors
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

func TestSegmenting(t *testing.T) {
	prog := NewProgram([]byte{byte(PUSH1), 0x1, byte(PUSH1), 0x1, 0x0})
	err := CompileProgram(prog)
	if err != nil {
		t.Fatal(err)
	}

	if instr, ok := prog.instructions[0].(pushSeg); ok {
		if len(instr.data) != 2 {
			t.Error("expected 2 element width pushSegment, got", len(instr.data))
		}
	} else {
		t.Errorf("expected instr[0] to be a pushSeg, got %T", prog.instructions[0])
	}

	prog = NewProgram([]byte{byte(PUSH1), 0x1, byte(PUSH1), 0x1, byte(JUMP)})
	err = CompileProgram(prog)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := prog.instructions[1].(jumpSeg); ok {
	} else {
		t.Errorf("expected instr[1] to be jumpSeg, got %T", prog.instructions[1])
	}

	prog = NewProgram([]byte{byte(PUSH1), 0x1, byte(PUSH1), 0x1, byte(PUSH1), 0x1, byte(JUMP)})
	err = CompileProgram(prog)
	if err != nil {
		t.Fatal(err)
	}
	if instr, ok := prog.instructions[0].(pushSeg); ok {
		if len(instr.data) != 2 {
			t.Error("expected 2 element width pushSegment, got", len(instr.data))
		}
	} else {
		t.Errorf("expected instr[0] to be a pushSeg, got %T", prog.instructions[0])
	}
	if _, ok := prog.instructions[2].(jumpSeg); ok {
	} else {
		t.Errorf("expected instr[1] to be jumpSeg, got %T", prog.instructions[1])
	}
}

func TestCompiling(t *testing.T) {
	prog := NewProgram([]byte{0x60, 0x10})
	err := CompileProgram(prog)
	if err != nil {
		t.Error("didn't expect compile error")
	}

	if len(prog.instructions) != 1 {
		t.Error("expected 1 compiled instruction, got", len(prog.instructions))
	}
}

func TestPcMappingToInstruction(t *testing.T) {
	program := NewProgram([]byte{byte(PUSH2), 0xbe, 0xef, byte(ADD)})
	CompileProgram(program)
	if program.mapping[3] != 1 {
		t.Error("expected mapping PC 4 to me instr no. 2, got", program.mapping[4])
	}
}
