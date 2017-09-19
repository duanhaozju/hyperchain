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
	"reflect"
	"testing"
)

func TestMerkleNodeEncode(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	n := newMerkleNode(&Position{4, 1}, NewTestLog())
	n.children[0] = []byte{1, 2, 3}
	n.children[1] = []byte{2, 3, 4}
	n.children[2] = []byte{3, 4, 5}
	expect := []byte{3, 1, 2, 3, 3, 2, 3, 4, 3, 3, 4, 5, 0, 0, 0, 0, 0, 0, 0}
	if bytes.Compare(n.encode(), expect) != 0 {
		t.Error("expect to be same after encode")
	}

	n2 := decodeMerkleNode(NewTestLog(), &Position{4, 1}, expect)
	if !reflect.DeepEqual(n, n2) {
		t.Error("expect to be same after decode")
	}
}

func TestMerkleNodeMerge(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	n := newMerkleNode(&Position{4, 1}, NewTestLog())
	n.children[0] = []byte{1, 2, 3}
	n.children[1] = []byte{2, 3, 4}

	n2 := newMerkleNode(&Position{4, 1}, NewTestLog())
	n2.children[1] = []byte{3, 4, 5}
	n2.children[2] = []byte{4, 5, 6}
	n2.dirty[1] = true
	n2.dirty[2] = true

	n2.merge(n)
	n3 := newMerkleNode(&Position{4, 1}, NewTestLog())
	n3.children[0] = []byte{1, 2, 3}
	n3.children[1] = []byte{3, 4, 5}
	n3.children[2] = []byte{4, 5, 6}
	if !reflect.DeepEqual(n2.children, n3.children) {
		t.Error("expect to be same after merge")
	}
}

func TestMerkleNodeComputeCrypto(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	node := newMerkleNode(&Position{4, 1}, NewTestLog())
	if !merkleNodeComputeCrypto(node, nil) {
		t.Error("invalid hash crypto for empty node")
	}
	node.setChild(&Position{5, 1}, []byte{1, 2, 3})
	if !merkleNodeComputeCrypto(node, []byte{1, 2, 3}) {
		t.Error("invalid hash crypto for single child node")
	}

	node.setChild(&Position{5, 2}, []byte{2, 3, 4})
	if !merkleNodeComputeCrypto(node, []byte{54, 196, 50, 247, 78}) {
		t.Error("invalid hash crypto for single child node")
	}

	node = newMerkleNode(&Position{2, 1}, NewTestLog())
	node.setChild(&Position{3, 1}, []byte{1, 2, 3})
	node.setChild(&Position{3, 2}, []byte{2, 3, 4})
	if !merkleNodeComputeCrypto(node, []byte{251, 77, 144, 124, 34, 137, 191, 79, 149, 65, 7, 209, 132, 113, 94, 174, 249, 65, 87, 219, 30, 189, 184, 121, 13, 99, 246, 40, 3, 185, 49, 29}) {
		t.Error("invalid hash crypto for single child node")
	}
}

func merkleNodeComputeCrypto(node *MerkleNode, expect []byte) bool {
	return bytes.Compare(node.computeCryptoHash(), expect) == 0
}

func TestMerkleNodeDelta(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	delta := newMerkleNodeDelta(NewTestLog())
	poses := []*Position{{4, 1}, {4, 2}, {4, 3}}
	for _, pos := range poses {
		delta.getOrCreate(pos)
	}
	for idx, n := range delta.getLevel(4) {
		if !n.pos.equals(poses[idx]) {
			t.Error("expect to be same got merkle node from delta")
		}
	}
}
