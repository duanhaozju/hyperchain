package state

import (
	"fmt"
	"math/big"
	"hyperchain/common"
	"hyperchain/core/crypto"
)

var emptyCodeHash = crypto.Keccak256(nil)

type Code []byte
type ABI []byte

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

type StateObject struct {
			   // Address belonging to this account
	address common.Address
			   // The balance of the account
	balance *big.Int
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

func NewStateObject(address common.Address) *StateObject {
	object := &StateObject{
		address:  address,
		balance:  new(big.Int),
		dirty:    true,
		codeHash: emptyCodeHash,
		storage:  make(Storage),
	}
	return object
}

func (self *StateObject) MarkForDeletion() {
	self.remove = true
	self.dirty = true
}


func (self *StateObject) Storage() Storage {
	return self.storage
}

func (self *StateObject) GetState(key common.Hash) common.Hash {
	value, _ := self.storage[key]
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
			continue
		}
		self.SetState(key, value)
	}
}

func (c *StateObject) AddBalance(amount *big.Int) {
	c.SetBalance(new(big.Int).Add(c.balance, amount))
}

func (c *StateObject) SubBalance(amount *big.Int) {
	c.SetBalance(new(big.Int).Sub(c.balance, amount))
}

func (c *StateObject) SetBalance(amount *big.Int) {
	c.balance = amount
	c.dirty = true
}

// Return the gas back to the origin. Used by the Virtual machine or Closures
func (c *StateObject) ReturnGas(gas, price *big.Int) {}

func (self *StateObject) Copy() *StateObject {
	stateObject := NewStateObject(self.Address())
	stateObject.balance.Set(self.balance)
	stateObject.codeHash = common.CopyBytes(self.codeHash)
	stateObject.nonce = self.nonce
	stateObject.code = common.CopyBytes(self.code)
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
	return self.balance
}

// Returns the address of the contract/account
func (c *StateObject) Address() common.Address {
	return c.address
}


func (self *StateObject) Code() []byte {
	return self.code
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

}
