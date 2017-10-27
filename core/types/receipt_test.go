// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package types

import (
	"github.com/hyperchain/hyperchain/common"
	"testing"
)

func TestReceipt_MakeBloom(t *testing.T) {
	logs := Logs{
		{
			Address: common.BytesToAddress([]byte("address")),
			Topics: []common.Hash{
				common.BytesToHash([]byte("topic1")),
				common.BytesToHash([]byte("topic2")),
				common.BytesToHash([]byte("topic3")),
			},
		},
		{
			Address: common.BytesToAddress([]byte("address")),
			Topics: []common.Hash{
				common.BytesToHash([]byte("topic3")),
				common.BytesToHash([]byte("topic4")),
				common.BytesToHash([]byte("topic5")),
			},
		},
	}

	receipt := &Receipt{}
	receipt.SetLogs(logs)
	receipt.Bloom, _ = CreateBloom([]*Receipt{receipt})

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
