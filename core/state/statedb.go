package state

import (
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"hyperchain/trie"
	"math/big"
	"time"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("state")
}

// The starting nonce determines the default nonce when new accounts are being
// created.
var StartingNonce uint64

// StateDBs within the ethereum protocol are used to store anything
// within the merkle trie. StateDBs take care of caching and storing
// nested states. It's the general query interface to retrieve:
// * Contracts
// * Accounts
type StateDB struct {
	db               hyperdb.Database
	trie             *trie.SecureTrie
	stateObjects     map[string]*StateObject
	refund           *big.Int
	thash, bhash     common.Hash
	txIndex          int
	logs             map[common.Hash]vm.Logs
	logSize          uint
	leastStateObject *StateObject
}

// Create a new state from a given trie
func New(root common.Hash, db hyperdb.Database) (*StateDB, error) {
	tr, err := trie.NewSecure(root, db)
	if err != nil {
		return nil, err
	}
	return &StateDB{
		db:           db,
		trie:         tr,
		stateObjects: make(map[string]*StateObject),
		refund:       new(big.Int),
		logs:         make(map[common.Hash]vm.Logs),
	}, nil

}

/*
func GetStateObjects(db hyperdb.Database) (map[string]*StateObject, error) {
	//var stateDb StateDB
	var stateDb map[string]*StateObject
	var b = make(map[string]*StateObject)
	data, err := db.Get(stateDbPrefix)

	if err != nil {

		return b, err
	}
	if err = json.Unmarshal(data, &stateDb); err != nil {

		return b, err
	}

	return stateDb, nil

}
*/
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
	*self = StateDB{
		db:           self.db,
		trie:         tr,
		stateObjects: make(map[string]*StateObject),
		refund:       new(big.Int),
		logs:         make(map[common.Hash]vm.Logs),
	}
	return nil
}

func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {
	self.thash = thash
	self.bhash = bhash
	self.txIndex = ti
}

func (self *StateDB) AddLog(log *vm.Log) {
	log.TxHash = self.thash
	log.BlockHash = self.bhash
	log.TxIndex = uint(self.txIndex)
	log.Index = self.logSize
	self.logs[self.thash] = append(self.logs[self.thash], log)
	self.logSize++
}

func (self *StateDB) GetLogs(hash common.Hash) vm.Logs {
	return self.logs[hash]
}

func (self *StateDB) Logs() vm.Logs {
	var logs vm.Logs
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

func (self *StateDB) GetAccount(addr common.Address) vm.Account {
	return self.GetStateObject(addr)
}

func (self *StateDB) GetLeastAccount() vm.Account {
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
	var n = 0
	for _, v := range self.GetAccounts() {
		log.Info("+++++++++++++++++++++the ", n, " account+++++++++++++++++++++++")
		log.Info("Account key:", common.ToHex(v.address.Bytes()), "----------value:", v)
		n = n + 1
	}

}

// Retrieve the BalanceData from the given address or 0 if object not found
func (self *StateDB) GetBalance(addr common.Address) *big.Int {
	stateObject := self.GetStateObject(addr)
	if stateObject != nil {
		return stateObject.BalanceData
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
	stateObject = self.stateObjects[addr.Str()]
	if stateObject != nil {
		if stateObject.deleted {
			stateObject = nil
		}
		return stateObject
	}

	data := self.trie.Get(addr[:])
	if len(data) == 0 {
		return nil
	}
	stateObject, err := DecodeObject(addr, self.db, data)
	if err != nil {
		return nil
	}
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

func (self *StateDB) CreateAccount(addr common.Address) vm.Account {
	return self.CreateStateObject(addr)
}

//
// Setting, copying of the state methods
//

func (self *StateDB) Copy() *StateDB {
	state, _ := New(common.Hash{}, self.db)
	state.trie = self.trie
	for k, stateObject := range self.stateObjects {
		state.stateObjects[k] = stateObject.Copy()
	}
	state.refund.Set(self.refund)

	for hash, logs := range self.logs {
		state.logs[hash] = make(vm.Logs, len(logs))
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

//var stateObjectPrefix = []byte("stateObject-")
//var stateDbPrefix = []byte("stateDb-")

/*
// Commit commits all state changes to the database.
// we should not use statedb.db because it is useless
// TODO test
func (s *StateDB) Commit() (root common.Hash, err error) {
	//batch := s.db.NewBatch()
	db, _ := hyperdb.GetLDBDatabase()
	bacth := db.NewBatch()
	stateDbData, err := json.Marshal(s.stateObjects)
	for addr, stateObject := range s.stateObjects {
		if stateObject.dirty {
			data, err := json.Marshal(stateObject)
			if err != nil {
				// err
			}
			keyFact := append(stateObjectPrefix, []byte(addr)...)
			log.Notice("save the data")
			bacth.Put(keyFact, data)
			//log.Notice("save already")
		}
	}
	db.Put(stateDbPrefix, stateDbData)

	return common.Hash{}, bacth.Write()
}
*/

func (s *StateDB) Commit() (common.Hash, error) {
	batch := s.db.NewBatch()
	root, _ := s.commit(batch)
	return root, batch.Write()
}

func (s *StateDB) commit(db trie.DatabaseWriter) (common.Hash, error) {
	s.refund = new(big.Int)
	for _, stateObject := range s.stateObjects {
		if stateObject.remove {
			// If the object has been removed, don't bother syncing it
			// and just mark it for deletion in the trie.
			s.DeleteStateObject(stateObject)
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
		}
		stateObject.dirty = false
	}
	return s.trie.CommitTo(db)
}
