package core

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/core/vm/params"
)

// Call executes within the given contract
func Call(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	ret, _, err = exec(env, caller, &addr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value)
	return ret, err
}

// CallCode executes the given address' code as the given contract address
func CallCode(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	callerAddr := caller.Address()
	ret, _, err = exec(env, caller, &callerAddr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value)
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
	ret, address, err = exec(env, caller, nil, nil, nil, code, gas, gasPrice, value)
	// Here we get an error if we run into maximum stack depth,
	// See: https://github.com/ethereum/yellowpaper/pull/131
	// and YP definitions for CREATE instruction
	if err != nil {
		return nil, address, err
	}
	return ret, address, err
}

func exec(env vm.Environment, caller vm.ContractRef, address, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	// gas and gasPrice should be set after

	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	// 深度检查代码,如果超过limit则只消耗gas
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)

		return nil, common.Address{}, vm.DepthError
	}
	// 判断是否能够交易,转移
	if !env.CanTransfer(caller.Address(), value) {
		caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, nil
	}

	var (
		evm = env.Vm()
		statedb = env.Db()
		input []byte        // input 应该存在tx中
		fromAddress = caller.Address()
		toAddress = *address
		fromAccount = statedb.GetAccount(fromAddress)
		toAccount vm.Account
		createAccount bool
	)

	if toAddress == nil {
		// create a new addr of to save contract
		// TODO 设置db的nonce,生成新的address
		toAddress = ([]byte)(nil)
		toAccount = statedb.CreateAccount(toAddress)
		createAccount = true
	}else {
		if !statedb.Exist(toAddress) {
			toAccount = statedb.CreateAccount(toAddress)
		} else {
			toAccount = statedb.GetAccount(toAddress)
		}
		createAccount = false
	}
	env.Transfer(fromAccount,toAccount,value)
	// get a snapshot to recovery
	snapshotPreTransfer := env.MakeSnapshot()

	contract := vm.NewContract(fromAccount, toAccount, value, nil, nil)
	contract.SetCallCode(codeAddr, statedb.GetCode(toAddress))
	defer contract.Finalise()

	ret, err = evm.Run(contract, input)

	if err == nil && createAccount {
		dataGas := big.NewInt(int64(len(ret)))
		dataGas.Mul(dataGas, params.CreateDataGas)
		if contract.UseGas(dataGas) {
			env.Db().SetCode(*address, ret)
		} else {
			err = vm.CodeStoreOutOfGasError
		}
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	// 当一个error被EVM返回时,或者设置创建代码在我们回复到快照时,消耗剩余gas。
	// 此外,当我们处于homestead,这个也会计算code storage gas errors
	if err != nil && (env.RuleSet().IsHomestead(env.BlockNumber()) || err != vm.CodeStoreOutOfGasError) {
		contract.UseGas(contract.Gas)
		env.SetSnapshot(snapshotPreTransfer)
	}

	return ret,toAddress,err
}

func execDelegateCall(env vm.Environment, caller vm.ContractRef, originAddr, toAddr, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	evm := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	//if env.Depth() > int(params.CallCreateDepth.Int64()) {
	//	caller.ReturnGas(gas, gasPrice)
	//	return nil, common.Address{}, vm.DepthError
	//}

	snapshot := env.MakeSnapshot()

	var to vm.Account
	if !env.Db().Exist(*toAddr) {
		to = env.Db().CreateAccount(*toAddr)
	} else {
		to = env.Db().GetAccount(*toAddr)
	}

	// Iinitialise a new contract and make initialise the delegate values
	contract := vm.NewContract(caller, to, value, gas, gasPrice).AsDelegate()
	contract.SetCallCode(codeAddr, code)
	defer contract.Finalise()

	ret, err = evm.Run(contract, input)
	if err != nil {
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
