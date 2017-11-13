package version1_1

import (
	"github.com/hyperchain/hyperchain/common"
	"math/big"
)

func (tv *TransactionValue) GetPayload() []byte {
	return common.CopyBytes(tv.Payload)
}

func (tv *TransactionValue) GetGas() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.GasLimit))
}

func (tv *TransactionValue) GetGasPrice() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.Price))
}

func (tv *TransactionValue) GetAmount() *big.Int {
	return new(big.Int).Set(big.NewInt(tv.Amount))
}
