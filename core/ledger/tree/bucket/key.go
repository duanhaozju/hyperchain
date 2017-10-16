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
	"fmt"
	"github.com/golang/protobuf/proto"
)

// Position represents the node position in hash tree.
type Position struct {
	level int // The node level in tree
	index int // The index of the level
}

// newPosition creates a postion struct with level and index.
func newPosition(level int, index int) *Position {
	if level > conf.getLowestLevel() || level < 0 {
		panic(fmt.Errorf("Invalid Level [%d] for bucket key. Level can be between 0 and [%d]", level, conf.lowest))
	}

	if index < 1 || index > conf.getMaxIndex(level) {
		panic(fmt.Errorf("Invalid bucket number [%d]. Bucket index at level [%d] can be between 1 and [%d]", index, level, conf.getMaxIndex(level)))
	}
	return &Position{level, index}
}

// newLowestPosition creates a position at the lowest level with given index.
func newLowestPosition(index int) *Position {
	return newPosition(conf.getLowestLevel(), index)
}

// rootPostion returns the position information of root node.
func rootPosition() *Position {
	return newPosition(0, 1)
}

// getParent returns parent node's position.
func (pos *Position) getParent() *Position {
	return newPosition(pos.level-1, conf.computeParentIndex(pos.index))
}

// equals checks whether the given position is same with itself.
func (pos *Position) equals(another *Position) bool {
	return pos.level == another.level && pos.index == another.index
}

func (pos *Position) getChildIndex(childKey *Position) int {
	first := ((pos.index - 1) * conf.getAggreation()) + 1
	last := pos.index * conf.getAggreation()
	if childKey.index < first || childKey.index > last {
		panic(fmt.Errorf("[%#v] is not a valid child bucket of [%#v]", childKey, pos))
	}
	return childKey.index - first
}

func (pos *Position) getChildKey(index int) *Position {
	first := ((pos.index - 1) * conf.getAggreation()) + 1
	number := first + index - 1
	return newPosition(pos.level+1, number)
}

// encode encodes position information into byte format.
// Position encode rules:
//    1. add a zero byte as the leading flag.
//    2. encode level and index using proto serially.
func (pos *Position) encode() []byte {
	buf := []byte{}
	buf = append(buf, byte(0))
	buf = append(buf, proto.EncodeVarint(uint64(pos.level))...)
	buf = append(buf, proto.EncodeVarint(uint64(pos.index))...)
	return buf
}

// String returns a string to represents the position.
func (pos *Position) String() string {
	return fmt.Sprintf("level=[%d], index=[%d]", pos.level, pos.index)
}

// clone clones a new position.
func (pos *Position) clone() *Position {
	return newPosition(pos.level, pos.index)
}

// EntryKey represents a key of the entry data.
type EntryKey struct {
	pos *Position // the position in tree
	key []byte    // encoded with prefix and actual key
}

func newEntryKey(prefix string, key string) *EntryKey {
	compositeKey := ConstructCompositeKey(prefix, key)
	h := conf.computeIndex(compositeKey[:len(compositeKey)-1])
	// Adding one because - we start bucket-numbers 1 onwards
	idx := int(h%uint32(conf.getCapacity()) + 1)
	ekey := &EntryKey{newLowestPosition(idx), compositeKey}
	return ekey
}

func (key *EntryKey) getPosition() *Position {
	return key.pos
}

func encodeIndex(index int) []byte {
	return EncodeOrderPreservingVarUint64(uint64(index))
}

func decodeIndex(enc []byte) (int, int) {
	index, n := DecodeOrderPreservingVarUint64(enc)
	return int(index), n
}

func (key *EntryKey) encode() []byte {
	encodedBytes := encodeIndex(key.pos.index)
	encodedBytes = append(encodedBytes, key.key...)
	return encodedBytes
}

func newEntryKeyFromBytes(enc []byte) *EntryKey {
	idx, l := decodeIndex(enc)
	key := make([]byte, len(enc)-l)
	copy(key, enc[l:])
	return &EntryKey{newLowestPosition(idx), key}
}

func (key *EntryKey) String() string {
	return fmt.Sprintf("bucketKey=[%s], compositeKey=[%s]", key.pos, string(key.key))
}

func (key *EntryKey) clone() *EntryKey {
	clone := &EntryKey{key.pos.clone(), key.key}
	return clone
}
