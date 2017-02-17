////Hyperchain License
////Copyright (C) 2016 The Hyperchain Authors.
package hyperdb
//
//import (
//	"github.com/willf/bloom"
//	"github.com/syndtr/goleveldb/leveldb"
//)
//
////IndexManager manage indexes by the namespace
//type IndexManager struct {
//	indexs map[string] Index
//}
//
//type Index interface {
//	Namespace() string
//	Contains(key interface{}) bool
//	MayContains(key interface{}) bool
//	GetIndex(key interface{}) interface{}
//	AddIndexForKey(key interface{}) error
//	Init() error
//	Rebuild() error
//}
//
////index implement for key-value store
////based on BloomFilter
//type KeyIndex struct {
//	namespace string
//	bf        *bloom.BloomFilter
//	db        *leveldb.DB
//}
//
////Namespace get namespace of this keyindex instance
//func (ki KeyIndex) Namespace() string  {
//	return ki.namespace
//}
//
////Contains judge whether the key is stored in the underlying database
//func (ki KeyIndex) Contains(key []byte) bool {
//	//TODO: Contains(key []byte)
//	return false
//}
//
////MayContains if return true then the key may stored in the underlying database
////of return false, the key must not stored in the underlying database
//func (ki KeyIndex) MayContains(key []byte) bool {
//	return ki.bf.Test(key)
//}
//
////GetIndex get the index of the specify key
//func (ki KeyIndex) GetIndex(key interface{}) interface {} {
//	log.Debugf("Get index for key: %v", key)
//	return nil
//}
//
////Init the index
//func (ki KeyIndex) Init() error  {
//	log.Debug("Init KeyIndex")
//	return nil
//}
//
////Rebuild rebuild the index
//func (ki KeyIndex) Rebuild() error {
//
//
//	return nil
//}
//
////NewKeyIndex new KeyIndex instance
//func NewKeyIndex(ns string, db *leveldb.DB) Index {
//	filter := bloom.New(10000 * 10000, 3)
//	return &KeyIndex{
//		namespace:ns,
//		bf:filter,
//		db:db,
//	}
//}
//
//func (ki KeyIndex) AddIndexForKey(key interface{}) error {
//	ki.bf.Add(key)
//	return nil
//}