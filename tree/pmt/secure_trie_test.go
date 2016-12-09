//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package pmt

import (
	"bytes"
	"fmt"
	"hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"testing"
)

func newEmptySecure() *SecureTrie {
	db, _ := hyperdb.NewMemDatabase()
	trie, _ := NewSecure(common.Hash{}, db)
	return trie
}

func TestSecureDelete(t *testing.T) {
	trie := newEmptySecure()
	vals := []struct{ k, v string }{
		{"do", "verb"},
		{"ether", "wookiedoo"},
		{"horse", "stallion"},
		{"shaman", "horse"},
		{"doge", "coin"},
		{"ether", ""},
		{"dog", "puppy"},
		{"shaman", ""},
	}
	for _, val := range vals {
		if val.v != "" {
			trie.Update([]byte(val.k), []byte(val.v))
		} else {
			trie.Delete([]byte(val.k))
		}
	}
	root, _ := trie.Commit()

	ClearGlobalCache()
	trie, _ = NewSecure(root, trie.db)
	traverse(trie)
}

func TestSecureGetKey(t *testing.T) {
	trie := newEmptySecure()
	trie.Update([]byte("foo"), []byte("bar"))

	key := []byte("foo")
	value := []byte("bar")
	seckey := crypto.Keccak256(key)

	if !bytes.Equal(trie.Get(key), value) {
		t.Errorf("Get did not return bar")
	}
	if k := trie.GetKey(seckey); !bytes.Equal(k, key) {
		t.Errorf("GetKey returned %q, want %q", k, key)
	}
}

func traverse(trie *SecureTrie) {
	it := NewIterator(trie.Trie)
	for it.Next() {
		fmt.Printf("TRIE key: %x, val : %v\n", string(it.Key), string(it.Value))
	}
}
