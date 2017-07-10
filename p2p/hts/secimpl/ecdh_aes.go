package secimpl

import (
	"crypto"
	"io"
)
// this secimpl implements the Security interface

type ECDHWithAES struct{
}
func NewECDHWithAES()*ECDHWithAES{
	return &ECDHWithAES{}
}

// ExchangeKey part
func(ea *ECDHWithAES)GeneratePrivateKey(io.Reader)(crypto.PrivateKey,crypto.PublicKey,error){
	return nil,nil,nil
}

func(ea *ECDHWithAES)Marshal(crypto.PublicKey)[]byte{
	return nil
}

func(ea *ECDHWithAES)UnMarshal([]byte)(crypto.PublicKey,error){
	return nil,nil
}

func(ea *ECDHWithAES)GenerateSharedKey(crypto.PrivateKey,crypto.PublicKey)(sharedKey []byte,err error){
	return nil,nil
}

func(ea *ECDHWithAES)Encrypt(key, originMsg []byte)(encryptedMsg []byte,err error){
	return originMsg,nil
}
func(ea *ECDHWithAES)Decrypt(key, encryptedMsg []byte)(originMsg []byte,err error){
	return encryptedMsg,nil
}
