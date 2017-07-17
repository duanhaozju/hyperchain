// Security defined the security layer issues
// includes the Diffie-Hellman key Exchange algorithm
// generally algorithm is ECDH or SM2DH
// and symmetric Security:
// generally algorithm is AES/3DES/SM4
//
package hts

import (
	"crypto"
	"io"
)

type Security interface {
	KeyExchange
	SymmetricSecurity
}

type KeyExchange interface {

	// VerifySign verify the sign is generate by this cert
	// this method will not check the cert is valid
	//verify the cert and it's signature
	VerifySign(sign []byte,cert []byte)(bool,error)
	//GenerateShareKey generate the shared key
	//priKey self side private key
	//rand []byte CSPRN
	//cert remote certificate
	GenerateShareKey(priKey []byte,rand []byte,cert []byte)(sharedKey []byte,err error)
}

type SymmetricSecurity interface {
	Encrypt(key, originMsg []byte)(encryptedMsg []byte,err error)
	Decrypt(key, encryptedMsg []byte)(originMsg []byte,err error)
}
