package hpc

import (
	"testing"
	"hyperchain/hyperdb"
	"hyperchain/core"
	"hyperchain/core/types"
	"github.com/golang/protobuf/proto"
	"fmt"
	"hyperchain/common"
	"encoding/json"
	"os"
)

func TestPublicTransactionAPI_GetTransactionByHash(t *testing.T) {


	api :=  NewPublicTransactionAPI(nil,nil)

	db, err := hyperdb.GetLDBDatabase()

	if err != nil {
		t.Error(err)
	}

	txValue := types.NewTransactionValue(100,100,1,nil)

	value, err := proto.Marshal(txValue)

	if err != nil {
		t.Error(err)
	}

	tx := types.NewTransaction(common.FromHex("0x000f1a7a08ccc48e5d30f80850cf1cf283aa3abd"), common.FromHex("0x0000000000000000000000000000000000000003"), value)

	hash := tx.BuildHash()

	core.PutTransaction(db, hash[:], tx)

	txGet,err := api.GetTransactionByHash(common.ToHex(hash[:]))

	if err != nil {
		t.Error(err)
	}

	fmt.Printf("tx hash: %#v\n",common.ToHex(txGet.Hash[:]))
	e := json.NewEncoder(os.Stdout)
	e.Encode(txGet)
}