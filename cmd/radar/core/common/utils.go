package common

import "math/big"

func LeftPaddingZeroByte(temp []byte) []byte {
	var res []byte
	for i := 0; i < 32-len(temp); i++ {
		res = append(res, 0)
	}
	for i := 0; i < len(temp); i++ {
		res = append(res, temp[i])
	}
	return res
}

func LeftPaddingZero(startAddress *big.Int) []byte {
	temp := startAddress.Bytes()
	var res []byte
	for i := 0; i < 32-len(temp); i++ {
		res = append(res, 0)
	}
	for i := 0; i < len(temp); i++ {
		res = append(res, temp[i])
	}
	return res
}
