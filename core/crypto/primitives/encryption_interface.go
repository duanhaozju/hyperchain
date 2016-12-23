package primitives


type Encryption interface {

	Sign(payload []byte,pri interface{})([]byte,error)

	VerifySign(verKey interface{}, msg, signature []byte) (bool, error)
}
