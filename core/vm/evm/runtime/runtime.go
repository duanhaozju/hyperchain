//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package runtime

import (
	"hyperchain/common"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm/evm"
	"hyperchain/crypto"
	"hyperchain/hyperdb/db"
	"math/big"
	"time"
)

// Config is a basic type specifying certain configuration flags for running
// the EVM.
type Config struct {
	Difficulty  *big.Int
	Origin      common.Address
	Receiver    common.Address
	Coinbase    common.Address
	BlockNumber *big.Int
	Time        *big.Int
	GasLimit    *big.Int
	GasPrice    *big.Int
	Value       *big.Int
	DisableJit  bool // "disable" so it's enabled by default
	Debug       bool

	State     *hyperstate.StateDB
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
	if cfg.conf == nil {
		cfg.conf = InitConf()
	}
}

// Execute executes the code using the input as call data during the execution.
// It returns the EVM's return value, the new state and an error if it failed.
//
// Executes sets up a in memory, temporarily, environment for the execution of
// the given code. It enabled the JIT by default and make sure that it's restored
// to it's original state afterwards.
func Execute(db db.Database, code, input []byte, cfg *Config) ([]byte, *hyperstate.StateDB, []evm.StructLog, error) {
	if cfg == nil {
		cfg = new(Config)
	}
	setDefaults(cfg)

	if cfg.State == nil {
		cfg.State = hyperstate.NewRaw(db, 0, "global", cfg.conf)
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
		cfg.State = hyperstate.NewRaw(db, 0, "global", cfg.conf)
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
		0,
	)
	return ret, err
}

func InitConf() *common.Config {
	conf := common.NewRawConfig()
	conf.Set(hyperstate.StateObjectBucketSize, 1000003)
	conf.Set(hyperstate.StateObjectBucketLevelGroup, 5)
	conf.Set(hyperstate.StateObjectBucketCacheSize, 100000)
	conf.Set(hyperstate.StateObjectDataNodeCacheSize, 100000)

	conf.Set(hyperstate.StateBucketSize, 1000003)
	conf.Set(hyperstate.StateBucketLevelGroup, 5)
	conf.Set(hyperstate.StateBucketCacheSize, 100000)
	conf.Set(hyperstate.StateDataNodeCacheSize, 100000)

	conf.Set(hyperstate.GlobalDataNodeCacheSize, 100000)
	conf.Set(hyperstate.GlobalDataNodeCacheLength, 20)
	return conf
}
