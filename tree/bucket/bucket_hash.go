package bucket

import (
	"crypto/sha1"
)

type bucketHashCalculator struct {
	hashingData []byte
}

func newBucketHashCalculator() *bucketHashCalculator {
	return &bucketHashCalculator{nil}
}
func (c *bucketHashCalculator) setHashingData(hashingData []byte) {
	c.hashingData = hashingData
}

func (c *bucketHashCalculator) computeCryptoHash() []byte {
	if c.hashingData == nil || len(c.hashingData) == 0 {
		return nil
	}
	h := sha1.New()
	h.Write([]byte(c.hashingData))
	bs := h.Sum(nil)
	return bs
}
