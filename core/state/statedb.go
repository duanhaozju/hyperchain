//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"fmt"
	lru "github.com/hashicorp/golang-lru"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/hyperdb/db"
	"hyperchain/tree/pmt"
	"math/big"
	"sync"
)

var (
	log              *logging.Logger                 // package-level logger
	codeCache        *lru.Cache                      // TODO is it could be faster?  it can be put outside of the StateDB
	abiCache         *lru.Cache                      // TODO is it could be faster?  it can be put outside of the StateDB
	stateObjectCache map[common.Address]*StateObject // TODO is it should be used as the lru.Cache
)

const (
	// Number of codehash->size associations to keep
	codeSizeCacheSize = 10000
	abiSizeCacheSize  = 10000
)

func init() {
	log = logging.MustGetLogger("state")
	// 1.init the codeSizeCache
	csc, err := lru.New(codeSizeCacheSize)
	if err != nil {
		codeCache = csc
	} else {
	}
	// 2.init the abiSizeCache
	asc, err := lru.New(abiSizeCacheSize)
	if err != nil {
		abiCache = asc
	} else {
	}
	// 2.init the stateObjectSizeCache
	// stateObjectCache = make(map[common.Address] *StateObject)
}

// The starting nonce determines the default nonce when new accounts are being
// created.
var StartingNonce uint64

// StateDBs within the hyperchain protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db   db.Database
	trie *pmt.SecureTrie

	//this map holds 'live' objects, which will get modified while processing a state transition
	stateObjects     map[string]*StateObject
	stateObjectDirty map[common.Address]struct{}

	refund           *big.Int
	thash, bhash     common.Hash
	txIndex          int
	logs             map[common.Hash]types.Logs
	logSize          uint
	leastStateObject *StateObject
	lock             sync.Mutex
}

// Create a new state from a given trie
func New(root common.Hash, db db.Database) (*StateDB, error) {
	tr, err := pmt.NewSecure(root, db)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:               db,
		trie:             tr,
		stateObjects:     make(map[string]*StateObject),
		stateObjectDirty: make(map[common.Address]struct{}),
		refund:           new(big.Int),
		logs:             make(map[common.Hash]types.Logs),
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
		logs:             make(map[common.Hash]types.Logs),
	}, nil
}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) ResetTo(root common.Hash) error {
	var (
		err error
		tr  = self.trie
	)
	if self.trie.Hash() != root {
		if tr, err = pmt.NewSecure(root, self.db); err != nil {
			return err
		}
	}
	self.trie = tr
	self.stateObjects = make(map[string]*StateObject)
	self.stateObjectDirty = make(map[common.Address]struct{})
	self.thash = common.Hash{}
	self.bhash = common.Hash{}
	self.logs = make(map[common.Hash]types.Logs)
	self.logSize = 0
	self.refund = new(big.Int)
	// if reset we will clear all stateObjectSizeCache
	// stateObjectCache =  make(map[common.Address] *StateObject)
	return nil
}

func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

// doesn't assign block hash now
// because the blcok hash hasn't been calculated
// correctly block  and block hash will be assigned in the commit phase
func (self *StateDB) AddLog(log *types.Log) {
	log.TxHash = self.thash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) types.Logs {
	return self.logs[hash]
}

func (self *StateDB) Logs() types.Logs {
	var logs types.Logs
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

func (self *StateDB) GetAccount(addr common.Address) *StateObject {
	return self.GetStateObject(addr)
}

func (self *StateDB) GetLeastAccount() *StateObject {
	return self.leastStateObject
}

// todo
func (self *StateDB) SetLeastAccount(account *vm.Account) {
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

func (self *StateDB) ForEachAccounts() {
}

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

func (self *StateDB) GetState(a common.Address, b common.Hash) (bool, common.Hash) {
	stateObject := self.GetStateObject(a)
	if stateObject != nil {
		return stateObject.GetState(b)
	}
	return false, common.Hash{}
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
		stateObject.SetCode(common.Hash{}, code)
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

	// TODO fix in the next version
	// 2.we will find the stateObject from stateObjectSizeCache
	//stateObject = stateObjectCache[addr]
	//if stateObject != nil{
	//	if stateObject.deleted{
	//		stateObject = nil
	//	}else {
	//		self.SetStateObject(stateObject)
	//		return stateObject
	//	}
	//}

	// 3.we will find the stateObject from the db by trie
	data := self.trie.Get(addr[:])
	if len(data) == 0 {
		return nil
	}
	stateObject, err := DecodeObject(addr, self.db, data)
	if err != nil {
		return nil
	}
	// stateObjectCache[addr] = stateObject
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
	// stateObjectCache[addr] = stateObject
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

func (self *StateDB) CreateAccount(addr common.Address) *StateObject {
	return self.CreateStateObject(addr)
}

//
// Setting, copying of the state methods
//

// TODO it could be better
// 1.from the stateObjectCacheSize
func (self *StateDB) Copy() *StateDB {
	state, _ := New(common.Hash{}, self.db)
	state.trie = self.trie
	for k, stateObject := range self.stateObjects {
		state.stateObjects[k] = stateObject.Copy()
	}
	state.refund.Set(self.refund)

	for hash, logs := range self.logs {
		state.logs[hash] = make(types.Logs, len(logs))
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
func (s *StateDB) commit(db pmt.DatabaseWriter) (common.Hash, error) {
	s.refund = new(big.Int)

	for _, stateObject := range s.stateObjects {
		if stateObject.remove {
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.DeleteStateObject(stateObject)
			stateObject.dirty = false
			//delete(stateObjectCache,stateObject.Address())
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
			// stateObjectCache[stateObject.Address()] = stateObject
		}
	}
	return s.trie.CommitTo(db)
}

func (self *StateDB) Snapshot() *StateDB {
	return self.Copy()
}

func (self *StateDB) RevertToSnapshot(statedb *StateDB) {
	self.Set(statedb)
}
