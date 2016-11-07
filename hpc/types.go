package hpc

import (
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"
	"strings"
	"hyperchain/core"
	"errors"
)

type Number int64

const (
	//latestBlockNumber  = 0
	pendingBlockNumber = 1
	earliestBlockNumber = 2
	//maxBlockNumber
)

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

// UnmarshalJSON parses a hash in its hex from to a number. It supports:
// - "latest", "earliest" or "pending" as string arguments
// - number
func (n *Number) UnmarshalJSON(data []byte) error {

	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	if len(input) == 0 {
		*n = *NewUint64ToNumber(core.GetChainCopy().Height)
		//*n = Number(latestBlockNumber)
		return nil
	}

	in := new(big.Int)
	_, ok := in.SetString(input, 0)

	if !ok { // test if user supplied string tag

		strBlockNumber := input
		if strBlockNumber == "latest" {
			*n = *NewUint64ToNumber(core.GetChainCopy().Height)
			//*n = Number(latestBlockNumber)
			return nil
		}

		if strBlockNumber == "earliest" {
			*n = Number(earliestBlockNumber)
			return nil
		}

		if strBlockNumber == "pending" {
			*n = Number(pendingBlockNumber)
			return nil
		}

		return fmt.Errorf(`invalid number %s`, data)
	}

	if v, err := strconv.ParseInt(input, 0, 0);err != nil {
		return errors.New("number out of range")
	} else if (v < 0) {
		return errors.New("number can't be negative")
	} else {
		*n = *NewInt64ToNumber(v)
		return nil
	}

	return nil
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
