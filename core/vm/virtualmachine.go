package vm

import (
	"hyperchain/common"
	"math/big"
)


// Vm is the basic interface for an implementation of the virtual machine.
type Vm interface {
	// Run should execute the given contract with the input given in in
	// and return the contract execution return bytes or an error if it
	// failed.
	Run(c VmContext, in []byte) ([]byte, error)
}

const (
	CtxAttr_Op = iota
	CtxAttr_Opcode
	CtxAttr_CallerAddr
	CtxAttr_Caller
	CtxAttr_CodeAddr
	CtxAttr_Code
	CtxAttr_Env
	CtxAttr_Gas
	CtxAttr_GasPrice
	CtxAttr_Input
)
type VmContext interface {
	ContractRef
	AsDelegate() VmContext

	GetOp(uint64) byte
	GetOpCode() int32

	Caller() common.Address
	GetCaller() ContractRef

	UseGas(*big.Int) bool
	GetGas() *big.Int
	GetPrice() *big.Int

	GetCode() []byte
	GetCodeHash() common.Hash
	SetCallCode(*common.Address, []byte)

	SetInput([]byte)
	GetInput() []byte

	Finalise()
	GetCodeAddr() *common.Address
	GetJumpdests() interface{}
	GetEnv()   Environment

	IsCreation() bool
	GetCodePath() string
	// GetAttribute(int, interface{}) interface{}
	// SetAttribute(int, ...interface{})
}
