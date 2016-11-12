//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"hyperchain/core"
	"math/big"
)

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

	if v, err := strconv.ParseUint(input, 0, 0);err != nil {
		return fmt.Errorf("number %v may be out of range",input)
	} else if (v < 0) {
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

func (n Number) ToInt() int {
	if n <= 0 {
		return 0
	}
	return int(n)
}

type BlockNumber uint64

const (
	//latestBlockNumber  = 0
	pendingBlockNumber = 1
	earliestBlockNumber = 2
	//maxBlockNumber
)

func NewUint64ToBlockNumber(n uint64) *BlockNumber {
	num := BlockNumber(n)
	return &num
}

func (n BlockNumber) Hex() string { return "0x" + strconv.FormatInt(int64(n), 16) }

// MarshalJSON serialize given number to JSON
func (n BlockNumber) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.Hex())
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

	latest_number := core.GetChainCopy().Height

	if !ok { // test if user supplied string tag

		strBlockNumber := input
		if strBlockNumber == "latest" {
			*n = *NewUint64ToBlockNumber(latest_number)
			//*n = BlockNumber(latestBlockNumber)
			return nil
		}

		if strBlockNumber == "earliest" {
			*n = BlockNumber(earliestBlockNumber)
			return nil
		}

		if strBlockNumber == "pending" {
			*n = BlockNumber(pendingBlockNumber)
			return nil
		}

		return fmt.Errorf(`invalid block number %s`, data)
	}

	if v, err := strconv.ParseUint(input, 0, 0);err != nil {
		return fmt.Errorf("block number %v may be out of range",input)
	} else if (v <= 0) {
		return fmt.Errorf("block number can't be negative or zero, but get %v", input)
	} else if v > latest_number{
		return fmt.Errorf("block number is out of range, and now latest block number is %d", latest_number)
	} else {
		*n = *NewUint64ToBlockNumber(v)
		return nil
	}
}

func (n BlockNumber) ToUint64() uint64 {
	if n <= 0 {
		return 0
	}
	return uint64(n)
}
