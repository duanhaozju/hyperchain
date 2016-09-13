package tests

import (
	"testing"
	"github.com/golang/protobuf/proto"
	"hyperchain/core/types"
	"hyperchain/core/vm/api"
)

func Test_FormatTx(t *testing.T){
	var from = ""
	var to = ""
	var tx_value = &types.TransactionValue{Price:100000,GasLimit:100000,Amount:100,Payload:[]byte(to)}

	value,err := proto.Marshal(tx_value)
	if err != nil{
		t.Log("the test transaction has error")
	}
	tx := types.NewTransaction([]byte(from),[]byte(to),value)
	api.ExecTransaction(*tx)
}

