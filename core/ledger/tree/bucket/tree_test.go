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
	"hyperchain/common"
	"hyperchain/hyperdb/mdb"
	"testing"
)

type HashTestCase struct {
	input  Entries
	expect []byte
}

func TestBucketTree_ComputeHash(t *testing.T) {
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	tree := NewBucketTree(db, "prefix")
	for _, c := range generateHashTestCases() {
		tree.Initialize(NewTestConfig())
		tree.Prepare(c.input)
		h, err := tree.Process()
		if err != nil {
			t.Error(err.Error())
		}
		if bytes.Compare(h, c.expect) != 0 {
			t.Error("expect to be same")
		}
		tree.Clear()
	}
}

func TestBucketTree_ComputeWithBase(t *testing.T) {
	var (
		db, _  = mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
		batch  = db.NewBatch()
		cases  = generateHashTestCases()
		tree   = NewBucketTree(db, "prefix")
		hash   []byte
		expect = []byte{35, 120, 81, 252, 124, 230, 225, 223, 198, 131, 17, 47, 236, 15, 143, 177, 60, 107, 103, 94, 17, 250, 164, 77, 81, 228, 149, 64, 55, 152, 247, 213}
	)
	tree.Initialize(NewTestConfig())
	tree.Prepare(cases[0].input)
	tree.Process()
	tree.Commit(batch)
	batch.Write()

	tree.Prepare(cases[1].input)
	hash, _ = tree.Process()
	tree.Commit(batch)

	if bytes.Compare(hash, expect) != 0 {
		t.Error("expect to be same")
	}
}

func generateHashTestCases() []HashTestCase {
	return []HashTestCase{
		{
			// share the same bucket
			Entries{"key1": []byte("value1"), "key2": []byte("value2"), "key3": []byte("value3")},
			[]byte{240, 2, 195, 167, 87},
		}, {
			Entries{"key100": []byte("value100"), "key200": []byte("value200"), "key300": []byte("value300")},
			[]byte{205, 215, 10, 4, 248, 44, 79, 183, 117, 33, 238, 14, 99, 189, 215, 177, 62, 230, 194, 36, 121, 123, 206, 198, 128, 239, 238, 52, 85, 25, 197, 162},
		},
	}
}
