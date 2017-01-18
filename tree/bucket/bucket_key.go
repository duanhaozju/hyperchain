package bucket

import (
	"fmt"
	"github.com/golang/protobuf/proto"
)

type BucketKey struct {
	level        int
	bucketNumber int
}

func newBucketKey(level int, bucketNumber int) *BucketKey {
	if level > conf.getLowestLevel() || level < 0 {
		panic(fmt.Errorf("Invalid Level [%d] for bucket key. Level can be between 0 and [%d]", level, conf.lowestLevel))
	}

	if bucketNumber < 1 || bucketNumber > conf.getNumBuckets(level) {
		panic(fmt.Errorf("Invalid bucket number [%d]. Bucket nuber at level [%d] can be between 1 and [%d]", bucketNumber, level, conf.getNumBuckets(level)))
	}
	return &BucketKey{level, bucketNumber}
}

func newBucketKeyAtLowestLevel(bucketNumber int) *BucketKey {
	return newBucketKey(conf.getLowestLevel(), bucketNumber)
}

func constructRootBucketKey() *BucketKey {
	return newBucketKey(0, 1)
}

// TODO the Decode function should be changed
func decodeBucketKey(keyBytes []byte) BucketKey {
	level, numBytesRead := proto.DecodeVarint(keyBytes[1:])
	bucketNumber, _ := proto.DecodeVarint(keyBytes[numBytesRead+1:])
	return BucketKey{int(level), int(bucketNumber)}
}

func (bucketKey *BucketKey) getParentKey() *BucketKey {
	return newBucketKey(bucketKey.level-1, conf.computeParentBucketNumber(bucketKey.bucketNumber))
}

func (bucketKey *BucketKey) equals(anotherBucketKey *BucketKey) bool {
	return bucketKey.level == anotherBucketKey.level && bucketKey.bucketNumber == anotherBucketKey.bucketNumber
}

func (bucketKey *BucketKey) getChildIndex(childKey *BucketKey) int {
	bucketNumberOfFirstChild := ((bucketKey.bucketNumber - 1) * conf.getMaxGroupingAtEachLevel()) + 1
	bucketNumberOfLastChild := bucketKey.bucketNumber * conf.getMaxGroupingAtEachLevel()
	if childKey.bucketNumber < bucketNumberOfFirstChild || childKey.bucketNumber > bucketNumberOfLastChild {
		panic(fmt.Errorf("[%#v] is not a valid child bucket of [%#v]", childKey, bucketKey))
	}
	return childKey.bucketNumber - bucketNumberOfFirstChild
}

func (bucketKey *BucketKey) getChildKey(index int) *BucketKey {
	bucketNumberOfFirstChild := ((bucketKey.bucketNumber - 1) * conf.getMaxGroupingAtEachLevel()) + 1
	bucketNumberOfChild := bucketNumberOfFirstChild + index
	return newBucketKey(bucketKey.level+1, bucketNumberOfChild)
}

func (bucketKey *BucketKey) getEncodedBytes() []byte {
	encodedBytes := []byte{}
	encodedBytes = append(encodedBytes, byte(0))
	encodedBytes = append(encodedBytes, proto.EncodeVarint(uint64(bucketKey.level))...)
	encodedBytes = append(encodedBytes, proto.EncodeVarint(uint64(bucketKey.bucketNumber))...)
	return encodedBytes
}

func (bucketKey *BucketKey) String() string {
	return fmt.Sprintf("level=[%d], bucketNumber=[%d]", bucketKey.level, bucketKey.bucketNumber)
}

func (bucketKey *BucketKey) clone() *BucketKey {
	return newBucketKey(bucketKey.level, bucketKey.bucketNumber)
}
