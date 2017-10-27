package guomi

import (
	"github.com/hyperchain/hyperchain/common"
	"testing"
)

func Test_sm4Enc(t *testing.T) {
	hex := "12345678901234567890123456789012123456"
	key := common.Hex2Bytes(hex)
	//src := "0A123312312938239483749827834291E1098734872398AF1278675645340A123312312938239483749827834291E1098734872398AF1278675645340A123312312938239483749827834291E1098734872398AF1278675645340A123312312938239483749827834291E1098734872398AF1278675645340A123312312938239483749827834291E1098734872398AF127867564534"
	//srcByte := common.Hex2Bytes(src)
	srcByte := []byte{48, 120, 48, 56, 48, 50, 49, 50, 48, 101, 48, 56, 48, 49, 49, 48, 56, 99, 97, 99, 102, 99, 56, 54, 56, 99, 100, 100, 98, 48, 101, 51, 49, 52, 50, 48, 48, 50, 55, 56, 98, 100, 100, 56, 102, 100, 56, 54, 56, 99, 100, 100, 98, 48, 101, 51, 49, 52}
	t.Log(srcByte)
	dst, err := Sm4Enc(key, srcByte)
	if err != nil {
		t.Log(err)
	}
	dst, err = Sm4Dec(key, dst)
	if err != nil {
		t.Log(err)
	}
	t.Log(dst)
}

func Test_sm4Dec(t *testing.T) {
	enc := "0x5c2a7a9963"
	hex := "12345678901234567890123456789012"
	key := common.Hex2Bytes(hex)
	dst, err := Sm4Dec(key, common.Hex2Bytes(enc))
	if err != nil {
		t.Log(err)
	}

	t.Log(dst)
}
