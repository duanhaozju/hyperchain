package types

import (
	"hyperchain/common"
	"math/big"
	"github.com/golang/protobuf/proto"
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

func (tx *Transaction) GetNVPHash() (string, error) {
	var txExtra TxExtra
	err := proto.Unmarshal(tx.Extra, &txExtra)
	if err != nil {
		return "", err
	}
	return common.Bytes2Hex(txExtra.NodeHash), nil
}

func (tx *Transaction) SetNVPHash(hash string) error {
	txExtra := &TxExtra{
		NodeHash: common.Hex2Bytes(hash),
	}
	extra, err := proto.Marshal(txExtra)
	if err != nil {
		return err
	}
	tx.Extra = extra
	return nil
}