package crypto


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


