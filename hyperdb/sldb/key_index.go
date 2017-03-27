//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package sldb

import (
	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/util"
	"github.com/willf/bloom"
	"hyperchain/common"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	SLDB_INDEX_BLOOM_BIT_NUM = "dbConfig.sldb.index.bloombit"
	SLDB_INDEX_HASH_NUM      = "dbConfig.sldb.index.hashnum"
	SLDB_INDEX_DUMP_INTERVAL = "dbConfig.sldb.index.dumpinterval"
	SLDB_INDEX_DIR           = "dbConfig.sldb.index.dir"
)

//Index interface of the data index.
type Index interface {
	Namespace() string
	Contains(key []byte) bool
	MayContains(key []byte) bool
	GetIndex(key []byte) interface{}
	AddIndexForKey(key []byte, indexBatch *leveldb.Batch) error
	AddAndPersistIndexForKey(key []byte) error
	Init() error
	Rebuild() error
	Persist() error
}

//KeyIndex index implement for key-value store based on BloomFilter.
type KeyIndex struct {
	namespace          string
	bf                 *bloom.BloomFilter
	db                 *leveldb.DB
	keyPrefix          []byte
	keyPrefixLock      *sync.RWMutex
	lastStartKeyPrefix []byte
	currStartKeyPrefix []byte
	bloomPath          string
	conf               *common.Config
}

//NewKeyIndex new KeyIndex instance
func NewKeyIndex(conf *common.Config, ns string, db *leveldb.DB, path string) *KeyIndex {
	filter := bloom.New(uint(conf.GetInt(SLDB_INDEX_BLOOM_BIT_NUM)),
		uint(conf.GetInt(SLDB_INDEX_HASH_NUM)))

	index := &KeyIndex{
		namespace:     ns,
		bf:            filter,
		db:            db,
		keyPrefix:     []byte(ns + "_bloom_key."),
		bloomPath:     path,
		keyPrefixLock: new(sync.RWMutex),
	}
	index.Init()
	return index
}

//Namespace get namespace of this keyindex instance
func (ki *KeyIndex) Namespace() string {
	return ki.namespace
}

func (ki *KeyIndex) SetNamespace(ns string) {
	ki.namespace = ns
}

//Contains judge whether the key is stored in the underlying database
func (ki *KeyIndex) Contains(key []byte) bool {
	//TODO: Contains(key []byte)
	return true
}

//MayContains if return true then the key may stored in the underlying database
//of return false, the key must not stored in the underlying database
func (ki *KeyIndex) MayContains(key []byte) bool {
	return ki.bf.Test(key)
}

//GetIndex get the index of the specify key
func (ki *KeyIndex) GetIndex(key []byte) interface{} {
	log.Debugf("Get index for key: %v", key)
	return nil
}

//AddIndexForKey add index for a specify key
func (ki *KeyIndex) AddIndexForKey(key []byte, indexBatch *leveldb.Batch) error {
	ki.bf.Add(key)
	ki.addKeyIndexIntoBatch(key, indexBatch)
	return nil
}

//AddAndPersistIndexForKey add index for key and persist the key
func (ki *KeyIndex) AddAndPersistIndexForKey(key []byte) error {
	ki.bf.Add(key)
	ki.persistKey(key)
	return nil
}

//Init the index
func (ki *KeyIndex) Init() error {
	log.Debug("Init KeyIndex")
	var err error
	ki.lastStartKeyPrefix, err = ki.db.Get(ki.lastStartKey(), nil)
	firstStart := err == leveldb.ErrNotFound
	if err != nil && err != leveldb.ErrNotFound {
		log.Noticef("fetch lastStartKey error, %v", err)
		return err
	}
	log.Debugf("Last start key index: %v", string(ki.lastStartKeyPrefix))
	ki.currStartKeyPrefix = ki.checkPointKey()
	ki.keyPrefix = ki.currStartKeyPrefix
	log.Debugf("Current start key index: %v", string(ki.currStartKeyPrefix))
	err = ki.Rebuild()
	if err != nil {
		return err
	}
	if firstStart {
		ki.db.Put(ki.lastStartKey(), ki.currStartKeyPrefix, nil)
	}
	return err
}

//Rebuild rebuild the index
func (ki *KeyIndex) Rebuild() error {
	//1.load bloom from local file
	log.Noticef("load bloom from local file, file name: %s", ki.bloomPath)
	bfile, err := os.Open(ki.bloomPath)
	if err != nil {
		log.Warningf("load bloom filter with file %s error %v ", ki.bloomPath, err)
		err = nil
		//return err
	} else {
		size, err := ki.bf.ReadFrom(bfile)
		if err != nil {
			log.Errorf("read data from bloom file error %v", err)
			return err
		}
		log.Noticef("read %d bytes data from bloom file %s", size, ki.bloomPath)
	}

	//2.add the recent un_persist keys into the bloom
	if ki.lastStartKeyPrefix != nil {
		r := &util.Range{
			Start: ki.lastStartKeyPrefix,
			Limit: ki.currStartKeyPrefix,
		}
		it := ki.db.NewIterator(r, nil)
		log.Noticef("start load recent un_persist keys into bloom")
		for it.Next() {
			k := it.Key()
			v := it.Value()
			if !strings.HasPrefix(string(k), "chk.") {
				break
			}
			log.Debugf("add key %s, value: %s", string(k), string(v))
			ki.bf.Add(v)
		}
	}
	return nil
}

//persistKey add key into the batch
func (ki *KeyIndex) addKeyIndexIntoBatch(key []byte, keyBatch *leveldb.Batch) error {
	nkey := []byte(string(ki.keyPrefix) + string(key))
	keyBatch.Put(nkey, key)
	return nil
}

//persistKey add key into the db
func (ki *KeyIndex) persistKey(key []byte) error {
	ki.keyPrefixLock.RLock()
	defer ki.keyPrefixLock.RUnlock()
	nkey := []byte(string(ki.keyPrefix) + string(key))
	return ki.db.Put(nkey, key, nil)
}

//dropPreviousKey drop keys which have been added into bloom and persisted
//drop keys which is generated a day ago
//invoked after persistBloom
func (ki *KeyIndex) dropPreviousKey() error {
	limit := ki.newCheckPointKey(time.Now().UnixNano() - ki.conf.GetDuration(SLDB_INDEX_DUMP_INTERVAL).Nanoseconds())

	r := &util.Range{
		Start: ki.lastStartKeyPrefix,
		Limit: limit,
	}
	it := ki.db.NewIterator(r, nil)
	for it.Next() {
		key := it.Key()
		log.Debugf("delete key %s", string(key))
		ki.db.Delete(key, nil)
	}
	return nil
}

func (ki *KeyIndex) persistBloom() error {
	bloomDir := ki.conf.GetString(SLDB_INDEX_DIR)
	_, error := os.Stat(bloomDir)
	if !(error == nil || os.IsExist(error)) {
		err := os.MkdirAll(bloomDir, 0777)
		if err != nil {
			log.Errorf("make bloom file dir error %v", err)
		}
	}
	tmpName := ki.bloomPath + ".tmp." + strconv.FormatInt(time.Now().UnixNano(), 10)
	inputFile, err := os.Create(tmpName)
	if err != nil {
		log.Errorf("persist bloom filter error %v", err)
		return err
	}
	size, err := ki.bf.WriteTo(inputFile)
	log.Noticef("persist bloom filter for namespace: %s size: %d", ki.Namespace(), size)
	if err == nil {
		err = os.Rename(tmpName, ki.bloomPath)
		if err != nil {
			log.Noticef("rename file %v to %v, error: %v", tmpName, ki.bloomPath, err)
			return err
		}
	} else {
		log.Errorf("persist bloom filter for namespace: %s size: %d err %v", ki.Namespace(), size, err)
		return err
	}
	//warn: should reset lastStartKey only after the bloom filter has been persisted
	ki.keyPrefixLock.Lock()
	defer ki.keyPrefixLock.Unlock()
	err = ki.db.Put(ki.lastStartKey(), ki.currStartKeyPrefix, nil)
	ki.lastStartKeyPrefix = ki.currStartKeyPrefix
	ki.currStartKeyPrefix = ki.checkPointKey()
	ki.keyPrefix = ki.currStartKeyPrefix
	return err
}

//BloomFilter just for test now
func (ki *KeyIndex) Equals(otherKeyIndex *KeyIndex) bool {
	return ki.bf.Equal(otherKeyIndex.bf)
}

//Persist persist the key index
func (ki *KeyIndex) Persist() error {
	log.Noticef("Persist bloom filter")
	var err = ki.persistBloom()
	if err != nil {
		log.Errorf("Persist error %v", err)
		return err
	}
	log.Noticef("drop previous keys")
	err = ki.dropPreviousKey()
	if err != nil {
		log.Error(err.Error())
		return err
	}
	return err
}

//lastStartKey key the fetch last Start key
func (ki *KeyIndex) lastStartKey() []byte {
	return []byte(ki.namespace + "_start_key")
}

func (ki *KeyIndex) checkPointKey() []byte {
	return []byte("chk." + ki.namespace + "_bloom_key." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".")
}

func (ki *KeyIndex) newCheckPointKey(seq int64) []byte {
	return []byte("chk." + ki.namespace + "_bloom_key." + strconv.FormatInt(seq, 10) + ".")
}
