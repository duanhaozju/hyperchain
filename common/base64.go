package common

import "encoding/base64"

// DecodeBase64 decodes from Base64
func DecodeBase64(in string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(in)
}

// EncodeBase64 encodes to Base64
func EncodeBase64(in []byte) string {
	return base64.StdEncoding.EncodeToString(in)
}
