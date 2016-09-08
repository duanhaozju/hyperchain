package api

import (
	"hyperchain/core/types"
	"hyperchain/core/vm"
	glog "github.com/op/go-logging"
	"hyperchain/common"
	"math/big"
	"hyperchain/hyperdb"
	//"hyperchain/core/state"
)
type Code []byte
var logger = glog.Logger{}
var(
	//TODO set the vm.config
	memdb,err = hyperdb.NewMemDatabase()
	//statedb,_ := state.New(common.Hash{}, memdb)
	//env = core.NewEnv(statedb)
)
// 这个地方主要是执行交易里的代码,我们只考虑合约情况
// TODO 1 we don't have gas in tx when I program this func,but it should be add
// TODO 2 consider use a snapshot, so we can easily to recovery
//func ExecBlock(block types.Block,db,hashfucn)(err error){
// 得到虚拟机VM
func ExecBlock(block types.Block,env vm.Environment)(err error){
	if(err != nil || env == nil){
		return err
	}
	for _,tx := range block.Transactions{
		_,err = ExecTransaction(env,*tx)
	}
	return
}

// 这一块相当于ethereum里的TransitionDB
func ExecTransaction(env vm.Environment,tx types.Transaction)(ret []byte,err error) {

	var(
		sender = env.Db().GetAccount(common.BytesToAddress(tx.From))
		//sender = common.BytesToAddress(tx.From)
		to = common.BytesToAddress(tx.To)
		// TODO these there parameters should be added into the tx
		data = tx.Payload()
		gas = tx.Gas()
		gasPrice = tx.GasPrice()
		value = tx.Amount()
	)
	return Exec(env,sender,&to,data,gas,gasPrice,value)
}

func Exec(env vm.Environment,sender vm.ContractRef, to *common.Address, data []byte, gas,
	gasPrice, value *big.Int)(ret []byte,err error){

	contractCreation := (nil == to)
	//ret,err = env.Call(sender,*to,data,gas,gasPrice,value)
	// 判断是否能够交易,转移,这一步可以考虑在外部执行
	if contractCreation{
		ret,_,err = env.Create(sender,data,gas,gasPrice,value)
		if err != nil{
			ret = nil
			logger.Error("VM create err:",err)
		}
	} else {
		ret,err = env.Call(sender,*to,data,gas,gasPrice,value)
		if err != nil{
			logger.Error("VM call err:",err)
		}
	}
	return ret,err
}
