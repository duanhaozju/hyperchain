package hyperstate

import (
	"bytes"
	"fmt"
	"math/big"
	"hyperchain/crypto"
	"hyperchain/common"
	"encoding/json"
	"hyperchain/hyperdb"
	"github.com/pkg/errors"
	"hyperchain/tree/bucket"
)

var emptyCodeHash = crypto.Keccak256(nil)

type Code []byte

func (self Code) String() string {
	return string(self) //strings.Join(Disassemble(self), " ")
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
	dbErr error

	code Code             // contract bytecode, which gets set when code is loaded
	cachedStorage Storage // Storage entry cache to avoid duplicate reads
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

	bucketTree    *bucket.BucketTree      // a bucket tree use to calculate fingerprint of storage efficiency
	bucketConf    map[string]interface{}  // bucket tree config
	// Cache flags.
	// When an object is marked suicided it will be delete from the db
	// during the "update" phase of the state transition.
	dirtyCode bool // true if the code was updated
	suicided  bool
	touched   bool
	deleted   bool
	onDirty   func(addr common.Address) // Callback method to mark a state object newly dirty
}

// empty returns whether the account is considered empty.
func (s *StateObject) empty() bool {
	return s.data.Nonce == 0 && s.data.Balance.BitLen() == 0 && bytes.Equal(s.data.CodeHash, emptyCodeHash)
}

// Account is the hyperchain consensus representation of accounts.
type Account struct {
	Nonce    uint64
	Balance  *big.Int
	Root     common.Hash // merkle root of the storage
	CodeHash []byte
}

// MemAccount use for state object marshal and unmarshal in journal
type MemAccount struct {
	Address common.Address
	Account
}

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account, onDirty func(addr common.Address), setup bool, bktConf map[string]interface{}) *StateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	obj := &StateObject{db: db, address: address, data: data, cachedStorage: make(Storage), dirtyStorage: make(Storage), onDirty: onDirty}
	// initialize bucket tree
	if setup {
		prefix, _ := CompositeStorageBucketPrefix(address.Bytes())
		obj.bucketTree = bucket.NewBucketTree(string(prefix))
		obj.bucketTree.Initialize(bktConf)
		obj.bucketConf = bktConf
	}
	return obj
}


// marshal whole state object with address for journal usage
func (c *StateObject) MarshalJSON() ([]byte, error) {
	account := MemAccount{
		Address:  c.address,
		Account:  c.data,
	}
	return json.Marshal(account)
}
// marshal for state object persist
func (c *StateObject) Marshal() ([]byte, error) {
	log.Criticalf("[DEBUG] nonce %d, balance %d, root %s, codehash %s", c.data.Nonce, c.data.Balance, c.data.Root.Hex(), common.Bytes2Hex(c.data.CodeHash))
	account := &Account{
		Nonce:    c.data.Nonce,
		Balance:  c.data.Balance,
		Root:     c.data.Root,
		CodeHash: c.data.CodeHash,
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
// Unmarshal stateObject in disk
func Unmarshal(data []byte, v interface{}) error {
	account, ok := v.(*Account)
	if ok == false {
		return errors.New("invalid type")
	}
	err := json.Unmarshal(data, account)
	return err
}

// String
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

func (self *StateObject) markSuicided() {
	self.suicided = true
	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
	log.Infof("%x: #%d %v X\n", self.Address(), self.Nonce(), self.Balance())
}

func (c *StateObject) touch() {
	c.db.journal.JournalList = append(c.db.journal.JournalList, &TouchChange{
		Account: &c.address,
		Prev:    c.touched,
	})
	if c.onDirty != nil {
		c.onDirty(c.Address())
		c.onDirty = nil
	}
	c.touched = true
}

// GetState from live state object return a storage entry's value if existed
func (self *StateObject) GetState(key common.Hash) common.Hash {
	value, exists := self.cachedStorage[key]
	if exists {
		return value
	} else {
		return common.Hash{}
	}
}
// GetState from database return a storage entry's value if existed
func GetStateFromDB(db hyperdb.Database,address common.Address, key common.Hash) common.Hash {
	var value common.Hash
	// Load from DB in case it is missing.
	val, _ := db.Get(CompositeStorageKey(address.Bytes(), key.Bytes()))
	value.SetBytes(val)
	return value
}

// SetState updates a value in account storage.
func (self *StateObject) SetState(db hyperdb.Database, key, value common.Hash) {
	self.db.journal.JournalList = append(self.db.journal.JournalList, &StorageChange{
		Account:  &self.address,
		Key:      key,
		Prevalue: self.db.GetState(self.address, key),
	})
	self.setState(key, value)
}

func (self *StateObject) setState(key, value common.Hash) {
	// Write both cache
	log.Errorf("put storage item key %s value %s to storage cache and dirty cache", key.Hex(), value.Hex())
	self.cachedStorage[key] = value
	self.dirtyStorage[key] = value

	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

func (self *StateObject) Flush(db hyperdb.Batch) error {
	// IMPORTANT root should calculate first
	// otherwise dirty storage will be removed in persist phase
	self.GenerateFingerPrintOfStorage()
	for key, value := range self.dirtyStorage {
		delete(self.dirtyStorage, key)
		if (value == common.Hash{}) {
			// delete
			log.Errorf("flush dirty storage address [%s] delete item key: [%s]", self.address.Hex(), key.Hex())
			if err := db.Delete(CompositeStorageKey(self.address.Bytes(), key.Bytes())); err != nil {
				return err
			}
		} else {
			log.Errorf("flush dirty storage address [%s] put item key: [%s], value [%s]", self.address.Hex(), key.Hex(), value.Hex())
			if err := db.Put(CompositeStorageKey(self.address.Bytes(), key.Bytes()), value.Bytes()); err != nil {
				return err
			}
		}
	}
	// flush all bucket tree modified to batch
	self.bucketTree.AddChangesForPersistence(db, big.NewInt(int64(self.db.curSeqNo)))
	return nil
}

func (self *StateObject) GenerateFingerPrintOfStorage() common.Hash {
	var set ChangeSet
	if enableFakeHashFn {
		for k, v := range self.dirtyStorage {
			d := append(k.Bytes(), v.Bytes()...)
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
		workingSet := bucket.NewKVMap()
		for k, v := range self.dirtyStorage {
			log.Errorf("calculate state object %s storage hash, dirty entry key %s value %s", self.address.Hex(), k.Hex(), v.Hex())
			if (v == common.Hash{}) {
				workingSet[k.Hex()] = nil
			} else {
				workingSet[k.Hex()] = v.Bytes()
			}
		}
		log.Errorf("state object %s storage root hash %s before #%d", self.address.Hex(), prev.Hex(), self.db.curSeqNo)
		self.bucketTree.PrepareWorkingSet(workingSet, big.NewInt(int64(self.db.curSeqNo)))
		// 2. calculate hash
		hash, err := self.bucketTree.ComputeCryptoHash()
		if err != nil {
			log.Errorf("calculate storage hash for stateObject [%s] failed", self.address.Hex())
			return common.Hash{}
		}
		// 3. assign to self.ROOT
		self.SetRoot(common.BytesToHash(hash))
		log.Errorf("state object %s storage root hash %s after #%d", self.address.Hex(), self.Root().Hex(), self.db.curSeqNo)
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

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *StateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Cmp(common.Big0) == 0 {
		return
	}
	c.SetBalance(new(big.Int).Add(c.Balance(), amount))

	log.Infof("%x: #%d %v (+ %v)\n", c.Address(), c.Nonce(), c.Balance(), amount)
}

// SubBalance removes amount from c's balance.
// It is used to remove funds from the origin account of a transfer.
func (c *StateObject) SubBalance(amount *big.Int) {
	if amount.Cmp(common.Big0) == 0 {
		return
	}
	c.SetBalance(new(big.Int).Sub(c.Balance(), amount))

	log.Infof("%x: #%d %v (- %v)\n", c.Address(), c.Nonce(), c.Balance(), amount)
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

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *StateObject) ReturnGas(gas, price *big.Int) {}

func (self *StateObject) deepCopy(db *StateDB, onDirty func(addr common.Address)) *StateObject {
	stateObject := newObject(db, self.address, self.data, onDirty, true, self.bucketConf)
	stateObject.code = self.code
	stateObject.dirtyStorage = self.dirtyStorage.Copy()
	stateObject.cachedStorage = self.dirtyStorage.Copy()
	stateObject.suicided = self.suicided
	stateObject.dirtyCode = self.dirtyCode
	stateObject.deleted = self.deleted
	return stateObject
}

//
// Attribute accessors
//

// Returns the address of the contract/account
func (c *StateObject) Address() common.Address {
	return c.address
}

// Code returns the contract code associated with this object, if any.
func (self *StateObject) Code(db hyperdb.Database) []byte {
	if self.code != nil {
		return self.code
	}
	if bytes.Equal(self.CodeHash(), emptyCodeHash) {
		return nil
	}
	code, err := db.Get(CompositeCodeHash(self.address.Bytes(), self.CodeHash()))
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

func (self *StateObject) Balance() *big.Int {
	return self.data.Balance
}

func (self *StateObject) Nonce() uint64 {
	return self.data.Nonce
}

func (self *StateObject) Root() common.Hash {
	return self.data.Root
}

func (self *StateObject) SetRoot(root common.Hash) {
	self.data.Root = root
}
// Never called, but must be present to allow StateObject to be used
// as a vm.Account interface that also satisfies the vm.ContractRef
// interface. Interfaces are awesome.
func (self *StateObject) Value() *big.Int {
	panic("Value on StateObject should never be called")
}

func (self *StateObject) ForEachStorage(cb func(key, value common.Hash) bool) map[common.Hash]common.Hash {
	// When iterating over the storage check the cache first
	var ret map[common.Hash]common.Hash
	for h, value := range self.cachedStorage {
		cb(h, value)
		ret[h] = value
	}
	leveldb, ok := self.db.db.(*hyperdb.LDBDatabase)
	if ok == false {
		return ret
	}
	iter := leveldb.NewIteratorWithPrefix(GetStorageKeyPrefix(self.address.Bytes()))
	for iter.Next() {
		k, ok := SplitCompositeStorageKey(self.address.Bytes(), iter.Key())
		if ok == false {
			continue
		}
		key := common.BytesToHash(k)
		// ignore cached values
		if _, ok := self.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(iter.Value()))
			ret[key] = common.BytesToHash(iter.Value())
		}
	}
	return ret
}
