package crypto


type Encryption interface {
	Sign(hash []byte,  prv interface{})(sig []byte, err error)
	UnSign()([]byte, error)
}
