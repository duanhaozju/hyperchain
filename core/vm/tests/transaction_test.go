package tests

import (
	"testing"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/core"
	"hyperchain/common"
)
func Test_FormatTx(t *testing.T){
	var (
		from  = common.HexToAddress("0f572e5295c57f15886f9b263e2f6d2d6c7b5ec6")
		//input = common.FromHex("0x3ad14af300000000000000000000000000000000000000000000000000000000000000010000000000000000000000000000000000000000000000000000000000000002")
	)
	var tx_value1 = &types.TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:([]byte)(sourcecode)}
	value1,err := proto.Marshal(tx_value1)
	if err != nil{
		t.Log("the test transaction has error")
	}

	if err != nil{
		t.Log("the test transaction has error")
	}
	core.ExecTransaction(*types.NewTransaction(from.Bytes(),nil,value1))

	//fmt.Println("addrs----",core.GetVMEnv().State().GetLeastAccount().Address().Bytes())
	//fmt.Println("addrs----",core.GetVMEnv().State().GetLeastAccount().Address().Hex())

	core.ExecTransaction(*types.NewTestCallTransaction())
	core.ExecTransaction(*types.NewTestCallTransaction())

	log.Notice(len(core.GetVMEnv().State().GetAccounts()))
	var num = 0
	for k,v := range core.GetVMEnv().State().GetAccounts(){
		log.Notice("the num of accounts is-----------------------------",num)
		log.Notice("Account key:",[]byte(k),"----------value:",v.Code())
		log.Notice("Account addr:",v.Address().Hex())
		for a, v := range v.Storage() {
			log.Notice("storage key:",a,"----------value:",v)
		}
		num  = num+1
	}
}
