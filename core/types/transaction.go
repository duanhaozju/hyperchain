package types

import (
	"hyperchain/crypto"
	"hyperchain/common"
	"time"
	"crypto/ecdsa"
	"log"
)

func (tx *Transaction)Hash() common.Hash {
	this := *tx

	s256 := crypto.NewKeccak256Hash("Keccak256")

	return s256.Hash(this)
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

// WriteSignature returns the transaction with signature
func (tx *Transaction) WriteSignature(priv *ecdsa.PrivateKey) *Transaction{
	ee := crypto.NewEcdsaEncrypto("ECDSAEncryto")

	sig, err := ee.Sign(tx.Hash(), *priv)

	if err != nil {
		log.Fatalf("sign transaction error: %v",err)
	}

	tx.Signature = sig

	return tx

}

// VerifyTransaction is to verify balance of the tranaction
// If the balance is not enough, returns false
func (tx *Transaction) VerifyTransaction() bool{
	return false
}