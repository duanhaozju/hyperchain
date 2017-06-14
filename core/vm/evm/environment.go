//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"math/big"

	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"hyperchain/core/types"
)

// RuleSet is an interface that defines the current rule set during the
// execution of the EVM instructions (e.g. whether it's homestead)
type RuleSet interface {
	IsHomestead(*big.Int) bool
}

// Environment is an EVM requirement and helper which allows access to outside
// information such as states.
type Environment interface {
	// The current ruleset
	RuleSet() RuleSet
	// The state database
	Db() Database
	// Creates a restorable snapshot
	MakeSnapshot() interface{}
	// Set database to previous snapshot
	SetSnapshot(interface{})
	// Address of the original invoker (first occurrence of the VM invoker)
	Origin() common.Address
	// The block number this VM is invoked on
	BlockNumber() *big.Int
	// The n'th hash ago from this block number
	GetHash(uint64) common.Hash
	// The handler's address
	Coinbase() common.Address
	// The current time (block time)
	Time() *big.Int
	// Difficulty set on the current block
	Difficulty() *big.Int
	// The gas limit of the block
	GasLimit() *big.Int
	// Determines vm type
	VmType() Type
	// Env logger
	Logger() *logging.Logger
	// Determines whether it's possible to transact
	CanTransfer(from common.Address, balance *big.Int) bool
	// Transfers amount from one account to the other
	Transfer(from, to Account, amount *big.Int)
	// Adds a LOG to the state
	AddLog(*types.Log)
	// Type of the VM
	Vm() Vm
	// Dump vm runtime logs for debug
	DumpStructLog()
	// Get the curret calling depth
	Depth() int
	// Set the current calling depth
	SetDepth(i int)
	// Call another contract
	Call(me ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, op int32) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(me ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(me ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error)
	// Create a new contract
	Create(me ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error)
}

// Vm is the basic interface for an implementation of the EVM.
type Vm interface {
	// Run should execute the given contract with the input given in in
	// and return the contract execution return bytes or an error if it
	// failed.
	Run(c *Contract, in []byte) ([]byte, error)
	Finalise()
}

// Database is a EVM database for full state querying.
type Database interface {
	GetAccount(common.Address) Account
	CreateAccount(common.Address) Account

	AddBalance(common.Address, *big.Int)
	GetBalance(common.Address) *big.Int

	GetNonce(common.Address) uint64
	SetNonce(common.Address, uint64)

	GetCode(common.Address) []byte
	SetCode(common.Address, []byte)

	GetStatus(common.Address) int
	SetStatus(common.Address, int)

	AddDeployedContract(common.Address, common.Address)
	GetDeployedContract(common.Address) []string

	SetCreator(common.Address, common.Address)
	GetCreator(common.Address) common.Address

	SetCreateTime(common.Address, uint64)
	GetCreateTime(common.Address) uint64

	AddRefund(*big.Int)
	GetRefund() *big.Int

	GetState(common.Address, common.Hash) (bool, common.Hash)
	SetState(common.Address, common.Hash, common.Hash, int32)

	Delete(common.Address) bool
	Exist(common.Address) bool
	IsDeleted(common.Address) bool

	// Log
	StartRecord(common.Hash, common.Hash, int)
	AddLog(log *types.Log)
	GetLogs(hash common.Hash) types.Logs
	// Dump and Load
	Snapshot() interface{}
	RevertToSnapshot(interface{})
	RevertToJournal(uint64, uint64, []byte, db.Batch) error
	// Reset statuso
	Purge()

	Commit() (common.Hash, error)
	Reset() error
	// Query
	GetAccounts() map[string]Account
	Dump() []byte
	GetTree() interface{}
	// Atomic Related
	MarkProcessStart(uint64)
	MarkProcessFinish(uint64)

	FetchBatch(seqNo uint64) db.Batch
	DeleteBatch(seqNo uint64)
	MakeArchive(uint64)
	ShowArchive(common.Address, string) map[string]map[string]string

	RecomputeCryptoHash() (common.Hash, error)
	ResetToTarget(uint64, common.Hash)

	Apply(db.Database, db.Batch, common.Hash) error
}

// Account represents a contract or basic ethereum account.
type Account interface {
	SubBalance(amount *big.Int)
	AddBalance(amount *big.Int)
	SetBalance(*big.Int)
	SetNonce(uint64)
	Balance() *big.Int
	Address() common.Address
	ReturnGas(*big.Int, *big.Int)
	SetCode(common.Hash, []byte)
	ForEachStorage(cb func(key, value common.Hash) bool) map[common.Hash]common.Hash
	Value() *big.Int
}
