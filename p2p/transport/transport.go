//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package transport

import (
	"crypto/cipher"
	"crypto/elliptic"
	"github.com/op/go-logging"

	"bytes"
	"crypto"
	//"crypto/aes"
	"crypto/des"
	"crypto/rand"
	"encoding/hex"
	"hyperchain/p2p/transport/ecdh"
	//"crypto/aes"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/transport")
}

type TransportEncryptManager interface {
	GetLocalPublicKey() []byte
	GenerateSecret(remotePublicKey []byte, peerHash string) error
	EncWithSecret(message []byte, peerHash string) []byte
	DecWithSecret(message []byte, peerHash string) []byte
	GetSecret(peerHash string) string
	GetSceretPoolSize() int
	PrintAllSecHash()
}

type HandShakeManager struct {
	e            ecdh.ECDH
	privateKey   crypto.PrivateKey
	publicKey    crypto.PublicKey
	remotePubKey crypto.PublicKey
	secrets      map[string][]byte
}

//---------------------------------ECDH-------------------------------------------
func NewHandShakeManger() *HandShakeManager {
	var hSM HandShakeManager
	hSM.secrets = make(map[string][]byte)
	hSM.e = ecdh.NewEllipticECDH(elliptic.P384())
	var err error
	hSM.privateKey, hSM.publicKey, err = hSM.e.GenerateKey(rand.Reader)
	if err != nil{
		panic("Generate key failed, please restart the node!")
	}
	return &hSM
}
func (hSM *HandShakeManager) GetLocalPublicKey() []byte {
	return hSM.e.Marshal(hSM.publicKey)
}
func (hSM *HandShakeManager) GenerateSecret(remotePublicKey []byte, peerHash string) error {
	remotePubKey, _ := hSM.e.Unmarshal(remotePublicKey)
	var err error
	hSM.secrets[peerHash], err = hSM.e.GenerateSharedSecret(hSM.privateKey, remotePubKey)
	if err != nil {
		log.Error("Generate share secret failed!", err)
		return err
	} else {
		return nil
	}
}

func (hSM *HandShakeManager) EncWithSecret(message []byte, peerHash string) []byte {

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
	//if _,ok := hSM.secrets[peerHash];!ok{
	//	panic("the peer hasn't negotiate the share secret, and please restart this node")
	//	return []byte("")
	//}
	//key := hSM.secrets[peerHash][:16]
	//var iv = []byte(key)[:aes.BlockSize]
	//encrypted := make([]byte, len(message))
	//aesBlockEncrypter, _ := aes.NewCipher(key)
	//aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	//aesEncrypter.XORKeyStream(encrypted, []byte(message))
	//return encrypted
	return message

}

func (hSM *HandShakeManager) DecWithSecret(message []byte, peerHash string) []byte {

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
	//if _,ok := hSM.secrets[peerHash];!ok{
	//	panic("the peer hasn't negotiate the share secret, and please restart this node")
	//	return []byte("")
	//}
	//key := hSM.secrets[peerHash][:16]
	//var iv = []byte(key)[:aes.BlockSize]
	//decrypted := make([]byte, len(message))
	//aesBlockDecrypter, _ := aes.NewCipher([]byte(key))
	//aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	//aesDecrypter.XORKeyStream(decrypted, message)
	//return decrypted
	return message

}

func (this *HandShakeManager) GetSecret(peerHash string) string {
	if sc, ok := this.secrets[peerHash]; ok {
		return hex.EncodeToString(sc)
	} else {
		log.Error("无法取得相应秘钥", peerHash)
		return ""
	}

}

func (this *HandShakeManager) GetSceretPoolSize() int {
	return len(this.secrets)
}

func (this *HandShakeManager) PrintAllSecHash() {
	for hash, _ := range this.secrets {
		log.Notice(hash)
	}

}

// 3DES加密
func TripleDesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:8])
	crypted := make([]byte, len(origData))
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}
func DesEncrypt(origData, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	origData = PKCS5Padding(origData, block.BlockSize())
	// origData = ZeroPadding(origData, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key)
	crypted := make([]byte, len(origData))
	// 根据CryptBlocks方法的说明，如下方式初始化crypted也可以
	// crypted := origData
	blockMode.CryptBlocks(crypted, origData)
	return crypted, nil
}

func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

// 3DES解密
func TripleDesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewTripleDESCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key[:8])
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}
func DesDecrypt(crypted, key []byte) ([]byte, error) {
	block, err := des.NewCipher(key)
	if err != nil {
		return nil, err
	}
	blockMode := cipher.NewCBCDecrypter(block, key)
	origData := make([]byte, len(crypted))
	// origData := crypted
	blockMode.CryptBlocks(origData, crypted)
	origData = PKCS5UnPadding(origData)
	// origData = ZeroUnPadding(origData)
	return origData, nil
}
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
