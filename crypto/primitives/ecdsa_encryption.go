/**
author:Zhangkejie
date:16-12-22
*/

package primitives

import (
//"encoding/pem"
)

type EcdsaEncryption struct {
	name string
	id   string
}

func NewEcdsaEncrypto(name string) *EcdsaEncryption {
	ee := &EcdsaEncryption{name: name}
	return ee
}

func (ee *EcdsaEncryption) Sign(payload []byte, pri interface{}) ([]byte, error) {
	sign, err := ECDSASign(pri, payload)

	if err != nil {
		return nil, err
	}

	return sign, err

}

func (ee *EcdsaEncryption) VerifySign(verKey interface{}, msg, signature []byte) (bool, error) {

	result, err := ECDSAVerify(verKey, msg, signature)

	if err != nil {
		return false, err
	}
	return result, nil
}
