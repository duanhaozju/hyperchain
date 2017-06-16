package jvm

import (
	"hyperchain/core/vm"
	"math/big"
	"hyperchain/common"
)

type Context struct {
	caller     vm.ContractRef
	callee     vm.ContractRef
	env        vm.Environment
	code       []byte
	isCreation bool
	isUpdate   bool
	codePath   string
	codeHash   common.Hash
}


func NewContext(caller vm.ContractRef, callee vm.ContractRef, env vm.Environment, isCreation bool, isUpdate bool, codePath string, codeHash common.Hash) *Context {
	return &Context{
		caller:     caller,
		callee:     callee,
		env:        env,
		isCreation: isCreation,
		isUpdate:   isUpdate,
		codePath:   codePath,
		codeHash:   codeHash,
	}
}

func (ctx *Context) ReturnGas(*big.Int, *big.Int) {

}

func (ctx *Context) Address() common.Address {
	return ctx.callee.Address()
}

func (ctx *Context) Value() *big.Int {
	return nil
}

func (ctx *Context) SetCode(hash common.Hash, code []byte) {
	ctx.code = code
}

func (ctx *Context) ForEachStorage(cb func(key common.Hash, value []byte) bool) map[common.Hash][]byte {
	return ctx.caller.ForEachStorage(cb)
}

func (ctx *Context) AsDelegate() vm.VmContext {
	return nil
}

func (ctx *Context) GetOp(uint64) byte {
	return byte(0)
}

func (ctx *Context) GetOpCode() int32 {
	return 0
}

func (ctx *Context) Caller() common.Address {
	return ctx.caller.Address()
}

func (ctx *Context) GetCaller() vm.ContractRef {
	return ctx.caller
}


func (ctx *Context) UseGas(*big.Int) bool {
	return true
}
func (ctx *Context) GetGas() *big.Int {
	return nil
}
func (ctx *Context) GetPrice() *big.Int {
	return nil
}

func (ctx *Context) GetCode() []byte {
	return ctx.code
}

func (ctx *Context) GetCodePath() string {
	return ctx.codePath
}

func (ctx *Context) GetCodeHash() common.Hash {
	return ctx.codeHash
}
func (ctx *Context) SetCallCode(*common.Address, []byte) {

}

func (ctx *Context) SetInput([]byte) {

}
func (ctx *Context) GetInput() []byte {
	return nil
}

func (ctx *Context) Finalise() {

}
func (ctx *Context) GetCodeAddr() *common.Address {
	return nil
}
func (ctx *Context) GetJumpdests() interface{} {
	return nil
}

func (ctx *Context) GetEnv() vm.Environment {
	return ctx.env
}


func (ctx *Context) IsCreation() bool {
	return ctx.isCreation
}

func (ctx *Context) IsUpdate() bool {
	return ctx.isUpdate
}
