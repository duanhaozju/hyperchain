package hyperchain

import (
	"testing"
	"fmt"
	"hyperchain/core"
	"log"
	"hyperchain/common"
)

func TestSendTransaction(t *testing.T) {

	isSuccess := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	fmt.Println(isSuccess)

	balanceIns,err := core.GetBalanceIns()
	if err != nil {
		log.Fatalf("%v", err)
	}
	balanceIns.PutCacheBalance(common.BytesToAddress([]byte("addressFrom")),[]byte("13"))

	isSuccess2 := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	fmt.Println(isSuccess2)
}
