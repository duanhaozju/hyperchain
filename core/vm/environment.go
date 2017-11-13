// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package vm

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/op/go-logging"
	"math/big"
)

type Type byte

const (
	StdVmTy Type = iota // Default standard ethereum VM
	JavaVmTy
	MaxVmTy
)

// Environment is a VM requirement and helper which allows access to outside
// information such as states.
type Environment interface {
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
	// The current time (block time)
	Time() *big.Int
	// Determines vm type
	VmType() Type
	// Env logger
	Logger() *logging.Logger
	// Namespace
	Namespace() string
	// Current transaction hash
	TransactionHash() common.Hash
	// Adds a LOG to the state
	AddLog(*types.Log)
	// Dump vm execution trace info
	DumpStructLog()
	// Type of the VM
	Vm() Vm
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
