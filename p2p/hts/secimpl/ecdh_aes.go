package secimpl

import (
	"crypto"
	"io"
)
// this secimpl implements the Security interface

type ECDHWithAES struct{
}

// ExchangeKey part
func(ea *ECDHWithAES)GeneratePrivateKey(io.Reader)(crypto.PrivateKey,crypto.PublicKey,error){

}

func(ea *ECDHWithAES)Marshal(crypto.PublicKey)[]byte{

}

func(ea *ECDHWithAES)UnMarshal([]byte)(crypto.PublicKey,error){

}

func(ea *ECDHWithAES)GenerateSharedKey(crypto.PrivateKey,crypto.PublicKey)(sharedKey []byte,err error){

}

