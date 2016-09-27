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

	//var signingPubKey = []byte(`-----BEGIN PUBLIC KEY-----
	//MIICIDANBgkqhkiG9w0BAQEFAAOCAg0AMIICCAKCAgEApSmU3y4DzPhjnpOrdpPs
	//cIosWJ4zSV8h02b0abLW6nk7cnb5jSwBZKLrryAlF4vs+cF1mtMYjX0QKtEYq2V6
	//WVDnoXj3BeLYVbhsHuvxYmwXmAkNsSnhMfSCxsck9y6zuNeH0ovzBD90nISIJw+c
	//VAnUt0dzc7YKjBqThHRAvi8HoGZlzB7Ryb8ePSW+Mfr4jcH3Mio5T0OH3HTavN6Y
	//zpnohzQo0blwtwEXZOwrNPjQNrSigdPDrtvM32+hLTIJ75Z2NbIRLBjNlwznu7dQ
	//Asb/AiPTHXihxCRDm+dH70dps5JfT5Zg9LKsPhANk6fNK3e4wdN89ybQsBaswp9h
	//xzORVD3UiG4LuqP4LMCadjoEazShEiiveeRBgyiFlIldybuPwSq/gUuFveV5Jnqt
	//txNG6DnJBlIeYhVlA25XDMjxnJ3w6mi/pZyn9ZR9+hFic7Nm1ra7hRUoigfD/lS3
	//3AsDoRLy0xZqCWGRUbkhlo9VjDxo5znjv870Td1/+fp9QzSaESPfFAUBFcykDXIU
	//f1nVeKAkmhkEC9/jGF+VpUsuRV3pjjrLMcuI3+IimfWhWK1C56JJakfT3WB6nwY3
	//A92g4fyVGaWFKfj83tTNL2rzMkfraExPEP+VGesr8b/QMdBlZRR4WEYG3ObD2v/7
	//jgOS2Ol4gq8/QdNejP5J4wsCAQM=
	//-----END PUBLIC KEY-----`)
	//
	//block, _ := pem.Decode(signingPubKey)
	//if block == nil {
	//	fmt.Errorf("expected block to be non-nil %v", block)
	//	return
	//}
	//
	//var pubkey dsa.PublicKey
	//
	//_,err := asn1.Unmarshal(block.Bytes, &pubkey)
	//if (err != nil ){
	//	fmt.Errorf("could not unmarshall data: `%s`", err)
	//}
	//
	//fmt.Printf("public key param P: %d\n", pubkey.Parameters.P)
	//fmt.Printf("public key param Q: %d\n", pubkey.Parameters.Q)
	//fmt.Printf("public key param G: %d\n", pubkey.Parameters.G)
	//fmt.Printf("public key Y: %d\n", pubkey.Y)
	//
	//fmt.Println("done")
	//
	//
	//
	//// todo 得到一个靠谱的取得公钥的方法
	//// 目前以简单方式实现
	//return pubkey
	//ecdsa.GenerateKey(elliptic.Curve())
	return hSM.e.Marshal(hSM.publicKey)
}
func (hSM *HandShakeManager) GenerateSecret(remotePublicKey []byte, peerHash string) {
	remotePubKey, _ := hSM.e.Unmarshal(remotePublicKey)
	hSM.secrets[peerHash], _ = hSM.e.GenerateSharedSecret(hSM.privateKey, remotePubKey)
}

func (hSM *HandShakeManager) EncWithSecret(message []byte,peerHash string) []byte {
	log.Critical("enc len",len(hSM.secrets[peerHash]))
	log.Critical("enc len2",len(hSM.GetSecret(peerHash)))
	key := hSM.secrets[peerHash][:16]
	var iv = []byte(key)[:aes.BlockSize]
	encrypted := make([]byte, len(message))
	aesBlockEncrypter, _ := aes.NewCipher(key)
	aesEncrypter := cipher.NewCFBEncrypter(aesBlockEncrypter, iv)
	aesEncrypter.XORKeyStream(encrypted, []byte(message))
	//return encrypted
	return message
}

func (hSM *HandShakeManager) DecWithSecret(message []byte,peerHash string) []byte {
	key := hSM.secrets[peerHash][:16]
	var iv = []byte(key)[:aes.BlockSize]
	decrypted := make([]byte, len(message))
	aesBlockDecrypter, _ := aes.NewCipher([]byte(key))
	aesDecrypter := cipher.NewCFBDecrypter(aesBlockDecrypter, iv)
	aesDecrypter.XORKeyStream(decrypted, message)
	//return decrypted
	return message
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