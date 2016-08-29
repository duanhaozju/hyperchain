package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"
	"hyperchain/core"
	"log"
	"math/big"
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
func (tx *Transaction) VerifyBalance() bool{
	var balance big.Int
	var value big.Int

	balanceIns, err := core.GetBalanceIns()

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}

	bal := balanceIns.GetCacheBalance(common.BytesToAddress(tx.From))

	balance.SetString(string(bal), 10)
	value.SetString(string(tx.Value), 10)

	if value.Cmp(balance) == 1 {
		return false
	}

	return true
}