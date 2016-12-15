//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package tests

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/vm"
	"hyperchain/hyperdb"
)


type Account struct {
	Balance string
	Code    string
	Nonce   string
	Storage map[string]string
}


type VmTest struct {
	Callcreates interface{}
	//Env         map[string]string
	Env           VmEnv
	Exec          map[string]string
	Transaction   map[string]string
	Gas           string
	Out           string
	Post          map[string]Account
	Pre           map[string]Account
	PostStateRoot string
}



func StateObjectFromAccount(db hyperdb.Database, addr string, account Account) *state.StateObject {
	obj := state.NewStateObject(common.HexToAddress(addr),db)
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

