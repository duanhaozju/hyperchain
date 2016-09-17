package core

import (
"hyperchain/core/types"
"hyperchain/core/vm"
glog "github.com/op/go-logging"
"hyperchain/common"
"math/big"
"hyperchain/hyperdb"
"hyperchain/core/state"
"hyperchain/core/vm/params"
)
type Code []byte
var logger = glog.Logger{}
var(
	//TODO set the vm.config
	db,err = hyperdb.GetLDBDatabase()
	statedb,_  = state.New(db)
	env = make(map[string]string)
	vmenv = (*Env)(nil)
)

func init(){
	env["currentNumber"] = "1"
	env["currentGasLimit"] = "10000000"
	vm.Precompiled = make(map[string]*vm.PrecompiledAccount)
	vmenv = NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)
}

// TODO 1 we don't have gas in tx when I program this func,but it should be add
// TODO 2 consider use a snapshot, so we can easily to recovery
//func ExecBlock(block types.Block,db,hashfucn)(err error){
// 得到虚拟机VM
func ExecBlock(block *types.Block)(err error){
	if(err != nil || env == nil){
		return err
	}
	ExecTransaction(*types.NewTestCreateTransaction())

	//for _,tx := range block.Transactions{
	for _,_ = range block.Transactions{
		_,err = ExecTransaction(*types.NewTestCallTransaction())
		//_,err = ExecTransaction(*tx)
	}
	log.Notice("the sum of transactions is ",len(block.Transactions))
	log.Notice("the sum of accounts is :",len(vmenv.State().GetAccounts()))
	log.Notice("---------------------------------------------------------")
	for _,v := range vmenv.State().GetAccounts(){
		log.Notice("##################################################")
		v.ForEachStorage(func(key, value common.Hash) (bool) {
			log.Notice("the key is ",key,"       the value is ",value)
			return true
		})
		log.Notice("##################################################")
	}
	log.Notice("---------------------------------------------------------")

	//start := time.Now()
	//log.Error("we cost ")
	return
}

// 这一块相当于ethereum里的TransitionDB
func ExecTransaction(tx types.Transaction)(ret []byte,err error) {
	var(
		from = common.BytesToAddress(tx.From)
		//sender = common.BytesToAddress(tx.From)
		to = common.BytesToAddress(tx.To)
		// TODO these there parameters should be added into the tx
		data = tx.Payload()
		gas = tx.Gas()
		gasPrice = tx.GasPrice()
		amount = tx.Amount()
	)
	log.Notice("the to is ---------",to)
	log.Notice("the to is ---------",tx.To)
	if(tx.To == nil){
		return Exec(&from,nil,data,gas,gasPrice,amount)
	}

	return Exec(&from,&to,data,gas,gasPrice,amount)
}

func ExecSourceCode(from, to *common.Address, data []byte, gas,
gasPrice, value *big.Int)(ret []byte,err error){

	sender := vmenv.Db().GetAccount(*from)
	contractCreation := (nil == to)
	//ret,err = env.Call(sender,*to,data,gas,gasPrice,value)
	// 判断是否能够交易,转移,这一步可以考虑在外部执行

	if contractCreation{
		//logger.Notice("------create contract")
		ret,_,err = vmenv.Create(sender,data,gas,gasPrice,value)
		if err != nil{
			ret = nil
			logger.Error("VM create err:",err)
		}
	} else {
		ret,err = vmenv.Call(sender,*to,data,gas,gasPrice,value)
		if err != nil{
			logger.Error("VM call err:",err)
		}
	}
	return ret,err
}

func Exec(from, to *common.Address, data []byte, gas,
gasPrice, value *big.Int)(ret []byte,err error){

	sender := vmenv.Db().GetAccount(*from)
	contractCreation := (nil == to)
	//ret,err = env.Call(sender,*to,data,gas,gasPrice,value)
	// 判断是否能够交易,转移,这一步可以考虑在外部执行

	if contractCreation{
		//logger.Notice("------create contract")
		ret,_,err = vmenv.Create(sender,data,gas,gasPrice,value)
		if err != nil{
			ret = nil
			logger.Error("VM create err:",err)
		}
	} else {
		ret,err = vmenv.Call(sender,*to,data,gas,gasPrice,value)
		if err != nil{
			logger.Error("VM call err:",err)
		}
	}
	return ret,err
}

func CommitStatedbToBlockchain(){
	vmenv.State().Commit()
}

func SetVMEnv(new_env *Env)  {
	vmenv = new_env
}

func GetVMEnv()  *Env{
	return vmenv
}