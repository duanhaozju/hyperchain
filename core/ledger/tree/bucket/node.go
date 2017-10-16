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
	"github.com/gogo/protobuf/proto"
	"github.com/op/go-logging"
	"sync"
)

// MerkleNode merkleNode represents a tree node except the lowest level's hash bucket.
// Each node contains a list of children's hash. It's hash is derived from children's content.
// If the aggreation is larger(children number is increased), the size of a merkle node will increase too.
type MerkleNode struct {
	pos      *Position
	children [][]byte
	dirty    []bool
	deleted  bool
	lock     sync.RWMutex
	log      *logging.Logger
}

func newMerkleNode(pos *Position, log *logging.Logger) *MerkleNode {
	maxChildren := conf.getAggreation()
	return &MerkleNode{
		pos:      pos,
		children: make([][]byte, maxChildren),
		dirty:    make([]bool, maxChildren),
		deleted:  false,
		log:      log,
	}
}

// encode serializes a merkle node to bytes.
// Only children list will been serialized.
func (node *MerkleNode) encode() []byte {
	buffer := proto.NewBuffer([]byte{})
	for i := 0; i < conf.getAggreation(); i++ {
		buffer.EncodeRawBytes(node.children[i])
	}
	data := buffer.Bytes()
	return data
}

// decodeMerkleNode unserializes a merkle node from bytes.
func decodeMerkleNode(log *logging.Logger, pos *Position, bytes []byte) *MerkleNode {
	node := newMerkleNode(pos, log)
	buffer := proto.NewBuffer(bytes)
	for i := 0; i < conf.getAggreation(); i++ {
		child, err := buffer.DecodeRawBytes(false)
		if err != nil {
			panic(fmt.Errorf("this error should not occur: %s", err))
		}
		if len(child) != 0 {
			node.children[i] = child
		}
	}
	return node
}

// setChild updates child's content with given position and value.
func (node *MerkleNode) setChild(childKey *Position, hash []byte) {
	node.lock.Lock()
	defer node.lock.Unlock()
	i := node.pos.getChildIndex(childKey)
	node.children[i] = hash
	node.dirty[i] = true
}

// merge merges two merkle node into single one.
func (node *MerkleNode) merge(another *MerkleNode) {
	node.lock.Lock()
	defer node.lock.Unlock()
	if !node.pos.equals(another.pos) {
		panic(fmt.Errorf("Nodes with different keys can not be merged. BaseKey=[%#v], MergeKey=[%#v]", node.pos, another.pos))
	}
	for i, child := range another.children {
		if !node.dirty[i] {
			node.children[i] = child
		}
	}
}

// computeCryptoHash calculates a fingerprint for content representation and comparison.
func (node *MerkleNode) computeCryptoHash() []byte {
	node.lock.RLock()
	defer node.lock.RUnlock()
	buffer := []byte{}
	num := 0
	for i, child := range node.children {
		if child != nil {
			num++
			node.log.Debugf("Append child [%s] to buffer", node.pos.getChildKey(i+1))
			buffer = append(buffer, child...)
		}
	}
	if num == 0 {
		node.log.Debugf("Empty merkle node at [%s]", node.pos)
		node.deleted = true
		return nil
	}
	if num == 1 {
		node.log.Debugf("Propagating the single child node at [%s]", node.pos)
		return buffer
	}

	node.log.Debugf("Calculate hash for [%s] merkle node, with [%d] children", node.pos, num)
	if node.pos.level <= 2 {
		hasher := newHasher(0, MerkleHigh)
		defer returnHasher(hasher)
		return hasher.Hash(buffer)
	} else {
		// If the node is low enough, trim the content instead of using the whole one.
		hasher := newHasher(5, MerkleLow)
		defer returnHasher(hasher)
		return hasher.Hash(buffer)
	}
}

// String convert merkle node to string format.
func (node *MerkleNode) String() string {
	node.lock.RLock()
	defer node.lock.RUnlock()
	num := 0
	for i := range node.children {
		if node.children[i] != nil {
			num++
		}
	}
	str := fmt.Sprintf("Postion={%s} Children={%d}", node.pos, num)
	if num == 0 {
		return str
	}

	str = str + "Children hashes:\n"
	for i := range node.children {
		child := node.children[i]
		if child != nil {
			str = str + fmt.Sprintf("childIndex={%d}, hash={%x}\n", i, child)
		}
	}
	return str
}

type MerkleNodeDelta struct {
	data map[int]map[int]*MerkleNode
	lock sync.RWMutex
	log  *logging.Logger
}

func newMerkleNodeDelta(log *logging.Logger) *MerkleNodeDelta {
	return &MerkleNodeDelta{
		data: make(map[int]map[int]*MerkleNode),
		log:  log,
	}
}

func (delta *MerkleNodeDelta) getOrCreate(pos *Position) *MerkleNode {
	delta.lock.Lock()
	defer delta.lock.Unlock()
	levelNodes := delta.data[pos.level]
	if levelNodes == nil {
		levelNodes = make(map[int]*MerkleNode)
		delta.data[pos.level] = levelNodes
	}
	merkle := levelNodes[pos.index]
	if merkle == nil {
		merkle = newMerkleNode(pos, delta.log)
		levelNodes[pos.index] = merkle
	}
	return merkle
}

func (delta *MerkleNodeDelta) empty() bool {
	delta.lock.RLock()
	defer delta.lock.RUnlock()
	return delta.data == nil || len(delta.data) == 0
}

func (delta *MerkleNodeDelta) getLevel(level int) []*MerkleNode {
	delta.lock.RLock()
	defer delta.lock.RUnlock()
	nodes := []*MerkleNode{}
	levelNodes := delta.data[level]
	if levelNodes == nil {
		return nil
	}
	for _, n := range levelNodes {
		nodes = append(nodes, n)
	}
	return nodes
}

func (delta *MerkleNodeDelta) getRoot() *MerkleNode {
	levelNodes := delta.getLevel(0)
	if levelNodes == nil || len(levelNodes) == 0 {
		panic("This method should be called after processing is completed (i.e., the root node has been created)")
	}
	return levelNodes[0]
}
