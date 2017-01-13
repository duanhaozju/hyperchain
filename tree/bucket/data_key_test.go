package bucket

import (
	"hyperchain/tree/bucket/testutil"
	"testing"
)

func TestDataKey(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	dataKey := newDataKey("chaincodeID", "key")
	encodedBytes := dataKey.getEncodedBytes()
	dataKeyFromEncodedBytes := newDataKeyFromEncodedBytes(encodedBytes)
	testutil.AssertEquals(t, dataKey, dataKeyFromEncodedBytes)
}

func TestDataKeyGetBucketKey(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	newDataKey("chaincodeID1", "key1").getBucketKey()
	newDataKey("chaincodeID1", "key2").getBucketKey()
	newDataKey("chaincodeID2", "key1").getBucketKey()
	newDataKey("chaincodeID2", "key2").getBucketKey()
}
