package core

import (
	"fmt"
	checker "gopkg.in/check.v1"
	"hyperchain/common"
	"hyperchain/core/types"
	"math/big"
	"testing"
)

type ReceiptSuite struct {
}

func Test(t *testing.T) {
	checker.TestingT(t)
}

var _ = checker.Suite(&ReceiptSuite{})

func (s *ReceiptSuite) TestPut(c *checker.C) {
	receiptInst, _ := GetReceiptInst()
	r1 := NewReceipt(common.BytesToHash([]byte("txhash1")))
	r2 := NewReceipt(common.BytesToHash([]byte("txhash2")))
	r3 := NewReceipt(common.BytesToHash([]byte("txhash3")))
	r4 := NewReceipt(common.BytesToHash([]byte("txhash4")))
	receiptInst.PutReceipt(common.BytesToHash([]byte("txhash1")), r1)
	receiptInst.PutReceipt(common.BytesToHash([]byte("txhash2")), r2)
	receiptInst.PutReceipt(common.BytesToHash([]byte("txhash3")), r3)
	receiptInst.PutReceipt(common.BytesToHash([]byte("txhash4")), r4)

	receiptInst.DeleteReceipt(common.BytesToHash([]byte("txhash4")))

	err := receiptInst.Commit()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(GetReceipt(common.BytesToHash([]byte("txhash3"))))
	receiptInst.Clear()
	fmt.Println(receiptInst.GetAllReceipt())
}
func NewReceipt(txhash common.Hash) *types.Receipt {
	receipt := types.NewReceipt([]byte("root"), big.NewInt(123))
	receipt.TxHash = txhash.Bytes()
	return receipt
}
