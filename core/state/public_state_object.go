package state
/**
Author: ZhuoHaizhen
Date: 2016.10.18
 */

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

// TODO clear the the PublicStateObject which is less than frequence
func (self *PublicStateDB) ClearLowFrequence(frequence int){
	for addr,publicStateObject := range self.PublicStateObjectMap{
		if(publicStateObject.Frequence<frequence){
			delete(PublicDB.PublicStateObjectMap,addr)
		}
	}
}

// TODO clear the the PublicStateObject which is old than timestamp
func (self *PublicStateDB) ClearOldTimestamp(timestamp time.Time){
	for addr,publicStateObject := range self.PublicStateObjectMap{
		if(publicStateObject.Timestamp.Before(timestamp)){
			delete(PublicDB.PublicStateObjectMap,addr)
		}
	}
}




