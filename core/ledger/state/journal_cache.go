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
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/tree/bucket"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
)

const (
	StateWorkingSet = iota
	ObjectWorkingSet
)

// JournalCache journal cache is an auxiliary helper used in state revert.
// All temporary changes will saved here and commit to db together.
type JournalCache struct {
	stateObjects      map[common.Address]*StateObject
	db                db.Database
	stateWorkingSet   bucket.Entries
	objectsWorkingSet map[common.Address]bucket.Entries
	logger            *logging.Logger
}

// NewJournalCache creates a journal cache with given db.
func NewJournalCache(db db.Database, logger *logging.Logger) *JournalCache {
	return &JournalCache{
		db:                db,
		logger:            logger,
		stateObjects:      make(map[common.Address]*StateObject),
		stateWorkingSet:   bucket.NewEntries(),
		objectsWorkingSet: make(map[common.Address]bucket.Entries),
	}
}

// Add adds a modified object to dirty collection.
// If exists, replace it directly. Since we only care about the last status.
func (cache *JournalCache) Add(stateObject *StateObject) {
	cache.stateObjects[stateObject.address] = stateObject
}

// Fetch fetchs a object with specific address.
// load from database if not hit in cache.
func (cache *JournalCache) Fetch(address common.Address) *StateObject {
	if obj, existed := cache.stateObjects[address]; existed == true {
		return obj
	}
	// load from database
	data, err := cache.db.Get(CompositeAccountKey(address.Bytes()))
	if err != nil {
		cache.logger.Debugf("no state object been find")
		return nil
	}
	var account Account
	err = Unmarshal(data, &account)
	if err != nil {
		return nil
	}
	// Insert into the live set.
	cache.logger.Debugf("find state object %s in database, add it to live objects", address.Hex())
	obj := &StateObject{
		address:       address,
		data:          account,
		cachedStorage: make(Storage),
		dirtyStorage:  make(Storage),
		logger:        cache.logger,
	}
	cache.stateObjects[address] = obj
	return obj
}

// Create creates an empty object.
func (cache *JournalCache) Create(address common.Address, s *StateDB) *StateObject {
	obj := newObject(s, address, Account{}, nil, false, nil, s.logger)
	cache.stateObjects[address] = obj
	return obj
}

// Flush flushs all modification to batch.
func (cache *JournalCache) Flush(batch db.Batch) error {
	for _, stateObject := range cache.stateObjects {
		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			cache.logger.Debugf("state object %s been suicide or clearing out for empty", stateObject.address.Hex())
			workingSet := bucket.NewEntries()
			for key, value := range stateObject.dirtyStorage {
				delete(stateObject.dirtyStorage, key)
				if len(value) == 0 {
					// delete
					workingSet[key.Hex()] = nil
				} else {
					workingSet[key.Hex()] = value
				}
			}
			cache.deleteStateObject(batch, stateObject)
			cache.stateWorkingSet[stateObject.address.Hex()] = nil
			cache.objectsWorkingSet[stateObject.address] = workingSet
		} else {
			cache.logger.Debugf("state object %s been updated", stateObject.address.Hex())
			// Write any storage changes in the state object to its storage trie.
			workingSet := bucket.NewEntries()
			for key, value := range stateObject.dirtyStorage {
				delete(stateObject.dirtyStorage, key)
				if len(value) == 0 {
					// delete
					cache.logger.Debugf("flush dirty storage address [%s] delete item key: [%s]", stateObject.address.Hex(), key.Hex())
					if err := batch.Delete(CompositeStorageKey(stateObject.address.Bytes(), key.Bytes())); err != nil {
						return err
					}
					workingSet[key.Hex()] = nil
				} else {
					cache.logger.Debugf("flush dirty storage address [%s] put item key: [%s], value [%s]", stateObject.address.Hex(), key.Hex(), common.Bytes2Hex(value))
					if err := batch.Put(CompositeStorageKey(stateObject.address.Bytes(), key.Bytes()), value); err != nil {
						return err
					}
					workingSet[key.Hex()] = value
				}
			}
			// Update the object in the main account trie.
			cache.updateStateObject(batch, stateObject)
			d, _ := stateObject.MarshalJSON()
			cache.stateWorkingSet[stateObject.address.Hex()] = d
			cache.objectsWorkingSet[stateObject.address] = workingSet
		}
	}
	return nil
}

// GetWorkingSet returns required working set.
func (cache *JournalCache) getWorkingSet(workingSetType int, address common.Address) bucket.Entries {
	switch workingSetType {
	case StateWorkingSet:
		return cache.stateWorkingSet
	case ObjectWorkingSet:
		return cache.objectsWorkingSet[address]
	default:
		return nil
	}
}

// deleteStateObject removes the given object from the database.
func (cache *JournalCache) deleteStateObject(batch db.Batch, stateObject *StateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	batch.Delete(CompositeAccountKey(addr.Bytes()))
	// delete related storage content
	iter := cache.db.NewIterator(GetStorageKeyPrefix(stateObject.address.Bytes()))
	for iter.Next() {
		batch.Delete(iter.Key())
	}
}

// updateStateObject writes the given object to the database
func (cache *JournalCache) updateStateObject(batch db.Batch, stateObject *StateObject) []byte {
	addr := stateObject.Address()
	data, err := stateObject.Marshal()
	if err != nil {
		cache.logger.Error("marshal stateobject failed", addr.Hex())
		return nil
	}
	batch.Put(CompositeAccountKey(addr.Bytes()), data)
	return data
}
