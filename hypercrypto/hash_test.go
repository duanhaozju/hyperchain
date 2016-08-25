package hyperencrypt

import (
	"testing"
	"hyperchain-alpha/common"
	"math/big"
	"fmt"
)

func TestFix32Hash(t *testing.T) {
	tx := NewTransaction(common.Address{},big.NewInt(2))
	txs := make([]*Transaction,1)
	txs[0] = tx
	block := NewBlock(txs,"parenthash")

	fmt.Println(tx.txHash())
	fmt.Println(block.blockHash())
}
func (tx Transaction)txHash() common.Hash {
	return Fix32Hash(tx)
}
func (b Block)blockHash() common.Hash {
	return Fix32Hash(b)
}