package api

/*import (
	"testing"
	"hyperchain/core"
	"hyperchain/common"
	"github.com/stretchr/testify/assert"
)


func TestSendTransaction(t *testing.T) {
	balanceIns,err := core.GetBalanceIns()
	if err != nil {
		log.Fatalf("%v", err)
	}
	balanceIns.DeleteCacheBalance(common.HexToAddress("000000000000000000000000000000adressFrom"))

	isSuccess := SendTransaction(TxArgs{
		From: "000000000000000000000000000000adressFrom",
		To: "00000000000000000000000000000000adressTo",
		Value: "12",
	})
	assert.Equal(t,false, isSuccess, "they should be equal")


	balanceIns.PutCacheBalance(common.HexToAddress("000000000000000000000000000000adressFrom"),[]byte("13"))

	isSuccess2 := SendTransaction(TxArgs{
		From: "000000000000000000000000000000adressFrom",
		To: "00000000000000000000000000000000adressTo",
		Value: "12",
	})

	assert.Equal(t,true, isSuccess2, "they should be equal")
}*/

