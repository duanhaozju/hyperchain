//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"github.com/willf/bloom"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	"fmt"
	"strconv"
	"strings"
)

//IndexManager manage indexes by the namespace.
type IndexManager struct {
	indexs map[string] Index
}

var bloomPath = "./build/index/index.bloom.dat" //TODO: fix me

//GetIndex get index by the namespace.
func (im IndexManager) GetIndex(ns string) Index {
	if index, ok := im.indexs[ns]; ok {
		return index
	}
	return nil
}

//AddIndex add index if not existed.
func (im IndexManager) AddIndex(i Index)  {
	if _, ok := im.indexs[i.Namespace()]; ok {
		log.Warningf("index existed: %v", i)
	}
	im.indexs[i.Namespace()] = i
}

//Index interface of the data index.
type Index interface {
	Namespace() string
	Contains(key []byte) bool
	MayContains(key []byte) bool
	GetIndex(key []byte) interface{}
	AddIndexForKey(key []byte) error
	Init() error
	Rebuild() error
	Persist() error
	PersistKeyBatch() error
}

//KeyIndex index implement for key-value store based on BloomFilter.
type KeyIndex struct {
	namespace          string
	bf                 *bloom.BloomFilter
	db                 *leveldb.DB
	keyPrefix          []byte
	lastStartKeyPrefix []byte
	currStartKeyPrefix []byte
	keyBatch           *leveldb.Batch
}

//NewKeyIndex new KeyIndex instance
func NewKeyIndex(ns string, db *leveldb.DB) *KeyIndex {
	filter := bloom.New(1 * 10000 * 10000, 3) //todo: fix it
	index := &KeyIndex{
		namespace:ns,
		bf:filter,
		db:db,
		keyPrefix:[]byte(ns + "_bloom_key."),
	}
	index.keyBatch = new(leveldb.Batch)
	index.Init()
	return index
}

//Namespace get namespace of this keyindex instance
func (ki *KeyIndex) Namespace() string  {
	return ki.namespace
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
func (ki *KeyIndex) GetIndex(key []byte) interface {} {
	log.Debugf("Get index for key: %v", key)
	return nil
}

//AddIndexForKey add index for a specify key
func (ki *KeyIndex) AddIndexForKey(key []byte) error {
	ki.bf.Add(key)
	ki.persistKey(key)
	return nil
}

//Init the index
func (ki *KeyIndex) Init() error  {
	log.Debug("Init KeyIndex")
	var err error
	ki.lastStartKeyPrefix, err =  ki.db.Get(ki.lastStartKey(), nil)
	firstStart := (err == leveldb.ErrNotFound)
	if err != nil && err != leveldb.ErrNotFound {
		log.Noticef("fetch lastStartKey error, %v", err)
		return err
	}
	log.Debugf("Last start key index: %v", string(ki.lastStartKeyPrefix))
	ki.currStartKeyPrefix = ki.checkPointKey()
	ki.keyPrefix = ki.currStartKeyPrefix
	log.Debugf("Current start key index: %v", string(ki.currStartKeyPrefix))
	err = ki.Rebuild()
	if err != nil{
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
	log.Noticef("load bloom from local file, file name: %s", bloomPath)
	bfile, err := os.Open(bloomPath)
	if err != nil {
		log.Warningf("load bloom filter with file %s error %v ", bloomPath, err)
		err = nil
		//return err
	}else {
		size, err := ki.bf.ReadFrom(bfile)
		if err != nil {
			log.Errorf("read data from bloom file error %v", err)
			return err
		}
		log.Noticef("read %d bytes data from bloom file %s", size, bloomPath)
	}

	//2.add the recent un_persist keys into the bloom
	if ki.lastStartKeyPrefix != nil{
		r := &util.Range{
			Start:ki.lastStartKeyPrefix,
			Limit:ki.currStartKeyPrefix,
		}
		it := ki.db.NewIterator(r, nil)
		log.Noticef("start load recent un_persist keys into bloom")
		for ; it.Next(); {
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
func (ki *KeyIndex) persistKey(key []byte) error {
	nkey := []byte(string(ki.keyPrefix) + string(key))
	ki.keyBatch.Put(nkey, key)
	return nil
}

//dropPreviousKey drop keys which have been added into bloom and persisted
//drop keys which is generated a day ago
//invoked after persistBloom
func (ki *KeyIndex) dropPreviousKey() error  {

	start := ki.newCheckPointKey((time.Now().UnixNano() - 24 * time.Hour.Nanoseconds()))

	r := &util.Range{
		Start:start,
		Limit:ki.checkPointKey(),
	}
	it := ki.db.NewIterator(r, nil)
	for ; it.Next(); {
		key := it.Key()
		log.Debugf("delete key %s", string(key))
		ki.db.Delete(key, nil)
	}
	return nil
}

//persistBloom persist bloom filter
//1.persist current bloom into a tmp file
//2.rename tmp file
func (ki *KeyIndex) persistBloom () error {
	tmpName := bloomPath+".tmp." + strconv.FormatInt(time.Now().UnixNano(), 10)
	inputFile, err := os.Open(tmpName)
	if err == nil {
		size, err := ki.bf.WriteTo(inputFile)
		fmt.Printf("persist bloom filter for namespace: %s size: %d\n", ki.Namespace(), size)
		return err
	}else {//no such file
		err = os.MkdirAll("./build/index", 0777) //TODO: fix it
		if err != nil {
			log.Errorf("persist bloom error %v", err)
			return err
		}
		inputFile, err = os.Create(tmpName)
		if err == nil {
			size, err := ki.bf.WriteTo(inputFile)
			log.Debugf("persist bloom filter for namespace: %s size: %d", ki.Namespace(), size)
			if err == nil {
				err = os.Rename(tmpName, bloomPath)
			}
			return err
		}
	}
	//warn: should reset lastStartKey only after the bloom filter has been persisted
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
	if err != nil{
		log.Error(err.Error())
		return err
	}
	err = ki.dropPreviousKey()
	return err
}

//PersistKeyBatch persist keys at level of batch
func (ki *KeyIndex) PersistKeyBatch() error {
	err := ki.db.Write(ki.keyBatch, nil)
	if err == nil {// keyBatch write success
		ki.keyBatch.Reset()
	}
	return err
}

//lastStartKey key the fetch last Start key
func (ki *KeyIndex) lastStartKey() []byte  {
	return []byte(ki.namespace + "_start_key")
}

func (ki *KeyIndex) checkPointKey() []byte {
	return []byte("chk." + ki.namespace + "_bloom_key." + strconv.FormatInt(time.Now().UnixNano(), 10) + ".")
}

func (ki *KeyIndex) newCheckPointKey(seq int64) []byte {
	return []byte("chk." + ki.namespace + "_bloom_key." + strconv.FormatInt(seq, 10) + ".")
}