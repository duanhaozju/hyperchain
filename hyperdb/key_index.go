//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperdb

import (
	"github.com/willf/bloom"
	"github.com/syndtr/goleveldb/leveldb"
	"time"
	"github.com/syndtr/goleveldb/leveldb/util"
	"os"
	//"bytes"
)

//IndexManager manage indexes by the namespace.
type IndexManager struct {
	indexs map[string] Index
}

var bloomPath = "bloom.dat"

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
}

//KeyIndex index implement for key-value store based on BloomFilter.
type KeyIndex struct {
	namespace          string
	bf                 *bloom.BloomFilter
	db                 *leveldb.DB
	keyPrefix          []byte
	lastStartKeyPrefix []byte
	currStartKeyPrefix []byte
}

//Namespace get namespace of this keyindex instance
func (ki KeyIndex) Namespace() string  {
	return ki.namespace
}

//Contains judge whether the key is stored in the underlying database
func (ki KeyIndex) Contains(key []byte) bool {
	//TODO: Contains(key []byte)
	return true
}

//MayContains if return true then the key may stored in the underlying database
//of return false, the key must not stored in the underlying database
func (ki KeyIndex) MayContains(key []byte) bool {
	return ki.bf.Test(key)
}

//GetIndex get the index of the specify key
func (ki KeyIndex) GetIndex(key []byte) interface {} {
	log.Debugf("Get index for key: %v", key)
	return nil
}

func (ki KeyIndex) lastStartKey() []byte  {
	return []byte(ki.namespace + "_start_key")
}

//Init the index
func (ki KeyIndex) Init() error  {
	log.Debug("Init KeyIndex")
	var err error
	ki.lastStartKeyPrefix, err =  ki.db.Get(ki.lastStartKey(), nil)
	if err != nil {
		return err
	}
	log.Debugf("Last start key index: %v", ki.lastStartKeyPrefix)
	ki.currStartKeyPrefix = []byte(string(ki.keyPrefix) + string(time.Now().UnixNano()) + ".")
	ki.keyPrefix = ki.currStartKeyPrefix
	log.Debugf("Current start key index: %v", ki.currStartKeyPrefix)
	if ok, err := ki.db.Has(ki.lastStartKeyPrefix, nil); ok && err == nil {
		ki.Rebuild()
	}

	ki.db.Put(ki.lastStartKey(), ki.currStartKeyPrefix, nil)
	return err
}

//Rebuild rebuild the index
func (ki KeyIndex) Rebuild() error {
	it := ki.db.NewIterator(util.BytesPrefix(ki.lastStartKeyPrefix), nil)
	for ;it.Next(); {
		key := it.Key()
		ki.bf.Add(key)
	}
	return nil
}

//NewKeyIndex new KeyIndex instance
func NewKeyIndex(ns string, db *leveldb.DB) Index {
	filter := bloom.New(10000 * 10000, 3)
	return &KeyIndex{
		namespace:ns,
		bf:filter,
		db:db,
		keyPrefix:[]byte(ns + "_bloom_key."),
	}
}

//AddIndexForKey add index for a specify key
func (ki KeyIndex) AddIndexForKey(key []byte) error {
	ki.bf.Add(key)
	//ki.persistKey(key)
	return nil
}

//persist the key into the database
func (ki KeyIndex) persistKey(key []byte) error {
	//key = append(ki.keyPrefix, key)

	key = []byte(string(ki.keyPrefix) + string(key))
	return ki.db.Put(key, []byte(time.Now().String()), nil)
}

func (ki KeyIndex) dropPreviousKey() error  {
	it := ki.db.NewIterator(util.BytesPrefix(ki.lastStartKeyPrefix), nil)
	for ;it.Next(); {
		key := it.Key()
		ki.db.Delete(key, nil)
	}
	return nil
}

//persistBloom persist bloom filter
func (ki KeyIndex) persistBloom () error {
	inputFile, err := os.Open(bloomPath)
	if err == nil {
		ki.bf.WriteTo(inputFile)
	}else {
		inputFile, err = os.Create(bloomPath)
		if err != nil {
			ki.bf.WriteTo(inputFile)
		}

	}
	return err
}