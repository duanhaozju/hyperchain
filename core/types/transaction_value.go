package types

import (
	"hyperchain/common"
	"math/big"
	"encoding/json"
	"encoding/hex"
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

func ConstructInvokeArgs(method string, args []string) ([]byte, error) {
	var tmp [][]byte
	for _, arg := range args {
		v, err := hex.DecodeString(arg)
		if err != nil {
			return nil, err
		}
		tmp = append(tmp, v)
	}
	return json.Marshal(&InvokeArgs{
		MethodName:   method,
		Args:         tmp,
	})
}
