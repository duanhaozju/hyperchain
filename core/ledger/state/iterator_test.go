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
package state

import (
	"bytes"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/vm"
	"github.com/hyperchain/hyperchain/hyperdb/mdb"
	"github.com/op/go-logging"
	"math/big"
	"sync"
	"testing"
)

type KV struct {
	Key   common.Hash
	Value []byte
}

var (
	LogOnce sync.Once
	logger  *logging.Logger
)

func NewTestLog() *logging.Logger {
	LogOnce.Do(func() {
		conf := common.NewRawConfig()
		common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
		logger = common.GetLogger(common.DEFAULT_NAMESPACE, "state")
		common.GetLogger(common.DEFAULT_NAMESPACE, "buckettree")
		common.SetLogLevel(common.DEFAULT_NAMESPACE, "state", "WARNING")
		common.SetLogLevel(common.DEFAULT_NAMESPACE, "buckettree", "WARNING")
	})
	return logger
}

func NewTestConfig() *common.Config {
	conf := common.NewRawConfig()
	conf.Set(StateCapacity, 10009)
	conf.Set(StateAggreation, 10)
	conf.Set(StateMerkleCacheSize, 10000)
	conf.Set(StateBucketCacheSize, 10000)

	conf.Set(StateObjectCapacity, 10009)
	conf.Set(StateObjectAggreation, 10)
	conf.Set(StateObjectMerkleCacheSize, 10000)
	conf.Set(StateObjectBucketCacheSize, 10000)
	return conf
}

func TestMemIterator_Iter(t *testing.T) {
	testValue := map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key")):   []byte("value"),
		common.BytesToRightPaddingHash([]byte("key01")): []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key1")):  []byte("value1"),
		common.BytesToRightPaddingHash([]byte("key11")): []byte("value11"),
		common.BytesToRightPaddingHash([]byte("key2")):  []byte("value2"),
	}
	expect := []KV{
		{
			Key:   common.BytesToRightPaddingHash([]byte("key")),
			Value: []byte("value"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key01")),
			Value: []byte("value01"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key1")),
			Value: []byte("value1"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key11")),
			Value: []byte("value11"),
		},
		{
			Key:   common.BytesToRightPaddingHash([]byte("key2")),
			Value: []byte("value2"),
		},
	}
	cnt := 0
	iter := NewMemIterator(Storage(testValue), Storage{})
	defer iter.Release()
	for iter.Next() {
		if bytes.Compare(iter.Key(), expect[cnt].Key.Bytes()) != 0 {
			t.Error("expect to be same key")
		}
		if bytes.Compare(iter.Value(), expect[cnt].Value) != 0 {
			t.Error("expect to be same value")
		}
		cnt += 1
	}
	if cnt != len(expect) {
		t.Error("except to be same length")
	}
}

func TestStorageIteratorWithPrefix(t *testing.T) {
	stateDb := prepareTestCase()
	if !checkIterResult(stateDb, common.BytesToAddress([]byte("address001")),
		vm.BytesPrefix([]byte("key")).Start.Bytes(), vm.BytesPrefix([]byte("key")).Limit.Bytes(), expectIterResult()[1:]) {
		t.Error("iter with prefix failed")
	}
}

func TestStorageIteratorWithRange(t *testing.T) {
	stateDb := prepareTestCase()
	// Checking
	// 1. dump all
	if !checkIterResult(stateDb, common.BytesToAddress([]byte("address001")), nil, nil, expectIterResult()) {
		t.Error("nil start and nil limit, failed")
	}
	// 2. with start and limit
	if !checkIterResult(stateDb, common.BytesToAddress([]byte("address001")), []byte("key1"), []byte("key2"), expectIterResult()[5:7]) {
		t.Error("start and limit, failed")
	}
	// 3. with nil start and limit
	if !checkIterResult(stateDb, common.BytesToAddress([]byte("address001")), nil, []byte("key2"), expectIterResult()[:7]) {
		t.Error("nil start and limit, failed")
	}
	// 4. with start and nil limit
	if !checkIterResult(stateDb, common.BytesToAddress([]byte("address001")), []byte("key1"), nil, expectIterResult()[5:]) {
		t.Error("start and nil limit, failed")
	}
}

func prepareTestCase() *StateDB {
	NewTestLog()
	db, _ := mdb.NewMemDatabase(common.DEFAULT_NAMESPACE)
	stateDb, _ := New(common.Hash{}, db, db, NewTestConfig(), 0)

	stateDb.MarkProcessStart(1)
	stateDb.CreateAccount(common.BytesToAddress([]byte("address001")))
	stateDb.SetBalance(common.BytesToAddress([]byte("address001")), big.NewInt(100))
	for k, v := range testData1() {
		stateDb.SetState(common.BytesToAddress([]byte("address001")), k, v, 0)
	}
	stateDb.Commit()
	stateDb.FetchBatch(1, BATCH_NORMAL).Write()
	stateDb.MarkProcessFinish(1)
	stateDb.Purge()
	stateDb.MarkProcessStart(2)
	for k, v := range testData2() {
		stateDb.SetState(common.BytesToAddress([]byte("address001")), k, v, 0)
	}
	return stateDb
}

func checkIterResult(stateDb *StateDB, addr common.Address, start, limit []byte, expect []KV) (equal bool) {
	defer func() {
		if r := recover(); r != nil {
			equal = false
		}
	}()
	var (
		startH, limitH *common.Hash
		cnt            int
	)
	if start != nil {
		tmp := common.BytesToRightPaddingHash(start)
		startH = &tmp
	}
	if limit != nil {
		tmp := common.BytesToRightPaddingHash(limit)
		limitH = &tmp
	}
	iter, err := stateDb.NewIterator(addr, &vm.IterRange{startH, limitH})
	if err != nil {
		return false
	}
	defer iter.Release()

	validity := true
	for iter.Next() {
		if bytes.Compare(iter.Key(), expect[cnt].Key.Bytes()) != 0 {
			validity = false
			return validity
		}
		if bytes.Compare(iter.Value(), expect[cnt].Value) != 0 {
			validity = false
			return validity
		}
		cnt += 1
	}
	if cnt != len(expect) {
		validity = false
	}
	return validity
}

func testData1() map[common.Hash][]byte {
	return map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key01")): []byte("value01"),
		common.BytesToRightPaddingHash([]byte("key02")): []byte("value02"),
		common.BytesToRightPaddingHash([]byte("key03")): []byte("value03"),
		common.BytesToRightPaddingHash([]byte("key1")):  []byte("value1"),
		common.BytesToRightPaddingHash([]byte("key10")): []byte("value10"),
		common.BytesToRightPaddingHash([]byte("key3")):  []byte("value3"),
		common.BytesToRightPaddingHash([]byte("key")):   []byte("value"),
		common.BytesToRightPaddingHash([]byte("-")):     []byte("value-"),
	}
}

func testData2() map[common.Hash][]byte {
	return map[common.Hash][]byte{
		common.BytesToRightPaddingHash([]byte("key02")): []byte("newvalue02"),
		common.BytesToRightPaddingHash([]byte("key1")):  []byte("newvalue1"),
		common.BytesToRightPaddingHash([]byte("key2")):  []byte("newvalue2"),
		common.BytesToRightPaddingHash([]byte("key04")): []byte("newvalue04"),
		common.BytesToRightPaddingHash([]byte("key")):   nil,
	}
}

func expectIterResult() []KV {
	expect := []KV{
		{
			Key:   common.BytesToRightPaddingHash([]byte("-")),
			Value: []byte("value-"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key01")),
			Value: []byte("value01"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key02")),
			Value: []byte("newvalue02"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key03")),
			Value: []byte("value03"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key04")),
			Value: []byte("newvalue04"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key1")),
			Value: []byte("newvalue1"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key10")),
			Value: []byte("value10"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key2")),
			Value: []byte("newvalue2"),
		}, {
			Key:   common.BytesToRightPaddingHash([]byte("key3")),
			Value: []byte("value3"),
		},
	}
	return expect
}
