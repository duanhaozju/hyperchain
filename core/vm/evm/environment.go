//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"math/big"
	"hyperchain/common"
	"github.com/op/go-logging"
	"hyperchain/core/vm"
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
	Db() vm.Database
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
	Transfer(from, to vm.Account, amount *big.Int)
	// Adds a LOG to the state
	AddLog(vm.Log)
	// Type of the VM
	Vm() vm.Vm
	// Get the curret calling depth
	Depth() int
	// Set the current calling depth
	SetDepth(i int)
	// Call another contract
	Call(me vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, op int32) ([]byte, error)
	// Take another's contract code and execute within our own context
	CallCode(me vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error)
	// Same as CallCode except sender and value is propagated from parent to child scope
	DelegateCall(me vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error)
	// Create a new contract
	Create(me vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error)
}

