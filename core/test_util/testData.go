package test_util

import (
	"hyperchain/core/types"
	"time"
)

var TransactionCases = []*types.Transaction{
	&types.Transaction{
		Version:         []byte(TransactionVersion),
		From:            []byte("6201cb0448964ac597faf6fdf1f472edf2a22b89"),
		To:              []byte("0000000000000000000000000000000000000003"),
		Value:           []byte("100"),
		Timestamp:       time.Now().UnixNano() - int64(time.Second),
		Signature:       []byte("signature1"),
		Id:              1,
		TransactionHash: []byte("transactionHash1"),
		Nonce:           1,
	},
	&types.Transaction{
		Version:         []byte(TransactionVersion),
		From:            []byte("0000000000000000000000000000000000000001"),
		To:              []byte("0000000000000000000000000000000000000002"),
		Value:           []byte("100"),
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("signature2"),
		Id:              2,
		TransactionHash: []byte("transactionHash2"),
		Nonce:           2,
	},
	&types.Transaction{
		Version:         []byte(TransactionVersion),
		From:            []byte("0000000000000000000000000000000000000002"),
		To:              []byte("0000000000000000000000000000000000000003"),
		Value:           []byte("700"),
		Timestamp:       time.Now().UnixNano(),
		Signature:       []byte("signature3"),
		Id:              3,
		TransactionHash: []byte("transactionHash3"),
		Nonce:           3,
	},
}

var BlockCases = types.Block{
	Version:      []byte(BlockVersion),
	ParentHash:   []byte("parentHash"),
	BlockHash:    []byte("blockHash"),
	Transactions: TransactionCases,
	Timestamp:    1489387222,
	MerkleRoot:   []byte("merkleRoot"),
	TxRoot:       []byte("txRoot"),
	ReceiptRoot:  []byte("receiptRoot"),
	Number:       1,
	WriteTime:    1489387223,
	CommitTime:   1489387224,
	EvmTime:      1489387225,
}

var TransactionMeta = types.TransactionMeta{
	BlockIndex: 1,
	Index:      1,
}

var Receipt = types.Receipt{
	Version:           []byte(ReceiptVersion),
	PostState:         []byte("postState"),
	CumulativeGasUsed: 1,
	TxHash:            []byte("12345678901234567890123456789012"),
	ContractAddress:   []byte("contractAddress"),
	GasUsed:           1,
	Ret:               []byte("ret"),
	Logs:              []byte("logs"),
	Status:            types.Receipt_SUCCESS,
	Message:           []byte("message"),
}
