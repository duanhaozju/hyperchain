// Cryptographically Secure Pseudo-Random Number Generator
package csprng

import (
	crand "crypto/rand"
	"io"
)

var Reader io.Reader = &randEntropy{}

type randEntropy struct {
}

func (*randEntropy) Read(bytes []byte) (n int, err error) {
	readBytes := CSPRNGRandom(len(bytes))
	copy(bytes, readBytes)
	return len(bytes), nil
}

func CSPRNGRandom(n int) []byte {
	mainBuff := make([]byte, n)
	_, err := io.ReadFull(crand.Reader, mainBuff)
	if err != nil {
		panic("reading from crypto/rand failed: " + err.Error())
	}
	return mainBuff
}
