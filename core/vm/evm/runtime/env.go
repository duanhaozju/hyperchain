//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package runtime

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/ledger/state"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm"
	"math/big"
)

// Env is a basic runtime environment required for running the EVM.
type Env struct {
	namespace string
	depth     int
	state     *state.StateDB
	origin    common.Address
	number    *big.Int
	time      *big.Int
	logs      []evm.StructLog
	getHashFn func(uint64) common.Hash
	evm       *evm.EVM
}

// NewEnv returns a new vm.Environment
func NewEnv(cfg *Config, state *state.StateDB) vm.Environment {
	env := &Env{
		state:  state,
		origin: cfg.Origin,
		number: cfg.BlockNumber,
		time:   cfg.Time,
	}
	env.evm = evm.New(env, evm.EVMConfig{
		Debug:     cfg.Debug,
		EnableJit: !cfg.DisableJit,
		ForceJit:  !cfg.DisableJit,
		Logger: evm.LogConfig{
			Collector:      env,
			DisableStorage: cfg.DisableStorage,
			DisableMemory:  cfg.DisableMemory,
			DisableStack:   cfg.DisableStack,
		},
	})
	return env
}

func (self *Env) StructLogs() []evm.StructLog    { return self.logs }
func (self *Env) AddStructLog(log evm.StructLog) { self.logs = append(self.logs, log) }

func (self *Env) Vm() vm.Vm                    { return self.evm }
func (self *Env) Origin() common.Address       { return self.origin }
func (self *Env) BlockNumber() *big.Int        { return self.number }
func (self *Env) Time() *big.Int               { return self.time }
func (self *Env) Db() vm.Database              { return self.state }
func (self *Env) VmType() vm.Type              { return vm.StdVmTy }
func (self *Env) GetHash(n uint64) common.Hash { return self.getHashFn(n) }
func (self *Env) AddLog(log *types.Log)        { self.state.AddLog(log) }
func (self *Env) Depth() int                   { return self.depth }
func (self *Env) SetDepth(i int)               { self.depth = i }
func (self *Env) MakeSnapshot() interface{}    { return self.state.Snapshot() }
func (self *Env) SetSnapshot(copy interface{}) { self.state.RevertToSnapshot(copy) }
func (self *Env) Namespace() string            { return self.namespace }
func (self *Env) TransactionHash() common.Hash { return common.Hash{} }
func (self *Env) DumpStructLog()               { evm.StdErrFormat(self.logs) }
func (self *Env) Logger() *logging.Logger      { return logging.MustGetLogger("runtime") }

func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, opcode int32) ([]byte, error) {
	return evm.Call(self, caller, addr, data, gas, price, value, types.TransactionValue_Opcode(opcode))
}
func (self *Env) CallCode(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return evm.CallCode(self, caller, addr, data, gas, price, value)
}

func (self *Env) DelegateCall(me vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return evm.DelegateCall(self, me, addr, data, gas, price)
}

func (self *Env) Create(caller vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return evm.Create(self, caller, data, gas, price, value)
}
