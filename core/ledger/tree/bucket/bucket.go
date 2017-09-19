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

package bucket

import (
	"bytes"
	"fmt"
	"github.com/golang/protobuf/proto"
	"github.com/pkg/errors"
	"hyperchain/common"
	"sort"
)

// DataEntry data entry is the minimal data unit in tree.
// A data entry is consists of a key-value pair.
// The actual key is encoded in a full key with the tree prefix.
type DataEntry struct {
	key   *EntryKey
	value []byte
}

func newDataEntry(key *EntryKey, value []byte) *DataEntry {
	return &DataEntry{key, value}
}

func (entry *DataEntry) getCompositeKey() []byte {
	return entry.key.key
}

func (entry *DataEntry) getActuallyKey() []byte {
	_, key := entry.getKeyElements()
	return []byte(key)
}

func (entry *DataEntry) isDelete() bool {
	return entry.value == nil || len(entry.value) == 0 || common.EmptyHash(common.BytesToHash(entry.value))
}

func (entry *DataEntry) getKeyElements() (string, string) {
	return DecodeCompositeKey(entry.getCompositeKey())
}

func (entry *DataEntry) getValue() []byte {
	return entry.value
}

func (entry *DataEntry) String() string {
	return fmt.Sprintf("dataKey=[%s], value=[%s]", entry.key, common.Bytes2Hex(entry.value))
}

type Bucket []*DataEntry

func (bucket Bucket) Len() int {
	return len(bucket)
}

func (bucket Bucket) Swap(i, j int) {
	bucket[i], bucket[j] = bucket[j], bucket[i]
}

func (bucket Bucket) Less(i, j int) bool {
	return bytes.Compare(bucket[i].key.key, bucket[j].key.key) < 0
}

func (bucket Bucket) Encode() []byte {
	buffer := proto.NewBuffer([]byte{})
	buffer.EncodeFixed64(uint64(len(bucket)))
	for _, entry := range bucket {
		buffer.EncodeRawBytes(entry.getActuallyKey())
		buffer.EncodeRawBytes(entry.getValue())
	}
	return buffer.Bytes()
}

func DecodeBucket(prefix string, pos *Position, data []byte, v interface{}) error {
	bucket, ok := v.(*Bucket)
	if ok == false {
		return errors.New("invalid type")
	}

	buffer := proto.NewBuffer(data)
	length, err := buffer.DecodeFixed64()
	if err != nil {
		return err
	}
	for i := 0; i < int(length); i++ {
		key, err := buffer.DecodeRawBytes(false)
		if err != nil {
			panic(fmt.Errorf("this error should not occur: %s", err))
		}
		value, err := buffer.DecodeRawBytes(false)
		if err != nil {
			panic(fmt.Errorf("this error should not occur: %s", err))
		}
		dataKey := &EntryKey{pos, ConstructCompositeKey(prefix, string(key))}
		*bucket = append(*bucket, newDataEntry(dataKey, value))
	}
	return err
}

func (bucket Bucket) merge(another Bucket) Bucket {
	var (
		i, j int
		tmp  Bucket
	)
	for i < len(another) && j < len(bucket) {
		newNode := another[i]
		oldNode := bucket[j]
		c := bytes.Compare(newNode.key.key, oldNode.key.key)
		var next *DataEntry
		switch c {
		case -1:
			next = newNode
			i++
		case 0:
			next = newNode
			i++
			j++
		case 1:
			next = oldNode
			j++
		}
		if !next.isDelete() {
			tmp = append(tmp, next)
		}
	}

	var remain Bucket
	if i < len(another) {
		remain = another[i:]
		for _, n := range remain {
			if !n.isDelete() {
				tmp = append(tmp, n)
			}
		}
	} else if j < len(bucket) {
		remain = bucket[j:]
		tmp = append(tmp, remain...)
	}
	return tmp
}

func (bucket Bucket) computeCryptoHash() []byte {
	buffer := make([][]byte, len(bucket))
	for i, entry := range bucket {
		buffer[i] = entry.getValue()
	}
	hasher := newHasher(5, BucketType)
	defer returnHasher(hasher)
	return hasher.Hash(JoinBytes(buffer, ""))
}

type Buckets struct {
	data map[Position]Bucket
}

func newBuckets(prefix string, entries Entries) *Buckets {
	buckets := &Buckets{make(map[Position]Bucket)}
	for key, value := range entries {
		buckets.add(prefix, key, value)
	}
	for _, bucket := range buckets.data {
		sort.Sort(bucket)
	}
	return buckets
}

func (buckets *Buckets) add(prefix string, key string, value []byte) {
	dataKey := newEntryKey(prefix, key)
	pos := dataKey.getPosition()
	dataNode := newDataEntry(dataKey, value)
	buckets.data[*pos] = append(buckets.data[*pos], dataNode)
}

func (buckets *Buckets) get(pos *Position) Bucket {
	return buckets.data[*pos]
}

func (buckets *Buckets) getAllPosition() []*Position {
	changed := []*Position{}
	for pos := range buckets.data {
		changed = append(changed, pos.clone())
	}
	return changed
}
