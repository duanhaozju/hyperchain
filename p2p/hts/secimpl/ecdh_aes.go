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
	"crypto/cipher"
	"crypto/aes"
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

func(ea *ECDHWithAES)Encrypt(key, originMsg []byte)(encryptedMsg []byte,err error){
	return AesEnc(key,originMsg)
}
func(ea *ECDHWithAES)Decrypt(key, encryptedMsg []byte)(originMsg []byte,err error){
	return AesDec(key,encryptedMsg)
}

func AesEnc(key,src []byte)([]byte,error){
	if len(key) != 32 {
		return nil, errors.New("the secret len must be 32")
	}

	block,err := aes.NewCipher(key)
	if err!= nil{
		return nil,err
	}
	msg := PKCS5Padding(src,block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block,key[:block.BlockSize()])
	crypted := make([]byte,len(msg))
	blockMode.CryptBlocks(crypted,msg)
	return crypted,nil
}

func AesDec(key,src []byte)([]byte,error){
	if len(key) != 32 {
		return nil, errors.New("the secret len must be 32")
	}
	block,err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:block.BlockSize()])
	//log.Criticalf("dec block size:%d , src len %d, %d",blockMode.BlockSize(),len(src),len(src)%block.BlockSize())
	origData := make([]byte, len(src))
	blockMode.CryptBlocks(origData, src)
	origData = PKCS5UnPadding(origData)
	return origData, nil
}