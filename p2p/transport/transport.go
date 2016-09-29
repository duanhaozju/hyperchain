// p2p handshake and communicate
// author: Lizhong kuang
// date: 2016-09-08

package transport

import (
	"crypto/cipher"
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
	GenerateSecret(remotePublicKey []byte, peerHash string)
	EncWithSecret(message []byte,peerHash string) []byte
	DecWithSecret(message []byte,peerHash string) []byte
	GetSecret(peerHash string)string
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
func NewHandShakeManger() *HandShakeManager{
	var hSM HandShakeManager
	hSM.secrets = make(map[string][]byte)
	hSM.e = ecdh.NewEllipticECDH(elliptic.P384())
	hSM.privateKey, hSM.publicKey, _ = hSM.e.GenerateKey(rand.Reader)
	return &hSM
}
func (hSM *HandShakeManager) GetLocalPublicKey() []byte {
	return hSM.e.Marshal(hSM.publicKey)
}
func (hSM *HandShakeManager) GenerateSecret(remotePublicKey []byte, peerHash string) {
	remotePubKey, _ := hSM.e.Unmarshal(remotePublicKey)
	hSM.secrets[peerHash], _ = hSM.e.GenerateSharedSecret(hSM.privateKey, remotePubKey)
}

func (hSM *HandShakeManager) EncWithSecret(message []byte,peerHash string) []byte {
	key := hSM.secrets[peerHash][:16]
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(message))
	aesBlockEncrypter, _ := aes.NewCipher(key)
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(message))
	return encrypted
	//return message
}

func (hSM *HandShakeManager) DecWithSecret(message []byte,peerHash string) []byte {
	key := hSM.secrets[peerHash][:16]
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(message))
	aesBlockDecrypter, _ := aes.NewCipher([]byte(key))
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, message)
	return decrypted
	//return message
}

func (this *HandShakeManager) GetSecret(peerHash string)string{
	if sc , ok := this.secrets[peerHash];ok{
		return hex.EncodeToString(sc)
	}else{
		log.Error("无法取得相应秘钥", peerHash)
		return ""
	}

}
func (this *HandShakeManager) GetSceretPoolSize()int{
	return len(this.secrets)
}

func (this *HandShakeManager) PrintAllSecHash(){
	for hash,_ := range this.secrets{
		log.Notice(hash)
	}

}