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
	retString :="\n======================BLOCK<STRAT>==============\n"
	retString +="= ParentHasht\t: " + hex.EncodeToString([]byte(this.ParentHash))+"\n"
	retString +="= BlockHash\t: "+ hex.EncodeToString([]byte(this.BlockHash))+"\n"
	retString +="= Transactions\t: "	+ strconv.Itoa(len(this.Transactions))+"\n"
	retString +="= TimeStramp\t: "+ strconv.FormatInt(this.TimeStramp,10)+"\n"
	retString +="= CoinBase\t: "+ this.CoinBase.String()+"\n"
	retString +="= MerkleRoot\t: "+ this.MerkleRoot	+"\n"
	retString +="======================BLOCK<END>================\n"
	return retString
}

//-- 将Block序列化
func (self Block) MarshalJSON() ([]byte, error) {
	type BlockJSON struct {
		ParentHash string
		BlockHash string
		Transactions [][]byte  //-- []byte表示一个交易
		TimeStramp int64 //unix时间戳
		CoinBase node.Node // 打包该Block的地址
		MerkleRoot string // merkleRoot 的hash值
	}

	var trans [][]byte
	for _, t := range self.Transactions {
		tstrb, _ := t.MarshalJSON()
		trans = append(trans, tstrb)
	}

	bJSON := BlockJSON{
		ParentHash: self.ParentHash,
		BlockHash: self.BlockHash,
		Transactions: trans,
		TimeStramp: self.TimeStramp,
		CoinBase: self.CoinBase,
		MerkleRoot: self.MerkleRoot,
	}
	return json.Marshal(bJSON)
}


//-- 将BlocK反序列化， self为接收容器
func (self *Block) UnMarShalJSON(data []byte) error {
	type BlockJSON struct {
		ParentHash string
		BlockHash string
		Transactions [][]byte  //-- []byte表示一个交易
		TimeStramp int64 //unix时间戳
		CoinBase node.Node // 打包该Block的地址
		MerkleRoot string // merkleRoot 的hash值
	}
	var bJSON BlockJSON
	json.Unmarshal(data, &bJSON)

	var ts []Transaction
	for _, v := range bJSON.Transactions  {
		var t Transaction
		if err := t.UnMarShalJSON(v); err != nil {
			return err
		}
		ts = append(ts, t)
	}
	self.ParentHash = bJSON.ParentHash
	self.BlockHash = bJSON.BlockHash
	self.Transactions = ts
	self.TimeStramp = bJSON.TimeStramp
	self.CoinBase = bJSON.CoinBase
	self.MerkleRoot = bJSON.MerkleRoot
	return nil
}