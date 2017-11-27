package api

import (
	"encoding/json"
	"strings"
	"math/big"
	"fmt"
	"strconv"
	"math"
)

type BlockNumber string

func uint64ToBlockNumber(n uint64) *BlockNumber {
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
				return 0, fmt.Errorf("there is no block generated!")
			} else {
				return latest, nil
			}
		} else {
			return 0, fmt.Errorf("invalid block number %s", input)
		}
	}

	if num, err := strconv.ParseUint(input, 0, 64); err != nil {
		return 0, fmt.Errorf("block number %v may be out of range", input)
	} else if num <= 0 {
		return 0, fmt.Errorf("block number can't be negative or zero, but get %v", input)
	} else if num > latest {
		return 0, fmt.Errorf("block number %v is out of range, and now latest block number is %d", input, latest)
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

// UnmarshalJSON parses a hash from its hex to a number. It supports:
// - "latest" as string arguments
// - number
func (n *BlockNumber) UnmarshalJSON(data []byte) error {

	input := strings.TrimSpace(string(data))
	if len(input) >= 2 && input[0] == '"' && input[len(input)-1] == '"' {
		input = input[1 : len(input)-1]
	}

	in := new(big.Int)
	_, ok := in.SetString(input, 0)

	if !ok && input != "latest" {
		return fmt.Errorf("invalid block number %s", data)
	} else {
		*n = BlockNumber(input)
		return nil
	}
}


