package bucket

import (
	"bytes"
	"crypto/sha1"
	"encoding/binary"
	"fmt"
	"github.com/golang/protobuf/proto"
	"hyperchain/crypto"
)

var stateKeyDelimiter = []byte{0x00}

// ComputeCryptoHash should be used in openchain code so that we can change the actual algo used for crypto-hash at one place
func ComputeCryptoHashWithTrim(data []byte) []byte {
	if data == nil || len(data) == 0 {
		return nil
	}
	h := sha1.New()
	h.Write([]byte(data))
	bs := h.Sum(nil)[:5]
	return bs
}

// ComputeCryptoHash should be used in openchain code so that we can change the actual algo used for crypto-hash at one place
func ComputeCryptoHash(data []byte) []byte {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	hash := kec256Hash.Hash(data)
	return hash.Bytes()
}

// EncodeOrderPreservingVarUint64 returns a byte-representation for a uint64 number such that
// all zero-bits starting bytes are trimmed in order to reduce the length of the array
// For preserving the order in a default bytes-comparison, first byte contains the number of remaining bytes.
// The presence of first byte also allows to use the returned bytes as part of other larger byte array such as a
// composite-key representation in db
func EncodeOrderPreservingVarUint64(number uint64) []byte {
	bytes := make([]byte, 8)
	binary.BigEndian.PutUint64(bytes, number)
	startingIndex := 0
	size := 0
	for i, b := range bytes {
		if b != 0x00 {
			startingIndex = i
			size = 8 - i
			break
		}
	}
	sizeBytes := proto.EncodeVarint(uint64(size))
	if len(sizeBytes) > 1 {
		panic(fmt.Errorf("[]sizeBytes should not be more than one byte because the max number it needs to hold is 8. size=%d", size))
	}
	encodedBytes := make([]byte, size+1)
	encodedBytes[0] = sizeBytes[0]
	copy(encodedBytes[1:], bytes[startingIndex:])
	return encodedBytes
}

// DecodeOrderPreservingVarUint64 decodes the number from the bytes obtained from method 'EncodeOrderPreservingVarUint64'.
// Also, returns the number of bytes that are consumed in the process
func DecodeOrderPreservingVarUint64(bytes []byte) (uint64, int) {
	s, _ := proto.DecodeVarint(bytes)
	size := int(s)
	decodedBytes := make([]byte, 8)
	copy(decodedBytes[8-size:], bytes[1:size+1])
	numBytesConsumed := size + 1
	return binary.BigEndian.Uint64(decodedBytes), numBytesConsumed
}

// ConstructCompositeKey returns a []byte that uniquely represents a given chaincodeID and key.
// This assumes that chaincodeID does not contain a 0x00 byte, but the key may
// TODO:enforce this restriction on chaincodeID or use length prefixing here instead of delimiter
func ConstructCompositeKey(chaincodeID string, key string) []byte {
	return bytes.Join([][]byte{[]byte(chaincodeID), []byte(key)}, stateKeyDelimiter)
}

// DecodeCompositeKey decodes the compositeKey constructed by ConstructCompositeKey method
// back to the original chaincodeID and key form
func DecodeCompositeKey(compositeKey []byte) (string, string) {
	split := bytes.SplitN(compositeKey, stateKeyDelimiter, 2)
	return string(split[0]), string(split[1])
}
