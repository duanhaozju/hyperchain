//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
	"math"
	"math/big"
	"strconv"
	"strings"
)

// API describes the set of methods offered over the RPC interface
type API struct {
	Srvname string      // srvname under which the rpc methods of Service are exposed
	Version string      // api version
	Service interface{} // receiver instance which holds the methods
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

func NewInt64ToNumber(n int64) *Number {
	num := Number(n)
	return &num
}

func NewUint64ToNumber(n uint64) *Number {
	num := Number(n)
	return &num
}

func NewIntToNumber(n int) *Number {
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
		*n = *NewUint64ToNumber(v)
		return nil
	}
}

func (n *Number) ToInt64() int64 {
	if n == nil {
		return 0
	}
	return int64(*n)
}

func (n Number) ToUint64() uint64 {
	if n <= 0 {
		return 0
	}
	return uint64(n)
}

func (n Number) Int() int {
	if n <= 0 {
		return 0
	}
	return int(n)
}

type BlockNumber string

func Uint64ToBlockNumber(n uint64) *BlockNumber {
	number := BlockNumber(strconv.FormatUint(n, 10))
	return &number
}

func (n BlockNumber) BlockNumberToUint64(latest uint64) (uint64, error) {
	input := strings.TrimSpace(string(n))
	in := new(big.Int)
	_, ok := in.SetString(input, 0)

	if !ok {
		if input == "latest" {
			if latest == 0 {
				return 0, &common.InvalidParamsError{Message: "There is no block generated!"}
			} else {
				return latest, nil
			}
		} else if input == "earliest" {
			//TODO
			return 0, &common.InvalidParamsError{Message: "Support later..."}
		} else if input == "pending" {
			//TODO
			return 0, &common.InvalidParamsError{Message: "Support later..."}
		} else {
			return 0, &common.InvalidParamsError{Message: fmt.Sprintf("invalid block number %s", input)}
		}
	}

	if num, err := strconv.ParseUint(input, 0, 64); err != nil {
		return 0, &common.InvalidParamsError{Message: fmt.Sprintf("block number %v may be out of range", input)}
	} else if num <= 0 {
		return 0, &common.InvalidParamsError{Message: fmt.Sprintf("block number can't be negative or zero, but get %v", input)}
	} else if num > latest {
		return 0, &common.InvalidParamsError{Message: fmt.Sprintf("block number is out of range, and now latest block number is %d", latest)}
	} else {
		return num, nil
	}
}

func (n BlockNumber) Hex() (string, error) {
	uint64Num, err := n.BlockNumberToUint64(math.MaxUint64)
	if err != nil {
		return "", err
	}

	return "0x" + strconv.FormatUint(uint64Num, 16), nil
}

// MarshalJSON serialize given number to JSON
func (n BlockNumber) MarshalJSON() ([]byte, error) {
	hexNum, err := n.Hex()
	if err != nil {
		return nil, err
	}
	return json.Marshal(hexNum)
}

// UnmarshalJSON parses a hash in its hex from to a number. It supports:
// - "latest", "earliest" or "pending" as string arguments
// - number
func (n *BlockNumber) UnmarshalJSON(data []byte) error {

	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	in := new(big.Int)
	_, ok := in.SetString(input, 0)

	if !ok && input != "latest" && input != "earliest" && input != "pending" {
		return fmt.Errorf("invalid block number %s", data)
	} else {
		*n = BlockNumber(input)
		return nil
	}
}
