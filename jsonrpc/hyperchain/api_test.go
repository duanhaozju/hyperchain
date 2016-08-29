package hyperchain

import (
	"testing"
	"fmt"
)

func TestSendTransaction(t *testing.T) {

	isSuccess := SendTransaction(TxArgs{
		From: "addressFrom",
		To: "addressTo",
		Value: "12",
	})

	fmt.Println(isSuccess)
}
