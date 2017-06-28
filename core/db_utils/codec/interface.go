package codec

type TransactionCodec interface {
	Encode() []byte
}
