package core

import (
	glog "github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm"

	"math/big"
	"time"
)

type Code []byte

var logger = glog.Logger{}


// 这一块相当于ethereum里的TransitionDB
func ExecTransaction(tx types.Transaction, env vm.Environment) (receipt *types.Receipt, ret []byte, addr common.Address, err error) {
	exec_start_time := time.Now()
	var (
		from = common.BytesToAddress(tx.From)
		//sender = common.BytesToAddress(tx.From)
		to = common.BytesToAddress(tx.To)
		// TODO these there parameters should be added into the tx
		data       = tx.Payload()
		gas        = tx.Gas()
		gasPrice   = tx.GasPrice()
		amount     = tx.Amount()
		statedb, _ = env.Db().(*state.StateDB)
	)
	// TODO ZHZ_TEST the time of above will cost 10us

	exec_begin_time := time.Now()
	// TODO ZHZ_TEST this will cost 200us
	if tx.To == nil {
		ret, addr, err = Exec(env, &from, nil, data, gas, gasPrice, amount)
		log.Noticef("the time of Exec in create contract",time.Since(exec_begin_time))
	} else {
		ret, _, err = Exec(env, &from, &to, data, gas, gasPrice, amount)
		log.Noticef("the time of Exec in Call contract",time.Since(exec_begin_time))
	}


	// TODO ZHZ_TEST the time of below will cost 10us
	receipt = types.NewReceipt(nil, gas)
	receipt.ContractAddress = addr.Bytes()
	//todo add tx hash in tx struct
	receipt.TxHash = tx.GetTransactionHash().Bytes()
	// todo replace the gasused
	receipt.GasUsed = 100000
	receipt.Ret = ret
	receipt.SetLogs(statedb.GetLogs(common.BytesToHash(receipt.TxHash)))

	if err != nil && IsValueTransferErr(err) {
		receipt.Status = types.Receipt_OUTOFBALANCE
		receipt.Message = []byte(err.Error())
	} else {
		receipt.Status = types.Receipt_SUCCESS
		receipt.Message = nil
	}
	log.Noticef("the time of ExecTransaction in contract",time.Since(exec_start_time))
	return receipt, ret, addr, err
}

func Exec(vmenv vm.Environment, from, to *common.Address, data []byte, gas,
	gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {
	var sender vm.Account

	if !(vmenv.Db().Exist(*from)) {
		//createAccount_time := time.Now()
		sender = vmenv.Db().CreateAccount(*from)
		//log.Notice("createAccount_time is",time.Since(createAccount_time))
	} else {
		//getAccount_time := time.Now()
		sender = vmenv.Db().GetAccount(*from)
		//log.Notice("getAccount_time is",time.Since(getAccount_time))
	}
	contractCreation := (nil == to)

	//ret,err = env.Call(sender,*to,data,gas,gasPrice,value)
	// 判断是否能够交易,转移,这一步可以考虑在外部执行
	if contractCreation {
		log.Debug("------create contract")
		ret, addr, err = vmenv.Create(sender, data, gas, gasPrice, value)
		if err != nil {
			ret = nil
			log.Error("VM create err:", err)
		}
	} else {
		log.Debug("------call contract")
		ret, err = vmenv.Call(sender, *to, data, gas, gasPrice, value)
		if err != nil {
			log.Error("VM call err:", err)
		}
	}
	// todo replace the gasused
	// todo just for test

	return ret, addr, err
}

