package secimpl

import (
	"bytes"
	"crypto/cipher"
	"crypto/des"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/asn1"
	"errors"
	"fmt"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/crypto/primitives"
	"github.com/hyperchain/hyperchain/crypto/sha3"
	"math/big"
)

// this secimpl implements the Security interface

type ECDHWith3DES struct {
}

func NewECDHWith3DES() *ECDHWith3DES {
	return &ECDHWith3DES{}
}

func (ea *ECDHWith3DES) VerifySign(sign, data, rawcert []byte) (bool, error) {
	cert, err := primitives.ParseCertificate(rawcert)
	if err != nil {
		return false, err
	}
	pubkey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return false, errors.New(fmt.Sprintf("cannot complete the type conversationm reason: %s", err.Error()))
	}
	hash := crypto.Keccak256(data)
	ecdsasign := struct {
		R, S *big.Int
	}{}
	_, err = asn1.Unmarshal(sign, &ecdsasign)
	if err != nil {
		return false, err
	}
	return ecdsa.Verify(pubkey, hash, ecdsasign.R, ecdsasign.S), nil
}

func (ea *ECDHWith3DES) GenerateShareKey(priKey []byte, rand []byte, rawcert []byte) (sharedKey []byte, err error) {
	prikey, err := primitives.ParseKey(priKey)
	if err != nil {
		return nil, err
	}
	pri, ok := prikey.(*ecdsa.PrivateKey)
	if !ok {
		return nil, errors.New(fmt.Sprintf("cannot complete the type conversationm reason: %s", err.Error()))
	}
	cert, err := primitives.ParseCertificate(rawcert)
	if err != nil {
		return nil, err
	}
	pubkey, ok := cert.PublicKey.(*ecdsa.PublicKey)
	if !ok {
		return nil, errors.New(fmt.Sprintf("cannot complete the type conversationm reason: %s", err.Error()))
	}

	var sharekey = make([]byte, 0)
	//negotiate shared key
	curve := elliptic.P256()
	x, y := curve.ScalarMult(pubkey.X, pubkey.Y, pri.D.Bytes())
	//fmt.Printf("pubx : %s\n",common.ToHex(pubkey.X.Bytes()))
	//fmt.Printf("puby : %s\n",common.ToHex(pubkey.Y.Bytes()))
	//fmt.Printf("priD : %s\n",common.ToHex(pri.D.Bytes()))
	//fmt.Printf("self pubx : %s\n",common.ToHex(pri.X.Bytes()))
	//fmt.Printf("self puby : %s\n",common.ToHex(pri.Y.Bytes()))
	//fmt.Printf("x : %s\n",common.ToHex(x.Bytes()))
	//fmt.Printf("y : %s\n",common.ToHex(y.Bytes()))
	//fmt.Printf("rand : %s\n",common.ToHex(rand))

	sharekey = append(sharekey, x.Bytes()...)
	sharekey = append(sharekey, y.Bytes()...)
	sharekey = append(sharekey, rand...)
	hasher := sha3.NewKeccak256()
	hasher.Write(sharekey)
	sharedKey = hasher.Sum(nil)
	return sharedKey[:], nil

}

func (ea *ECDHWith3DES) Encrypt(key, originMsg []byte) (encryptedMsg []byte, err error) {
	return TripleDesEnc(key, originMsg)
}
func (ea *ECDHWith3DES) Decrypt(key, encryptedMsg []byte) (originMsg []byte, err error) {
	return TripleDesDec(key, encryptedMsg)
}

func TripleDesEncrypt8(origData, key []byte) ([]byte, error) {
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

func TripleDesDecrypt8(crypted, key []byte) ([]byte, error) {
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
	blockMode.CryptBlocks(crypted, msg)
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
