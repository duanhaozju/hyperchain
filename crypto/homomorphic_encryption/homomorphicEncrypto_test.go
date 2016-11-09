//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package homomorphic_encryption

import (
	"bytes"
	"encoding/binary"
	"fmt"
	//"math/big"
	"testing"
)

//整形转换成字节
func IntToBytes(n int) []byte {
	tmp := int32(n)
	bytesBuffer := bytes.NewBuffer([]byte{})
	binary.Write(bytesBuffer, binary.BigEndian, tmp)
	return bytesBuffer.Bytes()
}

//字节转换成整形
func BytesToInt(b []byte) int {

	if len(b) == 1 {
		bytesBuffer := bytes.NewBuffer(b)
		var tmp int8
		binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp)
	} else if len(b) == 2 {
		bytesBuffer := bytes.NewBuffer(b)
		var tmp int16
		binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp)
	} else if len(b) == 5 {
		bytesBuffer := bytes.NewBuffer(b[1:])
		var tmp int32
		binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp)
	} else {
		bytesBuffer := bytes.NewBuffer(b)
		var tmp int32
		binary.Read(bytesBuffer, binary.BigEndian, &tmp)
		return int(tmp)
	}
}

func TestNew_Paillier_Hmencryption(t *testing.T) {

	//普通的测试
	//	paillier_hm, _ := New_Paillier_Hmencryption()
	//	k, _ := paillier_hm.NewKey() //返回密钥对的指针

	//	var message1, message2 int
	//	message1 = -20
	//	message2 = 30

	//	m1 := IntToBytes(message1)
	//	fmt.Println(m1)
	//	m2 := IntToBytes(message2)
	//	fmt.Println(m2)
	//	ciphertext1, _ := paillier_hm.Encrypto_message(k, m1)
	//	m3 := paillier_hm.Decrypto_Ciphertext(k, ciphertext1)
	//	fmt.Println(m3)
	//	ciphertext2, _ := paillier_hm.Encrypto_message(k, m2)
	//	addresult, _ := paillier_hm.Calculator(k, "paillier", ciphertext1, ciphertext2)
	//	result := paillier_hm.Decrypto_Ciphertext(k, addresult)
	//	fmt.Println(result)
	//	var res int
	//	res = BytesToInt(result)
	//	fmt.Println(res)

	//----------------------------------------------------------------------------------
	//测试-10+30是否会等于15+5
	paillier_hm := New_Paillier_Hmencryption()
	k, _ := paillier_hm.NewKey() //返回密钥对的指针
	key := k.(*PaillierKey)
	publickey := &key.PaillierPublickey
	//fmt.Println(key.P)
	//fmt.Println(*key.P)

	var message1, message2, message3, message4 int
	message1 = 50
	message2 = 30
	message3 = 20
	message4 = 5

	m1 := IntToBytes(message1)
	m2 := IntToBytes(message2)
	m3 := IntToBytes(message3)
	m4 := IntToBytes(message4)

	en_message1, _ := paillier_hm.Encrypto_message(publickey, m1)
	//fmt.Println(len(en_message1))
	en_message2, _ := paillier_hm.Encrypto_message(publickey, m2)
	en_message3, _ := paillier_hm.Encrypto_message(publickey, m3)
	en_message4, _ := paillier_hm.Encrypto_message(publickey, m4)

	add_message12, _ := paillier_hm.Calculator(publickey, "paillier", en_message1, en_message2)
	add_message34, _ := paillier_hm.Calculator(publickey, "paillier", en_message3, en_message4)

	fmt.Println(add_message12)
	fmt.Println(add_message34)

	de_message12 := paillier_hm.Decrypto_Ciphertext(k, add_message12)
	fmt.Println(de_message12)
	de_message34 := paillier_hm.Decrypto_Ciphertext(k, add_message34)
	fmt.Println(de_message34)
	m12 := BytesToInt(de_message12)
	m34 := BytesToInt(de_message34)
	fmt.Println(m12)
	fmt.Println(m34)

}
