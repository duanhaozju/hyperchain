package bucket

import (
	"math/big"
	"hyperchain/common"
	"fmt"
	"hyperchain/core/crypto"
	"hyperchain/hyperdb"
	"bytes"
)

var emptyCodeHash = crypto.Keccak256(nil)
type Code []byte
type ABI []byte
type Storage map[common.Hash]common.Hash

type StateObject struct {
	address     common.Address   // Address belonging to this account
	db          hyperdb.Database    // State database for storing state changes
	BalanceData *big.Int         // The BalanceData of the account
	nonce       uint64           // The nonce of the account
	codeHash    []byte           // The code hash if code is present (i.e. a contract)
	code        Code             // The code for this account
	abi         ABI              // The ABI for this account
	storage     Storage          // Cached storage (flushed when updated)
	remove      bool
	deleted     bool
	dirty       bool
}

func NewStateObject(address common.Address, db hyperdb.Database) *StateObject {
	object := &StateObject{
		db:          db,
		address:     address,
		BalanceData: new(big.Int),
		dirty:       true,
		codeHash:    emptyCodeHash,
		storage:     make(Storage),
	}
	return object
}


func (self *StateObject) SubBalance(amount *big.Int) {
	self.SetBalance(new(big.Int).Sub(self.BalanceData, amount))
}

func (self *StateObject) AddBalance(amount *big.Int) {
	self.SetBalance(new(big.Int).Add(self.BalanceData, amount))
}

func (self *StateObject) SetBalance(amount *big.Int) {
	self.BalanceData = amount
	self.dirty = true
}

func (self *StateObject) SetNonce(nonce uint64) {
	self.nonce = nonce
	self.dirty = true
}

func (self *StateObject) Balance() *big.Int {
	return self.BalanceData
}

func (self *StateObject) Address() common.Address {
	return self.address
}

// TODO it should be discarded later
func (self *StateObject) ReturnGas(*big.Int, *big.Int) {}

func (self *StateObject) CodeHash() []byte{
	return self.codeHash
}
// Code returns the contract code associated with this object,if any
// TODO the code could be stored in cache
func (self *StateObject) Code(db hyperdb.Database) []byte {
	if self.code != nil{
		return self.code
	}
	if bytes.Equal(self.CodeHash(),emptyCodeHash){
		return nil
	}
	code,err := db.Get(self.CodeHash())
	if err != nil{
		log.Errorf("can't load code hash %x: %v",self.CodeHash(),err)
		return nil
	}
	self.code = code
	return code
}
func (self *StateObject) SetCode(code []byte) {
	self.code = code
	self.codeHash = crypto.Keccak256(code)
	self.dirty = true
}

func (self *StateObject) ForEachStorage(cb func(key, value common.Hash) bool) {
	// When iterating over the storage check the cache first
	for h, value := range self.storage {
		cb(h, value)
	}
}

func (self *StateObject) PrintStorages() {
	// When iterating over the storage check the cache first
	log.Info("the address of stateObject is ", common.ToHex(self.address.Bytes()))
	for k, v := range self.Storage() {
		log.Info("storage key is ", k, "value is ", v)
	}
}

func (self *StateObject) Value() *big.Int {
	panic("Value on StateObject should never be called")
}

func (self *StateObject) Storage() Storage {
	return self.storage
}

// TODO the key should be considered
func (self *StateObject) GetState(key common.Hash) common.Hash {
	value, exists := self.storage[key]
	if !exists {
		dbKey := append(self.Address().Bytes(),key.Bytes())
		dbValue,err := self.db.Get(dbKey)
		if (err != nil || len(dbValue)==0) {
			log.Debug("get value from db error or there is no data in db ")
		}else if(len(dbValue)>0){
			self.storage[key] = dbValue
		}
		return dbValue
	}else {
		return value
	}
}
func (self *StateObject) SetState(key, value common.Hash) {
	self.storage[key] = value
	self.dirty = true
}

// TODO update the storage
func (self *StateObject) Update() {
	for key, value := range self.storage {
		if (value == common.Hash{}) {
			log.Debug("the key of storage is ",key,"the value of sotrage is ",value )
			continue
		}
	}
}
// TODO maybe should be discarded later
func (self *StateObject) Copy() *StateObject {
	stateObject := NewStateObject(self.Address(), self.db)
	stateObject.BalanceData.Set(self.BalanceData)
	stateObject.codeHash = common.CopyBytes(self.codeHash)
	stateObject.nonce = self.nonce
	stateObject.code = common.CopyBytes(self.code)
	stateObject.abi = common.CopyBytes(self.abi)
	stateObject.storage = self.storage.Copy()
	stateObject.remove = self.remove
	stateObject.dirty = self.dirty
	stateObject.deleted = self.deleted
	return stateObject
}

func (self *StateObject) MarkForDeletion() {
	self.remove = true
	self.dirty = true
}



func (self Code) String() string {
	return string(self)
}

func (self Storage) String() (str string) {
	for key, value := range self {
		str += fmt.Sprintf("%X : %X\n", key, value)
	}
	return
}

func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}
	return cpy
}

