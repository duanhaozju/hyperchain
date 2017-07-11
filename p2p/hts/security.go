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
	GeneratePrivateKey(io.Reader)(crypto.PrivateKey,crypto.PublicKey,error)
	Marshal(crypto.PublicKey)[]byte
	UnMarshal([]byte)(crypto.PublicKey,error)
	GenerateSharedKey(crypto.PrivateKey,crypto.PublicKey)(sharedKey []byte,err error)
}

type SymmetricSecurity interface {
	Encrypt(key, originMsg []byte)(encryptedMsg []byte,err error)
	Decrypt(key, encryptedMsg []byte)(originMsg []byte,err error)
}
