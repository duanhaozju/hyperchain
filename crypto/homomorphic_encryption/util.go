package homomorphic_encryption

import (
	"bytes"
	"encoding/gob"
	"encoding/hex"
	//"io"
	"io/ioutil"
	//"os"
)

func Encode(data interface{}) ([]byte, error) {
	buf := bytes.NewBuffer(nil)
	enc := gob.NewEncoder(buf)
	err := enc.Encode(data)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func Decode(data []byte, to interface{}) error {
	buf := bytes.NewBuffer(data)
	dec := gob.NewDecoder(buf)
	return dec.Decode(to)
}

func Getnumber(file string) ([]byte, error) {
	b, err := ioutil.ReadFile(file)
	N, err := hex.DecodeString(string(b))

	return N, err
}
