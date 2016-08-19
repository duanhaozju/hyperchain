package encrypt

import (
	"testing"
)

func TestEncodeSignature(t *testing.T) {
	testString := "teststring"
	prikey := GetPrivateKey()
	//pubkey := GetPublicKey(prikey)
	sign,_ := Sign(prikey,[]byte(testString))
	EncodeSignature(&sign)
}
func TestDecodeSignature(t *testing.T) {
	testString := "teststring"
	prikey := GetPrivateKey()
	sign,_ := Sign(prikey,[]byte(testString))
	r := sign.R
	s := sign.S
	signString := EncodeSignature(&sign)
	var toSign Signature
	DecodeSignature(signString,&toSign)
	if r.Cmp(&toSign.R) != 0|| s.Cmp(&toSign.S)!= 0{
		t.Errorf("错误，数值不匹配")
	}
}