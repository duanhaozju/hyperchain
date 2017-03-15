//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package jsonrpc

import (
	"golang.org/x/net/context"
	"math/big"
	"reflect"
	"unicode"
	"unicode/utf8"
)

// Is this an exported - upper case - name?
func isExported(name string) bool {
	rune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(rune)
}

// Is this type exported or a builtin?
func isExportedOrBuiltinType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	// PkgPath will be non-empty even for an exported type,
	// so we need to check the type name as well.
	return isExported(t.Name()) || t.PkgPath() == ""
}

var contextType = reflect.TypeOf((*context.Context)(nil)).Elem()

// isContextType returns an indication if the given t is of context.Context or *context.Context type
func isContextType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == contextType
}

var errorType = reflect.TypeOf((*error)(nil)).Elem()

// Implements this type the error interface
func isErrorType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t.Implements(errorType)
}

// formatName will convert to first character to lower case
func formatName(name string) string {
	ret := []rune(name)
	if len(ret) > 0 {
		ret[0] = unicode.ToLower(ret[0])
	}
	return string(ret)
}

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
