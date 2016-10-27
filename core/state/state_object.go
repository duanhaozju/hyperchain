package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hyperchain/common"
	"hyperchain/core/crypto"
	"hyperchain/trie"
	"math/big"
)

var emptyCodeHash = crypto.Keccak256(nil)

type Code []byte
type ABI []byte

func (self Code) String() string {
	return string(self)
}

type Storage map[common.Hash]common.Hash

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

type StateObject struct {
	// Address belonging to this account
	address common.Address
	db      trie.Database // State database for storing state changes

	// DB error
	// the error which StateObject cannot be deal with will eventually be returned by StateDB.Commit
	dbErr error

	// Used to store account Storage
	trie *trie.SecureTrie
	// The BalanceData of the account
	BalanceData *big.Int
	// The nonce of the account
	nonce uint64
	// The code hash if code is present (i.e. a contract)
	codeHash []byte
	// The code for this account
	code Code
	// The ABI for this account
	abi ABI
	// Cached storage (flushed when updated)
	storage Storage

	remove  bool
	deleted bool
	dirty   bool
}

func NewStateObject(address common.Address, db trie.Database) *StateObject {
	object := &StateObject{
		db:          db,
		address:     address,
		BalanceData: new(big.Int),
		dirty:       true,
		codeHash:    emptyCodeHash,
		storage:     make(Storage),
	}
	object.trie, _ = trie.NewSecure(common.Hash{}, db)
	return object
}

func (self *StateObject) MarkForDeletion() {
	self.remove = true
	self.dirty = true
}

func (c *StateObject) getAddr(addr common.Hash) common.Hash {
	ret := c.trie.Get(addr[:])
	return common.BytesToHash(ret)
}

func (c *StateObject) setAddr(addr, value common.Hash) {
	c.trie.Update(addr[:], value[:])
}

func (self *StateObject) Storage() Storage {
	return self.storage
}

func (self *StateObject) GetState(key common.Hash) common.Hash {
	value, exists := self.storage[key]
	if !exists {
		value = self.getAddr(key)
		if (value != common.Hash{}) {
			self.storage[key] = value
		}
	}
	return value
}

func (self *StateObject) SetState(key, value common.Hash) {
	self.storage[key] = value
	self.dirty = true
}

// Update updates the current cached storage to the trie
func (self *StateObject) Update() {
	for key, value := range self.storage {
		if (value == common.Hash{}) {
			self.trie.Delete(key[:])
			continue
		}
		self.setAddr(key, value)
	}
}

// setError remembers the first non-nil error it is called with
func (self *StateObject) setError(err error){
	if self.dbErr == nil{
		self.dbErr = err
	}
}

func (c *StateObject) AddBalance(amount *big.Int) {
	c.SetBalance(new(big.Int).Add(c.BalanceData, amount))
}

func (c *StateObject) SubBalance(amount *big.Int) {
	c.SetBalance(new(big.Int).Sub(c.BalanceData, amount))
}

func (c *StateObject) SetBalance(amount *big.Int) {
	c.BalanceData = amount
	c.dirty = true
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *StateObject) ReturnGas(gas, price *big.Int) {}

func (self *StateObject) Copy() *StateObject {
	stateObject := NewStateObject(self.Address(), self.db)
	stateObject.BalanceData.Set(self.BalanceData)
	stateObject.codeHash = common.CopyBytes(self.codeHash)
	stateObject.nonce = self.nonce
	stateObject.trie = self.trie
	stateObject.code = common.CopyBytes(self.code)
	stateObject.abi = common.CopyBytes(self.abi)
	stateObject.storage = self.storage.Copy()
	stateObject.remove = self.remove
	stateObject.dirty = self.dirty
	stateObject.deleted = self.deleted
	return stateObject
}

//
// Attribute accessors
//

func (self *StateObject) Balance() *big.Int {
	return self.BalanceData
}

// Returns the address of the contract/account
func (c *StateObject) Address() common.Address {
	return c.address
}

func (self *StateObject) Trie() *trie.SecureTrie {
	return self.trie
}

func (self *StateObject) Root() []byte {
	return self.trie.Root()
}

func (self *StateObject) ABI() []byte {
	return self.abi
}

func (self *StateObject) SetABI(abi []byte) {
	self.abi = abi
	self.dirty = true
}

// Code returns the contract code associated with this object,if any
// TODO the code could be stored in cache
func (self *StateObject) Code(db trie.Database) []byte {
	if self.code != nil{
		return self.code
	}
	if bytes.Equal(self.CodeHash(),emptyCodeHash){
		return nil
	}
	code,err := db.Get(self.CodeHash())
	if err != nil{
		self.setError(fmt.Errorf("can't load code hash %x: %v",self.CodeHash(),err))
	}
	self.code = code
	return code
}


func (self *StateObject) SetCode(code []byte) {
	self.code = code
	self.codeHash = crypto.Keccak256(code)
	self.dirty = true
}

func (self *StateObject) SetNonce(nonce uint64) {
	self.nonce = nonce
	self.dirty = true
}

func (self *StateObject) Nonce() uint64 {
	return self.nonce
}

// Never called, but must be present to allow StateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (self *StateObject) Value() *big.Int {
	panic("Value on StateObject should never be called")
}

func (self *StateObject) ForEachStorage(cb func(key, value common.Hash) bool) {
	// When iterating over the storage check the cache first
	for h, value := range self.storage {
		cb(h, value)
	}
	it := self.trie.Iterator()
	for it.Next() {
		key := common.BytesToHash(self.trie.GetKey(it.Key))
		if _, ok := self.storage[key]; !ok {
			cb(key, common.BytesToHash(it.Value))
		}
	}
}

// just for test
func (self *StateObject) PrintStorages() {
	// When iterating over the storage check the cache first
	log.Info("+++++++++++++++++the address of stateObject is,", common.ToHex(self.address.Bytes()), "+++++++++++++++++")
	for k, v := range self.Storage() {
		log.Info("storage key is ------", k, "value is -----", v)
	}
}

func (self *StateObject) CodeHash() []byte{
	return self.codeHash
}

type extStateObject struct {
	Nonce       uint64
	BalanceData *big.Int
	Root        common.Hash
	CodeHash    []byte
	Abi         ABI
}

func (self *StateObject) EncodeObject() ([]byte, error) {
	ext := extStateObject{
		Nonce:       self.nonce,
		BalanceData: self.BalanceData,
		Root:        self.trie.Hash(),
		CodeHash:    self.codeHash,
		Abi:         self.abi,
	}
	//self.trie.CommitTo(self.db)
	//self.db.Put(self.codeHash, self.code)
	return json.Marshal(ext)
}

func DecodeObject(address common.Address, db trie.Database, data []byte) (*StateObject, error) {
	var (
		obj = &StateObject{
			address: address,
			db:      db,
			storage: make(Storage),
		}
		ext extStateObject
		err error
	)
	err = json.Unmarshal(data, &ext)
	if err != nil {
		return nil, err
	}
	if obj.trie, err = trie.NewSecure(ext.Root, db); err != nil {
		return nil, err
	}
	if !bytes.Equal(ext.CodeHash, emptyCodeHash) {
		if obj.code, err = db.Get(ext.CodeHash); err != nil {
			return nil, fmt.Errorf("can't get code for hash %x: %v", ext.CodeHash, err)
		}
	}
	obj.nonce = ext.Nonce
	obj.BalanceData = ext.BalanceData
	obj.codeHash = ext.CodeHash
	obj.abi = ext.Abi
	return obj, nil
}
