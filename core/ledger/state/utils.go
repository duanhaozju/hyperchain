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
	"hyperchain/common"
	"strconv"
)

const (
	storagePrefix = "-storage" // storagePrefix + address + key -> contract storage entry
	accountPrefix = "-account" // accountPrefix + address -> account
	codePrefix    = "-code"    // codePrefix + address + codeHash -> code
	journalPrefix = "-journal" // journalPrefix + block number -> state journal
	treePrefix    = "-bucket"
)

const (
	TreeCapacity        = "capacity"
	TreeAggreation      = "aggreation"
	MerkleNodeCacheSize = "merkleNodeCache"
	BucketCacheSize     = "bucketCache"
)

const (
	STATEDB              = "state"
	StateCapacity        = "executor.buckettree.state.capacity"
	StateAggreation      = "executor.buckettree.state.aggreation"
	StateMerkleCacheSize = "executor.buckettree.state.merklenode_cache"
	StateBucketCacheSize = "executor.buckettree.state.bucket_cache"

	STATEOBJ                   = "stateObject"
	StateObjectCapacity        = "executor.buckettree.storage.capacity"
	StateObjectAggreation      = "executor.buckettree.storage.aggreation"
	StateObjectMerkleCacheSize = "executor.buckettree.storage.merklenode_cache"
	StateObjectBucketCacheSize = "executor.buckettree.storage.bucket_cache"
)

// CompositeStorageKey constructs contract storage entry's key.
func CompositeStorageKey(address []byte, key []byte) []byte {
	ret := append([]byte(storagePrefix), address...)
	ret = append(ret, key...)
	return ret
}

// GetStorageKeyPrefix constructs contract storage prefix by contract address.
// It is offen used in database traverse.
func GetStorageKeyPrefix(address []byte) []byte {
	ret := append([]byte(storagePrefix), address...)
	return ret
}

// SplitCompositeStorageKey splits the composite key to get the actual key.
func SplitCompositeStorageKey(address []byte, key []byte) ([]byte, bool) {
	prefix := append([]byte(storagePrefix), address...)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) {
		return key[prefixLen:], true
	} else {
		return nil, false
	}
}

// RetrieveAddrFromStorageKey splits the composite key to get address info.
func RetrieveAddrFromStorageKey(key []byte) ([]byte, bool) {
	prefix := []byte(storagePrefix)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) && len(key) > prefixLen+common.AddressLength {
		return key[prefixLen : prefixLen+common.AddressLength], true
	} else {
		return nil, false
	}
}

// CompositeCodeHash constructs code key with given address and codehash.
func CompositeCodeHash(address []byte, codeHash []byte) []byte {
	ret := append([]byte(codePrefix), address...)
	return append(ret, codeHash...)
}

// SplitCompositeCodeHash splits the composite key to retrieve codehash.
func SplitCompositeCodeHash(address []byte, key []byte) ([]byte, bool) {
	prefix := append([]byte(codePrefix), address...)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) && len(key) >= prefixLen {
		return key[prefixLen:], true
	} else {
		return nil, false
	}
}

// RetrieveAddrFromCodeHash splits the composite key to retrieve address.
func RetrieveAddrFromCodeHash(key []byte) ([]byte, bool) {
	prefix := []byte(codePrefix)
	prefixLen := len(prefix)
	if bytes.HasPrefix(key, prefix) && len(key) > prefixLen+common.AddressLength {
		return key[prefixLen : prefixLen+common.AddressLength], true
	} else {
		return nil, false
	}
}

// CompositeAccountKey constructs account key with given address.
func CompositeAccountKey(address []byte) []byte {
	return append([]byte(accountPrefix), address...)
}

// SplitCompositeAccountKey splits composite key to retrieve address.
func SplitCompositeAccountKey(key []byte) ([]byte, bool) {
	identifierLen := len([]byte(accountPrefix))
	if bytes.HasPrefix(key, []byte(accountPrefix)) {
		return key[identifierLen:], true
	} else {
		return nil, false
	}
}

func CompositeStateBucketPrefix() []byte {
	return append([]byte(treePrefix), []byte("-state")...)
}

func CompositeStorageBucketPrefix(address string) []byte {
	return []byte(treePrefix + address)
}

// InitTreeConfig constructs bucket tree configuration.
func InitTreeConfig(cap, aggreation, msize, bsize int) map[string]interface{} {
	ret := make(map[string]interface{})
	ret[TreeCapacity] = cap
	ret[TreeAggreation] = aggreation
	ret[MerkleNodeCacheSize] = msize
	ret[BucketCacheSize] = bsize
	return ret
}

// CompositeJournalKey constructs journal key with prefix and block number.
func CompositeJournalKey(blockNumber uint64) []byte {
	s := strconv.FormatUint(blockNumber, 10)
	return append([]byte(journalPrefix), []byte(s)...)
}

/*
	Configuration
*/

// GetCapacity reads tree capacity.
func (stateDB *StateDB) GetCapacity(typ string) int {
	switch typ {
	case STATEDB:
		return stateDB.conf.GetInt(StateCapacity)
	case STATEOBJ:
		return stateDB.conf.GetInt(StateObjectCapacity)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJ)
		return 0
	}
}

// GetAggreation reads tree aggreation.
func (stateDB *StateDB) GetAggreation(typ string) int {
	switch typ {
	case STATEDB:
		return stateDB.conf.GetInt(StateAggreation)
	case STATEOBJ:
		return stateDB.conf.GetInt(StateObjectAggreation)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJ)
		return 0
	}
}

// GetMerkleCacheSize gets tree merkle cache max size.
func (stateDB *StateDB) GetMerkleCacheSize(typ string) int {
	switch typ {
	case STATEDB:
		return stateDB.conf.GetInt(StateMerkleCacheSize)
	case STATEOBJ:
		return stateDB.conf.GetInt(StateObjectMerkleCacheSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJ)
		return 0
	}
}

// GetMerkleCacheSize gets tree bucket cache max size.
func (stateDB *StateDB) GetBucketCacheSize(typ string) int {
	switch typ {
	case STATEDB:
		return stateDB.conf.GetInt(StateBucketCacheSize)
	case STATEOBJ:
		return stateDB.conf.GetInt(StateObjectBucketCacheSize)
	default:
		stateDB.logger.Errorf("no choice specified. %s or %s", STATEDB, STATEOBJ)
		return 0
	}
}
