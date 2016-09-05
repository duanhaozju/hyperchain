package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"
	"fmt"
	"math/big"
	"github.com/golang/protobuf/proto"
)

func (self *Transaction)Hash(ch crypto.CommonHash) common.Hash {
	return ch.Hash(self)
}

func (self *Transaction)SighHash(ch crypto.CommonHash) common.Hash {
	return ch.Hash([]interface{}{
		self.Value,
		self.TimeStamp,
		self.From,
		self.To,
	})
}

func (self *Transaction)FString() string {
	return fmt.Sprintf(`
	From : %s
	To : %s
	Value : %s
	TimeStamp : %d
	Signature : %s`,
		self.From,
		self.To,
		self.Value,
		self.TimeStamp,
		self.Signature)
}

// NewTransaction returns a new transaction
func NewTransaction(from []byte,to []byte,value []byte) *Transaction{

	transaction := &Transaction{
		From: from,
		To: to,
		Value: value,
		TimeStamp: time.Now().UnixNano(),
	}

	return transaction
}



func (tx *Transaction) Payload() []byte       {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value(),transactionValue)
	return common.CopyBytes(transactionValue.Payload)
}
func (tx *Transaction) Gas() *big.Int      {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value(),transactionValue)
	return new(big.Int).Set(transactionValue.GasLimit)
}
func (tx *Transaction) GasPrice() *big.Int {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value(),transactionValue)
	return new(big.Int).Set(transactionValue.Price)
}
func (tx *Transaction) Amount() *big.Int    {
	transactionValue := &TransactionValue{}
	proto.Unmarshal(tx.Value(),transactionValue)
	return new(big.Int).Set(transactionValue.Amount)
}
