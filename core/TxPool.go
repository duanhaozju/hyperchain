package core

type TxPool struct {
	MaxCapacity int
	Transactions []types.Transaction
}
