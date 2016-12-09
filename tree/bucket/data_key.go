package buckettree

import (
	"fmt"
)

type dataKey struct {
	bucketKey    *bucketKey
	compositeKey []byte
}

func newDataKey(chaincodeID string, key string) *dataKey {
	logger.Debugf("Enter - newDataKey. chaincodeID=[%s], key=[%s]", chaincodeID, key)
	compositeKey := ConstructCompositeKey(chaincodeID, key)
	bucketHash := conf.computeBucketHash(compositeKey)
	// Adding one because - we start bucket-numbers 1 onwards
	bucketNumber := int(bucketHash)%conf.getNumBucketsAtLowestLevel() + 1
	dataKey := &dataKey{newBucketKeyAtLowestLevel(bucketNumber), compositeKey}
	logger.Debugf("Exit - newDataKey=[%s]", dataKey)
	return dataKey
}

func minimumPossibleDataKeyBytesFor(bucketKey *bucketKey) []byte {
	min := encodeBucketNumber(bucketKey.bucketNumber)
	min = append(min, byte(0))
	return min
}

func minimumPossibleDataKeyBytes(bucketNumber int, chaincodeID string, key string) []byte {
	b := encodeBucketNumber(bucketNumber)
	b = append(b, ConstructCompositeKey(chaincodeID, key)...)
	return b
}

func (key *dataKey) getBucketKey() *bucketKey {
	return key.bucketKey
}

func encodeBucketNumber(bucketNumber int) []byte {
	return EncodeOrderPreservingVarUint64(uint64(bucketNumber))
}

func decodeBucketNumber(encodedBytes []byte) (int, int) {
	bucketNum, bytesConsumed := DecodeOrderPreservingVarUint64(encodedBytes)
	return int(bucketNum), bytesConsumed
}

func (key *dataKey) getEncodedBytes() []byte {
	encodedBytes := encodeBucketNumber(key.bucketKey.bucketNumber)
	encodedBytes = append(encodedBytes, key.compositeKey...)
	return encodedBytes
}

func newDataKeyFromEncodedBytes(encodedBytes []byte) *dataKey {
	bucketNum, l := decodeBucketNumber(encodedBytes)
	compositeKey := encodedBytes[l:]
	return &dataKey{newBucketKeyAtLowestLevel(bucketNum), compositeKey}
}

func (key *dataKey) String() string {
	return fmt.Sprintf("bucketKey=[%s], compositeKey=[%s]", key.bucketKey, string(key.compositeKey))
}

func (key *dataKey) clone() *dataKey {
	clone := &dataKey{key.bucketKey.clone(), key.compositeKey}
	return clone
}
