//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package rbft

import (
	"encoding/base64"

	"hyperchain/core/types"

	"github.com/golang/protobuf/proto"
	"golang.org/x/crypto/sha3"
)

// encode the []byte to string
func hash(msg interface{}) string {
	var raw []byte
	switch converted := msg.(type) {
	case *types.Transaction:
		raw, _ = proto.Marshal(converted)
	case *TransactionBatch:
		raw, _ = proto.Marshal(converted)
	default:
		return ""
	}

	return base64.StdEncoding.EncodeToString(computeCryptoHash(raw))
}

// computeCryptoHash computes hash using sha256
func computeCryptoHash(data []byte) (hash []byte) {
	hash = make([]byte, 64)
	sha3.ShakeSum256(hash, data)
	return
}

// hashByte returns the base64 encoded result of CryptoHash
func hashByte(data []byte) string {
	return base64.StdEncoding.EncodeToString(computeCryptoHash(data))
}

// byteToString returns the base64 encoded result of data
func byteToString(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}
