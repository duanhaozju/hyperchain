//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package homomorphic_encryption

import (
	//"crypto/rand"
	"errors"

	"math/big"
	"strings"
)

type Paillier_Hmencryption struct {
	rand *big.Int
	name string //algorithm name
}

func New_Paillier_Hmencryption() *Paillier_Hmencryption {
	//	r, err := rand.Prime(rand.Reader, 256)
	r := big.NewInt(1)
	//	if err != nil {
	//		return nil, err
	//	}
	phm := &Paillier_Hmencryption{rand: r, name: "paillier"}
	return phm
}

func (phm *Paillier_Hmencryption) NewKey() (interface{}, error) {
	key, err := Generate_paillierKey()
	return key, err
}

//将传入的数据设置为[]byte类型，传出也是一个[]byte
func (phm *Paillier_Hmencryption) Encrypto_message(publickey interface{}, message []byte) ([]byte, error) {
	plaintext := new(big.Int)
	plaintext = plaintext.SetBytes(message)
	PaillierPublickey := publickey.(*PaillierPublickey)
	r := phm.rand
	ciphertext, err := PaillierPublickey.Paillier_Encrypto(plaintext, r)
	ciphertext_byte := ciphertext.Bytes()
	return ciphertext_byte, err

}

//传入的密文是[]byte类型，传出的也是[]byte类型
func (phm *Paillier_Hmencryption) Decrypto_Ciphertext(key interface{}, emessage []byte) []byte {

	paillierkey := key.(*PaillierKey)
	ciphertext := new(big.Int)
	ciphertext = ciphertext.SetBytes(emessage)
	//cliphertext := emessage.(*big.Int)
	message := paillierkey.Paillier_Decrypto(ciphertext)
	message_byte := message.Bytes()
	return message_byte

}

func (phm *Paillier_Hmencryption) Calculator(publickey interface{}, name string, args ...[]byte) ([]byte, error) {
	if strings.EqualFold(phm.name, name) == false {
		err := errors.New("the two algorithm don't match")
		return nil, err
	}
	paillierpulickey := publickey.(*PaillierPublickey)
	length := len(args)
	s := big.NewInt(0)
	//r, _ := rand.Prime(rand.Reader, 256)
	r := phm.rand
	sum, _ := paillierpulickey.Paillier_Encrypto(s, r)
	for i := 0; i < length; i++ {
		ciphertext := new(big.Int)
		ciphertext = ciphertext.SetBytes(args[i])
		sum = paillierpulickey.Paillier_Cipher_add(sum, ciphertext)
	}
	sum_byte := sum.Bytes()

	return sum_byte, nil
}
