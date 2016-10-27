package core

import (
	"hyperchain/common"
	"hyperchain/core/crypto"
	"hyperchain/core/vm"
	"hyperchain/core/vm/params"
	"math/big"
	//"hyperchain/core/vm/compiler"
	"hyperchain/core/state"
	"time"
)

// Call executes within the given contract
func Call(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	//fmt.Println("call")
	ret, _, err = exec(env, caller, &addr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value)
	return ret, err
}

// CallCode executes the given address' code as the given contract address
func CallCode(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	//fmt.Println("callcode")

	callerAddr := caller.Address()
	ret, _, err = exec(env, caller, &callerAddr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value)
	return ret, err
}

// DelegateCall is equivalent to CallCode except that sender and value propagates from parent scope to child scope
func DelegateCall(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice *big.Int) (ret []byte, err error) {
	//fmt.Println("DelegateCall")

	callerAddr := caller.Address()
	originAddr := env.Origin()
	callerValue := caller.Value()
	ret, _, err = execDelegateCall(env, caller, &originAddr, &callerAddr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, callerValue)
	return ret, err
}

// Create creates a new contract with the given code
func Create(env vm.Environment, caller vm.ContractRef, code []byte, gas, gasPrice, value *big.Int) (ret []byte, address common.Address, err error) {
	//fmt.Println("Create")
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
	another_test_time := time.Now()
	evm := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)

		return nil, common.Address{}, vm.DepthError
	}

	// TODO
	if !env.CanTransfer(caller.Address(), value) {
		//caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, ValueTransferErr("insufficient funds to transfer value. Req %v, has %v", value, env.Db().GetBalance(caller.Address()))
	}

	var createAccount bool
	if address == nil {
		// Create a new account on the state
		nonce := env.Db().GetNonce(caller.Address())
		env.Db().SetNonce(caller.Address(), nonce+1)
		addr = crypto.CreateAddress(caller.Address(), nonce)
		address = &addr
		createAccount = true
	}
	statedb, _ := env.Db().(*state.StateDB)
	snapshotPreTransfer := env.MakeSnapshot()
	var (
		from = env.Db().GetAccount(caller.Address())
		to   vm.Account
	)
	if createAccount {
		to = env.Db().CreateAccount(*address)
	} else {
		if !env.Db().Exist(*address) {
			to = env.Db().CreateAccount(*address)
			env.Transfer(from, to, value)
		} else {

			to = env.Db().GetAccount(*address)
			if statedb.GetCode(to.Address()) == nil {

				env.Transfer(from, to, value)
			}
		}
	}
	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := vm.NewContract(caller, to, value, gas, gasPrice)
	contract.SetCallCode(codeAddr, code)
	defer contract.Finalise()
	log.Noticef("another test time cost is",time.Since(another_test_time))

	evm_run_begin_time := time.Now()
	ret, err = evm.Run(contract, input)
	log.Noticef("evm_run_begin_time is ",time.Since(evm_run_begin_time))
	/*
		fmt.Println("---------------------------------------")
		fmt.Println("caller.address",caller.Address())
		fmt.Println("address",address)
		fmt.Println("codeaddress",codeAddr)
		fmt.Println("input",input)
		fmt.Println("code",code)
		fmt.Println("---------------------------------------")
		fmt.Println("ret",ret)
	*/

	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	test2_time := time.Now()
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
	if err != nil && (env.RuleSet().IsHomestead(env.BlockNumber()) || err != vm.CodeStoreOutOfGasError) {
		contract.UseGas(contract.Gas)

		env.SetSnapshot(snapshotPreTransfer)
	}
	log.Noticef("test2_time is ",time.Since(test2_time))  // the time is less than 1us
	return ret, addr, err
}

func execDelegateCall(env vm.Environment, caller vm.ContractRef, originAddr, toAddr, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	//fmt.Println("execDelegateCall")
	evm := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, vm.DepthError
	}

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
