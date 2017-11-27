//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"encoding/json"
	"strings"
	"math/big"
	"fmt"
	"strconv"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Svcname string      // service name
	Version string      // api version
	Service interface{} // api service instance which holds the methods
	Public  bool        // indication if the methods must be considered safe for public use
}

var Apis map[string]*API

func GetApiObjectByNamespace(name string) *API {
	if api, ok := Apis[name]; ok {
		return api
	}
	return nil
}

type Number int64

func Int64ToNumber(n int64) *Number {
	num := Number(n)
	return &num
}

func Uint64ToNumber(n uint64) *Number {
	num := Number(n)
	return &num
}

func IntToNumber(n int) *Number {
	num := Number(n)
	return &num
}

func (n Number) Hex() string { return "0x" + strconv.FormatInt(int64(n), 16) }

// MarshalJSON serialize given number to JSON
func (n Number) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Hex())
}

// UnmarshalJSON parses a hash in its hex from to a number.
func (n *Number) UnmarshalJSON(data []byte) error {

	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	in := new(big.Int)
	_, ok := in.SetString(input, 0)

	if !ok { // test if user supplied string tag
		return fmt.Errorf(`invalid number %s`, data)
	}

	if v, err := strconv.ParseUint(input, 0, 0); err != nil {
		return fmt.Errorf("number %v may be out of range", input)
	} else if v < 0 {
		return fmt.Errorf("number can't be negative or zero, but get %v", input)
	} else {
		*n = *Uint64ToNumber(v)
		return nil
	}
}

func (n Number) Int64() int64 {
	if n <= 0 {
		return 0
	}
	return int64(n)
}

func (n Number) Int() int {
	if n <= 0 {
		return 0
	}
	return int(n)
}
