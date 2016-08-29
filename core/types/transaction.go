package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"

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


// NewTransaction returns a new transaction
func NewTransaction(from []byte,to []byte,value []byte) *Transaction{

	transaction := &Transaction{
		From: from,
		To: to,
		Value: value,
		TimeStamp: time.Now().Unix(),
	}

	return transaction
}


// VerifyTransaction is to verify balance of the tranaction
// If the balance is not enough, returns false
func (tx *Transaction) VerifyTransaction() bool{
	return false
}