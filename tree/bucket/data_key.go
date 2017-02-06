package bucket

import (
	"fmt"
)

type DataKey struct {
	bucketKey    *BucketKey
	compositeKey []byte
}

func newDataKey(treePrefix string, key string) *DataKey {
	compositeKey := ConstructCompositeKey(treePrefix, key)
	// TODO hash can be replaced
	bucketHash := conf.computeBucketHash(compositeKey)
	// Adding one because - we start bucket-numbers 1 onwards
	bucketNumber := int(bucketHash)%conf.getNumBucketsAtLowestLevel() + 1
	dataKey := &DataKey{newBucketKeyAtLowestLevel(bucketNumber), compositeKey}
	return dataKey
}

func (key *DataKey) getBucketKey() *BucketKey {
	return key.bucketKey
}

func encodeBucketNumber(bucketNumber int) []byte {
	return EncodeOrderPreservingVarUint64(uint64(bucketNumber))
}

func decodeBucketNumber(encodedBytes []byte) (int, int) {
	bucketNum, bytesConsumed := DecodeOrderPreservingVarUint64(encodedBytes)
	return int(bucketNum), bytesConsumed
}

// TODO maybe could change the bucketNum and the compositeKey
func (key *DataKey) getEncodedBytes() []byte {
	encodedBytes := encodeBucketNumber(key.bucketKey.bucketNumber)
	encodedBytes = append(encodedBytes, key.compositeKey...)
	return encodedBytes
}

func newDataKeyFromEncodedBytes(encodedBytes []byte) *DataKey {
	bucketNum, l := decodeBucketNumber(encodedBytes)
	compositeKey := make([]byte, len(encodedBytes)-l)
	copy(compositeKey, encodedBytes[l:])
	return &DataKey{newBucketKeyAtLowestLevel(bucketNum), compositeKey}
}

func (key *DataKey) String() string {
	return fmt.Sprintf("bucketKey=[%s], compositeKey=[%s]", key.bucketKey, string(key.compositeKey))
}

func (key *DataKey) clone() *DataKey {
	clone := &DataKey{key.bucketKey.clone(), key.compositeKey}
	return clone
}
