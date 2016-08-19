package main

import (
	//"hyperchain.cn/app/jsonrpc/model"
	//"time"
	"fmt"
	//"encoding/json"
	//"os"
	//"hyperchain-alpha/hyperdb"
	//"hyperchain-alpha/core/types"
	"hyperchain-alpha/core"
)

func main() {
	// t := model.Transaction{
	//	 From:"jddd",
	//	 To:"snail",
	//	 Value: 3243,
	//	 TimeStamp:time.Now(),
	// }
	//
	//model.PutTransactionToDB("dd", t)

	/*t, _ := model.GetTransactionFromDB("dd")
	fmt.Println(t)*/


	/*type ddd struct {
		D []byte `json:"d"`
	}

	l := ddd{[]byte("ddddd")}
	fmt.Println(string(l.D))

	bstr, _ := json.Marshal(l)

	fmt.Println((bstr))
	fmt.Println(string(bstr))

	var k ddd;
	json.Unmarshal(bstr, &k)
	fmt.Println(string(k.D))*/


	//filepath, _ := os.Getwd()
	//filepath = filepath + "/ldb/8001"
	//ldb, _ := hyperdb.NewLDBDataBase(filepath)

	core.InitDB(8001)
	/*var t = types.Transaction{
		"dte",
		9899,
	}
	core.PutTransactionToLDB("key1", t)

	tget, _ := core.GetTransactionFromLDB("key")
	fmt.Println(tget)*/

	trans, _ := core.GetAllTransactionFromLDB();
	fmt.Println(trans)

}