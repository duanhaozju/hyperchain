package hyperstate

import (
	"bytes"
	"fmt"
	"math/big"
	"hyperchain/crypto"
	"hyperchain/common"
	"encoding/json"
	"hyperchain/hyperdb"
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

	root common.Hash      // storage fingerprint
	code Code             // contract bytecode, which gets set when code is loaded

	cachedStorage Storage // Storage entry cache to avoid duplicate reads
	dirtyStorage  Storage // Storage entries that need to be flushed to disk

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

// newObject creates a state object.
func newObject(db *StateDB, address common.Address, data Account, onDirty func(addr common.Address)) *StateObject {
	if data.Balance == nil {
		data.Balance = new(big.Int)
	}
	if data.CodeHash == nil {
		data.CodeHash = emptyCodeHash
	}
	return &StateObject{db: db, address: address, data: data, cachedStorage: make(Storage), dirtyStorage: make(Storage), onDirty: onDirty}
}


// Marshal stateObject
func (c *StateObject) Marshal() ([]byte, error) {
	account := Account{
		Nonce:    c.data.Nonce,
		Balance:  c.data.Balance,
		Root:     c.data.Root,
		CodeHash: c.data.CodeHash,
	}
	return json.Marshal(account)
}
// Unmarshal stateObject
func Unmarshal(data []byte) (Account, error) {
	var account Account
	err := json.Unmarshal(data, &account)
	return account, err
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
	c.db.journal = append(c.db.journal, touchChange{
		account: &c.address,
		prev:    c.touched,
	})
	if c.onDirty != nil {
		c.onDirty(c.Address())
		c.onDirty = nil
	}
	c.touched = true
}

// GetState returns a value in account storage.
func (self *StateObject) GetState(db hyperdb.Database, key common.Hash) common.Hash {
	value, exists := self.cachedStorage[key]
	if exists {
		return value
	}
	// Load from DB in case it is missing.
	val, _ := db.Get(CompositeStorageKey(self.address.Bytes(), key.Bytes()))
	value.SetBytes(val)

	if (value != common.Hash{}) {
		self.cachedStorage[key] = value
	}
	return value
}

// SetState updates a value in account storage.
func (self *StateObject) SetState(db hyperdb.Database, key, value common.Hash) {
	self.db.journal = append(self.db.journal, storageChange{
		account:  &self.address,
		key:      key,
		prevalue: self.GetState(db, key),
	})
	self.setState(key, value)
}

func (self *StateObject) setState(key, value common.Hash) {
	// Write both cache
	self.cachedStorage[key] = value
	self.dirtyStorage[key] = value

	if self.onDirty != nil {
		self.onDirty(self.Address())
		self.onDirty = nil
	}
}

func (self *StateObject) Flush(db hyperdb.Batch) error {
	for key, value := range self.dirtyStorage {
		delete(self.dirtyStorage, key)
		if (value == common.Hash{}) {
			// delete
			if err := db.Put(CompositeStorageKey(self.address.Bytes(), key.Bytes()), nil); err != nil {
				return err
			}
		} else {
			if err := db.Put(CompositeStorageKey(self.address.Bytes(), key.Bytes()), value.Bytes()); err != nil {
				return err
			}
		}
	}
	return nil
}

func (self *StateObject) GenerateFingerPrintOfStorage() common.Hash {
	var set ChangeSet
	if enableFakeHashFn {
		for k, v := range self.dirtyStorage {
			d := append(k.Bytes(), v.Bytes()...)
			set = append(set, d)
		}
		self.root = kec256Hash.Hash([]interface{}{
			self.root,
			set,
		})
		return self.root
	}
	return common.Hash{}
}

// AddBalance removes amount from c's balance.
// It is used to add funds to the destination account of a transfer.
func (c *StateObject) AddBalance(amount *big.Int) {
	// EIP158: We must check emptiness for the objects such that the account
	// clearing (0,0,0 objects) can take effect.
	if amount.Cmp(common.Big0) == 0 {
		if c.empty() {
			c.touch()
		}
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
	self.db.journal = append(self.db.journal, balanceChange{
		account: &self.address,
		prev:    new(big.Int).Set(self.data.Balance),
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
	stateObject := newObject(db, self.address, self.data, onDirty)
	stateObject.code = self.code
	stateObject.dirtyStorage = self.dirtyStorage.Copy()
	stateObject.cachedStorage = self.dirtyStorage.Copy()
	stateObject.suicided = self.suicided
	stateObject.dirtyCode = self.dirtyCode
	stateObject.deleted = self.deleted
	stateObject.root = self.root
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
	code, err := db.Get(self.CodeHash())
	if err != nil {
		self.setError(fmt.Errorf("can't load code hash %x: %v", self.CodeHash(), err))
	}
	self.code = code
	return code
}

func (self *StateObject) SetCode(codeHash common.Hash, code []byte) {
	prevcode := self.Code(self.db.db)
	self.db.journal = append(self.db.journal, codeChange{
		account:  &self.address,
		prevhash: self.CodeHash(),
		prevcode: prevcode,
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
	self.db.journal = append(self.db.journal, nonceChange{
		account: &self.address,
		prev:    self.data.Nonce,
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
	iter := leveldb.NewIterator()
	for ok := iter.Seek(GetStorageKeyPrefix(self.address.Bytes())); ok; ok = iter.Next() {
		key := common.BytesToHash(SplitCompositeStorageKey(self.address.Bytes(), iter.Key()))
		// ignore cached values
		if _, ok := self.cachedStorage[key]; !ok {
			cb(key, common.BytesToHash(iter.Value()))
			ret[key] = common.BytesToHash(iter.Value())
		}
	}
	return ret
}
