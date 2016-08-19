package main

import (
	//"time"
	"fmt"
	"hyperchain-alpha/jsonrpc/model"
)

func main() {

	//model.InitDB(8989)
	// t := model.Transaction{
	//	 From:"jddd",
	//	 To:"snail",
	//	 Value: 3243,
	//	 TimeStamp:time.Now(),
	// }
	//
	//model.PutTransactionToDB("dd", t)

	t, _ := model.GetTransactionFromDB("dd")
	fmt.Println(t)
}