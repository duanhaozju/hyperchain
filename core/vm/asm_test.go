//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
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
