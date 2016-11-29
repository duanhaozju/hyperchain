//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"time"
	"hyperchain/common"
)
var (
	PublicDB *PublicStateDB
)
type PublicStateObject struct{
	State_object	*StateObject
	Frequence	int
	Timestamp	time.Time
	LatestBlockNum	int
}

type PublicStateDB struct{
	PublicStateObjectMap	map[common.Address]*PublicStateObject
}

func init(){
	PublicDB = &PublicStateDB{}
}

// clear all objects
func (self *PublicStateDB) ClearAll(){
	PublicDB = &PublicStateDB{}
}

// clear all objects
func (self *PublicStateDB) Delete(addr common.Address){
	delete(self.PublicStateObjectMap,addr)
}

func (self *PublicStateDB) Update(addr common.Address,publicStateObject PublicStateObject)  {
	self.PublicStateObjectMap[addr] = &publicStateObject
}

func (self *PublicStateDB) GetPublicStateObject(addr common.Address) *PublicStateObject {
	if _,ok := self.PublicStateObjectMap[addr];ok{
		return self.PublicStateObjectMap[addr]
	}else{
		return nil
	}
}

// clear old PublicStateObject by frequence
func (self *PublicStateDB) ClearStateObjectByFrequence(frequence int){
	for addr,publicStateObject := range self.PublicStateObjectMap{
		if(publicStateObject.Frequence<frequence){
			delete(PublicDB.PublicStateObjectMap,addr)
		}
	}
}

// clear old PublicStateObject by timestamp
func (self *PublicStateDB) ClearStateObjectByTimestamp(timestamp time.Time){
	for addr,publicStateObject := range self.PublicStateObjectMap{
		if(publicStateObject.Timestamp.Before(timestamp)){
			delete(PublicDB.PublicStateObjectMap,addr)
		}
	}
}

// clear old PublicStateObject by LatestBlockNum
func (self *PublicStateDB) ClearStateObjectByLatestBlockNum(blockNum int){
	for addr,publicStateObject := range self.PublicStateObjectMap{
		if(publicStateObject.LatestBlockNum < blockNum){
			delete(PublicDB.PublicStateObjectMap,addr)
		}
	}
}