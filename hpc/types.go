package hpc

import (
	"math/big"
	"fmt"
	"strings"
)

type Number int64

func NewInt64ToNumber(n int64) *Number{
	num := Number(n)
	return &num
}

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

	*n = Number(in.Int64())

	return nil
}

// MarshalJSON serialize the hex number instance to a hex representation.
func (n *Number) MarshalJSON() ([]byte, error) {
	if n != nil {
		hn := big.NewInt(int64(*n))
		if hn.BitLen() == 0 {
			return []byte(`"0x0"`), nil
		}
		return []byte(fmt.Sprintf(`"0x%x"`, hn)), nil
	}
	return nil, nil
}

func (n *Number) ToInt64() int64{
	if n == nil {
		return 0
	}
	return int64(*n)
}