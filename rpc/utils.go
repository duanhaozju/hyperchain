//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"encoding/json"
	admin "github.com/hyperchain/hyperchain/api/admin"
	"math/big"
	"reflect"
	"strings"
)

var bigIntType = reflect.TypeOf((*big.Int)(nil)).Elem()

// Indication if this type should be serialized in hex
func IsHexNum(t reflect.Type) bool {
	if t == nil {
		return false
	}
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	return t == bigIntType
}

func SplitRawMessage(args json.RawMessage) ([]string, error) {
	str := string(args[:])
	length := len(str)
	if length < 2 {
		return nil, admin.ErrInvalidParamFormat
	} else {
		if str[0] != '[' || str[length-1] != ']' {
			return nil, admin.ErrInvalidParamFormat
		}
	}

	if length == 2 {
		return nil, nil
	} else if length < 4 {
		return nil, admin.ErrInvalidParams
	} else {
		str = str[2 : len(str)-2]
		splitstr := strings.Split(str, ",")
		return splitstr, nil
	}
}
