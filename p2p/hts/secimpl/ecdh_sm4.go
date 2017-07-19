package secimpl

import (
	"crypto"
	"crypto/ecdsa"
	"errors"
	"math/big"
	"encoding/asn1"
	"fmt"
	"crypto/elliptic"
	"hyperchain/crypto/sha3"
	"hyperchain/crypto/primitives"
	"hyperchain/crypto/guomi"
)
// this secimpl implements the Security interface

type ECDHWithSM4 struct{
}
func NewECDHWithSM4()*ECDHWithSM4{
	return &ECDHWithSM4{}
}

func (ea *ECDHWithSM4)VerifySign(sign,data,rawcert []byte)(bool,error){
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

func(ea *ECDHWithSM4)GenerateShareKey(priKey []byte,rand []byte,rawcert []byte)(sharedKey []byte,err error){
	prikey,err := primitives.ParseKey(priKey)
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
	x,y := curve.ScalarMult(pubkey.X,pubkey.Y,pri.D.Bytes())
	//fmt.Printf("pubx : %s\n",common.ToHex(pubkey.X.Bytes()))
	//fmt.Printf("puby : %s\n",common.ToHex(pubkey.Y.Bytes()))
	//fmt.Printf("priD : %s\n",common.ToHex(pri.D.Bytes()))
	//fmt.Printf("self pubx : %s\n",common.ToHex(pri.X.Bytes()))
	//fmt.Printf("self puby : %s\n",common.ToHex(pri.Y.Bytes()))
	//fmt.Printf("x : %s\n",common.ToHex(x.Bytes()))
	//fmt.Printf("y : %s\n",common.ToHex(y.Bytes()))
	//fmt.Printf("rand : %s\n",common.ToHex(rand))

	sharekey = append(sharekey,x.Bytes()...)
	sharekey = append(sharekey,y.Bytes()...)
	sharekey = append(sharekey,rand...)
	hasher := sha3.NewKeccak256()
	hasher.Write(sharekey)
	sharedKey = hasher.Sum(nil)
	return sharedKey[:], nil

}

func(ea *ECDHWithSM4)Encrypt(key, originMsg []byte)(encryptedMsg []byte,err error){
	return Sm4Encrypt(key,originMsg)
}
func(ea *ECDHWithSM4)Decrypt(key, encryptedMsg []byte)(originMsg []byte,err error){
	return Sm4Decrypt(key,encryptedMsg)
}

func Sm4Encrypt(key,originMsg []byte)([]byte, error){
	msg := PKCS5Padding(originMsg,16)
	return guomi.Sm4Enc(key,msg);
}

func Sm4Decrypt(key, src []byte) ([]byte, error) {
	msg,_ := guomi.Sm4Dec(key,src)
	msg = PKCS5UnPadding(msg)
	return msg, nil
}