//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
/**
author:zhangkejie
log:重新实现新的HandShakeManager
 */
package transport

import (
	"crypto/elliptic"
	//"github.com/op/go-logging"

	"crypto"
	//"crypto/aes"
	"crypto/rand"
	"encoding/hex"
	"hyperchain/p2p/transport/ecdh"
	"hyperchain/core/crypto/primitives"
	//"crypto/aes"
	"crypto/ecdsa"
	"crypto/aes"
	"crypto/cipher"
	"errors"
)

type HandShakeManagerNew struct {
	e            ecdh.ECDH
	privateKey   crypto.PrivateKey
	publicKey    crypto.PublicKey
	remotePubKey crypto.PublicKey
	secrets      map[string][]byte
	signPublickey map[string]crypto.PublicKey
	isVerified map[string]bool
 }

//---------------------------------ECDH-------------------------------------------
func NewHandShakeMangerNew() *HandShakeManagerNew {
	var hSMN HandShakeManagerNew
	hSMN.secrets = make(map[string][]byte)
	hSMN.signPublickey = make(map[string]crypto.PublicKey)
	hSMN.isVerified = make(map[string]bool)
	hSMN.e = ecdh.NewEllipticECDH(elliptic.P384())
	var err error
	contentPri,getErr1 := primitives.GetConfig("../../config/cert/server/eca.priv")
	//contenrPub,getErr2 := primitives.GetConfig("../../config/cert/server/eca.cert")

	//若无私钥，相当于无ecert,但为确保节点启动，自动生产公私钥对
	if(getErr1!=nil){
		hSMN.privateKey, hSMN.publicKey, err = hSMN.e.GenerateKey(rand.Reader)
		if err!=nil {
			panic("GenerateKey failed,please restart the node.")
		}
	}else {
		//var pri *ecdsa.PrivateKey
		pri,err1 := primitives.ParseKey(contentPri)
		privateKey := pri.(*ecdsa.PrivateKey)
		//cert := primitives.ParseCertificate(contenrPub)
		if err1!=nil {
			panic("Parse PrivateKey or Ecert failed,please check the privateKey or Ecert and restart the node!")
		}else {
			hSMN.privateKey = privateKey
			hSMN.publicKey = (*privateKey).PublicKey
		}
	}
	return &hSMN
}
func (hSMN *HandShakeManagerNew) GetLocalPublicKey() []byte {
	return hSMN.e.Marshal(hSMN.publicKey)
}
func (hSMN *HandShakeManagerNew) GenerateSecret(remotePublicKey []byte, peerHash string) error {
	remotePubKey, _ := hSMN.e.Unmarshal(remotePublicKey)
	var err error
	hSMN.secrets[peerHash], err = hSMN.e.GenerateSharedSecret(hSMN.privateKey, remotePubKey)
	if err != nil {
		log.Error("Generate share secret failed!", err)
		return err
	} else {
		return nil
	}
}

func (hSMN *HandShakeManagerNew) SetSignPublicKey(pub crypto.PublicKey,peerHash string){
	hSMN.isVerified[peerHash] = pub
}

func (hSMN *HandShakeManagerNew) SetIsVerified(is_verified bool,peerHash string){
	hSMN.signPublickey[peerHash] = is_verified
}

func (hSMN *HandShakeManagerNew) EncWithSecret(message []byte, peerHash string)  ([]byte,error) {

	// 3DES
	//key := []byte("sfe023f_sefiel#fi32lf3e!")
	////log.Critical("密钥长度",len(key))
	//
	//
	//encrypted,err := TripleDesEncrypt(message,key)
	//if err !=nil{
	//	log.Error(err)
	//	return nil
	//}
	//return encrypted


	//aes
	if _,ok := hSMN.secrets[peerHash];!ok{
		//panic("the peer hasn't negotiate the share secret, and please restart this node")

		return []byte(""),errors.New("the peer hasn't negotiate the share secret, and please restart this node")
	}
	key := hSMN.secrets[peerHash][:16]
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(message))
	aesBlockEncrypter, _ := aes.NewCipher(key)
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(message))
	return encrypted,nil
	//return message

}

func (hSMN *HandShakeManagerNew) DecWithSecret(message []byte, peerHash string)  ([]byte,error){

	//3DES
	//key := []byte("sfe023f_sefiel#fi32lf3e!")
	////log.Critical("密钥长度",len(key))
	//decrypted,err := TripleDesDecrypt(message,key)
	//if err !=nil{
	//	log.Error(err)
	//	return nil
	//}
	//return decrypted

	//aes
	//
	if _,ok := hSMN.secrets[peerHash];!ok{
		//panic("the peer hasn't negotiate the share secret, and please restart this node")
		return []byte(""),errors.New("the peer hasn't negotiate the share secret, and please restart this node")
	}
	key := hSMN.secrets[peerHash][:16]
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(message))
	aesBlockDecrypter, _ := aes.NewCipher([]byte(key))
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, message)
	return decrypted,nil
	//return message

}

func (this *HandShakeManagerNew) GetSecret(peerHash string) string {
	if sc, ok := this.secrets[peerHash]; ok {
		return hex.EncodeToString(sc)
	} else {
		log.Error("无法取得相应秘钥", peerHash)
		return ""
	}

}

func (this *HandShakeManagerNew) GetSceretPoolSize() int {
	return len(this.secrets)
}

func (this *HandShakeManagerNew) PrintAllSecHash() {
	for hash, _ := range this.secrets {
		log.Notice(hash)
	}

}


