package state

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
	"github.com/op/go-logging"
	"encoding/json"
)

var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("p2p")
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
	stateObjects     map[string]*StateObject
	refund           *big.Int
	thash, bhash     common.Hash
	txIndex          int
	logs             map[common.Hash]vm.Logs
	logSize          uint
	leastStateObject *StateObject
}

// Create a new state from a given trie
func New(db hyperdb.Database) (*StateDB, error) {
	return &StateDB{
		db:           db,
		stateObjects: make(map[string]*StateObject),
		refund:       new(big.Int),
		logs:         make(map[common.Hash]vm.Logs),
	}, nil

}

func GetStateObjects(db hyperdb.Database) (map[string]*StateObject ,error) {
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


	return stateDb,nil

}

// Reset clears out all emphemeral state objects from the state db, but keeps
// the underlying state trie to avoid reloading data for the next operations.
func (self *StateDB) Reset(root common.Hash) error {
	return nil
}

func (self *StateDB) StartRecord(thash, bhash common.Hash, ti int) {}

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

func (self *StateDB) Exist(addr common.Address) bool {
	return self.GetStateObject(addr) != nil
}

func (self *StateDB) GetAccount(addr common.Address) vm.Account {
	return self.GetOrNewStateObject(addr)
}

func (self *StateDB) GetLeastAccount() vm.Account {
	return self.leastStateObject
}

// todo
func (self *StateDB) SetLeastAccount(account *vm.Account) {
	//self.leastStateObject.abi = account.Address()
}
func (self *StateDB) GetAccounts() map[string]*StateObject {
	return self.stateObjects
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

func (self *StateDB) SetState(addr common.Address, key common.Hash, value common.Hash) {
	stateObject := self.GetOrNewStateObject(addr)
	if stateObject != nil {
		stateObject.SetState(key, value)
	}
}

func (self *StateDB) Delete(addr common.Address) bool {
	// TODO delete the StateObject from the
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
	// TODO update the StateDB
}

// Delete the given state object and delete it from the state trie
func (self *StateDB) DeleteStateObject(stateObject *StateObject) {
	stateObject.deleted = true
}

// Retrieve a state object given my the address. Nil if not found
func (self *StateDB) GetStateObject(addr common.Address) (stateObject *StateObject) {
	stateObject = self.stateObjects[addr.Hex()]
	if stateObject != nil {
		if stateObject.deleted {
			stateObject = nil
		}
	}
	return stateObject
}

func (self *StateDB) SetStateObject(object *StateObject) {
	self.stateObjects[object.Address().Hex()] = object
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
	stateObject := NewStateObject(addr)
	stateObject.SetNonce(StartingNonce)
	self.stateObjects[addr.Hex()] = stateObject

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
	return self
}

func (self *StateDB) Set(state *StateDB) {

}

func (self *StateDB) GetRefund() *big.Int {
	return self.refund
}

var stateObjectPrefix = []byte("stateObject-")
var stateDbPrefix = []byte("stateDb-")
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
