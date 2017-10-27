// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package state

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/ledger/tree/bucket"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/hyperchain/hyperchain/hyperdb/db"
	"github.com/op/go-logging"
	"github.com/pkg/errors"
	"math/big"
	"sort"
	"sync"
)

var (
	emptyCodeHash = crypto.Keccak256(nil)
	kec256Hash    = crypto.NewKeccak256Hash("keccak256")
)

type Code []byte

func (self Code) String() string {
	return string(self)
}

type Storage map[common.Hash][]byte

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

// StateObject represents an hyperchain account which is being modified.
//
// The usage pattern is as follows:
// First you need to obtain a state object.
// Account values can be accessed and modified through the object.
// Finally, call Commit to write the modified storage into a database.
type StateObject struct {
	address common.Address // Hyperchain address of this account
	data    Account
	db      *StateDB

	// DB error.
	// State objects are used by the consensus core and VM which are
	// unable to deal with database-level errors. Any error that occurs
	// during a database read is memoized here and will eventually be returned
	// by StateDB.Commit.
	dbErr          error
	code           Code                   // Contract bytecode, which gets set when code is loaded
	cachedStorage  Storage                // Storage entry cache to avoid duplicate reads
	dirtyStorage   Storage                // Storage entries that need to be flushed to disk
	archiveStorage Storage                // Storage entries that need to be archive
	evictList      map[common.Hash]uint64 // Record storage entry latest block number being modified

	tree       *bucket.BucketTree     // a bucket tree use to calculate fingerprint of storage efficiency
	treeConfig map[string]interface{} // bucket tree config
	// Cache flags.
	// When an object is marked suicided it will be delete from the db
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	deleted   bool
	onDirty   func(addr common.Address) // Callback method to mark a state object dirty
	logger    *logging.Logger
}

// empty returns whether the account is considered empty.
func (s *StateObject) empty() bool {
	return s.data.Nonce == 0 && s.data.Balance.BitLen() == 0 && bytes.Equal(s.data.CodeHash, emptyCodeHash)
}

// Account is the hyperchain consensus representation format.
type Account struct {
	Nonce             uint64         `json:"nonce"`
	Balance           *big.Int       `json:"balance"`
	Root              common.Hash    `json:"merkleRoot"`
	CodeHash          []byte         `json:"codeHash"`
	DeployedContracts []string       `json:"contracts"`
	Creator           common.Address `json:"creator"`
	Status            int            `json:"status"`
	CreateTime        uint64         `json:"createTime"`
}

// MemAccount used for state object marshal and unmarshal in journal.
type MemAccount struct {
	Address common.Address
	Account
}

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account, onDirty func(addr common.Address), setup bool, conf map[string]interface{}, logger *logging.Logger) *StateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	obj := &StateObject{
		db:             db,
		address:        address,
		data:           data,
		cachedStorage:  make(Storage),
		dirtyStorage:   make(Storage),
		archiveStorage: make(Storage),
		onDirty:        onDirty,
		logger:         logger,
		evictList:      make(map[common.Hash]uint64),
	}
	// Initialize bucket tree
	if setup {
		prefix := compositeStorageBucketPrefix(address.Hex())
		obj.tree = bucket.NewBucketTree(db.db, string(prefix))
		obj.tree.Initialize(conf)
		obj.treeConfig = conf
	}
	return obj
}

// marshal whole state object with address for journal usage
func (c *StateObject) MarshalJSON() ([]byte, error) {
	account := MemAccount{
		Address: c.address,
		Account: c.data,
	}
	return json.Marshal(account)
}

// unmarshal state object for journal usage
func (c *StateObject) UnmarshalJSON(data []byte) error {
	account := &MemAccount{}
	err := json.Unmarshal(data, account)
	c.data = account.Account
	c.address = account.Address
	return err
}

// marshal for state object persist
func (c *StateObject) Marshal() ([]byte, error) {
	account := &Account{
		Nonce:             c.data.Nonce,
		Balance:           c.data.Balance,
		Root:              c.data.Root,
		CodeHash:          c.data.CodeHash,
		DeployedContracts: c.data.DeployedContracts,
		Creator:           c.data.Creator,
		Status:            c.data.Status,
		CreateTime:        c.data.CreateTime,
	}
	return json.Marshal(account)
}

// Unmarshal unserialize bytes to account.
func Unmarshal(data []byte, v interface{}) error {
	account, ok := v.(*Account)
	if ok == false {
		return errors.New("invalid type")
	}
	err := json.Unmarshal(data, account)
	return err
}

func (c *StateObject) String() string {
	var str string
	str = fmt.Sprintf("stateObject: %x nonce [%d], balance [%d], root [%s], codeHash [%s]\n", c.address, c.Nonce(), c.Balance(), c.Root().Hex(), common.BytesToHash(c.CodeHash()).Hex())
	return str
}

// setError remembers the first non-nil error it is called with.
func (self *StateObject) setError(err error) {
	if self.dbErr == nil {
		self.dbErr = err
	}
}

// markSuicided marks the object has been deleted.
func (self *StateObject) markSuicided() {
	self.suicided = true
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
	self.logger.Infof("%x: #%d %v X\n", self.Address(), self.Nonce(), self.Balance())
}

// GetState from live state object return a storage entry's value if existed
func (self *StateObject) GetState(key common.Hash) (bool, []byte) {
	value, exists := self.cachedStorage[key]
	if exists {
		return true, value
	} else {
		return false, nil
	}
}

// GetState loads from database return a storage entry's value if existed
func GetStateFromDB(db db.Database, address common.Address, key common.Hash) (bool, []byte) {
	// Load from DB in case it is missing.
	val, err := db.Get(compositeStorageKey(address.Bytes(), key.Bytes()))
	if err != nil {
		return false, nil
	}
	return true, val
}

// SetState updates a value in account storage.
func (self *StateObject) SetState(db db.Database, key common.Hash, value []byte, opcode int32) {
	exist, previous := self.db.GetState(self.address, key)
	self.db.journal.JournalList = append(self.db.journal.JournalList, &StorageChange{
		Account:    &self.address,
		Key:        key,
		Prevalue:   previous,
		Exist:      exist,
		LastModify: self.evictList[key],
	})
	self.logger.Debugf("set state key %s value %s existed %v", key.Hex(), common.Bytes2Hex(value), exist)
	self.setState(key, value)
	if isArchive(opcode) {
		self.archive(key, previous, value)
	}
}

func (self *StateObject) setState(key common.Hash, value []byte) {
	// Write both cache
	self.logger.Debugf("put storage item key %s value %s to storage cache and dirty cache", key.Hex(), common.Bytes2Hex(value))
	self.cachedStorage[key] = value
	self.dirtyStorage[key] = value
	self.evictList[key] = self.db.curSeqNo
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

// remove storage entry from storage cache instead of setting empty value
// otherwise an entry will been leave in database
// only use in undo `add storage entry` vm operation
func (self *StateObject) removeState(key common.Hash) {
	delete(self.cachedStorage, key)
	delete(self.dirtyStorage, key)
	delete(self.evictList, key)
}

func (self *StateObject) Flush(db db.Batch, archieveDb db.Batch) error {
	var wg sync.WaitGroup
	wg.Add(2)
	go self.onEvict(&wg)
	go self.doArchive(&wg)

	// IMPORTANT root should calculate first
	// otherwise dirty storage will be removed in persist phase
	self.GenerateFingerPrintOfStorage()
	self.logger.Debugf("begin to flush dirty storage")
	for key, value := range self.dirtyStorage {
		delete(self.dirtyStorage, key)
		if len(value) == 0 {
			// delete
			self.logger.Debugf("flush dirty storage address [%s] delete item key: [%s]", self.address.Hex(), key.Hex())
			if err := db.Delete(compositeStorageKey(self.address.Bytes(), key.Bytes())); err != nil {
				return err
			}
		} else {
			self.logger.Debugf("flush dirty storage address [%s] put item key: [%s], value [%s]", self.address.Hex(), key.Hex(), common.Bytes2Hex(value))
			if err := db.Put(compositeStorageKey(self.address.Bytes(), key.Bytes()), value); err != nil {
				return err
			}
		}
	}
	// flush all bucket tree modified to batch
	self.tree.Commit(db)
	wg.Wait()
	self.archiveStorage = make(Storage)
	return nil
}

func (self *StateObject) GenerateFingerPrintOfStorage() common.Hash {
	var set ChangeSet
	if enableFakeHashFn {
		for k, v := range self.dirtyStorage {
			d := append(k.Bytes(), v...)
			set = append(set, d)
		}
		// if no change happen, it's no need to re-calculate root
		if len(set) != 0 {
			self.SetRoot(kec256Hash.Hash([]interface{}{
				self.Root(),
				set,
			}))
		}
		return self.Root()
	} else {
		// use bucket tree
		// 1. convert dirty storage entries to working set
		prev := self.Root()
		workingSet := bucket.NewEntries()
		for k, v := range self.dirtyStorage {
			self.logger.Debugf("calculate state object %s storage hash, dirty entry key %s value %s", self.address.Hex(), k.Hex(), common.Bytes2Hex(v))
			if len(v) == 0 {
				workingSet[k.Hex()] = nil
			} else {
				// Padding the value to at least 32 bytes.
				if len(v) <= common.HashLength {
					workingSet[k.Hex()] = common.BytesToHash(v).Bytes()
				} else {
					workingSet[k.Hex()] = v
				}
				self.logger.Debugf("working set key %s value %s", k.Hex(), common.Bytes2Hex(v))
			}
		}
		self.tree.Prepare(workingSet)
		// 2. calculate hash
		hash, err := self.tree.Process()
		if err != nil {
			self.logger.Errorf("calculate storage hash for stateObject [%s] failed", self.address.Hex())
			return common.Hash{}
		}
		// 3. assign to self.ROOT
		self.SetRoot(common.BytesToHash(hash))
		self.logger.Debugf("state object %s storage root hash %s after #%d", self.address.Hex(), self.Root().Hex(), self.db.curSeqNo)
		if (self.Root() != common.Hash{} || prev != common.Hash{}) {
			// storage hash changed
			self.db.journal.JournalList = append(self.db.journal.JournalList, &StorageHashChange{
				Account: &self.address,
				Prev:    prev.Bytes(),
			})
		}
		return self.Root()

	}
	return common.Hash{}
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *StateObject) ReturnGas(gas, price *big.Int) {}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (c *StateObject) Address() common.Address {
	return c.address
}

func (self *StateObject) Balance() *big.Int {
	return self.data.Balance
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *StateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Cmp(common.Big0) == 0 {
		return
	}
	c.SetBalance(new(big.Int).Add(c.Balance(), amount))

	c.logger.Infof("%x: #%d %v (+ %v)\n", c.Address(), c.Nonce(), c.Balance(), amount)
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (c *StateObject) SubBalance(amount *big.Int) {
	if amount.Cmp(common.Big0) == 0 {
		return
	}
	c.SetBalance(new(big.Int).Sub(c.Balance(), amount))

	c.logger.Infof("%x: #%d %v (- %v)\n", c.Address(), c.Nonce(), c.Balance(), amount)
}

func (self *StateObject) SetBalance(amount *big.Int) {
	self.db.journal.JournalList = append(self.db.journal.JournalList, &BalanceChange{
		Account: &self.address,
		Prev:    new(big.Int).Set(self.data.Balance),
	})
	self.setBalance(amount)
}

func (self *StateObject) setBalance(amount *big.Int) {
	self.data.Balance = amount
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

// Code returns the contract code associated with this object, if any.
func (self *StateObject) Code(db db.Database) []byte {
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.Get(compositeCodeHash(self.address.Bytes(), self.CodeHash()))
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
	}
	self.code = code
	return code
}

func (self *StateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := self.Code(self.db.db)
	self.db.journal.JournalList = append(self.db.journal.JournalList, &CodeChange{
		Account:  &self.address,
		Prevhash: self.CodeHash(),
		Prevcode: prevcode,
	})
	self.setCode(codeHash, code)
}

func (self *StateObject) setCode(codeHash common.Hash, code []byte) {
	self.code = code
	self.data.CodeHash = codeHash[:]
	self.dirtyCode = true
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

func (self *StateObject) Nonce() uint64 {
	return self.data.Nonce
}

func (self *StateObject) SetNonce(nonce uint64) {
	self.db.journal.JournalList = append(self.db.journal.JournalList, &NonceChange{
		Account: &self.address,
		Prev:    self.data.Nonce,
	})
	self.setNonce(nonce)
}

func (self *StateObject) setNonce(nonce uint64) {
	self.data.Nonce = nonce
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

func (self *StateObject) CodeHash() []byte {
	return self.data.CodeHash
}

func (self *StateObject) Root() common.Hash {
	return self.data.Root
}

func (self *StateObject) SetRoot(root common.Hash) {
	self.data.Root = root
}

func (self *StateObject) Creator() common.Address {
	return self.data.Creator
}

func (self *StateObject) SetCreator(addr common.Address) {
	self.db.journal.JournalList = append(self.db.journal.JournalList, &SetCreatorChange{
		Account: &self.address,
		Prev:    self.data.Creator,
	})
	self.setCreator(addr)
}

func (self *StateObject) setCreator(addr common.Address) {
	if bytes.Compare(self.data.Creator.Bytes(), addr.Bytes()) == 0 {
		self.logger.Warningf("state object %s set creator, same with the origin, ignore.", self.address.Hex())
		return
	}
	self.data.Creator = addr
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

func (self *StateObject) CreateTime() uint64 {
	return self.data.CreateTime
}

func (self *StateObject) SetCreateTime(time uint64) {
	self.db.journal.JournalList = append(self.db.journal.JournalList, &SetCreateTimeChange{
		Account: &self.address,
		Prev:    self.data.CreateTime,
	})
	self.setCreateTime(time)
}

func (self *StateObject) setCreateTime(time uint64) {
	if self.data.CreateTime == time {
		self.logger.Warningf("state object %s set create time, same with the origin, ignore.", self.address.Hex())
		return
	}
	self.data.CreateTime = time
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

// DeployedContracts - return a list of deployed contracts.
func (self *StateObject) DeployedContracts() []string {
	return self.data.DeployedContracts
}

// Status - return the status of state object.
func (self *StateObject) Status() int {
	return self.data.Status
}

// SetStatus - set the status of state object.
func (self *StateObject) SetStatus(status int) {
	if self.data.Status == status {
		self.logger.Noticef("state object %s set status, same with the origin status, ignore.", self.address.Hex())
		return
	}
	self.db.journal.JournalList = append(self.db.journal.JournalList, &StatusChange{
		Account: &self.address,
		Prev:    self.data.Status,
	})
	self.setStatus(status)
}

func (self *StateObject) setStatus(status int) {
	self.data.Status = status
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

// Never called, but must be present to allow StateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (self *StateObject) Value() *big.Int {
	panic("Value on StateObject should never be called")
}

func (self *StateObject) ForEachStorage(cb func(key common.Hash, value []byte) bool) map[common.Hash][]byte {
	// When iterating over the storage check the cache first
	ret := make(map[common.Hash][]byte)
	for h, value := range self.cachedStorage {
		cb(h, value)
		ret[h] = value
	}
	iter := self.db.db.NewIterator(getStorageKeyPrefix(self.address.Bytes()))
	defer iter.Release()
	for iter.Next() {
		k, ok := splitCompositeStorageKey(self.address.Bytes(), iter.Key())
		if ok == false {
			continue
		}
		key := common.BytesToHash(k)
		// ignore cached values
		if _, ok := self.cachedStorage[key]; !ok {
			cb(key, iter.Value())
			ret[key] = iter.Value()
		}
	}
	return ret
}

// appendDeployedContract - add a contract address to maintained list.
func (self *StateObject) AppendDeployedContract(address common.Address) {
	self.db.journal.JournalList = append(self.db.journal.JournalList, &DeployedContractChange{
		Account: &self.address,
		Prev:    &address,
	})
	self.appendDeployedContract(address)
}

func (self *StateObject) appendDeployedContract(address common.Address) {
	if idx := sort.SearchStrings(self.data.DeployedContracts, address.Hex()); idx < len(self.data.DeployedContracts) && self.data.DeployedContracts[idx] == address.Hex() {
		// already exist
		self.logger.Warningf("smart contract %s already in creator %s's list", address.Hex(), self.address.Hex())
		return
	} else {
		self.data.DeployedContracts = append(self.data.DeployedContracts, address.Hex())
		sort.Strings(self.data.DeployedContracts)
		if self.onDirty != nil {
			self.onDirty(self.Address())
			self.onDirty = nil
		}
	}
}

// removeDeployedContract - remove a contract address from maintained list.
func (self *StateObject) removeDeployedContract(address common.Address) bool {
	if idx := sort.SearchStrings(self.data.DeployedContracts, address.Hex()); idx < len(self.data.DeployedContracts) && self.data.DeployedContracts[idx] == address.Hex() {
		self.data.DeployedContracts = append(self.data.DeployedContracts[:idx], self.data.DeployedContracts[idx+1:]...)
		if len(self.data.DeployedContracts) == 0 {
			self.data.DeployedContracts = nil
		} else {
			sort.Strings(self.data.DeployedContracts)
		}
		if self.onDirty != nil {
			self.onDirty(self.Address())
			self.onDirty = nil
		}
		return true
	}
	return false
}

func (self *StateObject) archive(key common.Hash, originValue, value []byte) {
	// only record the initial value once.
	// Reason:
	//     for some small variables (e.g. bool, uint8), they are compacted to be stored in
	//     a single slot. If those varaibles try to archive themself continuously，it can lead only the latest
	//     slot value been record.
	// e.g.
	//     Initial slot value: [A, B, C]
	//     Final slot value:   [nil, nil, nil]
	//     if rewrite previous value each time, the recorded is [nil, nil, C], var A, B is dropped!
	// Additional:
	//     We store all initial key-value entries, skip judgement whether they are archived by a delete operation
	//     (via check len(value) == 0). This method can avoid data missing.
	if _, exist := self.archiveStorage[key]; !exist {
		self.archiveStorage[key] = originValue
	}
}

func isArchive(opcode int32) bool {
	return opcode == opcodeArchive
}

// onEvict drops `old enough` keyvalue pairs to avoid too much memory occupation.
func (self *StateObject) onEvict(wg *sync.WaitGroup) {
	defer func() {
		(*wg).Done()
	}()
	for key, brith := range self.evictList {
		if brith < self.db.oldestSeqNo {
			delete(self.evictList, key)
			delete(self.cachedStorage, key)
		}
	}
}

// doArchive flushs all archived entries to historical database.
func (self *StateObject) doArchive(wg *sync.WaitGroup) {
	defer func() {
		(*wg).Done()
	}()
	batch := self.db.archiveDb.NewBatch()
	for key, value := range self.archiveStorage {
		if len(value) != 0 {
			self.logger.Debugf("flush dirty storage address [%s] add key: [%s] to archieve db, value [%s]", self.address.Hex(), key.Hex(), common.BytesToHash(value).Hex())
			batch.Put(compositeStorageKey(self.address.Bytes(), key.Bytes()), value)
		}
	}
	batch.Write()
}
