//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pbft

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
		//logger.Errorf("Asked to hash non-supported message type, ignoring")
		return ""
	}

	return base64.StdEncoding.EncodeToString(computeCryptoHash(raw))
}

func computeCryptoHash(data []byte) (hash []byte) {
	hash = make([]byte, 64)
	sha3.ShakeSum256(hash, data)
	return
}

func hashByte(data []byte) (string) {
	return base64.StdEncoding.EncodeToString(computeCryptoHash(data))
}

func byteToString(data []byte) (re string) {

	re = base64.StdEncoding.EncodeToString(data)
	return
}

func stringToByte(data string) (re []byte, err error) {

	re, err = base64.StdEncoding.DecodeString(data)
	return
}
