package bucket

import 	(
	"testing"
	"BucketTree/bucket/testutil"
)

func TestBucketNodeComputeHash(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketNode := newBucketNode(newBucketKey(2, 7))
	testutil.AssertEquals(t, bucketNode.computeCryptoHash(), nil)

	childKey1 := newBucketKey(3, 19)
	bucketNode.setChildCryptoHash(childKey1, []byte("cryptoHashChild1"))
	testutil.AssertEquals(t, bucketNode.computeCryptoHash(), []byte("cryptoHashChild1"))

	childKey3 := newBucketKey(3, 21)
	bucketNode.setChildCryptoHash(childKey3, []byte("cryptoHashChild3"))
	testutil.ComputeCryptoHash([]byte("cryptoHashChild1cryptoHashChild3"))
	testutil.AssertEquals(t, bucketNode.computeCryptoHash(), testutil.ComputeCryptoHash([]byte("cryptoHashChild1cryptoHashChild3")))

	childKey2 := newBucketKey(3, 20)
	bucketNode.setChildCryptoHash(childKey2, []byte("cryptoHashChild2"))
	testutil.AssertEquals(t, bucketNode.computeCryptoHash(), testutil.ComputeCryptoHash([]byte("cryptoHashChild1cryptoHashChild2cryptoHashChild3")))
}

func TestBucketNodeMerge(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketNode := newBucketNode(newBucketKey(2, 7))
	bucketNode.childrenCryptoHash[0] = []byte("cryptoHashChild1")
	bucketNode.childrenUpdated[0] = true
	bucketNode.childrenCryptoHash[2] = []byte("cryptoHashChild3")
	bucketNode.childrenUpdated[2] = true

	dbBucketNode := newBucketNode(newBucketKey(2, 7))
	dbBucketNode.childrenCryptoHash[0] = []byte("DBcryptoHashChild1")
	dbBucketNode.childrenCryptoHash[1] = []byte("DBcryptoHashChild2")

	bucketNode.mergeBucketNode(dbBucketNode)
	testutil.AssertEquals(t, bucketNode.childrenCryptoHash[0], []byte("cryptoHashChild1"))
	testutil.AssertEquals(t, bucketNode.childrenCryptoHash[1], []byte("DBcryptoHashChild2"))
	testutil.AssertEquals(t, bucketNode.childrenCryptoHash[2], []byte("cryptoHashChild3"))
}

func TestBucketNodeMarshalUnmarshal(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketNode := newBucketNode(newBucketKey(2, 7))
	childKey1 := newBucketKey(3, 19)
	bucketNode.setChildCryptoHash(childKey1, []byte("cryptoHashChild1"))

	childKey3 := newBucketKey(3, 21)
	bucketNode.setChildCryptoHash(childKey3, []byte("cryptoHashChild3"))

	serializedBytes := bucketNode.marshal()
	deserializedBucketNode := unmarshalBucketNode(newBucketKey(2, 7), serializedBytes)
	testutil.AssertEquals(t, bucketNode.bucketKey, deserializedBucketNode.bucketKey)
	testutil.AssertEquals(t, bucketNode.childrenCryptoHash, deserializedBucketNode.childrenCryptoHash)
}
