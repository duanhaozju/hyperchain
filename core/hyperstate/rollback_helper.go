package hyperstate

import (
	"hyperchain/common"
	"hyperchain/tree/bucket"
	"hyperchain/hyperdb/db"
)

const (
	WORKINGSET_TYPE_STATE = 0
	WORKINGSET_TYPE_STATEOBJECT = 1
)

type JournalCache struct {
	stateObjects      map[common.Address]*StateObject
	db                db.Database
	stateWorkingSet   bucket.K_VMap
	stateObjectsWorkingSet map[common.Address]bucket.K_VMap
}


func NewJournalCache(db db.Database) *JournalCache {
	return &JournalCache{
		db: db,
		stateObjects: make(map[common.Address]*StateObject),
		stateWorkingSet: bucket.NewKVMap(),
		stateObjectsWorkingSet: make(map[common.Address]bucket.K_VMap),
	}
}

func (cache *JournalCache) Add(stateObject *StateObject) {
	cache.stateObjects[stateObject.address] = stateObject
}

// Fetch - fetch a object with specific address.
// load from database if not hit in cache.
func (cache *JournalCache) Fetch (address common.Address) *StateObject {
	if obj, existed := cache.stateObjects[address]; existed == true {
		return obj
	}
	// load from database
	// Load the object from the database.
	data, err := cache.db.Get(CompositeAccountKey(address.Bytes()))
	if err != nil {
		log.Debugf("no state object been find")
		return nil
	}
	var account Account
	err = Unmarshal(data, &account)
	if err != nil {
		return nil
	}
	// Insert into the live set.
	log.Debugf("find state object %s in database, add it to live objects", address.Hex())
	obj := &StateObject{
		address: address,
		data:	  account,
		cachedStorage: make(Storage),
		dirtyStorage: make(Storage),
	}
	cache.stateObjects[address] = obj
	return obj
}

// Flush - flush all modification to batch.
func (cache *JournalCache) Flush(batch db.Batch) error {
	for _, stateObject := range cache.stateObjects {
		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			log.Debugf("state object %s been suicide or clearing out for empty", stateObject.address.Hex())
			workingSet := bucket.NewKVMap()
			for key, value := range stateObject.dirtyStorage {
				delete(stateObject.dirtyStorage, key)
				if (value == common.Hash{}) {
					// delete
					workingSet[key.Hex()] = nil
				} else {
					workingSet[key.Hex()] = value.Bytes()
				}
			}
			cache.deleteStateObject(batch, stateObject)
			cache.stateWorkingSet[stateObject.address.Hex()] = nil
			cache.stateObjectsWorkingSet[stateObject.address] = workingSet
		} else {
			log.Debugf("state object %s been updated", stateObject.address.Hex())
			// Write any storage changes in the state object to its storage trie.
			workingSet := bucket.NewKVMap()
			for key, value := range stateObject.dirtyStorage {
				delete(stateObject.dirtyStorage, key)
				if (value == common.Hash{}) {
					// delete
					log.Debugf("flush dirty storage address [%s] delete item key: [%s]", stateObject.address.Hex(), key.Hex())
					if err := batch.Delete(CompositeStorageKey(stateObject.address.Bytes(), key.Bytes())); err != nil {
						return err
					}
					workingSet[key.Hex()] = nil
				} else {
					log.Debugf("flush dirty storage address [%s] put item key: [%s], value [%s]", stateObject.address.Hex(), key.Hex(), value.Hex())
					if err := batch.Put(CompositeStorageKey(stateObject.address.Bytes(), key.Bytes()), value.Bytes()); err != nil {
						return err
					}
					workingSet[key.Hex()] = value.Bytes()
				}
			}
			// Update the object in the main account trie.
			data := cache.updateStateObject(batch, stateObject)
			cache.stateWorkingSet[stateObject.address.Hex()] = data
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
	data, err := stateObject.MarshalJSON()
	if err != nil {
		log.Error("marshal stateobject failed", addr.Hex())
		return nil
	}
	batch.Put(CompositeAccountKey(addr.Bytes()), data)
	return data
}
