package crypto

import (
	"testing"
	"hyperchain/common"
	"math/big"
	"fmt"
)

type Block struct {
	ParentHash string
	Transactions []*Transaction
}

func NewBlock(txs []*Transaction, ParentHash string) *Block {
	block := &Block{
		ParentHash: ParentHash,
		Transactions: make([]*Transaction,len(txs)),
	}
	copy(block.Transactions,txs)

	return block
}

func TestHash(t *testing.T) {
	tx := NewTransaction(common.Address{},big.NewInt(2))
	txs := make([]*Transaction,1)
	txs[0] = tx
	block := NewBlock(txs,"parenthash")

	s256 := NewKeccak256Hash("Keccak256")

	fmt.Println("tx hash:")
	fmt.Println(s256.Hash(tx))
	fmt.Println("tx hash with part data")
	fmt.Println(s256.Hash([]interface{}{tx.data.Amount,tx.data.Recipient}))
	fmt.Println("block hash")
	fmt.Println(s256.Hash(block))
}

