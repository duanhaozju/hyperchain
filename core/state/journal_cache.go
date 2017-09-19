package state

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"hyperchain/tree/bucket"
)

const (
	WORKINGSET_TYPE_STATE       = 0
	WORKINGSET_TYPE_STATEOBJECT = 1
)

type JournalCache struct {
	stateObjects           map[common.Address]*StateObject
	db                     db.Database
	stateWorkingSet        bucket.K_VMap
	stateObjectsWorkingSet map[common.Address]bucket.K_VMap
	logger                 *logging.Logger
}

func NewJournalCache(db db.Database, logger *logging.Logger) *JournalCache {
	return &JournalCache{
		db:                     db,
		logger:                 logger,
		stateObjects:           make(map[common.Address]*StateObject),
		stateWorkingSet:        bucket.NewKVMap(),
		stateObjectsWorkingSet: make(map[common.Address]bucket.K_VMap),
	}
}

func (cache *JournalCache) Add(stateObject *StateObject) {
	cache.stateObjects[stateObject.address] = stateObject
}

// Fetch - fetch a object with specific address.
// load from database if not hit in cache.
func (cache *JournalCache) Fetch(address common.Address) *StateObject {
	if obj, existed := cache.stateObjects[address]; existed == true {
		return obj
	}
	// load from database
	// Load the object from the database.
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

// Create - create a empty object.
func (cache *JournalCache) Create(address common.Address, s *StateDB) *StateObject {
	obj := newObject(s, address, Account{}, nil, false, nil, s.logger)
	cache.stateObjects[address] = obj
	return obj
}

// Flush - flush all modification to batch.
func (cache *JournalCache) Flush(batch db.Batch) error {
	for _, stateObject := range cache.stateObjects {
		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			cache.logger.Debugf("state object %s been suicide or clearing out for empty", stateObject.address.Hex())
			workingSet := bucket.NewKVMap()
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
			cache.stateObjectsWorkingSet[stateObject.address] = workingSet
		} else {
			cache.logger.Debugf("state object %s been updated", stateObject.address.Hex())
			// Write any storage changes in the state object to its storage trie.
			workingSet := bucket.NewKVMap()
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
			cache.stateObjectsWorkingSet[stateObject.address] = workingSet
		}
	}
	return nil
}

// GetWorkingSet - return required working set.
func (cache *JournalCache) GetWorkingSet(workingSetType int, address common.Address) bucket.K_VMap {
	switch workingSetType {
	case WORKINGSET_TYPE_STATE:
		return cache.stateWorkingSet
	case WORKINGSET_TYPE_STATEOBJECT:
		return cache.stateObjectsWorkingSet[address]
	default:
		return nil
	}
}

// deleteStateObject - removes the given object from the database.
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

// updateStateObject - writes the given object to the database
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
