//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"fmt"
	"math/big"

	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/crypto"
	"os"
)

// Config are the configuration options for the EVM
type Config struct {
	Debug              bool
	EnableJit          bool
	ForceJit           bool
	Logger             LogConfig
	DisableGasMetering bool
}

// EVM is used to run Ethereum based contracts and will utilise the
// passed environment to query external sources for state information.
// The EVM will run the byte code VM or JIT VM based on the passed
// configuration.
type EVM struct {
	env       vm.Environment
	jumpTable vmJumpTable
	cfg       Config

	logger *Logger
}

// New returns a new instance of the EVM.
func New(env vm.Environment, cfg Config) *EVM {
	var logger *Logger
	if cfg.Debug {
		logger = NewLogger(cfg.Logger, env)
	}

	return &EVM{
		env:       env,
		jumpTable: newJumpTable(),
		cfg:       cfg,
		logger:    logger,
	}
}

// Run loops and evaluates the contract's code with the given input data
func (evm *EVM) Run(context vm.VmContext, input []byte) (ret []byte, err error) {
	var contract *Contract
	var ok bool
	if contract, ok = context.(*Contract); !ok {
		return nil, nil
	}
	// 1.设置虚拟机深度+1
	evm.env.SetDepth(evm.env.Depth() + 1)
	defer evm.env.SetDepth(evm.env.Depth() - 1)

	// 2.判断CodeAddr是否为空,如果不为空就去找已编译好的合约地址,然后运行该原生的智能合约
	if contract.CodeAddr != nil {
		if p := Precompiled[contract.CodeAddr.Str()]; p != nil {
			return evm.RunPrecompiled(p, input, contract)
		}
	}

	// Don't bother with the execution if there's no code.
	// 3.如果合约代码为空则返回空
	if len(contract.Code) == 0 {
		return nil, nil
	}

	var (
		codehash = crypto.Keccak256Hash(contract.Code) // codehash is used when doing jump dest caching
		program  *Program
	)
	// 4.如果可以使用JIT运行时编译
	if evm.cfg.EnableJit {
		// If the JIT is enabled check the status of the JIT program,
		// if it doesn't exist compile a new program in a separate
		// goroutine or wait for compilation to finish if the JIT is
		// forced.
		switch GetProgramStatus(codehash) {
		// 判断是否已经可用,如果可用直接用找
		case progReady:
			return RunProgram(evm, GetProgram(codehash), evm.env, contract, input)
		case progUnknown:
			// 如果不可用,且强制jit,则顺序执行且立刻执行
			if evm.cfg.ForceJit {
				// Create and compile program
				program = NewProgram(contract.Code)
				perr := CompileProgram(program)
				if perr == nil {
					return RunProgram(evm, program, evm.env, contract, input)
				}
			} else {
				// 否则可以另开一个线程
				// create and compile the program. Compilation
				// is done in a separate goroutine
				program = NewProgram(contract.Code)
				go func() {
					err := CompileProgram(program)
					if err != nil {
						return
					}
				}()
			}
		}
	}

	var (
		caller     = contract.caller
		code       = contract.Code
		instrCount = 0

		op      OpCode         // current opcode
		mem     = NewMemory()  // bound memory
		stack   = newstack()   // local stack
		statedb = evm.env.Db() // current state
		// For optimisation reason we're using uint64 as the program counter.
		// It's theoretically possible to go above 2^64. The YP defines the PC to be uint256. Practically much less so feasible.
		pc = uint64(0) // program counter

		// jump evaluates and checks whether the given jump destination is a valid one
		// if valid move the `pc` otherwise return an error.
		jump = func(from uint64, to *big.Int) error {
			if !contract.jumpdests.has(codehash, code, to) {
				nop := contract.GetOp(to.Uint64())
				return fmt.Errorf("invalid jump destination (%v) %v", nop, to)
			}

			pc = to.Uint64()

			return nil
		}

		newMemSize *big.Int
		cost       *big.Int
	)
	contract.Input = input

	// User defer pattern to check for an error and, based on the error being nil or not, use all gas and return.
	defer func() {
		if err != nil && evm.cfg.Debug {
			evm.logger.captureState(pc, op, contract.Gas, cost, mem, stack, contract, evm.env.Depth(), err)
		}
	}()

	// 一个指令一个指令执行
	for ; ; instrCount++ {
		/*
			if EnableJit && it%100 == 0 {
				if program != nil && progStatus(atomic.LoadInt32(&program.status)) == progReady {
					// move execution
					fmt.Println("moved", it)
					glog.V(logger.Info).Infoln("Moved execution to JIT")
					return runProgram(program, pc, mem, stack, evm.env, contract, input)
				}
			}
		*/

		// Get the memory location of pc
		// 得到pc的内存地址
		op = contract.GetOp(pc)
		// calculate the new memory size and gas price for the current executing opcode
		// 对当前执行的操作码计算新的内存大小和gas价格
		newMemSize, cost, err = calculateGasAndSize(evm.env, contract, caller, op, statedb, mem, stack)
		if err != nil {
			return nil, err
		}

		// Use the calculated gas. When insufficient gas is present, use all gas and return an
		// Out Of Gas error
		if !contract.UseGas(cost) {
			return nil, OutOfGasError
		}

		// Resize the memory calculated previously
		mem.Resize(newMemSize.Uint64())
		// Add a log message
		if evm.cfg.Debug {
			evm.logger.captureState(pc, op, contract.Gas, cost, mem, stack, contract, evm.env.Depth(), nil)
		}

		if opPtr := evm.jumpTable[op]; opPtr.valid {
			if opPtr.fn != nil {
				opPtr.fn(instruction{}, &pc, evm.env, contract, mem, stack)
				//log.Info("----------opPtr--------------0",op)
			} else {
				//log.Info("----------opPtr--------------1",op)
				switch op {
				case PC:
					opPc(instruction{data: new(big.Int).SetUint64(pc)}, &pc, evm.env, contract, mem, stack)
				case JUMP:
					if err := jump(pc, stack.pop()); err != nil {
						return nil, err
					}

					continue
				case JUMPI:
					pos, cond := stack.pop(), stack.pop()

					if cond.Cmp(common.BigTrue) >= 0 {
						if err := jump(pc, pos); err != nil {
							return nil, err
						}

						continue
					}
				case RETURN:
					offset, size := stack.pop(), stack.pop()
					ret := mem.GetPtr(offset.Int64(), size.Int64())
					return ret, nil
				case SUICIDE:
					opSuicide(instruction{}, nil, evm.env, contract, mem, stack)

					fallthrough
				case STOP: // Stop the contract
					return nil, nil
				}
			}
		} else {
			return nil, fmt.Errorf("Invalid opcode %x", op)
		}

		pc++

	}
}

// calculateGasAndSize calculates the required given the opcode and stack items calculates the new memorysize for
// the operation. This does not reduce gas or resizes the memory.
func calculateGasAndSize(env vm.Environment, contract *Contract, caller vm.ContractRef, op OpCode, statedb vm.Database, mem *Memory, stack *stack) (*big.Int, *big.Int, error) {
	var (
		//gas                 = new(big.Int)
		newMemSize *big.Int = new(big.Int)
	)
	err := baseCheck(op, stack)
	if err != nil {
		return nil, nil, err
	}

	// stack Check, memory resize & gas phase
	switch op {
	case SWAP1, SWAP2, SWAP3, SWAP4, SWAP5, SWAP6, SWAP7, SWAP8, SWAP9, SWAP10, SWAP11, SWAP12, SWAP13, SWAP14, SWAP15, SWAP16:
		n := int(op - SWAP1 + 2)
		err := stack.require(n)
		if err != nil {
			return nil, nil, err
		}
	case DUP1, DUP2, DUP3, DUP4, DUP5, DUP6, DUP7, DUP8, DUP9, DUP10, DUP11, DUP12, DUP13, DUP14, DUP15, DUP16:
		n := int(op - DUP1 + 1)
		err := stack.require(n)
		if err != nil {
			return nil, nil, err
		}
	case LOG0, LOG1, LOG2, LOG3, LOG4:
		n := int(op - LOG0)
		err := stack.require(n + 2)
		if err != nil {
			return nil, nil, err
		}
		mSize, mStart := stack.data[stack.len()-2], stack.data[stack.len()-1]

		newMemSize = calcMemSize(mStart, mSize)
	case EXP:
	//gas.Add(gas, new(big.Int).Mul(big.NewInt(int64(len(stack.data[stack.len()-2].Bytes()))), params.ExpByteGas))
	case SSTORE:
	//err := stack.require(2)
	//if err != nil {
	//	return nil, nil, err
	//}
	//
	//var g *big.Int
	//y, x := stack.data[stack.len()-2], stack.data[stack.len()-1]
	//_, val := statedb.GetState(contract.Address(), common.BigToHash(x))
	//
	//// This checks for 3 scenario's and calculates gas accordingly
	//// 1. From a zero-value address to a non-zero value         (NEW VALUE)
	//// 2. From a non-zero value address to a zero-value address (DELETE)
	//// 3. From a non-zero to a non-zero                         (CHANGE)
	//if common.EmptyHash(val) && !common.EmptyHash(common.BigToHash(y)) {
	//	// 0 => non 0
	//	g = params.SstoreSetGas
	//} else if !common.EmptyHash(val) && common.EmptyHash(common.BigToHash(y)) {
	//	statedb.AddRefund(params.SstoreRefundGas)
	//
	//	g = params.SstoreClearGas
	//} else {
	//	// non 0 => non 0 (or 0 => 0)
	//	g = params.SstoreClearGas
	//}
	//gas.Set(g)
	case SUICIDE:
		// Do nothing
	case MLOAD:
		newMemSize = calcMemSize(stack.peek(), u256(32))
	case MSTORE8:
		newMemSize = calcMemSize(stack.peek(), u256(1))
	case MSTORE:
		newMemSize = calcMemSize(stack.peek(), u256(32))
	case RETURN:
		newMemSize = calcMemSize(stack.peek(), stack.data[stack.len()-2])
	case SHA3:
		newMemSize = calcMemSize(stack.peek(), stack.data[stack.len()-2])

	//words := toWordSize(stack.data[stack.len()-2])
	//gas.Add(gas, words.Mul(words, params.Sha3WordGas))
	case CALLDATACOPY:
		newMemSize = calcMemSize(stack.peek(), stack.data[stack.len()-3])

	//words := toWordSize(stack.data[stack.len()-3])
	//gas.Add(gas, words.Mul(words, params.CopyGas))
	case CODECOPY:
		newMemSize = calcMemSize(stack.peek(), stack.data[stack.len()-3])

	//words := toWordSize(stack.data[stack.len()-3])
	//gas.Add(gas, words.Mul(words, params.CopyGas))
	case EXTCODECOPY:
		newMemSize = calcMemSize(stack.data[stack.len()-2], stack.data[stack.len()-4])

	//words := toWordSize(stack.data[stack.len()-4])
	//gas.Add(gas, words.Mul(words, params.CopyGas))

	case CREATE:
		newMemSize = calcMemSize(stack.data[stack.len()-2], stack.data[stack.len()-3])
	case CALL, CALLCODE:
		//gas.Add(gas, stack.data[stack.len()-1])

		//if op == CALL {
		//	if !env.Db().Exist(common.BigToAddress(stack.data[stack.len()-2])) {
		//		gas.Add(gas, params.CallNewAccountGas)
		//	}
		//}

		//if len(stack.data[stack.len()-3].Bytes()) > 0 {
		//	gas.Add(gas, params.CallValueTransferGas)
		//}

		x := calcMemSize(stack.data[stack.len()-6], stack.data[stack.len()-7])
		y := calcMemSize(stack.data[stack.len()-4], stack.data[stack.len()-5])

		newMemSize = common.BigMax(x, y)
	case DELEGATECALL:
		//gas.Add(gas, stack.data[stack.len()-1])

		x := calcMemSize(stack.data[stack.len()-5], stack.data[stack.len()-6])
		y := calcMemSize(stack.data[stack.len()-3], stack.data[stack.len()-4])

		newMemSize = common.BigMax(x, y)
	}
	quadMemGas(mem, newMemSize)

	return newMemSize, big.NewInt(1000), nil
}

// RunPrecompile runs and evaluate the output of a precompiled contract defined in contracts.go
func (evm *EVM) RunPrecompiled(p *PrecompiledAccount, input []byte, contract *Contract) (ret []byte, err error) {
	gas := p.Gas(len(input))
	if contract.UseGas(gas) {
		ret = p.Call(input)
		return ret, nil
	} else {
		return nil, OutOfGasError
	}
}

func (evm *EVM) Finalize() {
	if evm.cfg.Debug {
		fmt.Fprintf(os.Stdout, "[[   Dirty Accounts %08d:   ]]\n", len(evm.logger.changedValues))
		for addr, entries := range evm.logger.changedValues {
			fmt.Fprintf(os.Stdout, "### address %s ###\n", addr.Hex())
			for key, value := range entries {
				fmt.Fprintf(os.Stdout, "%s => %s\n", key.Hex(), value.Hex())
			}
			fmt.Fprint(os.Stdout, "### done ###\n")
		}
		evm.env.DumpStructLog()
	}
}
