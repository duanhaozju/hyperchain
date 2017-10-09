//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
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
