// Copyright 2015 The go-ethereum Authors
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
package evm

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/types"
	"github.com/hyperchain/hyperchain/core/vm"
	"github.com/hyperchain/hyperchain/crypto"
	"github.com/op/go-logging"
	"math/big"
)

var (
	ForceJit    bool
	EnableJit   bool
	EnableDebug bool
)

func init() {
	EnableJit = true
	ForceJit = true
	EnableDebug = false
}

type Env struct {
	depth     int
	state     vm.Database
	origin    common.Address
	number    *big.Int
	time      *big.Int
	traceLog  []StructLog
	logs      []*types.Log
	logger    *logging.Logger
	namespace string
	txHash    common.Hash
	evm       *EVM
}

func NewEnv(state vm.Database, number, timestamp int64, logger *logging.Logger, namespace string, txHash common.Hash) *Env {
	env := &Env{
		state:     state,
		logger:    logger,
		time:      big.NewInt(timestamp),
		number:    big.NewInt(number),
		namespace: namespace,
		txHash:    txHash,
	}
	var cfg EVMConfig
	if EnableDebug {
		cfg = EVMConfig{
			EnableJit: EnableJit,
			ForceJit:  ForceJit,
			Debug:     EnableDebug,
			Logger: LogConfig{
				Collector: env,
			},
		}
	} else {
		cfg = EVMConfig{
			EnableJit: EnableJit,
			ForceJit:  ForceJit,
		}
	}
	env.evm = New(env, cfg)
	return env
}

func (self *Env) Vm() vm.Vm                    { return self.evm }
func (self *Env) Origin() common.Address       { return self.origin }
func (self *Env) BlockNumber() *big.Int        { return self.number }
func (self *Env) Time() *big.Int               { return self.time }
func (self *Env) Db() vm.Database              { return self.state }
func (self *Env) VmType() vm.Type              { return vm.StdVmTy }
func (self *Env) Logger() *logging.Logger      { return self.logger }
func (self *Env) Namespace() string            { return self.namespace }
func (self *Env) TransactionHash() common.Hash { return self.txHash }
func (self *Env) GetHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
func (self *Env) AddLog(log *types.Log)        { self.Db().AddLog(log) }
func (self *Env) AddStructLog(log StructLog)   { self.traceLog = append(self.traceLog, log) }
func (self *Env) DumpStructLog()               { StdErrFormat(self.traceLog) }
func (self *Env) Depth() int                   { return self.depth }
func (self *Env) SetDepth(i int)               { self.depth = i }
func (self *Env) MakeSnapshot() interface{}    { return self.state.Snapshot() }
func (self *Env) SetSnapshot(copy interface{}) { self.state.RevertToSnapshot(copy) }

func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, op int32) ([]byte, error) {
	ret, err := Call(self, caller, addr, data, gas, price, value, types.TransactionValue_Opcode(op))
	return ret, err

}
func (self *Env) CallCode(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return CallCode(self, caller, addr, data, gas, price, value)
}

func (self *Env) DelegateCall(caller vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return DelegateCall(self, caller, addr, data, gas, price)
}
func (self *Env) Create(caller vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return Create(self, caller, data, gas, price, value)
}
