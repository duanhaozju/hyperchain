// Copyright 2015 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package evm

import (
	"hyperchain/common"
	er "hyperchain/core/errors"
	"hyperchain/core/ledger/state"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/evm/params"
	"hyperchain/crypto"
	"math/big"
)

// Call executes within the given contract
func Call(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int, op types.TransactionValue_Opcode) (ret []byte, err error) {
	ret, _, err = exec(env, caller, &addr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value, op)
	return ret, err
}

// CallCode executes the given address' code as the given contract address
func CallCode(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	callerAddr := caller.Address()
	ret, _, err = exec(env, caller, &callerAddr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value, 0)
	return ret, err
}

// DelegateCall is equivalent to CallCode except that sender and value propagates from parent scope to child scope
func DelegateCall(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice *big.Int) (ret []byte, err error) {
	callerAddr := caller.Address()
	originAddr := env.Origin()
	callerValue := caller.Value()
	ret, _, err = execDelegateCall(env, caller, &originAddr, &callerAddr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, callerValue)
	return ret, err
}

// Create creates a new contract with the given code
func Create(env vm.Environment, caller vm.ContractRef, code []byte, gas, gasPrice, value *big.Int) (ret []byte, address common.Address, err error) {
	ret, address, err = exec(env, caller, nil, nil, nil, code, gas, gasPrice, value, 0)
	if err != nil {
		return nil, address, err
	}
	return ret, address, err
}

func exec(env vm.Environment, caller vm.ContractRef, address, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int, op types.TransactionValue_Opcode) (ret []byte, addr common.Address, err error) {
	var (
		createAccount bool
		from          = env.Db().GetAccount(caller.Address())
		to            vm.Account
	)
	virtualMachine := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	snapshotPreTransfer := env.MakeSnapshot()
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, er.ExecContractErr(1, "Max call depth exceeded 1024")
	}

	if env.Db().GetBalance(caller.Address()).Cmp(value) < 0 {
		return nil, common.Address{}, er.ValueTransferErr("insufficient funds to transfer value. Req %v, has %v", value, env.Db().GetBalance(caller.Address()))
	}

	// create address
	if address == nil {
		// Create a new account on the state
		// Increase account nonce if a new account has been created.
		nonce := env.Db().GetNonce(caller.Address())
		env.Db().SetNonce(caller.Address(), nonce+1)
		addr = crypto.CreateAddress(caller.Address(), nonce)
		address = &addr
		createAccount = true
	}
	if createAccount {
		to = env.Db().CreateAccount(*address)
	} else {
		if !env.Db().Exist(*address) {
			to = env.Db().CreateAccount(*address)
		} else {
			to = env.Db().GetAccount(*address)
		}
	}
	// Account status management.
	// Short circuit no matter success or not, don't bother vm execution.
	if isFreeze(op) || isUnFreeze(op) {
		if err := manageAccount(env, op, to); err != nil {
			env.SetSnapshot(snapshotPreTransfer)
			return nil, common.Address{}, err
		} else {
			// No error occurs, break the execution frame immediately.
			return nil, common.Address{}, nil
		}
	}
	Transfer(from, to, value)

	/*
		RUN VM
	*/
	if env.Db().GetStatus(to.Address()) != state.OBJ_NORMAL {
		env.Logger().Debugf("account %s has been frozen", to.Address().Hex())
		env.SetSnapshot(snapshotPreTransfer)
		return nil, common.Address{}, er.ExecContractErr(1, "Try to invoke a frozen contract")
	}
	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := NewContract(caller, to, value, gas, gasPrice, int32(op))
	if isUpdate(op) {
		// using the new code to execute
		// otherwise errors could occur
		contract.SetCallCode(codeAddr, input)
	} else {
		// using the origin code to execute
		contract.SetCallCode(codeAddr, code)
	}
	defer contract.Finalise()

	// evm finalise
	ret, err = virtualMachine.Run(contract, input)

	if err == nil && (createAccount || isUpdate(op)) {
		dataGas := big.NewInt(int64(len(ret)))
		dataGas.Mul(dataGas, params.CreateDataGas)
		switch {
		case createAccount:
			if contract.UseGas(dataGas) {
				env.Db().SetCode(*address, ret)
				env.Db().AddDeployedContract(caller.Address(), *address)
				env.Db().SetCreator(*address, caller.Address())
				env.Db().SetCreateTime(*address, env.BlockNumber().Uint64())
			} else {
				err = CodeStoreOutOfGasError
			}
		case isUpdate(op):
			// if code ran successfully and no errors were returned
			// and this transaction is a update code operation
			// replace contract code with given one
			// undo all changes during the vm execution(construct function)
			env.SetSnapshot(snapshotPreTransfer)
			if contract.UseGas(dataGas) {
				env.Db().SetCode(*address, ret)
			} else {
				err = CodeStoreOutOfGasError
			}
		}
	}
	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		contract.UseGas(contract.Gas)
		env.SetSnapshot(snapshotPreTransfer)
		if createAccount {
			err = er.ExecContractErr(0, "contract creation failed, error msg", err.Error())
		} else {
			err = er.ExecContractErr(1, "contract invocation failed, error msg:", err.Error())
		}
	}
	return ret, addr, err
}

func execDelegateCall(env vm.Environment, caller vm.ContractRef, originAddr, toAddr, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	virtualMachine := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, DepthError
	}

	snapshot := env.MakeSnapshot()

	var to vm.Account
	if !env.Db().Exist(*toAddr) {
		to = env.Db().CreateAccount(*toAddr)
	} else {
		to = env.Db().GetAccount(*toAddr)
	}

	// Iinitialise a new contract and make initialise the delegate values
	contract := NewContract(caller, to, value, gas, gasPrice, 0).AsDelegate()
	contract.SetCallCode(codeAddr, code)
	defer contract.Finalise()

	ret, err = virtualMachine.Run(contract, input)
	if err != nil {
		// use all gas left in caller
		contract.UseGas(contract.Gas)

		env.SetSnapshot(snapshot)
	}

	return ret, addr, err
}

// generic transfer method
func Transfer(from, to vm.Account, amount *big.Int) {
	from.SubBalance(amount)
	to.AddBalance(amount)
}

// isUpdate judges the opcode is contract updation op.
func isUpdate(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_UPDATE
}

// isFreeze judges the opcode is contract freeze op.
func isFreeze(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_FREEZE
}

// isUnfreeze judges the opcode is contract unfreeze op.
func isUnFreeze(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_UNFREEZE
}

func isSpecialOperation(op types.TransactionValue_Opcode) bool {
	return op == types.TransactionValue_UPDATE || op == types.TransactionValue_FREEZE || op == types.TransactionValue_UNFREEZE
}

// manageAccount freeze or unfreeze a account with given op.
func manageAccount(env vm.Environment, op types.TransactionValue_Opcode, to vm.Account) error {
	switch {
	case isFreeze(op):
		if env.Db().GetStatus(to.Address()) == state.OBJ_FROZON {
			env.Logger().Warningf("try to freeze a frozen account %s", to.Address().Hex())
			return er.ExecContractErr(1, "duplicate freeze operation")
		}
		if env.Db().GetCode(to.Address()) == nil {
			env.Logger().Warningf("try to freeze a non-contract account %s", to.Address().Hex())
			return er.ExecContractErr(1, "freeze a non-contract account")
		}
		env.Logger().Debugf("freeze account %s", to.Address().Hex())
		env.Db().SetStatus(to.Address(), state.OBJ_FROZON)
	case isUnFreeze(op):
		if env.Db().GetStatus(to.Address()) == state.OBJ_NORMAL {
			env.Logger().Warningf("try to unfreeze a normal account %s", to.Address().Hex())
			return er.ExecContractErr(1, "duplicate unfreeze operation")
		}
		if env.Db().GetCode(to.Address()) == nil {
			env.Logger().Warningf("try to unfreeze a non-contract account %s", to.Address().Hex())
			return er.ExecContractErr(1, "unfreeze a non-contract account")
		}
		env.Logger().Debugf("unfreeze account %s", to.Address().Hex())
		env.Db().SetStatus(to.Address(), state.OBJ_NORMAL)
	}
	return nil
}
