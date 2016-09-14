package state

import (
	"testing"
	"math/big"
	"hyperchain/common"
	"encoding/json"
	"fmt"
)

var stateObjectCase = StateObject{
	address : common.HexToAddress("0000000000000000000000000000000000000001"),
	balance : big.NewInt(123),
	nonce : 123,
	codeHash : []byte("123345"),
	code : []byte("1324345"),
	abi : []byte("asdfsdf"),
	storage : Storage{
		common.StringToHash("sdfdsfsdfsdaf") : common.StringToHash("0000000000000000000000000000000000000001"),
	},

	remove : true,
	deleted : true,
	dirty  : true,
}

func TestMarshalJSON(t *testing.T) {
	data , err := json.Marshal(stateObjectCase)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(data))
	var state StateObject
	json.Unmarshal(data, &state)
	fmt.Printf("%#v\n", state)
	fmt.Printf("%#v\n", stateObjectCase)
}
