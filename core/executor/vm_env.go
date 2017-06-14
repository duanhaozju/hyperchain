package executor

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/crypto"
	"hyperchain/hyperdb/db"
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

type Account struct {
	Balance string
	Code    string
	Nonce   string
	Storage map[string]string
}

type Log struct {
	AddressF string   `json:"address"`
	DataF    string   `json:"data"`
	TopicsF  []string `json:"topics"`
	BloomF   string   `json:"bloom"`
}

func (self Log) Address() []byte      { return common.Hex2Bytes(self.AddressF) }
func (self Log) Data() []byte         { return common.Hex2Bytes(self.DataF) }
func (self Log) RlpData() interface{} { return nil }
func (self Log) Topics() [][]byte {
	t := make([][]byte, len(self.TopicsF))
	for i, topic := range self.TopicsF {
		t[i] = common.Hex2Bytes(topic)
	}
	return t
}

func StateObjectFromAccount(db db.Database, addr string, account Account) *state.StateObject {
	obj := state.NewStateObject(common.HexToAddress(addr), nil)
	obj.SetBalance(common.Big(account.Balance))

	if common.IsHex(account.Code) {
		account.Code = account.Code[2:]
	}
	obj.SetCode(common.Hash{}, common.Hex2Bytes(account.Code))
	obj.SetNonce(common.Big(account.Nonce).Uint64())

	return obj
}

type VmEnv struct {
	CurrentCoinbase   string
	CurrentDifficulty string
	CurrentGasLimit   string
	CurrentNumber     string
	CurrentTimestamp  interface{}
	PreviousHash      string
}

type VmTest struct {
	Callcreates interface{}
	//Env         map[string]string
	Env           VmEnv
	Exec          map[string]string
	Transaction   map[string]string
	Logs          []Log
	Gas           string
	Out           string
	Post          map[string]Account
	Pre           map[string]Account
	PostStateRoot string
}

type RuleSet struct {
	HomesteadBlock *big.Int
	DAOForkBlock   *big.Int
	DAOForkSupport bool
}

func (r RuleSet) IsHomestead(n *big.Int) bool {
	return true
	return n.Cmp(r.HomesteadBlock) >= 0
}

type Env struct {
	ruleSet    RuleSet
	depth      int
	state      vm.Database
	Gas        *big.Int
	origin     common.Address
	coinbase   common.Address
	number     *big.Int
	time       *big.Int
	difficulty *big.Int
	gasLimit   *big.Int

	logs   []vm.StructLog
	logger *logging.Logger

	vmTest bool

	evm *vm.EVM
}

func NewEnv(ruleSet RuleSet, state vm.Database) *Env {
	env := &Env{
		ruleSet: ruleSet,
		state:   state,
	}
	return env
}

func NewEnvFromMap(ruleSet RuleSet, state vm.Database, envValues map[string]string, logger *logging.Logger) *Env {
	env := NewEnv(ruleSet, state)
	env.time = common.Big(envValues["currentTimestamp"])
	env.gasLimit = common.Big(envValues["currentGasLimit"])
	env.number = common.Big(envValues["currentNumber"])
	env.Gas = new(big.Int)
	var cfg vm.Config
	if EnableDebug {
		cfg = vm.Config{
			EnableJit: EnableJit,
			ForceJit:  ForceJit,
			Debug:     EnableDebug,
			Logger: vm.LogConfig{
				Collector: env,
			},
		}
	} else {
		cfg = vm.Config{
			EnableJit: EnableJit,
			ForceJit:  ForceJit,
		}
	}
	env.evm = vm.New(env, cfg)
	env.logger = logger

	return env
}

func (self *Env) RuleSet() vm.RuleSet      { return self.ruleSet }
func (self *Env) Vm() vm.Vm                { return self.evm }
func (self *Env) Origin() common.Address   { return self.origin }
func (self *Env) BlockNumber() *big.Int    { return self.number }
func (self *Env) Coinbase() common.Address { return self.coinbase }
func (self *Env) Time() *big.Int           { return self.time }
func (self *Env) Difficulty() *big.Int     { return self.difficulty }
func (self *Env) Db() vm.Database          { return self.state }
func (self *Env) GasLimit() *big.Int       { return self.gasLimit }
func (self *Env) VmType() vm.Type          { return vm.StdVmTy }
func (self *Env) Logger() *logging.Logger  { return self.logger }
func (self *Env) GetHash(n uint64) common.Hash {
	return common.BytesToHash(crypto.Keccak256([]byte(big.NewInt(int64(n)).String())))
}
func (self *Env) AddLog(log *types.Log) {
	self.state.AddLog(log)
}
func (self *Env) Depth() int     { return self.depth }
func (self *Env) SetDepth(i int) { self.depth = i }
func (self *Env) CanTransfer(from common.Address, balance *big.Int) bool {
	return self.state.GetBalance(from).Cmp(balance) >= 0
}
func (self *Env) AddStructLog(log vm.StructLog) {
	self.logs = append(self.logs, log)
}
func (self *Env) DumpStructLog() {
	vm.StdErrFormat(self.logs)
}

func (self *Env) MakeSnapshot() interface{} {
	return self.state.Snapshot()
}
func (self *Env) SetSnapshot(copy interface{}) {
	self.state.RevertToSnapshot(copy)
}

func (self *Env) Transfer(from, to vm.Account, amount *big.Int) {
	Transfer(from, to, amount)
}

func (self *Env) Call(caller vm.ContractRef, addr common.Address, data []byte, gas, price, value *big.Int, op int32) ([]byte, error) {
	ret, err := Call(self, caller, addr, data, gas, price, value, types.TransactionValue_Opcode(op))
	self.Gas = gas
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
