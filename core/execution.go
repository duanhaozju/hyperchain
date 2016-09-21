package core

import (
	"math/big"
	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/core/crypto"
	"hyperchain/core/vm/params"
	//"hyperchain/core/vm/compiler"
	"fmt"
)

// Call executes within the given contract
func Call(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	//fmt.Println("Call")
	ret, _, err = exec(env, caller, &addr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value)
	return ret, err
}

// CallCode executes the given address' code as the given contract address
func CallCode(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int) (ret []byte, err error) {
	fmt.Println("CallCode")

	callerAddr := caller.Address()
	ret, _, err = exec(env, caller, &callerAddr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value)
	return ret, err
}

// DelegateCall is equivalent to CallCode except that sender and value propagates from parent scope to child scope
func DelegateCall(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice *big.Int) (ret []byte, err error) {
	fmt.Println("DelegateCall")

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
	if err != nil {
		log.Error("it is err",err)
		return nil, address, err
	}
	return ret, address, err
}

func exec(env vm.Environment, caller vm.ContractRef, toAddress, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	evm := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	// 深度检查代码,如果超过limit则只消耗gas
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, vm.DepthError
	}

	var createAccount bool
	if toAddress == nil {
		// Create a new account on the state
		nonce := env.Db().GetNonce(caller.Address())
		env.Db().SetNonce(caller.Address(), nonce+1)
		addr = crypto.CreateAddress(caller.Address(), nonce)
		toAddress = &addr
		createAccount = true
	}

	snapshotPreTransfer := env.MakeSnapshot()
	var (
		from = env.Db().GetAccount(caller.Address())
		to   vm.Account
	)
	if createAccount {
		to = env.Db().CreateAccount(*toAddress)
	} else {
		if !env.Db().Exist(*toAddress) {
			to = env.Db().CreateAccount(*toAddress)
		} else {
			to = env.Db().GetAccount(*toAddress)
		}
	}
	env.Transfer(from, to, value)

	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	contract := vm.NewContract(caller, to, value, gas, gasPrice)
	if createAccount{
		//abis,bins,err := compiler.CompileSourcefile(string(code))
		// TODO this is only for one contract
		//contract.SetABI(common.FromHex(abis[0]))
		//contract.SetCallCode(codeAddr,common.FromHex(bins[0]))
		//if err != nil{
		//	return nil,common.Address{},err
		//}
		//log.Info("abis:",abis)
		//log.Info("bins:",bins)
		//log.Info("to:",to.Address())
		//log.Info("--------------")
		// todo finish the method
		//vmenv.state.SetLeastAccount(&to)
		contract.SetCallCode(codeAddr,code)
	}else {
		contract.SetCallCode(codeAddr, code)
	}
	defer contract.Finalise()

	// very important 执行contract和input
	ret, err = evm.Run(contract, input)
	//log.Info("--------------------------------------")
	//log.Info("ret:",ret)
	//log.Info("caller.address:",caller.Address())
	//log.Info("address:",toAddress)
	//log.Info("codeaddress:",codeAddr)
	//log.Info("input:",input)
	//log.Info("code:",code)
	if(err!=nil){
		log.Info("there is a error")
	}
	//log.Info("--------------------------------------")
	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil && createAccount {
		dataGas := big.NewInt(int64(len(ret)))
		dataGas.Mul(dataGas, params.CreateDataGas)
		if contract.UseGas(dataGas) {
			env.Db().SetCode(*toAddress, ret)
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

	return ret, addr, err
}

func execDelegateCall(env vm.Environment, caller vm.ContractRef, originAddr, toAddr, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	log.Error("execDelegateCall")
	evm := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		//
		//caller.ReturnGas(gas, gasPrice)
		//return nil, common.Address{}, vm.DepthError
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
