// p2p handshake and communicate
// author: Lizhong kuang
// date: 2016-09-08

package transport

import (
	"crypto/cipher"
	"crypto/des"
	"bytes"
	"github.com/op/go-logging"
	"crypto/aes"
	"crypto/elliptic"

	"hyperchain/p2p/transport/ecdh"
	"crypto/rand"
	"crypto"
	"encoding/hex"
)
var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p/transport")
}

type TransportEncryptManager interface {
	GetLocalPublicKey() []byte
	GenerateSecret(remotePublicKey []byte)
	EncWithSecret(message []byte) []byte
	DecWithSecret(message []byte) []byte
	GetSecret() string
}

type HandShakeManager struct {
	e            ecdh.ECDH
	privateKey   crypto.PrivateKey
	publicKey    crypto.PublicKey
	remotePubKey crypto.PublicKey
	secret       []byte
}

//---------------------------------3des---------------------------------------
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
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}


//---------------------------------ECDH-------------------------------------------
func NewHandShakeManger() *HandShakeManager{
	var hSM HandShakeManager
	hSM.e = ecdh.NewEllipticECDH(elliptic.P384())
	hSM.privateKey, hSM.publicKey, _ = hSM.e.GenerateKey(rand.Reader)
	return &hSM
}
func (hSM *HandShakeManager) GetLocalPublicKey() []byte {
	return hSM.e.Marshal(hSM.publicKey)
}
func (hSM *HandShakeManager) GenerateSecret(remotePublicKey []byte) {
	remotePubKey, _ := hSM.e.Unmarshal(remotePublicKey)
	hSM.secret, _ = hSM.e.GenerateSharedSecret(hSM.privateKey, remotePubKey)
}

func (hSM *HandShakeManager) EncWithSecret(message []byte) []byte {
	key := hSM.secret[:16]
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(message))
	aesBlockEncrypter, _ := aes.NewCipher(key)
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(message))
	return encrypted
}

func (hSM *HandShakeManager) DecWithSecret(message []byte) []byte {
	key := hSM.secret[:16]
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(message))
	aesBlockDecrypter, _ := aes.NewCipher([]byte(key))
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, message)
	return decrypted
}

func (this *HandShakeManager) GetSecret()string{
	return hex.EncodeToString(this.secret)
}