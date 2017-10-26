package jvm

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/vm"
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

func (ctx *Context) Address() common.Address {
	return ctx.callee.Address()
}

func (ctx *Context) ForEachStorage(cb func(key common.Hash, value []byte) bool) map[common.Hash][]byte {
	return ctx.caller.ForEachStorage(cb)
}

func (ctx *Context) GetCaller() vm.ContractRef {
	return ctx.caller
}

func (ctx *Context) GetCodePath() string {
	return ctx.codePath
}

func (ctx *Context) GetCodeHash() common.Hash {
	return ctx.codeHash
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
