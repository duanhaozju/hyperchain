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
	"encoding/json"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/hyperdb/mdb"
	"io/ioutil"
	"testing"
)

type KV struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type BucketTestCase struct {
	Base   []KV              `json:"base"`
	Input  map[string]string `json:"input"`
	Expect string            `json:"expect"`
}

func ReadTestCase() ([]BucketTestCase, error) {
	blob, err := ioutil.ReadFile("testcases.json")
	if err != nil {
		return nil, err
	}
	var cases []BucketTestCase = make([]BucketTestCase, 0)
	if err := json.Unmarshal(blob, &cases); err != nil {
		return nil, err
	}
	return cases, nil
}

func TestBucketTree_Process(t *testing.T) {
	cases, err := ReadTestCase()
	if err != nil {
		t.Error(err.Error())
	}
	for i, c := range cases {
		var (
			db      *mdb.MemDatabase
			entries Entries = NewEntries()
			hash    []byte
		)
		db, _ = mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
		for _, kv := range c.Base {
			db.Set(common.Hex2Bytes(kv.Key), common.Hex2Bytes(kv.Value))
		}
		for k, v := range c.Input {
			entries[k] = []byte(v)
		}
		tree := NewBucketTree(db, "prefix")
		tree.Initialize(NewTestConfig())
		tree.Prepare(entries)
		hash, _ = tree.Process()
		if bytes.Compare(hash, common.Hex2Bytes(c.Expect)) != 0 {
			t.Errorf("expect to be same hash result for case [%d]", i)
		}
		batch := db.NewBatch()
		tree.Commit(batch)
		batch.Write()
		// Test content reload
		tree.Clear()
		tree.Initialize(NewTestConfig())
		hash, _ = tree.Process()
		if bytes.Compare(hash, common.Hex2Bytes(c.Expect)) != 0 {
			t.Errorf("expect to be same hash result for case [%d]", i)
		}
	}
}
