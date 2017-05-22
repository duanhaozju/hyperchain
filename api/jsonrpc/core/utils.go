//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"math/big"
	"reflect"
)

var bigIntType = reflect.TypeOf((*big.Int)(nil)).Elem()

// Indication if this type should be serialized in hex
func isHexNum(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t == bigIntType
}
