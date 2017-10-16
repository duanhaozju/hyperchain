//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package common

// IntArrayEquals checks if the arrays of ints are the same
func IntArrayEquals(a []int, b []int) bool {
	if len(a) != len(b) {
		return false
	}
	for i, v := range a {
		if v != b[i] {
			return false
		}
	}
	return true
}

// Clone clones the passed slice
func Clone(src []byte) []byte {
	clone := make([]byte, len(src))
	copy(clone, src)

	return clone
}

// TransportEncode transfer string to hex
func TransportEncode(str string) string {
	b := []byte(str)
	return ToHex(b)
}

// TransportDecode transfer hex to string
func TransportDecode(str string) string {
	if len(str) >= 2 && str[0:2] == "0x" {
		str = str[2:]
	}
	b := Hex2Bytes(str)
	return string(b)
}
