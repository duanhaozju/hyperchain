//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"fmt"
	"math/big"
	"os"
	"unicode"

	"hyperchain/common"
	"hyperchain/core/vm"
)

type Storage map[common.Hash]common.Hash

func (self Storage) Copy() Storage {
	cpy := make(Storage)
	for key, value := range self {
		cpy[key] = value
	}

	return cpy
}

// StructLogCollector is the basic interface to capture emited logs by the EVM logger.
type StructLogCollector interface {
	// Adds the structured log to the collector.
	AddStructLog(StructLog)
}

// LogConfig are the configuration options for structured logger the EVM
type LogConfig struct {
	DisableMemory  bool               // disable memory capture
	DisableStack   bool               // disable stack capture
	DisableStorage bool               // disable storage capture
	FullStorage    bool               // show full storage (slow)
	Collector      StructLogCollector // the log collector
}

// StructLog is emitted to the Environment each cycle and lists information about the current internal state
// prior to the execution of the statement.
type StructLog struct {
	Pc      uint64
	Op      OpCode
	Gas     *big.Int
	GasCost *big.Int
	Memory  []byte
	Stack   []*big.Int
	Storage map[common.Hash]common.Hash
	Depth   int
	Err     error
}

// Logger is an EVM state logger and implements VmLogger.
//
// Logger can capture state based on the given Log configuration and also keeps
// a track record of modified storage which is used in reporting snapshots of the
// contract their storage.
type Logger struct {
	cfg LogConfig

	env           vm.Environment
	changedValues map[common.Address]Storage
}

// newLogger returns a new logger
func NewLogger(cfg LogConfig, env vm.Environment) *Logger {
	return &Logger{
		cfg:           cfg,
		env:           env,
		changedValues: make(map[common.Address]Storage),
	}
}

// captureState logs a new structured log message and pushes it out to the environment
//
// captureState also tracks SSTORE ops to track dirty values.
func (l *Logger) captureState(pc uint64, op OpCode, gas, cost *big.Int, memory *Memory, stack *stack, contract *Contract, depth int, err error) {
	// short circuit if no log collector is present
	if l.cfg.Collector == nil {
		return
	}

	// initialise new changed values storage container for this contract
	// if not present.
	if l.changedValues[contract.Address()] == nil {
		l.changedValues[contract.Address()] = make(Storage)
	}

	// capture SSTORE opcodes and determine the changed value and store
	// it in the local storage container. NOTE: we do not need to do any
	// range checks here because that's already handler prior to calling
	// this function.
	switch op {
	case SSTORE:
		var (
			value   = common.BigToHash(stack.data[stack.len()-2])
			address = common.BigToHash(stack.data[stack.len()-1])
		)
		l.changedValues[contract.Address()][address] = value
	}

	// copy a snapstot of the current memory state to a new buffer
	var mem []byte
	if !l.cfg.DisableMemory {
		mem = make([]byte, len(memory.Data()))
		copy(mem, memory.Data())
	}

	// copy a snapshot of the current stack state to a new buffer
	var stck []*big.Int
	if !l.cfg.DisableStack {
		stck = make([]*big.Int, len(stack.Data()))
		for i, item := range stack.Data() {
			stck[i] = new(big.Int).Set(item)
		}
	}

	// Copy the storage based on the settings specified in the log config. If full storage
	// is disabled (default) we can use the simple Storage.Copy method, otherwise we use
	// the state object to query for all values (slow process).
	var storage Storage
	if !l.cfg.DisableStorage {
		if l.cfg.FullStorage {
			storage = make(Storage)
			// Get the contract account and loop over each storage entry. This may involve looping over
			// the trie and is a very expensive process.
			l.env.Db().GetAccount(contract.Address()).ForEachStorage(func(key common.Hash, value []byte) bool {
				storage[key] = common.BytesToHash(value)
				// Return true, indicating we'd like to continue.
				return true
			})
		} else {
			// copy a snapshot of the current storage to a new container.
			storage = l.changedValues[contract.Address()].Copy()
		}
	}
	// create a new snaptshot of the EVM.
	log := StructLog{pc, op, new(big.Int).Set(gas), cost, mem, stck, storage, l.env.Depth(), err}
	// Add the log to the collector
	l.cfg.Collector.AddStructLog(log)
}

// StdErrFormat formats a slice of StructLogs to human readable format
func StdErrFormat(logs []StructLog) {
	fmt.Fprintf(os.Stdout, "VM STAT %d OPs\n", len(logs))
	for _, log := range logs {
		fmt.Fprintf(os.Stdout, "PC %08d: [%s] GAS: %v COST: %v", log.Pc, log.Op.String(), log.Gas, log.GasCost)
		if log.Err != nil {
			fmt.Fprintf(os.Stdout, " ERROR: %v", log.Err)
		}
		fmt.Fprintf(os.Stdout, "\n")

		fmt.Fprintln(os.Stdout, "STACK =", len(log.Stack))

		for i := len(log.Stack) - 1; i >= 0; i-- {
			fmt.Fprintf(os.Stdout, "%04d: %x\n", len(log.Stack)-i-1, common.LeftPadBytes(log.Stack[i].Bytes(), 32))
		}

		// const maxMem = 10
		addr := 0
		fmt.Fprintln(os.Stdout, "MEM =", len(log.Memory))
		for i := 0; i+16 <= len(log.Memory); /*&& addr < maxMem*/ i += 16 {
			data := log.Memory[i : i+16]
			str := fmt.Sprintf("%04d: % x  ", addr*16, data)
			for _, r := range data {
				if r == 0 {
					str += "."
				} else if unicode.IsPrint(rune(r)) {
					str += fmt.Sprintf("%s", string(r))
				} else {
					str += "?"
				}
			}
			addr++
			fmt.Fprintln(os.Stdout, str)
		}

		fmt.Fprintln(os.Stdout, "STORAGE =", len(log.Storage))
		for h, item := range log.Storage {
			fmt.Fprintf(os.Stdout, "%x: %x\n", h, item)
		}
		fmt.Fprintln(os.Stdout)
	}
}
