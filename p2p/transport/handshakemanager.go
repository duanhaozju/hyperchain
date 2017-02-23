//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
/**
author:zhangkejie
log:重新实现新的HandShakeManager
*/
package transport

import (
	"crypto/elliptic"
	"crypto"
	"encoding/hex"
	"hyperchain/core/crypto/primitives"
	"hyperchain/p2p/transport/ecdh"
	//"crypto/aes"
	//"crypto/cipher"
	"crypto/ecdsa"
	"errors"
	"hyperchain/admittance"
	"sync"
)

type HandShakeManagerNew struct {
	e          ecdh.ECDH
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey
	secrets       map[string][]byte
	signPublickey map[string][]byte
	isVerified    map[string]bool
	verifiedLock sync.RWMutex
	sharedSecretLock sync.RWMutex


 }

//---------------------------------ECDH-------------------------------------------
func NewHandShakeMangerNew(cm *admittance.CAManager) *HandShakeManagerNew {
	var hSMN HandShakeManagerNew
	hSMN.secrets = make(map[string][]byte)
	hSMN.signPublickey = make(map[string][]byte)
	hSMN.isVerified = make(map[string]bool)
	hSMN.e = ecdh.NewEllipticECDH(elliptic.P256())
	//若无私钥，相当于无ecert,但为确保节点启动，自动生产公私钥对
		//var pri *ecdsa.PrivateKey
	contentPri := cm.GetECertPrivateKeyByte()
	pri, err1 := primitives.ParseKey(string(contentPri))
	privateKey := pri.(*ecdsa.PrivateKey)
	//cert := primitives.ParseCertificate(contenrPub)
	if err1 != nil {
		panic("Parse PrivateKey or Ecert failed,please check the privateKey or Ecert and restart the node!")
	} else {
		hSMN.privateKey = privateKey
		hSMN.publicKey = (*privateKey).PublicKey
	}
	return &hSMN
}
func (hSMN *HandShakeManagerNew) GetLocalPublicKey() []byte {
	return hSMN.e.Marshal(hSMN.publicKey)
}
func (hSMN *HandShakeManagerNew) GenerateSecret(remotePublicKey []byte, peerHash string) error {
	remotePubKey, success := hSMN.e.Unmarshal(remotePublicKey)
	if !success {
		log.Error("unmarshal the share public key failed")
		return errors.New("unmarshal remote share publc fey failed")
	}
	var err error
	hSMN.sharedSecretLock.Lock()
	hSMN.secrets[peerHash], err = hSMN.e.GenerateSharedSecret(hSMN.privateKey, remotePubKey)
	hSMN.sharedSecretLock.Unlock()
	if err != nil {
		log.Error("Generate share secret failed!", err)
		return err
	} else {
		return nil
	}
}

func (hSMN *HandShakeManagerNew) SetSignPublicKey(payload []byte, peerHash string) {
	hSMN.signPublickey[peerHash] = payload
}

func (hSMN *HandShakeManagerNew) GetSignPublicKey(peerHash string) []byte {
	if pub, ok := hSMN.signPublickey[peerHash]; ok {
		return pub
	} else {
		log.Error("无法取得相应公钥", peerHash)
		return nil
	}
}

func (hSMN *HandShakeManagerNew) GetIsVerified(peerHash string) bool {
	if bol, ok := hSMN.isVerified[peerHash]; ok {
		return bol
	} else {
		log.Error("this node has not verified!", peerHash)
		return false
	}

}

func (hSMN *HandShakeManagerNew) SetIsVerified(is_verified bool, peerHash string) {
	hSMN.verifiedLock.Lock()
	defer hSMN.verifiedLock.Unlock()
	hSMN.isVerified[peerHash] = is_verified
}

func (hSMN *HandShakeManagerNew) EncWithSecret(message []byte, peerHash string) ([]byte, error) {

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
	//if _, ok := hSMN.secrets[peerHash]; !ok {
	//	//panic("the peer hasn't negotiate the share secret, and please restart this node")
	//
	//	return []byte(""), errors.New("the peer hasn't negotiate the share secret, and please restart this node")
	//}
	//key := hSMN.secrets[peerHash][:16]
	//var iv = []byte(key)[:aes.BlockSize]
	//encrypted := make([]byte, len(message))
	//aesBlockEncrypter, _ := aes.NewCipher(key)
	//aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	//aesEncrypter.XORKeyStream(encrypted, []byte(message))
	//return encrypted, nil
	//return message
	return message, nil

}

func (hSMN *HandShakeManagerNew) DecWithSecret(message []byte, peerHash string) ([]byte, error) {

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
	//if _, ok := hSMN.secrets[peerHash]; !ok {
	//	//panic("the peer hasn't negotiate the share secret, and please restart this node")
	//	return []byte(""), errors.New("the peer hasn't negotiate the share secret, and please restart this node")
	//}
	//key := hSMN.secrets[peerHash][:16]
	//var iv = []byte(key)[:aes.BlockSize]
	//decrypted := make([]byte, len(message))
	//aesBlockDecrypter, _ := aes.NewCipher([]byte(key))
	//aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	//aesDecrypter.XORKeyStream(decrypted, message)
	//return decrypted, nil
	//return message
	return message, nil
}

func (this *HandShakeManagerNew) GetSecret(peerHash string) string {
	if sc, ok := this.secrets[peerHash]; ok {
		return hex.EncodeToString(sc)
	} else {
		log.Warning("cannot get the share secret,ignore.", peerHash)
		return ""
	}

}

func (this *HandShakeManagerNew) GetSceretPoolSize() int {
	return len(this.secrets)
}

