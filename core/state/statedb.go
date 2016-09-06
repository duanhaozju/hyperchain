package state

import (
	"math/big"

	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"github.com/op/go-logging"
)

// The starting nonce determines the default nonce when new accounts are being
// created.
var StartingNonce uint64
var logger = *logging.Logger{}
// StateDBs are used to store anything within the merkle trie.
// StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db	hyperdb.Database
	stateObjects map[string]*StateObject
	// the information of current transaction
	txIndex      int
	thash, bhash common.Hash
	refund *big.Int
}

// Create a new state from a given trie
// TODO we don't add any trie now
func New(root common.Hash,db *hyperdb.Database) (*StateDB, error) {
	return &StateDB{
		db:	db,
		stateObjects: make(map[string]*StateObject),
		refund:       new(big.Int),
	}, nil
}

// Reset clears out all state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
// TODO
func (self *StateDB) Reset(root common.Hash) error {
	return nil
}

// start to record the txs
func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (self *StateDB) AddRefund(gas *big.Int) {
	self.refund.Add(self.refund, gas)
}


func (self *StateDB) HasAccount(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

func (self *StateDB) Exist(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

func (self *StateDB) GetAccount(addr common.Address) vm.Account {
	return self.GetStateObject(addr)
}

// Retrieve the balance from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.balance
	}

	return common.Big0
}

func (self *StateDB) GetNonce(addr common.Address) uint64 {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.nonce
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
		stateObject.balance = new(big.Int)

		return true
	}

	return false
}

//
// Setting, updating & deleting state object methods
//

// Update the given state object and apply it to state trie
// TODO 需要一棵树来存
func (self *StateDB) UpdateStateObject(stateObject *StateObject) {
}

// Delete the given state object and delete it from the state trie
// TODO
func (self *StateDB) DeleteStateObject(stateObject *StateObject) {
}

// Retrieve a state object given my the address. Nil if not found
// TODO
func (self *StateDB) GetStateObject(addr common.Address) (stateObject *StateObject) {
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

// NewStateObject create a state object whether it exist in the trie or not
func (self *StateDB) newStateObject(addr common.Address) *StateObject {
	return nil
}

// Creates creates a new state object and takes ownership. This is different from "NewStateObject"
func (self *StateDB) CreateStateObject(addr common.Address) *StateObject {
	// Get previous (if any)
	so := self.GetStateObject(addr)
	// Create a new one
	newSo := self.newStateObject(addr)

	// If it existed set the balance to the new account
	if so != nil {
		newSo.balance = so.balance
	}

	return newSo
}

func (self *StateDB) CreateAccount(addr common.Address) vm.Account {
	return self.CreateStateObject(addr)
}

//
// Setting, copying of the state methods
//

func (self *StateDB) Copy() *StateDB {
	return nil
}

func (self *StateDB) Set(state *StateDB) {
	self.stateObjects = state.stateObjects

	self.refund = state.refund
}

func (self *StateDB) GetRefund() *big.Int {
	return self.refund
}

// IntermediateRoot computes the current root hash of the state trie.
// It is called in between transactions to get the root hash that
// goes into transaction receipts.
// IntermediateRoot会计算当前状态树的root hash,它将会在交易在得到root hash的
// 这段时间内被调用(正在进入到交易receipt状态)
func (s *StateDB) IntermediateRoot() common.Hash {
	return nil
}

// DeleteSuicides flags the suicided objects for deletion so that it
// won't be referenced again when called / queried up on.
//
// DeleteSuicides should not be used for consensus related updates
// under any circumstances.
func (s *StateDB) DeleteSuicides() {
	// Reset refund so that any used-gas calculations can use
	// this method.
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

// Commit commits all state changes to the database.
func (s *StateDB) Commit() (root common.Hash, err error) {
	return root, nil
}

// CommitBatch commits all state changes to a write batch but does not
// execute the batch. It is used to validate state changes against
// the root hash stored in a block.
func (s *StateDB) CommitBatch() (root common.Hash, batch hyperdb.Batch) {
	return root, batch
}

func (s *StateDB) commit(db *hyperdb.LDBDatabase) (common.Hash, error) {
	s.refund = new(big.Int)
	return nil,nil
}

func (self *StateDB) Refunds() *big.Int {
	return self.refund
}
