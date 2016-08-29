// Encryption interface defined
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package crypto

import "crypto/ecdsa"

type Encryption interface {

	//sign byte
	Sign(hash []byte,  prv interface{})(sig []byte, err error)
	UnSign(args ...interface{})([]byte, error)
	//general private key and save into file
	//GeneralKey(path string) (interface{},error)
	GeneralKey(path string)(ecdsa.PrivateKey,error)
	GetKey()(*ecdsa.PrivateKey,error)

}


