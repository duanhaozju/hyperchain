package common

import (
	"math/rand"
	"strings"
)

/*
	Bytes utils
*/
// Padding empty bytes at the begin
func LeftPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}
	padded := make([]byte, l)
	copy(padded[l-len(slice):], slice)
	return padded
}

func RightPadBytes(slice []byte, l int) []byte {
	if l < len(slice) {
		return slice
	}
	padded := make([]byte, l)
	copy(padded[0:len(slice)], slice)
	return padded
}

// Padding 0 char at the begin of the string
func LeftPadString(str string, l int) string {
	if l < len(str) {
		return str
	}
	zero := ""
	for i := 0; i < l-len(str); i += 1 {
		zero = zero + "0"
	}
	return zero + str
}

// Padding 0 char at the end of the string
func RightPadString(str string, l int) string {
	if l < len(str) {
		return str
	}
	zero := ""
	for i := 0; i < l-len(str); i += 1 {
		zero = zero + "0"
	}
	return str + zero
}

// Padding 0 char at the begin of the string
func LeftPadStringWithChar(str string, l int, c string) string {
	if l < len(str) {
		return str
	}
	zero := ""
	for i := 0; i < l-len(str); i += 1 {
		zero = zero + c
	}
	return zero + str
}

// Padding 0 char at the end of the string
func RightPadStringWithChar(str string, l int, c string) string {
	if l < len(str) {
		return str
	}
	zero := ""
	for i := 0; i < l-len(str); i += 1 {
		zero = zero + c
	}
	return str + zero
}

func SplitStringByInterval(str string, interval int) []string {
	var ret []string
	idx := 0
	length := len(str)
	for {
		if idx < length {
			if idx+interval <= length {
				ret = append(ret, str[idx:idx+interval])
			} else {
				ret = append(ret, str[idx:length])
			}
			idx += interval
		} else {
			break
		}
	}
	return ret
}

// Strip returns a slice of the byte s with all leading and
// trailing Unicode code points contained in cutset removed.
func Strip(in []byte, cutset string) []byte {
	str := string(in)
	out := strings.Trim(str, cutset)
	return []byte(out)
}

/*
	Random utils
*/
// Generate random string with specific length
func RandomString(length int) string {
	var letters = []byte("abcdef0123456789")
	b := make([]byte, length)
	for i := 0; i < length; i += 1 {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

// Return a random choice with specific ratio, true of false
func RandomChoice(ratio float64) int {
	var tmp []int
	for i := 0; i < int(ratio*10); i += 1 {
		tmp = append(tmp, 0)
	}
	for i := int(ratio * 10); i < 10; i += 1 {
		tmp = append(tmp, 1)
	}
	return tmp[rand.Intn(len(tmp))]
}

// Return a random int in the specific range
func RandomInt64(lowerLimit, upperLimit int) int64 {
	return int64(rand.Intn(upperLimit-lowerLimit) + lowerLimit)
}

func RandomNonce() int64 {
	return rand.Int63()
}

func RandomAddress() string {
	AddrLen := 40
	addr := StringToHex(RandomString(AddrLen))
	return addr
}
