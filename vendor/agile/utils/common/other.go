package common

import (
	"encoding/hex"
	"github.com/pkg/errors"
	"math/big"
	"strconv"
)

/*
	Common
*/

func StringToHex(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s
	} else {
		return "0x" + s
	}
}

// Don't use the default 'String' method in case we want to overwrite
func Bytes2Hex(d []byte) string {
	return hex.EncodeToString(d)
}

func Hex2Bytes(str string) []byte {
	h, _ := hex.DecodeString(str)

	return h
}

func ToHex(b []byte) string {
	hex := Bytes2Hex(b)
	// Prefer output of "0x0" instead of "0x"
	if len(hex) == 0 {
		hex = "0"
	}
	return "0x" + hex
}

func FromHex(s string) []byte {
	if len(s) > 1 {
		if s[0:2] == "0x" {
			s = s[2:]
		}
		if len(s)%2 == 1 {
			s = "0" + s
		}
		return Hex2Bytes(s)
	}
	return nil
}

func ParseData(data ...interface{}) (ret []byte) {
	for _, item := range data {
		switch t := item.(type) {
		case string:
			var str []byte
			if IsHex(t) {
				str = Hex2Bytes(t[2:])
			} else {
				str = []byte(t)
			}

			ret = append(ret, RightPadBytes(str, 32)...)
		case []byte:
			ret = append(ret, LeftPadBytes(t, 32)...)
		}
	}

	return
}

func IsHex(str string) bool {
	l := len(str)
	return l >= 4 && l%2 == 0 && str[0:2] == "0x"
}

func BytesToBig(data []byte) *big.Int {
	n := new(big.Int)
	n.SetBytes(data)

	return n
}

func Bytes2Big(data []byte) *big.Int { return BytesToBig(data) }

func BigD(data []byte) *big.Int { return BytesToBig(data) }

func BytesToHash(b []byte) Hash {
	var h Hash
	h.SetBytes(b)
	return h
}

func HexToString(s string) string {
	if len(s) >= 2 && s[:2] == "0x" {
		return s[2:]
	} else {
		return s
	}
}

func CopyBytes(b []byte) (copiedBytes []byte) {
	copiedBytes = make([]byte, len(b))
	copy(copiedBytes, b)

	return
}

func ConvertStringSliceToIntSlice(input []string) []int {
	var ret []int
	for _, elem := range input {
		e, err := strconv.ParseInt(elem, 10, 0)
		if err != nil {
			logger.Error("invalid config args")
			return nil
		}
		ret = append(ret, int(e))
	}
	return ret
}

func ConvertStringSliceToBoolSlice(input []string) []bool {
	var ret []bool
	for _, elem := range input {
		e, err := strconv.ParseBool(elem)
		if err != nil {
			logger.Error("invalid config args")
			return nil
		}
		ret = append(ret, e)
	}
	return ret
}

func RemoveSubString(str string, begin, end int) (error, string) {
	length := len(str)
	if length <= end || length <= begin || begin < 0 || begin > end {
		return errors.New("invalid params"), ""
	}
	return nil, str[0:begin] + str[end+1:length]
}
