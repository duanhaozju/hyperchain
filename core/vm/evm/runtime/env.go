//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package runtime

import (
	"math/big"

	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/state"
	"hyperchain/core/evm"
)

// Env is a basic runtime environment required for running the EVM.
// 测试用
type Env struct {
	ruleSet    evm.RuleSet
	depth      int
	state      *state.StateDB

	origin     common.Address
	coinbase   common.Address

	number     *big.Int
	time       *big.Int
	difficulty *big.Int
	gasLimit   *big.Int

	logs       []evm.StructLog

	getHashFn  func(uint64) common.Hash

	evm        *evm.EVM
}

// NewEnv returns a new vm.Environment
func NewEnv(cfg *Config, state *state.StateDB) evm.Environment {
	env := &Env{
		ruleSet:    cfg.RuleSet,
		state:      state,
		origin:     cfg.Origin,
		coinbase:   cfg.Coinbase,
		number:     cfg.BlockNumber,
		time:       cfg.Time,
		difficulty: cfg.Difficulty,
		gasLimit:   cfg.GasLimit,
	}
	env.evm = evm.New(env, evm.Config{
		Debug:     cfg.Debug,
		EnableJit: !cfg.DisableJit,
		ForceJit:  !cfg.DisableJit,

		Logger: evm.LogConfig{
			Collector: env,
		},
	})

	return env
}

func (self *Env) StructLogs() []evm.StructLog {
	return self.logs
}

func (self *Env) AddStructLog(log evm.StructLog) {
	self.logs = append(self.logs, log)
}

func (self *Env) RuleSet() evm.RuleSet      { return self.ruleSet }
func (self *Env) Vm() evm.Vm                { return self.evm }
func (self *Env) Origin() common.Address   { return self.origin }
func (self *Env) BlockNumber() *big.Int    { return self.number }
func (self *Env) Coinbase() common.Address { return self.coinbase }
func (self *Env) Time() *big.Int           { return self.time }
func (self *Env) Difficulty() *big.Int     { return self.difficulty }
func (self *Env) Db() evm.Database          { return self.state }
func (self *Env) GasLimit() *big.Int       { return self.gasLimit }
func (self *Env) VmType() evm.Type          { return evm.StdVmTy }
func (self *Env) GetHash(n uint64) common.Hash {
	return self.getHashFn(n)
}
func (self *Env) AddLog(log *evm.Log) {
	self.state.AddLog(log)
}
func (self *Env) Depth() int     { return self.depth }
func (self *Env) SetDepth(i int) { self.depth = i }
func (self *Env) CanTransfer(from common.Address, balance *big.Int) bool {
	return self.state.GetBalance(from).Cmp(balance) >= 0
}
func (self *Env) MakeSnapshot() evm.Database {
	return self.state.Copy()
}
func (self *Env) SetSnapshot(copy evm.Database) {
	self.state.Set(copy.(*state.StateDB))
}

func (self *Env) Transfer(from, to evm.Account, amount *big.Int) {
	core.Transfer(from, to, amount)
}

func (self *Env) Call(caller evm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return core.Call(self, caller, addr, data, gas, price, value)
}
func (self *Env) CallCode(caller evm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return core.CallCode(self, caller, addr, data, gas, price, value)
}

func (self *Env) DelegateCall(me evm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return core.DelegateCall(self, me, addr, data, gas, price)
}

func (self *Env) Create(caller evm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return core.Create(self, caller, data, gas, price, value)
}
