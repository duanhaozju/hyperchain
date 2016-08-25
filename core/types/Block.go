// common event defined
// author: Lizhong kuang
// date: 2016-08-25
// last modified:2016-08-25
package types

import (

	//"time"

	"math/big"
	"hyperchain-alpha/common"

	"time"
	"hyperchain-alpha/crypto"
)

type Block struct {
	ParentHash common.Hash
	BlockHash common.Hash
	Transactions []Transaction
	TimeStamp int64 //unix时间戳
	MerkleRoot common.Hash // merkleRoot 的hash值
	Number      *big.Int       // The block number

}



//general block
func NewBlock(txs Transactions, ParentHash common.Hash) *Block {
	ParentHash = new(common.Hash)
	block := Block{
		ParentHash: ParentHash,
		Transactions: txs,
		TimeStamp: time.Now().Unix(),

		MerkleRoot: "root",
	}

	return &block
}

// hash block
func (h *Block) Hash(hash crypto.CommonHash) common.Hash {
	return hash.Hash(h)
}

/*
func (blk Block) String()string{
	this := blk
	retString :="\n======================BLOCK<STRAT>==============\n"
	retString +="= ParentHasht\t: " + hex.EncodeToString([]byte(this.ParentHash))+"\n"
	retString +="= BlockHash\t: "+ hex.EncodeToString([]byte(this.BlockHash))+"\n"
	retString +="= Transactions\t: "	+ strconv.Itoa(len(this.Transactions))+"\n"
	retString +="= TimeStramp\t: "+ strconv.FormatInt(this.TimeStramp,10)+"\n"
	retString +="= CoinBase\t: "+ this.CoinBase.String()+"\n"
	retString +="= MerkleRoot\t: "+ this.MerkleRoot	+"\n"
	retString +="======================BLOCK<END>================\n"
	return retString
}*/
