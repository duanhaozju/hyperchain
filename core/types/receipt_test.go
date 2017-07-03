package types

import (
	"testing"
	"hyperchain/common"
)

func TestReceipt_MakeBloom(t *testing.T) {
	logs := Logs{
		{
			Address:  common.BytesToAddress([]byte("address")),
			Topics:   []common.Hash{
				common.BytesToHash([]byte("topic1")),
				common.BytesToHash([]byte("topic2")),
				common.BytesToHash([]byte("topic3")),
			},
		},
		{
			Address:  common.BytesToAddress([]byte("address")),
			Topics:   []common.Hash{
				common.BytesToHash([]byte("topic3")),
				common.BytesToHash([]byte("topic4")),
				common.BytesToHash([]byte("topic5")),
			},
		},
	}

	receipt := Receipt{}
	receipt.SetLogs(logs)
	receipt.MakeBloom()

	positive := [][]byte{
		common.LeftPadBytes([]byte("topic1"), 32),
		common.LeftPadBytes([]byte("topic2"), 32),
		common.LeftPadBytes([]byte("topic3"), 32),
		common.LeftPadBytes([]byte("topic4"), 32),
		common.LeftPadBytes([]byte("topic5"), 32),
		common.LeftPadBytes([]byte("address"), 20),
	}

	negitive := [][]byte{
		common.LeftPadBytes([]byte("topic"), 32),
		common.LeftPadBytes([]byte("addr"), 20),
	}
	_, bloomFilter := receipt.BloomFilter()
	for _, content := range positive {
		if !BloomLookup(bloomFilter, content) {
			t.Error("expect true")
		}
	}

	for _, content := range negitive {
		if BloomLookup(bloomFilter, content) {
			t.Error("expect false")
		}
	}
}
