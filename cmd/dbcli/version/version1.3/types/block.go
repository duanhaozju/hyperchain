package version1_3

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/crypto"
)

func (self *Block) Hash() common.Hash {
	kec256Hash := crypto.NewKeccak256Hash("keccak256")
	return kec256Hash.Hash([]interface{}{
		self.ParentHash,
		self.Number,
		self.Timestamp,
		self.TxRoot,
		self.ReceiptRoot,
		self.MerkleRoot,
	})
}

func (self *Block) EncodeVerbose() string {
	var transactionView []*TransactionView
	for i := 0; i < len(self.Transactions); i++ {
		transactionView = append(transactionView, self.Transactions[i].ToTransactionView())
	}
	blockVerboseView := &BlockVerboseView{
		Version:      string(self.Version),
		ParentHash:   common.Bytes2Hex(self.ParentHash),
		BlockHash:    common.Bytes2Hex(self.BlockHash),
		Transactions: transactionView,
		Timestamp:    self.Timestamp,
		MerkleRoot:   common.Bytes2Hex(self.MerkleRoot),
		TxRoot:       common.Bytes2Hex(self.TxRoot),
		ReceiptRoot:  common.Bytes2Hex(self.ReceiptRoot),
		Number:       self.Number,
		WriteTime:    self.WriteTime,
		CommitTime:   self.CommitTime,
		EvmTime:      self.EvmTime,
		Bloom:        common.Bytes2Hex(self.Bloom),
	}
	res, err := json.MarshalIndent(blockVerboseView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}

func (self *Block) Encode() string {
	var transactionViewHash []*TransactionViewHash
	for i := 0; i < len(self.Transactions); i++ {
		transactionViewHash = append(transactionViewHash, self.Transactions[i].ToTransactionViewHash())
	}
	blockView := &BlockView{
		Version:      string(self.Version),
		ParentHash:   common.Bytes2Hex(self.ParentHash),
		BlockHash:    common.Bytes2Hex(self.BlockHash),
		Transactions: transactionViewHash,
		Timestamp:    self.Timestamp,
		MerkleRoot:   common.Bytes2Hex(self.MerkleRoot),
		TxRoot:       common.Bytes2Hex(self.TxRoot),
		ReceiptRoot:  common.Bytes2Hex(self.ReceiptRoot),
		Number:       self.Number,
		WriteTime:    self.WriteTime,
		CommitTime:   self.CommitTime,
		EvmTime:      self.EvmTime,
		Bloom:        common.Bytes2Hex(self.Bloom),
	}
	res, err := json.MarshalIndent(blockView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}

func (self *Block) EncodeTransaction(txIndex int) string {
	transactionView := self.Transactions[txIndex].ToTransactionView()
	res, err := json.MarshalIndent(transactionView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}
