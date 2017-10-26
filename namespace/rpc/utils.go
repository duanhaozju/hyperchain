//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package rpc

import (
	"context"
	"github.com/hyperchain/hyperchain/common"
	"reflect"
	"strconv"
	"unicode"
	"unicode/utf8"
)

// Is this an exported - upper case - name?
func isExported(name string) bool {
	decoderune, _ := utf8.DecodeRuneInString(name)
	return unicode.IsUpper(decoderune)
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

var IDType = reflect.TypeOf((*common.ID)(nil)).Elem()

// isIDType returns an indication if the given t is of ID or *ID type
func isIDType(t reflect.Type) bool {
	for t.Kind() == reflect.Ptr {
		t = t.Elem()
	}
	return t == IDType
}

// isPubSub tests whether the given method has a first argument whose type is
// context.Context or not.
func isPubSub(methodType reflect.Type) bool {
	// The first input param numIn(0) is the receiver type, so Subscription methods have
	// at least 2 input params(receiver type and context.Context).
	if methodType.NumIn() < 2 || methodType.NumOut() != 2 {
		return false
	}

	return isContextType(methodType.In(1)) &&
		isIDType(methodType.Out(0)) &&
		isErrorType(methodType.Out(1))
}

func isEmpty(v reflect.Value) bool {
	k := v.Kind()
	switch k {
	case reflect.String:
		return v.String() == ""
	case reflect.Ptr, reflect.Chan, reflect.Func, reflect.Map, reflect.Interface, reflect.Slice:
		return v.IsNil()
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64:
		return strconv.FormatInt(v.Int(), 10) == "0"
	case reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return strconv.FormatUint(v.Uint(), 10) == "0"
	case reflect.Bool:
		return v.Bool() == false
	default:
		if addr, ok := v.Interface().(common.Address); ok {
			return addr.IsZero()
		} else if hash, ok := v.Interface().(common.Hash); ok {
			return common.EmptyHash(hash)
		}
		return true
	}
}

// suitableCallbacks iterates over the methods of the given type. It will determine if a
// method satisfies the criteria for a RPC callback or a subscription callback and adds
// it to the collection of callbacks or subscriptions.
func suitableCallbacks(rcvr reflect.Value, typ reflect.Type) (callbacks, subscriptions) {
	callbacks := make(callbacks)
	subscriptions := make(subscriptions)

METHODS:
	for m := 0; m < typ.NumMethod(); m++ {
		method := typ.Method(m)
		mtype := method.Type
		mname := formatName(method.Name)

		// method must be exported
		if method.PkgPath != "" {
			continue
		}

		var h callback
		h.isSubscribe = isPubSub(mtype)
		h.rcvr = rcvr
		h.method = method
		h.errPos = -1

		firstArg := 1
		numIn := mtype.NumIn()
		if numIn >= 2 && mtype.In(1) == contextType {
			h.hasCtx = true
			firstArg = 2
		}

		// process subscribe method.
		if h.isSubscribe {
			// skip receiver type
			h.argTypes = make([]reflect.Type, numIn-firstArg)
			for i := firstArg; i < numIn; i++ {
				argType := mtype.In(i)
				if isExportedOrBuiltinType(argType) {
					h.argTypes[i-firstArg] = argType
				} else {
					continue METHODS
				}
			}
			subscriptions[mname] = &h
			continue METHODS
		}

		// determine method arguments, ignore first arg since it's the receiver type
		// Arguments must be exported or builtin types
		h.argTypes = make([]reflect.Type, numIn-firstArg)
		for i := firstArg; i < numIn; i++ {
			argType := mtype.In(i)
			if !isExportedOrBuiltinType(argType) {
				continue METHODS
			}
			h.argTypes[i-firstArg] = argType
		}

		// check that all returned values are exported or builtin types
		for i := 0; i < mtype.NumOut(); i++ {
			if !isExportedOrBuiltinType(mtype.Out(i)) {
				continue METHODS
			}
		}

		// when a method returns an error it must be the last returned value
		h.errPos = -1
		for i := 0; i < mtype.NumOut(); i++ {
			if isErrorType(mtype.Out(i)) {
				h.errPos = i
				break
			}
		}

		if h.errPos >= 0 && h.errPos != mtype.NumOut()-1 {
			continue METHODS
		}

		switch mtype.NumOut() {
		case 0, 1:
			break
		case 2:
			// method must one return value and 1 error
			if h.errPos == -1 {
				continue METHODS
			}
			break
		default:
			continue METHODS
		}

		callbacks[mname] = &h
	}

	return callbacks, subscriptions
}
