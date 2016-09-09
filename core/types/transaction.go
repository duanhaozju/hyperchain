package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"

	"fmt"
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
//validates the signature
//if addr recovered from signature != tx.from return false
func (self *Transaction)ValidateSign(encryption crypto.Encryption,ch crypto.CommonHash)  bool{

	hash := self.SighHash(ch)
	addr,_ := encryption.UnSign(hash[:],self.Signature)

	return addr==self.From
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
