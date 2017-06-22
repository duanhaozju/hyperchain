package transport

import (
	"bytes"
	"crypto"
	"crypto/cipher"
	"crypto/des"
	"crypto/ecdsa"
	"errors"
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/admittance"
	"hyperchain/common"
	"hyperchain/core/crypto/primitives"
	pb "hyperchain/p2p/message"
	"sync"
	"hyperchain/p2p/hts/crypto/ecdh"
)

// Init the log setting
//var log *logging.Logger // package-level logger
//func init() {
//	log = logging.MustGetLogger("p2p")
//}

type sharedSecret struct {
	shareSec         []byte
	remotePubkey     crypto.PrivateKey
	remotePubkeyByte []byte
}

//Transport Manager to manage the p2p transport encrypt
type TransportManager struct {
	// ecdh elliptic curve Diffie-Hellman share secret exchange algorithm
	ecdh       ecdh.ECDH
	privateKey crypto.PrivateKey
	publicKey  crypto.PublicKey

	// share secret hash table
	shareSecTable map[string]*sharedSecret
	shareSecMux   sync.Mutex

	//symmetrical encryption algorithm callback
	encAlgo func(secret []byte, msg []byte) ([]byte, error)
	decAlgo func(secret []byte, msg []byte) ([]byte, error)

	//CAManager
	cm     *admittance.CAManager
	logger *logging.Logger
}

//NewTransportManager return a initialized transport manager
func NewTransportManager(cm *admittance.CAManager, namespace string) (*TransportManager, error) {
	logger := common.GetLogger(namespace, "p2p")
	contentPri := cm.GetECertPrivateKeyByte()
	pri, err := primitives.ParseKey(string(contentPri))
	if err != nil {
		return nil, err
	}
	privateKey := pri.(*ecdsa.PrivateKey)
	publicKey := (*privateKey).PublicKey
	var encAlgo func(key, src []byte) ([]byte, error)
	var decAlgo func(key, src []byte) ([]byte, error)
	if cm.EnableSymmetrical {
		encAlgo = TripleDesEnc
		decAlgo = TripleDesDec
	} else {
		logger.Warning("disable the Symmetrical Encryption.")
		encAlgo = pureEnc
		decAlgo = pureDec
	}
	return &TransportManager{
		privateKey:    privateKey,
		publicKey:     publicKey,
		shareSecTable: make(map[string]*sharedSecret),
		encAlgo:       encAlgo,
		decAlgo:       decAlgo,
		cm:            cm,
		logger:        logger,
	}, nil
}

// GetLocalPublicKey get local node public Key
func (tm *TransportManager) GetLocalPublicKey() []byte {
	return tm.ecdh.Marshal(tm.publicKey)
}

//NegoShareSecret negotiate share secret
func (tm *TransportManager) NegoShareSecret(remotePub []byte, addr *pb.Endpoint) error {
	remotePubKey, success := tm.ecdh.Unmarshal(remotePub)
	if !success {
		tm.logger.Error("unmarshal the remote share public key failed")
		return errors.New("unmarshal remote share publc fey failed")
	}
	tm.shareSecMux.Lock()
	defer tm.shareSecMux.Unlock()
	shareSec, err := tm.ecdh.GenerateSharedSecret(tm.privateKey, remotePubKey)
	if err != nil {
		tm.logger.Error("generate the share secret failed.")
		return errors.New("generate the share secret failed.")
	}
	tm.shareSecTable[string(addr.UUID)] = &sharedSecret{
		shareSec:         shareSec,
		remotePubkey:     remotePubKey,
		remotePubkeyByte: remotePub,
	}
	return nil
}

//Encrypt the message
func (tm *TransportManager) Encrypt(message []byte, addr *pb.Endpoint) ([]byte, error) {
	if shareSecret, ok := tm.shareSecTable[string(addr.UUID)]; ok {
		return tm.encAlgo(shareSecret.shareSec, message)
	}
	return nil, errors.New(fmt.Sprintf("can not get a shared secret for peer %d", addr.Hostname))
}

//decryption the message
func (tm *TransportManager) Decrypt(message []byte, addr *pb.Endpoint) ([]byte, error) {
	if shareSecret, ok := tm.shareSecTable[string(addr.UUID)]; ok {
		return tm.decAlgo(shareSecret.shareSec, message)
	}
	return nil, errors.New(fmt.Sprintf("can not get a shared secret for peer %d", addr.Hostname))
}

//SignMsg use the certificate's public sign the msg
func (tm *TransportManager) SignMsg(msg *pb.Message) (pb.Message, error) {
	//sign := &pb.Signature{}
	retmsg := pb.Message{
		From:         msg.From,
		Payload:      msg.Payload,
		MsgTimeStamp: msg.MsgTimeStamp,
		MessageType:  msg.MessageType,
		//Signature:    sign,
	}
	//retmsg.Signature.ECert = tm.cm.GetECertByte()
	//retmsg.Signature.RCert = tm.cm.GetRCertByte()
	if !tm.cm.IsCheckSign() {
		return retmsg, nil
	}
	ecdsaEncrypto := primitives.NewEcdsaEncrypto("ecdsa")
	_, err := ecdsaEncrypto.Sign(msg.Payload, tm.cm.GetECertPrivKey())
	// stupid bug 20170321!
	if err != nil {
		tm.logger.Critical(err)
		return retmsg, err
	}
	//retmsg.Signature.Signature = signa
	return retmsg, nil
}

// verify the msg is valid or not
func (tm *TransportManager) VerifyMsg(msg *pb.Message) (bool, error) {
	if !tm.cm.IsCheckSign() {
		return true, nil
	}
	//if msg.Signature == nil {
	//	tm.logger.Warning("The msg Signature is nil, msg from", msg.From.ID)
	//	return false, errors.New("invalid signature")
	//}
	//1. check the ECert is valid or not
	//f, e := tm.cm.VerifyECert(string(msg.Signature.ECert))
	//if e != nil {
	//	return f, e
	//}
	//2. check the signature is valid or not
	//return tm.cm.VerifyCertSign(string(msg.Signature.ECert), msg.Payload, msg.Signature.Signature)
	return true,nil
}

func (tm *TransportManager) VerifyRCert(msg *pb.Message) (bool, error) {
	//if msg.Signature == nil || msg.Signature.RCert == nil {
	//	tm.logger.Warning("The msg Signature is nil, msg from", msg.From.ID)
	//	return false, errors.New("invalid msg, signature is nil,or signature.rcert is nil")
	//}
	//rcertb := msg.Signature.RCert
	//再验证证书合法性
	//return tm.cm.VerifyRCert(string(rcertb))
	return true,nil
}

// 3DES encryption algorithm implements
func TripleDesEnc(key, src []byte) ([]byte, error) {
	if len(key) < 24 {
		return nil, errors.New("the secret len is less than 24")
	}
	block, err := des.NewTripleDESCipher(key[:24])
	if err != nil {
		return nil, err
	}
	msg := PKCS5Padding(src, block.BlockSize())
	blockMode := cipher.NewCBCEncrypter(block, key[:block.BlockSize()])
	crypted := make([]byte, len(msg))
	//log.Criticalf("block size:%d , src len %d, %d",blockMode.BlockSize(),len(msg),len(msg)%block.BlockSize())
	blockMode.CryptBlocks(crypted, msg)
	//log.Criticalf("after encrypt msg is : %s",common.ToHex(crypted))
	return crypted, nil
}

// 3DES decryption algorithm implements
func TripleDesDec(key, src []byte) ([]byte, error) {
	//log.Criticalf("to descrypt msg is : %s",common.ToHex(src))
	if len(key) < 24 {
		return nil, errors.New("the secret len is less than 24")
	}
	block, err := des.NewTripleDESCipher(key[:24])
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

// pure enc will not encrypt the message
func pureEnc(sec, msg []byte) ([]byte, error) {
	return msg, nil
}

// pure dec will return origin message
func pureDec(sec, msg []byte) ([]byte, error) {
	return msg, nil
}

//PKCS5Padding padding with pkcs5
func PKCS5Padding(ciphertext []byte, blockSize int) []byte {
	padding := blockSize - len(ciphertext)%blockSize
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

//PKCS5UnPadding unpadding with pkcs5
func PKCS5UnPadding(origData []byte) []byte {
	length := len(origData)
	// 去掉最后一个字节 unpadding 次
	unpadding := int(origData[length-1])
	return origData[:(length - unpadding)]
}
