package encrypt

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"crypto/dsa"
	"fmt"
)

func TestGetPrivateKey(t *testing.T) {
	privateKey := GetPrivateKey()
	if assert.IsType(t,dsa.PrivateKey{},privateKey){
		fmt.Printf("合法秘钥，%v\n",privateKey)
	}else{
		t.Errorf("秘钥生成不合法：%v",privateKey)
	}
}

func TestEncodePrivateKey(t *testing.T) {
	privateKey := GetPrivateKey()
	encodedStr := EncodePrivateKey(&privateKey)
	if assert.IsType(t,"",encodedStr){
		fmt.Printf("合法序列化秘钥，%v\n",encodedStr)
	}else{
		t.Errorf("秘钥生成不合法：%v",encodedStr)
	}
}
func TestDecodePrivateKey(t *testing.T) {
	privateKey := GetPrivateKey()
	encodedStr := EncodePrivateKey(&privateKey)
	privateKey2 := DecodePrivateKey(encodedStr)
	if assert.IsType(t,dsa.PrivateKey{},privateKey2){
		fmt.Printf("合法反序列化秘钥，%v\n",privateKey2)
	}else{
		t.Errorf("秘钥反序列化失败：%v",privateKey2)
	}
}