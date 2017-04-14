//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"reflect"
	"testing"
)

func TestDisasm(t *testing.T) {
	code := []byte{byte(PUSH2), 0x10, 0x20, byte(PUSH1), 0x1}
	result := []string{"PUSH2", "0x1020", "PUSH1", "0x01"}
	ret := Disasm(code)

	if !reflect.DeepEqual(result, ret) {
		t.Error("disasm error")
	}

}
