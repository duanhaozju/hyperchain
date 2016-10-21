package tests

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
)

func checkLogs(tlog []Log, logs vm.Logs) error {	return nil	}

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

func StateObjectFromAccount(db hyperdb.Database, addr string, account Account) *state.StateObject {
	obj := state.NewStateObject(common.HexToAddress(addr),db)
	obj.SetBalance(common.Big(account.Balance))

	if common.IsHex(account.Code) {
		account.Code = account.Code[2:]
	}
	obj.SetCode(common.Hex2Bytes(account.Code))
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
	ruleSet      RuleSet
	depth        int
	state        *state.StateDB
	Gas          *big.Int
	origin   common.Address
	coinbase common.Address
	number     *big.Int
	time       *big.Int
	difficulty *big.Int
	gasLimit   *big.Int

	logs []vm.StructLog

	vmTest bool

	evm *vm.EVM
}

