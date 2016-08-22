package types

import (
	"hyperchain-alpha/core/node"
	"hyperchain-alpha/encrypt"
	"time"
	"encoding/json"
	"strconv"
	"encoding/hex"
)

type Block struct {
	ParentHash string
	BlockHash string
	Transactions []Transaction
	TimeStramp int64 //unix时间戳
	CoinBase node.Node // 打包该Block的地址
	MerkleRoot string // merkleRoot 的hash值
}

//todo TimeStramp 写错了 应该该掉

//-- 根据Transactions 打包成一个block
func NewBlock(trans Transactions, ParentHash string, coinBase node.Node) *Block {
	//-- 打包创世块
	block := Block{
		ParentHash: ParentHash,
		Transactions: trans,
		TimeStramp: time.Now().Unix(),
		CoinBase: coinBase,
		MerkleRoot: "root",
	}
	txBStr, _ := json.Marshal(block.Transactions)
	coinbaseBStr , _ := json.Marshal(block.CoinBase)

	block.BlockHash = string(encrypt.GetHash([]byte(block.ParentHash + string(txBStr) + strconv.FormatInt(block.TimeStramp, 10) + string(coinbaseBStr) + string(block.MerkleRoot))))

	return &block
}

func (blk Block) String()string{
	this := blk
	retString :="\n======================Block=====================\n"
	retString +="= ParentHash\t\t:"		+ hex.EncodeToString([]byte(this.ParentHash))+"\t=\n"
	retString +="= BlockHash\t\t:"		+ this.BlockHash			+"\t=\n"
	retString +="= Transactions\t\t:"	+ strconv.Itoa(len(this.Transactions))	+"\t=\n"
	retString +="= TimeStramp\t\t:"		+ strconv.FormatInt(this.TimeStramp,10)	+"\t=\n"
	retString +="= CoinBase\t\t:"		+ this.CoinBase.String()		+"\t=\n"
	retString +="= MerkleRoot\t\t:"		+ this.MerkleRoot			+"\t=\n"
	retString +="======================Block=====================\n"
	return retString
}