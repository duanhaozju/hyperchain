package version1_3

import (
	"hyperchain/common"
	"math/big"
)

func (tv *TransactionValue) RetrievePayload() []byte {
	return common.CopyBytes(tv.Payload)
}

func (tv *TransactionValue) RetrieveGas() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.GasLimit))
}

func (tv *TransactionValue) RetrieveGasPrice() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.Price))
}

func (tv *TransactionValue) RetrieveAmount() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.Amount))
}
