//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package crypto

import (
	"bytes"
	"encoding/hex"
	"github.com/hyperchain/hyperchain/common"
	"math/big"
	"testing"
)

type Transaction struct {
	Recipient *common.Address
	Amount    *big.Int
	signature []byte
}

func NewTransaction(to common.Address, amount *big.Int) *Transaction {
	trans := Transaction{
		Recipient: &to,
		Amount:    new(big.Int),
	}
	if amount != nil {
		trans.Amount.Set(amount)
	}
	return &trans
}

var s256 = NewKeccak256Hash("Keccak256")

func TestHash(t *testing.T) {
	tx := NewTransaction(common.Address{}, big.NewInt(2))
	txHex := "5661b3c9f43aa179d80a1c7340aacff5da1425f9a9dac16e5f935ccb70f8568d"

	txHashnew := s256.Hash(tx).Bytes()
	txHashold, _ := hex.DecodeString(txHex)
	if bytes.Compare(txHashold, txHashnew) != 0 {
		t.Fatalf("hash mismatch: want: %x have: %x", txHashold, txHashnew)
	}

	msgHex := "7471aac94b229eaf10ebe70634c086b0c8c0887f00476c79001667e745199201"
	msgHexOld, _ := hex.DecodeString(msgHex)
	msgHashNew := s256.Hash([]interface{}{tx.Amount, tx.Recipient}).Bytes()

	if bytes.Compare(msgHexOld, msgHashNew) != 0 {
		t.Fatalf("hash mismatch: want: %x have: %x", msgHexOld, msgHashNew)
	}

}

func TestKeccak256Hash_ByteHash(t *testing.T) {
	msg := []byte("abc")
	msgHex := "4e03657aea45a94fc7d47ba826c8d667c0d1e6e33a64a036ec44f58fa12d6c45"

	newHash := s256.ByteHash(msg).Bytes()
	oldHash, _ := hex.DecodeString(msgHex)
	if !bytes.Equal(oldHash, newHash) {
		t.Fatalf("hash mismatch: want: %x have: %x", oldHash, newHash)
	}

	msgHex1 := "7830c5c592a6da328d2b1e060844c9f9b6daac9331f624407a4ef1b43a8574cc"
	newHash1 := s256.ByteHash([]byte{123}, []byte("abc")).Bytes()
	oldHash1, _ := hex.DecodeString(msgHex1)
	if !bytes.Equal(oldHash1, newHash1) {
		t.Fatalf("hash mismatch: want: %x have: %x", oldHash1, newHash1)
	}

}
