package tests

import (
	"testing"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/core/vm/api"
	"fmt"
	"hyperchain/common"
)
func Test_FormatTx(t *testing.T){
	var (
		from  = common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6")
		//input = common.FromHex("0x3ad14af300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002")
	)
	var tx_value1 = &types.TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:([]byte)(sourcecode)}
	//var tx_value2 = &types.TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:input}
	value1,err := proto.Marshal(tx_value1)
	if err != nil{
		t.Log("the test transaction has error")
	}

	if err != nil{
		t.Log("the test transaction has error")
	}
	tx_create := types.NewTransaction(from.Bytes(),nil,value1)
	api.ExecTransaction(*tx_create)

	//value2,err := proto.Marshal(tx_value2)
	//tx_call := types.NewTransaction(from.Bytes(),api.GetVMEnv().State().GetLeastAccount().Address().Bytes(),value2)
	fmt.Println("addrs----",api.GetVMEnv().State().GetLeastAccount().Address().Bytes())
	fmt.Println("addrs----",api.GetVMEnv().State().GetLeastAccount().Address().Hex())
	//api.ExecTransaction(*tx_call)
	api.ExecTransaction(*types.NewTestCallTransaction())
	api.ExecTransaction(*types.NewTestCallTransaction())

	fmt.Println(len(api.GetVMEnv().State().GetAccounts()))
	for k,v := range api.GetVMEnv().State().GetAccounts(){
		log.Info("Account key:",[]byte(k),"----------value:",v.Code())
		log.Info("Account addr:",v.Address().Hex())
		for a, v := range v.Storage() {
			log.Info("StateObject key:",a,"----------value:",v)
		}
	}
}

