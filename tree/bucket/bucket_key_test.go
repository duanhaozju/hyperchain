package bucket

import (
	"testing"
	"BucketTree/bucket/testutil"
)

func TestBucketKeyGetParentKey(t *testing.T) {
	conf = newConfig(26, 2, fnvHash)
	bucketKey := newBucketKey(5, 24)
	parentKey := bucketKey.getParentKey()
	testutil.AssertEquals(t, parentKey.level, 4)
	testutil.AssertEquals(t, parentKey.bucketNumber, 12)

	bucketKey = newBucketKey(5, 25)
	parentKey = bucketKey.getParentKey()
	testutil.AssertEquals(t, parentKey.level, 4)
	testutil.AssertEquals(t, parentKey.bucketNumber, 13)

	conf = newConfig(26, 3, fnvHash)
	bucketKey = newBucketKey(3, 24)
	parentKey = bucketKey.getParentKey()
	testutil.AssertEquals(t, parentKey.level, 2)
	testutil.AssertEquals(t, parentKey.bucketNumber, 8)

	bucketKey = newBucketKey(3, 25)
	parentKey = bucketKey.getParentKey()
	testutil.AssertEquals(t, parentKey.level, 2)
	testutil.AssertEquals(t, parentKey.bucketNumber, 9)
}

func TestBucketKeyEqual(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketKey1 := newBucketKey(1, 2)
	bucketKey2 := newBucketKey(1, 2)
	testutil.AssertEquals(t, bucketKey1.equals(bucketKey2), true)
	bucketKey2 = newBucketKey(2, 2)
	testutil.AssertEquals(t, bucketKey1.equals(bucketKey2), false)
	bucketKey2 = newBucketKey(1, 3)
	testutil.AssertEquals(t, bucketKey1.equals(bucketKey2), false)
	bucketKey2 = newBucketKey(2, 3)
	testutil.AssertEquals(t, bucketKey1.equals(bucketKey2), false)
}

func TestBucketKeyWrongLevelCausePanic(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	defer testutil.AssertPanic(t, "A panic should occur when asking for a level beyond a valid range")
	newBucketKey(4, 1)
}

func TestBucketKeyWrongBucketNumberCausePanic_1(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	defer testutil.AssertPanic(t, "A panic should occur when asking for a level beyond a valid range")
	newBucketKey(1, 4)
}

func TestBucketKeyWrongBucketNumberCausePanic_2(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	defer testutil.AssertPanic(t, "A panic should occur when asking for a level beyond a valid range")
	newBucketKey(3, 27)
}

func TestBucketKeyWrongBucketNumberCausePanic_3(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	defer testutil.AssertPanic(t, "A panic should occur when asking for a level beyond a valid range")
	newBucketKey(0, 2)
}

func TestBucketKeyGetChildIndex(t *testing.T) {
	conf = newConfig(26, 3, fnvHash)
	bucketKey := newBucketKey(3, 22)
	testutil.AssertEquals(t, bucketKey.getParentKey().getChildIndex(bucketKey), 0)

	bucketKey = newBucketKey(3, 23)
	testutil.AssertEquals(t, bucketKey.getParentKey().getChildIndex(bucketKey), 1)

	bucketKey = newBucketKey(3, 24)
	testutil.AssertEquals(t, bucketKey.getParentKey().getChildIndex(bucketKey), 2)
}

