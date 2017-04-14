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

type VmContext interface {
	AsDelegate() VmContext
	GetOp(uint64) byte
	GetByte(uint64) byte
	Caller() common.Address
	Finalise()
	UseGas(*big.Int) bool
	ReturnGas(*big.Int, *big.Int)
	Address() common.Address
	Value() *big.Int
	SetCode(common.Hash, []byte)
	SetABI([]byte)
	SetCallCode(*common.Address, []byte)
	ForEachStorage(func(common.Hash, []byte) bool) map[common.Hash][]byte
	GetOpCode() int32
	GetGas() *big.Int
	GetCodeAddr() *common.Address
	GetCode() []byte
	GetCaller() ContractRef
	GetJumpdests() interface{}
	SetInput([]byte)
	GetInput() []byte
	GetPrice() *big.Int
}
