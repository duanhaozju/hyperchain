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
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"hyperchain/hyperdb/mdb"
	"math/rand"
	"reflect"
	"testing"
	"time"
)

const (
	nMerkle byte = 1 << iota
	nBucket
)

var (
	rnd = rand.New(rand.NewSource(time.Now().UnixNano()))
)

const prefix = "treePrefix"

func NewTestMerkleNode() *MerkleNode {
	// Assume the config has been initialized.
	var (
		l        = rnd.Intn(conf.lowest + 1)
		i        = rnd.Intn(conf.getMaxIndex(l)) + 1
		children [][]byte
	)
	n := newMerkleNode(&Position{
		level: l,
		index: i,
	}, NewTestLog())
	if i%3 == 0 {
		// return empty node
		return n
	}
	for i := 0; i < conf.getAggreation(); i += 1 {
		if i%2 == 0 {
			children = append(children, nil)
		} else {
			children = append(children, common.RandBytes(32))
		}
	}
	n.children = children
	return n
}

func NewTestBucket() (Bucket, *Position) {
	var (
		l      = conf.getLowestLevel()
		i      = rnd.Intn(conf.getMaxIndex(l)) + 1
		size   = rnd.Intn(50)
		bucket Bucket
	)
	if i%3 == 0 {
		// return empty bucket
		return bucket, &Position{l, i}
	}
	for ii := 0; ii < size; ii += 1 {
		entry := newDataEntry(newEntryKey(prefix, string(common.RandBytes(32))), common.RandBytes(32))
		entry.key.pos.index = i
		bucket = append(bucket, entry)
	}
	return bucket, &Position{l, i}
}

func TestRetrieveMerkleNode(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	iterN := 100
	for i := 0; i < iterN; i += 1 {
		node := NewTestMerkleNode()
		if !compare(nMerkle, NewTestLog(), db, node.pos, node) {
			t.Errorf("expect to be same for merkle node persistence. merkle node %s", node.String())
		}
	}
}

func TestRetrieveBucket(t *testing.T) {
	initConfig(NewTestLog(), NewTestConfig())
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	iterN := 100
	for i := 0; i < iterN; i += 1 {
		b, pos := NewTestBucket()
		if !compare(nBucket, NewTestLog(), db, pos, b) {
			t.Errorf("expect to be same for bucket persistence. bucket %s, pos %s", b, pos.String())
		}
	}
}

func persist(typ byte, db db.Database, pos *Position, v interface{}) error {
	batch := db.NewBatch()
	switch typ {
	case nMerkle:
		if err := PersistMerkleNode(prefix, v.(*MerkleNode), batch); err != nil {
			return err
		}
	case nBucket:
		if err := PersistBucket(prefix, v.(Bucket), pos, batch); err != nil {
			return err
		}
	}
	if err := batch.Write(); err != nil {
		return err
	}
	return nil
}

func retrieve(typ byte, logger *logging.Logger, db db.Database, pos *Position) (interface{}, error) {
	switch typ {
	case nMerkle:
		return RetrieveMerkleNode(logger, db, prefix, pos)
	case nBucket:
		return RetrieveBucket(logger, db, prefix, pos)
	}
	return nil, nil
}

func compare(typ byte, logger *logging.Logger, db db.Database, pos *Position, v interface{}) bool {
	if err := persist(typ, db, pos, v); err != nil {
		return false
	}
	v2, err := retrieve(typ, logger, db, pos)
	if err != nil {
		return false
	}
	if !reflect.DeepEqual(v, v2) {
		logger.Errorf("different value. Original %s, current %s", v, v2)
		return false
	}
	return true

}
