//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package runtime

import (
	"math/big"
	"time"

	"hyperchain/common"
	"hyperchain/core/crypto"
	"hyperchain/core/state"
	"hyperchain/core/evm"
	//"hyperchain/hyperdb"
	"hyperchain/hyperdb"
)

// The default, always homestead, rule set for the vm env
type ruleSet struct{}

func (ruleSet) IsHomestead(*big.Int) bool { return true }

// Config is a basic type specifying certain configuration flags for running
// the EVM.
type Config struct {
	RuleSet     evm.RuleSet
	Difficulty  *big.Int
	Origin      common.Address
	Coinbase    common.Address
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    *big.Int
	GasPrice    *big.Int
	Value       *big.Int
	DisableJit  bool // "disable" so it's enabled by default
	Debug       bool

	State       *state.StateDB
	GetHashFn   func(n uint64) common.Hash
}

// sets defaults on the config
func setDefaults(cfg *Config) {
	if cfg.RuleSet == nil {
		cfg.RuleSet = ruleSet{}
	}

	if cfg.Difficulty == nil {
		cfg.Difficulty = new(big.Int)
	}
	if cfg.Time == nil {
		cfg.Time = big.NewInt(time.Now().Unix())
	}
	if cfg.GasLimit == nil {
		cfg.GasLimit = new(big.Int).Set(common.MaxBig)
	}
	if cfg.GasPrice == nil {
		cfg.GasPrice = new(big.Int)
	}
	if cfg.Value == nil {
		cfg.Value = new(big.Int)
	}
	if cfg.BlockNumber == nil {
		cfg.BlockNumber = new(big.Int)
	}
	if cfg.GetHashFn == nil {
		cfg.GetHashFn = func(n uint64) common.Hash {
			return common.BytesToHash(crypto.Keccak256([]byte(new(big.Int).SetUint64(n).String())))
		}
	}
}

// Execute executes the code using the input as call data during the execution.
// It returns the EVM's return value, the new state and an error if it failed.
//
// Executes sets up a in memory, temporarily, environment for the execution of
// the given code. It enabled the JIT by default and make sure that it's restored
// to it's original state afterwards.
func Execute(code, input []byte, cfg *Config) ([]byte, *state.StateDB, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	setDefaults(cfg)

	if cfg.State == nil {
		db, _ := hyperdb.NewMemDatabase()
		cfg.State, _ = state.New(common.Hash{}, db)
	}
	var (
		vmenv    = NewEnv(cfg, cfg.State)
		sender   = cfg.State.CreateAccount(cfg.Origin)
		receiver = cfg.State.CreateAccount(common.StringToAddress("contract"))
	)
	// set the receiver's (the executing contract) code for execution.
	receiver.SetCode(code)

	// Call the code with the given configuration.
	ret, err := vmenv.Call(
		sender,
		receiver.Address(),
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
	)

	return ret, cfg.State, err
}

// Call executes the code given by the contract's address. It will return the
// EVM's return value or an error if it failed.
//
// Call, unlike Execute, requires a config and also requires the State field to
// be set.
func Call(address common.Address, input []byte, cfg *Config) ([]byte, error) {
	setDefaults(cfg)

	vmenv := NewEnv(cfg, cfg.State)

	sender := cfg.State.GetOrNewStateObject(cfg.Origin)
	// Call the code with the given configuration.
	ret, err := vmenv.Call(
		sender,
		address,
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
	)

	return ret, err
}