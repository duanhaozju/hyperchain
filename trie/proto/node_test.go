package trie

import (
	"bytes"
	"github.com/golang/protobuf/proto"
	"testing"
)

func TestMarshal(t *testing.T) {
	hashnode := &HashNode{
		Hash: []byte("hash-hash"),
	}
	data, err := proto.Marshal(hashnode)
	if err != nil {
		t.Error(err)
	}

	hashnode_cmp := &HashNode{}
	err = proto.Unmarshal(data, hashnode_cmp)
	if err != nil {
		t.Error(err)
	}

	if bytes.Compare(hashnode_cmp.Hash, hashnode.Hash) != 0 {
		t.Error("mismatch")
	}

	valuenode := &ValueNode{
		Val: []byte("value-value"),
	}

	shortnode := &ShortNode{
		Key:   []byte("short-key"),
		Dirty: false,
		Hash:  []byte("short-hash"),
	}
	shortnode.SetVal(valuenode)
	fullnode := &FullNode{
		Dirty: true,
		Hash:  []byte("full-hash"),
	}
	fullnode.SetChild(shortnode, 0)

	data, err = proto.Marshal(fullnode)
	if err != nil {
		t.Error(err)
	}
	fullnode_cmp := &FullNode{}
	err = proto.Unmarshal(data, fullnode_cmp)
	if err != nil {
		t.Error(err)
	}
	if bytes.Compare(fullnode_cmp.Hash, fullnode.Hash) != 0 {
		t.Errorf("mismatch, %x, %x", fullnode_cmp.Hash, fullnode.Hash)
	}

	// generic unmarshal
	res := mustDecodeNode(data)
	switch n := res.(type) {
	case *ShortNode:
		t.Log(n)
	case *FullNode:
		t.Log(n)
	default:
		t.Errorf("Invalid node type")
	}

}
