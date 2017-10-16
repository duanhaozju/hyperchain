//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package runtime

import (
	"hyperchain/common"
	"hyperchain/core/ledger/state"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm"
	"hyperchain/crypto"
	"hyperchain/hyperdb/db"
	"math/big"
	"time"
)

// Config is a basic type specifying certain configuration flags for running
// the EVM.
type Config struct {
	Origin      common.Address
	Receiver    common.Address
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    *big.Int
	GasPrice    *big.Int
	Value       *big.Int
	DisableJit  bool // "disable" so it's enabled by default
	Debug       bool

	State     *state.StateDB
	GetHashFn func(n uint64) common.Hash

	logs           []evm.StructLog
	conf           *common.Config
	DisableStack   bool
	DisableMemory  bool
	DisableStorage bool
}

// sets defaults on the config
func setDefaults(cfg *Config) {
	if (cfg.Origin == common.Address{}) {
		cfg.Origin = common.StringToAddress("sender")
	}
	if (cfg.Receiver == common.Address{}) {
		cfg.Receiver = common.StringToAddress("receiver")
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
	if cfg.conf == nil {
		cfg.conf = DefaultConf()
	}
}

// Execute executes the code using the input as call data during the execution.
// It returns the EVM's return value, the new state and an error if it failed.
//
// Executes sets up a in memory, temporarily, environment for the execution of
// the given code. It enabled the JIT by default and make sure that it's restored
// to it's original state afterwards.
func Execute(db db.Database, code, input []byte, cfg *Config) ([]byte, *state.StateDB, []evm.StructLog, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	setDefaults(cfg)

	if cfg.State == nil {
		cfg.State, _ = state.New(common.Hash{}, db, db, cfg.conf, 0)
	}
	var (
		vmenv    = NewEnv(cfg, cfg.State)
		sender   = cfg.State.CreateAccount(cfg.Origin)
		receiver = cfg.State.CreateAccount(cfg.Receiver)
	)
	// set the receiver's (the executing contract) code for execution.
	receiver.SetCode(common.BytesToHash(crypto.Keccak256(code)), code)
	// Call the code with the given configuration.
	ret, err := vmenv.Call(
		sender,
		receiver.Address(),
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
		0,
	)

	return ret, cfg.State, vmenv.(*Env).StructLogs(), err
}

// Create executes the code using the EVM create method
func Create(db db.Database, input []byte, cfg *Config) ([]byte, common.Address, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	setDefaults(cfg)

	if cfg.State == nil {
		cfg.State, _ = state.New(common.Hash{}, db, db, cfg.conf, 0)
	}
	var (
		vmenv  = NewEnv(cfg, cfg.State)
		sender = cfg.State.CreateAccount(cfg.Origin)
	)

	// Call the code with the given configuration.
	code, address, err := vmenv.Create(
		sender,
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
	)
	return code, address, err
}

// Call executes the code given by the contract's address. It will return the
// EVM's return value or an error if it failed.
//
// Call, unlike Execute, requires a config and also requires the State field to
// be set.
func Call(address common.Address, input []byte, cfg *Config) ([]byte, error) {
	var sender vm.ContractRef
	setDefaults(cfg)

	vmenv := NewEnv(cfg, cfg.State)

	sender = cfg.State.GetStateObject(cfg.Origin)
	if sender == nil {
		sender = cfg.State.CreateAccount(cfg.Origin)
	}
	// Call the code with the given configuration.
	ret, err := vmenv.Call(
		sender,
		address,
		input,
		cfg.GasLimit,
		cfg.GasPrice,
		cfg.Value,
		0,
	)
	return ret, err
}

func DefaultConf() *common.Config {
	conf := common.NewRawConfig()
	conf.Set(state.StateCapacity, 1000003)
	conf.Set(state.StateAggreation, 5)
	conf.Set(state.StateMerkleCacheSize, 100000)
	conf.Set(state.StateBucketCacheSize, 100000)

	conf.Set(state.StateObjectCapacity, 1000003)
	conf.Set(state.StateObjectAggreation, 5)
	conf.Set(state.StateObjectMerkleCacheSize, 100000)
	conf.Set(state.StateObjectBucketCacheSize, 100000)
	return conf
}
