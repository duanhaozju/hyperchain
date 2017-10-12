package jvm

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/crypto"
	"math/big"
)

type Env struct {
	depth     int
	state     vm.Database
	Gas       *big.Int
	origin    common.Address
	number    *big.Int
	time      *big.Int
	logs      []*types.Log
	namespace string
	txHash    common.Hash
	logger    *logging.Logger
	jvm       vm.Vm
}

func NewEnv(state vm.Database, number, timestamp int64, logger *logging.Logger, namespace string, txHash common.Hash, jvmCli ContractExecutor) *Env {
	env := &Env{
		state:     state,
		logger:    logger,
		time:      big.NewInt(timestamp),
		number:    big.NewInt(number),
		namespace: namespace,
		txHash:    txHash,
		Gas:       new(big.Int),
		jvm:       jvmCli,
	}
	return env
}

func (self *Env) Vm() vm.Vm                    { return self.jvm }
func (self *Env) Origin() common.Address       { return self.origin }
func (self *Env) BlockNumber() *big.Int        { return self.number }
func (self *Env) Time() *big.Int               { return self.time }
func (self *Env) Db() vm.Database              { return self.state }
func (self *Env) VmType() vm.Type              { return vm.JavaVmTy }
func (self *Env) Logger() *logging.Logger      { return self.logger }
func (self *Env) Namespace() string            { return self.namespace }
func (self *Env) TransactionHash() common.Hash { return self.txHash }
func (self *Env) GetHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
func (self *Env) AddLog(log *types.Log)        { self.Db().AddLog(log) }
func (self *Env) DumpStructLog()               {}
func (self *Env) Depth() int                   { return self.depth }
func (self *Env) SetDepth(i int)               { self.depth = i }
func (self *Env) MakeSnapshot() interface{}    { return self.state.Snapshot() }
func (self *Env) SetSnapshot(copy interface{}) { self.state.RevertToSnapshot(copy) }

// Call java based contract invocation.
func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, op int32) ([]byte, error) {
	return Call(self, caller, addr, data, gas, price, value, types.TransactionValue_Opcode(op))

}

// Deprecate
func (self *Env) CallCode(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return nil, nil
}

// Deprecate
func (self *Env) DelegateCall(caller vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return nil, nil
}

// Create deploy a java based contract.
func (self *Env) Create(caller vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return Create(self, caller, data, gas, price, value)
}
