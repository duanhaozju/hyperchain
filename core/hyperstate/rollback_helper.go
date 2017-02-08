package hyperstate

import (
	"hyperchain/common"
	"hyperchain/hyperdb"
)

type JournalCache struct {
	stateObjects      map[common.Address]*StateObject
	db                hyperdb.Database
}


func NewJournalCache(db hyperdb.Database) *JournalCache {
	return &JournalCache{
		db: db,
		stateObjects: make(map[common.Address]*StateObject),
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
func (cache *JournalCache) Flush(batch hyperdb.Batch) error {
	for _, stateObject := range cache.stateObjects {
		if stateObject.suicided || (deleteEmptyObjects && stateObject.empty()) {
			log.Debugf("state object %s been suicide or clearing out for empty", stateObject.address.Hex())
			cache.deleteStateObject(batch, stateObject)
		} else {
			log.Debugf("state object %s been updated", stateObject.address.Hex())
			// Write any storage changes in the state object to its storage trie.
			for key, value := range stateObject.dirtyStorage {
				delete(stateObject.dirtyStorage, key)
				if (value == common.Hash{}) {
					// delete
					log.Debugf("flush dirty storage address [%s] delete item key: [%s]", stateObject.address.Hex(), key.Hex())
					if err := batch.Delete(CompositeStorageKey(stateObject.address.Bytes(), key.Bytes())); err != nil {
						return err
					}
				} else {
					log.Debugf("flush dirty storage address [%s] put item key: [%s], value [%s]", stateObject.address.Hex(), key.Hex(), value.Hex())
					if err := batch.Put(CompositeStorageKey(stateObject.address.Bytes(), key.Bytes()), value.Bytes()); err != nil {
						return err
					}
				}
			}
			// Update the object in the main account trie.
			cache.updateStateObject(batch, stateObject)
		}
	}
	return nil
}


// deleteStateObject - removes the given object from the database.
func (cache *JournalCache) deleteStateObject(batch hyperdb.Batch, stateObject *StateObject) {
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
func (cache *JournalCache) updateStateObject(batch hyperdb.Batch, stateObject *StateObject) {
	addr := stateObject.Address()
	data, err := stateObject.Marshal()
	if err != nil {
		log.Error("marshal stateobject failed", addr.Hex())
	}
	batch.Put(CompositeAccountKey(addr.Bytes()), data)
}
