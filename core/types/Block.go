package types

import (

	"time"

	"math/big"
	"hyperchain-alpha/common"
)

type Block struct {
	ParentHash string
	BlockHash common.Hash
	Transactions []Transaction
	TimeStramp int64 //unix时间戳
	MerkleRoot string // merkleRoot 的hash值
	Number      *big.Int       // The block number

}

//todo TimeStramp 写错了 应该该掉

//-- 根据Transactions 打包成一个block
func NewBlock(trans Transactions, ParentHash string) *Block {
	//-- 打包创世块
	block := Block{
		ParentHash: ParentHash,
		Transactions: trans,
		TimeStramp: time.Now().Unix(),

		MerkleRoot: "root",
	}
	//txBStr, _ := json.Marshal(block.Transactions)


	//block.BlockHash = string(encrypt.GetHash([]byte(block.ParentHash + string(txBStr) + strconv.FormatInt(block.TimeStramp, 10) + string(coinbaseBStr) + string(block.MerkleRoot))))

	return &block
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
