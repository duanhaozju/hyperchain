package pbft

import (
	"encoding/base64"

	"github.com/golang/protobuf/proto"
	"github.com/op/go-logging"
	"golang.org/x/crypto/sha3"
)

var logger *logging.Logger

func hash(msg interface{}) string {
	var raw []byte
	switch converted := msg.(type) {
	case *Request:
		raw, _ = proto.Marshal(converted)
	case *RequestBatch:
		raw, _ = proto.Marshal(converted)
	default:
		logger.Errorf("Asked to hash non-supported message type, ignoring")
		return ""
	}
	return base64.StdEncoding.EncodeToString(computeCryptoHash(raw))

}

func computeCryptoHash(data []byte) (hash []byte) {
	hash = make([]byte, 64)
	sha3.ShakeSum256(hash, data)
	return
}