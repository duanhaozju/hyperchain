package main

import (
	//"hyperchain.cn/app/jsonrpc/model"
	//"time"
	"fmt"
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

	var d []byte = []byte("中文")
	fmt.Println(string(d))
}