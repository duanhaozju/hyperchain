package core

import (
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/common"
	"math/big"
)

type VMEnv struct {
	state		*state.StateDB
	evm		*vm.Vm
	depth		int
	logs		[]vm.StructLog
	getHashFn	func(uint64) common.Hash
}

func NewEnv(state *state.StateDB,cfg vm.Config) *VMEnv {
	env := &VMEnv{
		state:		state,
	}
	if cfg.Logger.Collector == nil{
		cfg.Logger.Collector = env
	}
	env.evm = vm.New(env,cfg)
	return env
}
func (self *VMEnv) Vm() vm.Vm                { return self.evm }
func (self *VMEnv) Db() vm.Database          { return self.state }
func (self *VMEnv) Depth() int               { return self.depth }
func (self *VMEnv) SetDepth(i int)           { self.depth = i }
func (self *VMEnv) GetHash(n uint64) common.Hash {
	return self.getHashFn(n)
}

func (self *VMEnv) CanTransfer(from common.Address, balance *big.Int) bool {
	return self.state.GetBalance(from).Cmp(balance) >= 0
}

func (self *VMEnv) MakeSnapshot() vm.Database {
	return self.state.Copy()
}

func (self *VMEnv) SetSnapshot(copy vm.Database) {
	self.state.Set(copy.(*state.StateDB))
}

func (self *VMEnv) Transfer(from, to vm.Account, amount *big.Int) {
	Transfer(from, to, amount)
}

func (self *VMEnv) Call(me vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return Call(self, me, addr, data, gas, price, value)
}
func (self *VMEnv) CallCode(me vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return CallCode(self, me, addr, data, gas, price, value)
}

func (self *VMEnv) DelegateCall(me vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return DelegateCall(self, me, addr, data, gas, price)
}

func (self *VMEnv) Create(me vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return Create(self, me, data, gas, price, value)
}

func (self *VMEnv) StructLogs() []vm.StructLog {
	return self.logs
}

func (self *VMEnv) AddStructLog(log vm.StructLog) {
	self.logs = append(self.logs, log)
}
