package version1_2

import (
	"hyperchain/common"
	"hyperchain/crypto"
	//"fmt"
	//"encoding/json"
	"encoding/json"
	"fmt"
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

func (self *Block) Encode() string {
	var transactionView []*TransactionView
	for i := 0; i < len(self.Transactions); i++ {
		transactionView = append(transactionView, self.Transactions[i].ToTransactionView())
	}
	blockView := &BlockView{
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
	}
	res, err := json.MarshalIndent(blockView, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}
