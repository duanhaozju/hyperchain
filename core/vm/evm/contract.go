//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package evm

import (
	"math/big"

	"hyperchain/common"
	"hyperchain/core/vm"
)

// Contract represents an ethereum contract in the state database. It contains
// the the contract code, calling arguments. Contract implements ContractRef
type Contract struct {
	// CallerAddress is the result of the caller which initialised this
	// contract. However when the "call method" is delegated this value
	// needs to be initialised to that of the caller's caller.
	CallerAddress common.Address
	caller        vm.ContractRef
	self          vm.ContractRef
	env           vm.Environment

	jumpdests destinations // result of JUMPDEST analysis.

	Code     []byte
	Input    []byte
	CodeAddr *common.Address

	value, Gas, UsedGas, Price *big.Int

	Args []byte

	DelegateCall bool

	Opcode int32
}

// NewContract returns a new contract environment for the execution of EVM.
func NewContract(caller vm.ContractRef, object vm.ContractRef, value, gas, price *big.Int, opcode int32, env vm.Environment) vm.VmContext {
	c := &Contract{CallerAddress: caller.Address(), caller: caller, self: object, Args: nil, Opcode: opcode, env: env}

	if parent, ok := caller.(*Contract); ok {
		// Reuse JUMPDEST analysis from parent context if available.
		c.jumpdests = parent.jumpdests
	} else {
		c.jumpdests = make(destinations)
	}

	// Gas should be a pointer so it can safely be reduced through the run
	// This pointer will be off the state transition
	c.Gas = gas //new(big.Int).Set(gas)
	c.value = new(big.Int).Set(value)
	// In most cases price and value are pointers to transaction objects
	// and we don't want the transaction's values to change.
	c.Price = new(big.Int).Set(price)
	c.UsedGas = new(big.Int)

	return c
}

// AsDelegate sets the contract to be a delegate call and returns the current
// contract (for chaining calls)
func (c *Contract) AsDelegate() vm.VmContext {
	c.DelegateCall = true
	// NOTE: caller must, at all times be a contract. It should never happen
	// that caller is something other than a Contract.
	c.CallerAddress = c.caller.(*Contract).CallerAddress
	return c
}

// GetOp returns the n'th element in the contract's byte array
func (c *Contract) GetOp(n uint64) byte {
	return c.GetByte(n)
}

// GetByte returns the n'th byte in the contract's byte array
func (c *Contract) GetByte(n uint64) byte {
	if n < uint64(len(c.Code)) {
		return c.Code[n]
	}

	return 0
}

// Caller returns the caller of the contract.
//
// Caller will recursively call caller when the contract is a delegate
// call, including that of caller's caller.
func (c *Contract) Caller() common.Address {
	return c.CallerAddress
}

// Finalise finalises the contract and returning any remaining gas to the original
// caller.
func (c *Contract) Finalise() {
	// Return the remaining gas to the caller
	c.caller.ReturnGas(c.Gas, c.Price)
}

// UseGas attempts the use gas and subtracts it and returns true on success
func (c *Contract) UseGas(gas *big.Int) (ok bool) {
	ok = useGas(c.Gas, gas)
	if ok {
		c.UsedGas.Add(c.UsedGas, gas)
	}
	return
}

// ReturnGas adds the given gas back to itself.
func (c *Contract) ReturnGas(gas, price *big.Int) {
	// Return the gas to the context
	c.Gas.Add(c.Gas, gas)
	c.UsedGas.Sub(c.UsedGas, gas)
}

// Address returns the contracts address
func (c *Contract) Address() common.Address {
	return c.self.Address()
}

// Value returns the contracts value (sent to it from it's caller)
func (c *Contract) Value() *big.Int {
	return c.value
}

// SetCode sets the code to the contract
func (self *Contract) SetCode(hash common.Hash, code []byte) {
	self.Code = code
}

// SetCallCode sets the code of the contract and address of the backing data
// object
func (self *Contract) SetCallCode(addr *common.Address, code []byte) {
	self.Code = code
	self.CodeAddr = addr
}

// EachStorage iterates the contract's storage and calls a method for every key
// value pair.
func (self *Contract) ForEachStorage(cb func(key common.Hash, value []byte) bool) map[common.Hash][]byte {
	return self.caller.ForEachStorage(cb)
}

func (self *Contract) GetOpCode() int32 {
	return self.Opcode
}

func (self *Contract) GetGas() *big.Int {
	return self.Gas
}

func (self *Contract) GetCodeAddr() *common.Address {
	return self.CodeAddr
}

func (self *Contract) GetCode() []byte {
	return self.Code
}

func (self *Contract) GetCaller() vm.ContractRef {
	return self.caller
}

func (self *Contract) GetJumpdests() interface{} {
	return self.jumpdests
}

func (self *Contract) SetInput(input []byte) {
	self.Input = input
}

func (self *Contract) GetInput() []byte {
	return self.Input
}

func (self *Contract) GetPrice() *big.Int {
	return self.Price
}

func (self *Contract) GetEnv() vm.Environment {
	return self.env
}

func (self *Contract) IsCreation() bool {
	return false
}

func (self *Contract) GetCodePath() string {
	return ""
}


func (self *Contract) GetAttribute(t int, k interface{}) interface{} {
	switch t {
	case vm.CtxAttr_Op:
		tmp := k.(uint64)
		return self.GetOp(tmp)
	case vm.CtxAttr_Opcode:
		return self.GetOpCode()
	case vm.CtxAttr_Caller:
		return self.GetCaller()
	case vm.CtxAttr_CallerAddr:
		return self.Caller()
	case vm.CtxAttr_CodeAddr:
		return self.GetCodeAddr()
	case vm.CtxAttr_Code:
		return self.GetCode()
	case vm.CtxAttr_Env:
		return self.GetEnv()
	case vm.CtxAttr_Gas:
		return self.GetGas()
	case vm.CtxAttr_GasPrice:
		return self.GetPrice()
	case vm.CtxAttr_Input:
		return self.GetInput()
	default:
		return nil
	}
}

func (self *Contract) SetAttribute(t int, args ...interface{}) {
	switch t {
	case vm.CtxAttr_Input:
		if len(args) != 1 {
			return
		}
		if tmp, ok := args[0].([]byte); ok != false {
			self.SetInput(tmp)
		}
	case vm.CtxAttr_Code:
		if len(args) != 2 {
			return
		}
		addr, ok := args[0].(*common.Address)
		if ok == false {
			return
		}
		code, ok := args[1].([]byte)
		if ok == false {
			return
		}
		self.SetCallCode(addr, code)
	default:
		return
	}
}