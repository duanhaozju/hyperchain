package jvm
import (
	"hyperchain/common"
	"hyperchain/core/crypto"
	"hyperchain/core/vm/evm/params"
	"math/big"
	"hyperchain/core/types"
	"hyperchain/core/hyperstate"
	"hyperchain/core/vm"
	er "hyperchain/core/errors"
)

// Call executes within the given contract
func Call(env vm.Environment, caller vm.ContractRef, addr common.Address, input []byte, gas, gasPrice, value *big.Int, op types.TransactionValue_Opcode) (ret []byte, err error) {
	ret, _, err = exec(env, caller, &addr, &addr, input, env.Db().GetCode(addr), gas, gasPrice, value, op)
	return ret, err
}


func exec(env vm.Environment, caller vm.ContractRef, address, codeAddr *common.Address, input, code []byte, gas, gasPrice, value *big.Int, op types.TransactionValue_Opcode) (ret []byte, addr common.Address, err error) {
	virtualMachine := env.Vm()
	// Depth check execution. Fail if we're trying to execute above the
	// limit.
	snapshotPreTransfer := env.MakeSnapshot()
	if env.Depth() > int(params.CallCreateDepth.Int64()) {
		caller.ReturnGas(gas, gasPrice)
		return nil, common.Address{}, er.ExecContractErr(1, "Max call depth exceeded 1024")
	}

	if !env.CanTransfer(caller.Address(), value) {
		return nil, common.Address{}, er.ValueTransferErr("insufficient funds to transfer value. Req %v, has %v", value, env.Db().GetBalance(caller.Address()))
	}

	var createAccount bool
	// create address
	if address == nil {
		// Create a new account on the state
		nonce := env.Db().GetNonce(caller.Address())
		env.Db().SetNonce(caller.Address(), nonce+1)
		addr = crypto.CreateAddress(caller.Address(), nonce)
		address = &addr
		createAccount = true
	}
	var (
		// from = env.Db().GetAccount(caller.Address())
		to   vm.Account
	)
	if createAccount {
		to = env.Db().CreateAccount(*address)
	} else {
		if !env.Db().Exist(*address) {
			to = env.Db().CreateAccount(*address)
		} else {
			to = env.Db().GetAccount(*address)
		}
		if isSpecialOperation(op) && !isUpdate(op) {
			switch {
			case isFreeze(op):
				if env.Db().GetStatus(to.Address()) == hyperstate.STATEOBJECT_STATUS_FROZON {
					env.Logger().Warningf("try to freeze a frozen account %s", to.Address().Hex())
					env.SetSnapshot(snapshotPreTransfer)
					return nil, common.Address{}, er.ExecContractErr(1, "duplicate freeze operation")
				}
				if env.Db().GetCode(to.Address()) == nil {
					env.Logger().Warningf("try to freeze a non-contract account %s", to.Address().Hex())
					env.SetSnapshot(snapshotPreTransfer)
					return nil, common.Address{}, er.ExecContractErr(1, "freeze a non-contract account")
				}
				env.Logger().Debugf("freeze account %s", to.Address().Hex())
				env.Db().SetStatus(to.Address(), hyperstate.STATEOBJECT_STATUS_FROZON)
			case isUnFreeze(op):
				if env.Db().GetStatus(to.Address()) == hyperstate.STATEOBJECT_STATUS_NORMAL {
					env.Logger().Warningf("try to unfreeze a normal account %s", to.Address().Hex())
					env.SetSnapshot(snapshotPreTransfer)
					return nil, common.Address{}, er.ExecContractErr(1, "duplicate unfreeze operation")
				}
				if env.Db().GetCode(to.Address()) == nil {
					env.Logger().Warningf("try to unfreeze a non-contract account %s", to.Address().Hex())
					env.SetSnapshot(snapshotPreTransfer)
					return nil, common.Address{}, er.ExecContractErr(1, "unfreeze a non-contract account")
				}
				env.Logger().Debugf("unfreeze account %s", to.Address().Hex())
				env.Db().SetStatus(to.Address(), hyperstate.STATEOBJECT_STATUS_NORMAL)
			}
			return nil, common.Address{}, nil
		}
		// env.Transfer(from, to, value)
	}
	/*
		RUN VM
	 */
	if env.Db().GetStatus(to.Address()) != hyperstate.STATEOBJECT_STATUS_NORMAL {
		env.Logger().Debugf("account %s has been frozen", to.Address().Hex())
		env.SetSnapshot(snapshotPreTransfer)
		return nil, common.Address{}, er.ExecContractErr(1, "Try to invoke a frozen contract")
	}
	// initialise a new contract and set the code that is to be used by the
	// EVM. The contract is a scoped environment for this execution context
	// only.
	context := NewContext(caller, to, env)
	env.Logger().Notice("run in jvm")
	ret, err = virtualMachine.Run(context, input)
	env.Logger().Notice("run in jvm done")
	// if the contract creation ran successfully and no errors were returned
	// calculate the gas required to store the code. If the code could not
	// be stored due to not enough gas set an error and let it be handled
	// by the error checking condition below.
	if err == nil && createAccount {
		env.Db().SetCode(*address, ret)
		env.Db().AddDeployedContract(caller.Address(), *address)
		env.Db().SetCreator(*address, caller.Address())
		env.Db().SetCreateTime(*address, env.BlockNumber().Uint64())
	}

	// When an error was returned by the EVM or when setting the creation code
	// above we revert to the snapshot and consume any gas remaining. Additionally
	// when we're in homestead this also counts for code storage gas errors.
	if err != nil {
		env.Logger().Error("execute err", err.Error())
		env.SetSnapshot(snapshotPreTransfer)
		if createAccount {
			err = er.ExecContractErr(0, "contract creation failed, error msg", err.Error())
		} else {
			err = er.ExecContractErr(1, "contract invocation failed, error msg:", err.Error())
		}
	}
	env.Logger().Notice("execute result", common.Bytes2Hex(ret))
	return ret, addr, err
}


// generic transfer method
func Transfer(from, to vm.Account, amount *big.Int) {
	from.SubBalance(amount)
	to.AddBalance(amount)
}

func isUpdate(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_UPDATE
}

func isFreeze(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_FREEZE
}

func isUnFreeze(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_UNFREEZE
}

func isNormal(opcode types.TransactionValue_Opcode) bool {
	return opcode == types.TransactionValue_NORMAL
}

func isSpecialOperation(op types.TransactionValue_Opcode) bool {
	return op == types.TransactionValue_UPDATE || op == types.TransactionValue_FREEZE || op == types.TransactionValue_UNFREEZE
}
