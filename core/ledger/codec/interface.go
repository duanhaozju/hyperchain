package codec

// Package codec implement the encoder library for old version of transactions.

import "hyperchain/core/types"

// Encoder encodes the transaction and receipt
type Encoder interface {
	EncodeTransaction(tx *types.Transaction) ([]byte, error)
	EncodeReceipt(receipt *types.Receipt) ([]byte, error)
}
