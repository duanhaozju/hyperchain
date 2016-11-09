//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package homomorphic_encryption

type Homomorphic_Encrypton interface {

	//make a new key and return a PaillierKey
	NewKey() (interface{}, error)

	//encrypto a message by publickey
	Encrypto_message(publickey interface{}, message []byte) ([]byte, error)

	//decrypto a message use a PaillierKey
	Decrypto_Ciphertext(key interface{}, emessage []byte) []byte

	//calculate  ciphertexts
	Calculator(publickey interface{}, name string, args ...[]byte) ([]byte, error)
}
