package secimpl

import (
	"crypto"
	"crypto/ecdsa"
	"certgen/primitives"
	"errors"
	"math/big"
	"encoding/asn1"
	"fmt"
	"crypto/elliptic"
	"github.com/ethereum/go-ethereum/crypto/sha3"
)
// this secimpl implements the Security interface

type ECDHWithAES struct{
}
func NewECDHWithAES()*ECDHWithAES{
	return &ECDHWithAES{}
}

func (ea *ECDHWithAES)VerifySign(sign,data,rawcert []byte)(bool,error){
	cert,err := primitives.ParseCertificate(rawcert)
	if err != nil{
		return false,err
	}
	pubkey,ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false,errors.New(fmt.Sprintf("cannot complete the type conversationm reason: %s",err.Error()))
	}
	hasher := crypto.SHA3_256.New()
	hasher.Write(data)
	hash := hasher.Sum(nil)
	ecdsasign := struct {
		R,S *big.Int
	}{}
	_,err = asn1.Unmarshal(sign,&ecdsasign)
	if err != nil{
		return false,err
	}
	return ecdsa.Verify(pubkey,hash,ecdsasign.R,ecdsasign.S),nil
}

func(ea *ECDHWithAES)GenerateShareKey(priKey []byte,rand []byte,rawcert []byte)(sharedKey []byte,err error){
	prikey,err := primitives.ParsePriKey(priKey)
	if err != nil{
		return nil,err
	}
	pri,ok := prikey.(*ecdsa.PrivateKey)
	if !ok{
		return nil,errors.New(fmt.Sprintf("cannot complete the type conversationm reason: %s",err.Error()))
	}
	cert,err := primitives.ParseCertificate(rawcert)
	if err != nil{
		return nil,err
	}
	pubkey,ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil,errors.New(fmt.Sprintf("cannot complete the type conversationm reason: %s",err.Error()))
	}

	var sharekey = make([]byte,0)

	//negotiate shared key
	curve := elliptic.P256()
	x,y := curve.ScalarMult(pubkey.X,pubkey.Y,pri.X.Bytes())
	sharekey = append(sharedKey,x.Bytes()...)
	sharekey = append(sharedKey,y.Bytes()...)
	sharekey = append(sharedKey,rand...)
	hasher := sha3.New256()
	hasher.Write(sharekey)
	sharedKey = hasher.Sum(nil)
	return sharedKey[:], nil

}

func(ea *ECDHWithAES)Encrypt(key, originMsg []byte)(encryptedMsg []byte,err error){
	return originMsg,nil
}
func(ea *ECDHWithAES)Decrypt(key, encryptedMsg []byte)(originMsg []byte,err error){
	return encryptedMsg,nil
}
