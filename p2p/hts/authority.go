package hts

var AUTH_VER = int64(13)

func NewAuthority(identify, certificate, cipher, signature, keyExchangeParams, extends []byte) *Authority {
	return &Authority{
		Version:           AUTH_VER,
		Certificate:       certificate,
		Identify:          identify,
		Signature:         signature,
		KeyExchangeParams: keyExchangeParams,
		CipherSpec:        cipher,
		Extends:           extends,
	}
}
