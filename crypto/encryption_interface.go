// Encryption interface defined
// author: Lizhong kuang
// date: 2016-08-24
// last modified:2016-08-25
package crypto


type Encryption interface {
	Sign(hash []byte,  prv interface{})(sig []byte, err error)
	UnSign(args ...interface{})([]byte, error)
}


