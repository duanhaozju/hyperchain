package api

import (
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core"
	"hyperchain/hyperdb"
	glog "github.com/op/go-logging"
)
type Code []byte
var logger = glog.Logger{}
var(
	leveldb,err = hyperdb.GetLDBDatabase()
	env = core.NewEnv(leveldb,new(vm.Config))
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
		_,err = ExecTransaction(env,tx)
	}
	return
}

// 这一块相当于ethereum里的TransitionDB
func ExecTransaction(env vm.Environment,tx types.Transaction)(ret []byte,err error) {
	contractCreation := (tx.To == nil)

	var(
		sender = env.Db().GetAccount(tx.From)
		data = ([]byte)(nil)
		// TODO these there parameters should be added into the tx
		gas = tx.Gas()
		gasPrice = tx.GasPrice()
		value = tx.Amount()
	)

	// 判断是否能够交易,转移,这一步可以考虑在外部执行

	if contractCreation{
		ret,_,err = env.Create(sender,data,gas,gasPrice,value)
		if err != nil{
			ret = nil
			logger.Error("VM create err:",err)
		}
	} else {
		ret,err = env.Call(sender,tx.To,data,gas,gasPrice,value)
		if err != nil{
			logger.Error("VM call err:",err)
		}
	}
	return ret,err
}
