package jvm
import (
	"hyperchain/common"
	"hyperchain/crypto"
	"math/big"
	"github.com/op/go-logging"
	"hyperchain/core/vm"
)

type Account struct {
	Balance string
	Code    string
	Nonce   string
	Storage map[string]string
}

type Env struct {
	depth      int
	state      vm.Database
	Gas        *big.Int
	origin     common.Address
	coinbase   common.Address
	number     *big.Int
	time       *big.Int
	difficulty *big.Int
	gasLimit   *big.Int

	logger     *logging.Logger
	jvm        vm.Vm
}

func NewEnv(state vm.Database, setting map[string]string, logger *logging.Logger) *Env {
	env := &Env{
		state:     state,
		logger:    logger,
		time:      common.Big(setting["currentTimestamp"]),
		gasLimit:  common.Big(setting["currentGasLimit"]),
		number:    common.Big(setting["currentNumber"]),
		Gas:       new(big.Int),
	}
	// TODO new JVM executor
	//env.evm = New(env, Config{
	//	EnableJit: EnableJit,
	//	ForceJit:  ForceJit,
	//})
	return env
}

func (self *Env) Vm() vm.Vm                { return self.jvm }
func (self *Env) Origin() common.Address   { return self.origin }
func (self *Env) BlockNumber() *big.Int    { return self.number }
// Deprecate
func (self *Env) Coinbase() common.Address { return common.Address{}}
func (self *Env) Time() *big.Int           { return self.time }
// Deprecate
func (self *Env) Difficulty() *big.Int     { return nil}
func (self *Env) Db() vm.Database          { return self.state }
// Deprecate
func (self *Env) GasLimit() *big.Int       { return nil}
func (self *Env) VmType() vm.Type          { return vm.JavaVmTy }
func (self *Env) Logger() *logging.Logger  { return self.logger}
func (self *Env) GetHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
func (self *Env) AddLog(log vm.Log) {
	self.state.AddLog(log)
}
func (self *Env) Depth() int     { return self.depth }
func (self *Env) SetDepth(i int) { self.depth = i }
// Deprecate
func (self *Env) CanTransfer(from common.Address, balance *big.Int) bool {
	return true
}
func (self *Env) MakeSnapshot() interface{} {
	return self.state.Snapshot()
}
func (self *Env) SetSnapshot(copy interface{}) {
	self.state.RevertToSnapshot(copy)
}


// Deprecate
func (self *Env) Transfer(from, to vm.Account, amount *big.Int) {
}
// Deprecate
func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, op int32) ([]byte, error) {
	return nil, nil

}
// Deprecate
func (self *Env) CallCode(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int) ([]byte, error) {
	return nil, nil
}

// Deprecate
func (self *Env) DelegateCall(caller vm.ContractRef, addr common.Address, data []byte, gas, price *big.Int) ([]byte, error) {
	return nil, nil
}

// Deprecate
func (self *Env) Create(caller vm.ContractRef, data []byte, gas, price, value *big.Int) ([]byte, common.Address, error) {
	return nil, common.Address{}, nil
}
