package core

import (
	"fmt"
	glog "github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/core/state"
	"hyperchain/core/types"
	"hyperchain/core/vm"
	"hyperchain/core/vm/params"
	"hyperchain/crypto"
	"hyperchain/hyperdb"
	"math/big"
)

type Code []byte

var logger = glog.Logger{}
var (
	//TODO set the vm.config
	db, err    = hyperdb.GetLDBDatabase()
	statedb, _ = state.New(common.Hash{}, db)
	env        = make(map[string]string)
	vmenv      = (*Env)(nil)
)

func init() {
	//vm.Precompiled = make(map[string]*vm.PrecompiledAccount)
	env["currentNumber"] = "1"
	env["currentGasLimit"] = "10000000"
	vmenv = NewEnvFromMap(RuleSet{params.MainNetHomesteadBlock, params.MainNetDAOForkBlock, true}, statedb, env)
}

// TODO 1 we don't have gas in tx when I program this func,but it should be add
// TODO 2 consider use a snapshot, so we can easily to recovery
//func ExecBlock(block types.Block,db,hashfucn)(err error){
// 得到虚拟机VM
/*
func ExecBlock(block *types.Block) (err error) {
	if err != nil || env == nil {
		return err
	}
	//ExecTransaction(*types.NewTestCreateTransaction())

	//for _,tx := range block.Transactions{
	for _, _ = range block.Transactions {
		//	_, _, _, err = ExecTransaction(*types.NewTestCallTransaction())
		//_,err = ExecTransaction(*tx)
	}
	log.Notice("the sum of transactions is ", len(block.Transactions))
	log.Notice("the sum of accounts is :", len(vmenv.State().GetAccounts()))
	log.Notice("---------------------------------------------------------")
	for _, v := range vmenv.State().GetAccounts() {
		log.Notice("##################################################")
		v.ForEachStorage(func(key, value common.Hash) bool {
			log.Notice("the key is ", key, "       the value is ", value)
			return true
		})
		log.Notice("##################################################")
	}
	log.Notice("---------------------------------------------------------")

	//start := time.Now()
	//log.Error("we cost ")
	return
}
*/

func preCheck(tx types.Transaction) error {
	// check signature
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	if !(&tx).ValidateSign(encryption, kec256Hash) {
		fmt.Println("validate signature failed")
		fmt.Println("sign", tx.Signature)
		return SignatureError("")
	}
	// check balance
	return nil
}

// 这一块相当于ethereum里的TransitionDB
func ExecTransaction(tx types.Transaction, env vm.Environment) (receipt *types.Receipt, ret []byte, addr common.Address, err error) {
	var (
		from = common.BytesToAddress(tx.From)
		//sender = common.BytesToAddress(tx.From)
		to = common.BytesToAddress(tx.To)
		// TODO these there parameters should be added into the tx
		data     = tx.Payload()
		gas      = tx.Gas()
		gasPrice = tx.GasPrice()
		amount   = tx.Amount()
	)
	if err := preCheck(tx); err != nil {
		return nil, nil, common.Address{}, nil
	}
	//log.Notice("the to is ---------",to)
	//log.Notice("the to is ---------",tx.To)
	if tx.To == nil {
		ret, addr, err = Exec(env, &from, nil, data, gas, gasPrice, amount)
		//log.Info("the exetx addr is ",addr.Bytes())
		//log.Info("the exetx hash is ",common.ToHex(tx.BuildHash().Bytes()))
		//log.Info("the exetx create ret is ",ret)
		receipt.ContractAddress = addr.Bytes()
	} else {
		ret, _, err = Exec(env, &from, &to, data, gas, gasPrice, amount)
		//log.Info("the exetx to addr is ",to.Bytes())
		//log.Infof("the exetx tx.to is %#V",tx.To)
		//log.Info("the exetx hash is ",common.ToHex(tx.BuildHash().Bytes()))
		//log.Info("the exetx call ret is ",ret)
	}
	statedb, _ := env.Db().(*state.StateDB)
	receipt = types.NewReceipt(statedb.IntermediateRoot().Bytes(), gas)
	receipt.TxHash = tx.BuildHash().Bytes()
	// todo replace the gasused
	receipt.GasUsed = 100000
	receipt.Ret = ret
	receipt.SetLogs(statedb.GetLogs(common.BytesToHash(receipt.TxHash)))
	//fmt.Println("-----------------------")
	//fmt.Println("ret",ret)
	//fmt.Println("-----------------------")
	return receipt, ret, addr, err
}

func Exec(vmenv vm.Environment, from, to *common.Address, data []byte, gas,
	gasPrice, value *big.Int) (ret []byte, addr common.Address, err error) {

	sender := vmenv.Db().GetAccount(*from)
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
	receipt := types.NewReceipt(nil, gas)
	receipt.GasUsed = 100000
	receipt.ContractAddress = addr.Bytes()
	receipt.TxHash = common.Hash{}.Bytes()
	receipt.Ret = ret
	//WriteReceipts(types.Receipts{receipt,receipt,receipt})
	//fmt.Println("receipt from db",GetReceipt(common.Hash{}))
	return ret, addr, err
}

func CommitStatedbToBlockchain() {
	GetVMEnv().State().Commit()
}

func SetVMEnv(new_env *Env) {
	vmenv = new_env
}

func GetVMEnv() *Env {
	return vmenv
}
