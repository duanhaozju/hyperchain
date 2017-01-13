//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package vm

import (
	"bytes"
	"encoding/json"
	"fmt"
	"hyperchain/common"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"hyperchain/trie"
	"math/big"
	"sync"
)

type Code []byte
type ABI []byte

var (
	StartingNonce uint64
	emptyCodeHash = crypto.Keccak256(nil)
	//stateObjectCache        map[common.Address] *StateObject	// TODO is it should be used as the lru.Cache
	stateObjectCache = make(map[common.Address]*StateObject)
)

type StateDB struct {
	db   hyperdb.Database
	trie *trie.SecureTrie

	//this map holds 'live' objects, which will get modified while processing a state transition
	stateObjects     map[string]*StateObject
	stateObjectDirty map[common.Address]struct{}

	refund           *big.Int
	thash, bhash     common.Hash
	txIndex          int
	logs             map[common.Hash]Logs
	logSize          uint
	leastStateObject *StateObject
	lock             sync.Mutex
}

// Create a new state from a given trie
func NewStateDB(root common.Hash, db hyperdb.Database) (*StateDB, error) {
	tr, err := trie.NewSecure(root, db)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:               db,
		trie:             tr,
		stateObjects:     make(map[string]*StateObject),
		stateObjectDirty: make(map[common.Address]struct{}),
		refund:           new(big.Int),
		logs:             make(map[common.Hash]Logs),
	}, nil

}

func (self *StateDB) New(root common.Hash) (*StateDB, error) {
	// todo is needed?
	//self.lock.Lock()
	//defer self.lock.Unlock()

	return &StateDB{
		db:               self.db,
		stateObjects:     make(map[string]*StateObject),
		stateObjectDirty: make(map[common.Address]struct{}),
		refund:           new(big.Int),
		logs:             make(map[common.Hash]Logs),
	}, nil
}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset(root common.Hash) error {
	var (
		err error
		tr  = self.trie
	)
	if self.trie.Hash() != root {
		if tr, err = trie.NewSecure(root, self.db); err != nil {
			return err
		}
	}
	self.trie = tr
	self.stateObjects = make(map[string]*StateObject)
	self.stateObjectDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.logs = make(map[common.Hash]Logs)
	self.logSize = 0
	self.refund = new(big.Int)
	// if reset we will clear all stateObjectSizeCache
	stateObjectCache = make(map[common.Address]*StateObject)
	return nil
}

func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (self *StateDB) AddLog(log *Log) {
	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) Logs {
	return self.logs[hash]
}

func (self *StateDB) Logs() Logs {
	var logs Logs
	for _, lgs := range self.logs {
		logs = append(logs, lgs...)
	}
	return logs
}

func (self *StateDB) AddRefund(gas *big.Int) {
	self.refund.Add(self.refund, gas)
}

// Duplicate
func (self *StateDB) HasAccount(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

func (self *StateDB) Exist(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

func (self *StateDB) GetAccount(addr common.Address) Account {
	return self.GetStateObject(addr)
}

func (self *StateDB) GetLeastAccount() Account {
	return self.leastStateObject
}

// todo
func (self *StateDB) SetLeastAccount(account *Account) {
	//self.leastStateObject.abi = account.Address()
}

// return all StateObject saved in the trie instead of in CACHE
func (self *StateDB) GetAccounts() map[string]*StateObject {
	// return self.stateObjects
	ret := make(map[string]*StateObject)
	it := self.trie.Iterator()
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		stateObject, err := DecodeObject(common.BytesToAddress(addr), self.db, it.Value)
		if err != nil {
			continue
		}
		ret[common.BytesToAddress(addr).Hex()] = stateObject
	}
	return ret
}

func (self *StateDB) ForEachAccounts() {}

// Retrieve the BalanceData from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Balance()
	}

	return common.Big0
}

func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.Nonce()
	}
	return StartingNonce
}

func (self *StateDB) GetCode(addr common.Address) []byte {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.code
	}
	return nil
}

func (self *StateDB) GetState(a common.Address, b common.Hash) common.Hash {
	stateObject := self.GetStateObject(a)
	if stateObject != nil {
		return stateObject.GetState(b)
	}
	return common.Hash{}
}

func (self *StateDB) GetABI(addr common.Address) []byte {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.ABI()
	}
	return nil

}

func (self *StateDB) IsDeleted(addr common.Address) bool {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.remove
	}
	return false
}

/*
 * SETTERS
 */

func (self *StateDB) AddBalance(addr common.Address, amount *big.Int) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.AddBalance(amount)
	}
}

func (self *StateDB) SetNonce(addr common.Address, nonce uint64) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetNonce(nonce)
	}
}

func (self *StateDB) SetCode(addr common.Address, code []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetCode(code)
	}
	self.leastStateObject = stateObject
}

func (self *StateDB) SetABI(addr common.Address, abi []byte) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetABI(abi)
	}
}

func (self *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(key, value)
	}
}

func (self *StateDB) Delete(addr common.Address) bool {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		stateObject.MarkForDeletion()
		stateObject.BalanceData = new(big.Int)
		return true
	}

	return false
}

// Update the given state object and apply it to state trie
func (self *StateDB) UpdateStateObject(stateObject *StateObject) {
	addr := stateObject.Address()
	data, err := stateObject.EncodeObject()
	if err != nil {
		panic(fmt.Errorf("can't encode object at %x: %v", addr[:], err))
	}
	self.trie.Update(addr[:], data)
}

// Delete the given state object and delete it from the state trie
func (self *StateDB) DeleteStateObject(stateObject *StateObject) {
	stateObject.deleted = true
	addr := stateObject.Address()
	self.trie.Delete(addr[:])
}

// Retrieve a state object given my the address. Nil if not found
func (self *StateDB) GetStateObject(addr common.Address) (stateObject *StateObject) {
	// 1. we will find the stateObject from the StateDB firstly
	stateObject = self.stateObjects[addr.Str()]
	if stateObject != nil {
		if stateObject.deleted {
			stateObject = nil
		}
		return stateObject
	}

	// 2.we will find the stateObject from stateObjectSizeCache
	stateObject = stateObjectCache[addr]
	if stateObject != nil {
		if stateObject.deleted {
			stateObject = nil
		} else {
			self.SetStateObject(stateObject)
			return stateObject
		}
	}

	// 3.we will find the stateObject from the db by trie
	data := self.trie.Get(addr[:])
	if len(data) == 0 {
		return nil
	}
	stateObject, err := DecodeObject(addr, self.db, data)
	if err != nil {
		return nil
	}
	stateObjectCache[addr] = stateObject
	self.SetStateObject(stateObject)
	return stateObject
}

func (self *StateDB) SetStateObject(object *StateObject) {
	self.stateObjects[object.Address().Str()] = object
}

// Retrieve a state object or create a new state object if nil
func (self *StateDB) GetOrNewStateObject(addr common.Address) *StateObject {
	stateObject := self.GetStateObject(addr)
	if stateObject == nil || stateObject.deleted {
		stateObject = self.CreateStateObject(addr)
	}
	return stateObject
}

//TODO 修改addr的Str 为Hex
// NewStateObject create a state object whether it exist in the trie or not
func (self *StateDB) newStateObject(addr common.Address) *StateObject {
	stateObject := NewStateObject(addr, self.db)
	stateObject.SetNonce(StartingNonce)
	self.stateObjects[addr.Str()] = stateObject
	stateObjectCache[addr] = stateObject
	return stateObject
}

// Creates creates a new state object and takes ownership. This is different from "NewStateObject"
func (self *StateDB) CreateStateObject(addr common.Address) *StateObject {
	// Get previous (if any)
	so := self.GetStateObject(addr)
	// Create a new one
	newSo := self.newStateObject(addr)

	// If it existed set the BalanceData to the new account
	if so != nil {
		newSo.BalanceData = so.BalanceData
	}
	return newSo
}

func (self *StateDB) CreateAccount(addr common.Address) Account {
	return self.CreateStateObject(addr)
}

//
// Setting, copying of the state methods
//

// TODO it could be better
// 1.from the stateObjectCacheSize
func (self *StateDB) Copy() *StateDB {
	state, _ := NewStateDB(common.Hash{}, self.db)
	state.trie = self.trie
	for k, stateObject := range self.stateObjects {
		state.stateObjects[k] = stateObject.Copy()
	}
	state.refund.Set(self.refund)

	for hash, logs := range self.logs {
		state.logs[hash] = make(Logs, len(logs))
		copy(state.logs[hash], logs)
	}
	state.logSize = self.logSize
	return state
}

func (self *StateDB) Set(state *StateDB) {
	self.trie = state.trie
	self.stateObjects = state.stateObjects

	self.refund = state.refund
	self.logs = state.logs
	self.logSize = state.logSize
}

func (self *StateDB) GetRefund() *big.Int {
	return self.refund
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
func (s *StateDB) IntermediateRoot() common.Hash {
	s.refund = new(big.Int)
	for _, stateObject := range s.stateObjects {
		if stateObject.dirty {
			if stateObject.remove {
				s.DeleteStateObject(stateObject)
			} else {
				stateObject.Update()
				s.UpdateStateObject(stateObject)
			}
			stateObject.dirty = false
		}
	}
	return s.trie.Hash()
}

// DeleteSuicides flags the suicided objects for deletion so that it
// won't be referenced again when called / queried up on.
//
// DeleteSuicides should not be used for consensus related updates
// under any circumstances.

func (s *StateDB) DeleteSuicides() {
	// Reset refund so that any used-gas calculations can use
	// this method
	s.refund = new(big.Int)
	for _, stateObject := range s.stateObjects {
		if stateObject.dirty {
			// If the object has been removed by a suicide
			// flag the object as deleted.
			if stateObject.remove {
				stateObject.deleted = true
			}
			stateObject.dirty = false
		}
	}
}

func (s *StateDB) Commit() (common.Hash, error) {
	batch := s.db.NewBatch()
	root, _ := s.commit(batch)
	return root, batch.Write()
}

// TODO the logic could be optimized
func (s *StateDB) commit(db trie.DatabaseWriter) (common.Hash, error) {
	s.refund = new(big.Int)

	for _, stateObject := range s.stateObjects {
		if stateObject.remove {
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.DeleteStateObject(stateObject)
			stateObject.dirty = false
			delete(stateObjectCache, stateObject.Address())
		} else {
			if len(stateObject.code) > 0 {
				if err := db.Put(stateObject.codeHash, stateObject.code); err != nil {
					return common.Hash{}, err
				}
			}
			// Write any storage changes in the state object to its trie.
			stateObject.Update()

			// Commit the trie of the object to the batch.
			// This updates the trie root internally, so
			// getting the root hash of the storage transactionie
			// through UpdateStateObject is fast.
			if _, err := stateObject.trie.CommitTo(db); err != nil {
				return common.Hash{}, err
			}
			// Update the object in the account trie.
			s.UpdateStateObject(stateObject)
			stateObject.dirty = false
			stateObjectCache[stateObject.Address()] = stateObject
		}
	}
	return s.trie.CommitTo(db)
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
	storage Storage

	code Code
	// The ABI for this account
	abi     ABI
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
func (self *StateObject) setError(err error) {
	if self.dbErr == nil {
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
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.Get(self.CodeHash())
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
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
func (self *StateObject) PrintStorages() {}

func (self *StateObject) CodeHash() []byte {
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
