package main

import (
	//"hyperchain.cn/app/jsonrpc/model"
	//"time"
	"fmt"
	//"encoding/json"
	//"os"
	//"hyperchain-alpha/hyperdb"
	//"hyperchain-alpha/core/types"
	//"hyperchain-alpha/core"
	//"hyperchain-alpha/core/types"
	//"hyperchain-alpha/core"
	//"hyperchain-alpha/core/node"
	//"encoding/json"
	//"math/big"
	//"encoding/json"
	"hyperchain-alpha/core/types"
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

	//core.InitDB(8001)
	/*var t = types.Transaction{
		"dte",
		9899,
	}
	core.PutTransactionToLDB("key1", t)

	tget, _ := core.GetTransactionFromLDB("key")
	fmt.Println(tget)*/

//	trans, _ := core.GetAllTransactionFromLDB();
	//fmt.Println(trans)
	/*m1, _ := hyperdb.NewMemDatabase()
	m1.Put([]byte("key"), []byte("womendoushihaoxiangde"))
	bs,_ := m1.Get([]byte("key"))
	fmt.Println(string(bs))*/

	/*b1 := types.Balance{
		"yinyipeng",
		21,
	}
	b2 := types.Balance{
		"yinyipeng",
		22,
	}
	b3 := types.Balance{
		"yinyipeng",
		23,
	}
	core.PutBalanceToMEM("test1", b1)
	core.PutBalanceToMEM("test2", b2)
	core.PutBalanceToMEM("test3", b3)

	bs, _ := core.GetBalanceFromMEM("test2")
	fmt.Println(bs)

	bss, _ := core.GetAllBalanceFromMEM()

	fmt.Println(bss)*/

	/*balanceHeaderKey := "header"
	key := "headerJJJJdDD"
	//if (str2[:len(str1)])
	fmt.Println(key[:len(balanceHeaderKey)] == balanceHeaderKey)

	if key[:len(balanceHeaderKey)] == balanceHeaderKey {
		fmt.Println("pass")
	}*/

	/*type Person struct {
		Name string
		Age big.Int
	}

	*//*str,_  := json.Marshal(Person{"snail", 23})
	fmt.Println(string(str))*//*
	b := *big.NewInt(2437329598)
	p := Person{"snail", b}
	str, _ := json.Marshal(p)
	fmt.Println(string(str))*/
	/*core.InitDB(555)
	balance := types.Balance{"snail",8976}
	core.PutBalanceToMEM("yin", balance)
	core.PutBalanceToMEM("yid", balance)
	fmt.Println(core.GetBalanceFromMEM("yin"))
	fmt.Println(core.GetAllBalanceFromMEM())*/

	core.InitDB(1000)
	trans := types.Transaction{
		From:"345456",
	}
	core.AddTransactionToTxPool(trans)
	fmt.Println(core.GetTxPoolCapacity())
	core.AddTransactionToTxPool(trans)
	fmt.Println(core.GetTxPoolCapacity())
	core.AddTransactionToTxPool(trans)
	fmt.Println(core.GetTxPoolCapacity())
	core.AddTransactionToTxPool(trans)
	fmt.Println(core.GetTxPoolCapacity())
	core.AddTransactionToTxPool(trans)

	fmt.Println(core.GetTxPoolCapacity())

	fmt.Println(core.TxPoolIsFull())
	fmt.Println(core.GetTxPool())
	core.ClearTxPool()
	fmt.Println(core.GetTxPoolCapacity())
	core.AddTransactionToTxPool(trans)
	fmt.Println(core.GetTxPoolCapacity())
	core.AddTransactionToTxPool(trans)
	fmt.Println(core.GetTxPoolCapacity())

}