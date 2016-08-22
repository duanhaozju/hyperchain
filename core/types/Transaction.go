// hyperchain alpha version 2016
// Copyright 2014 hyperchain.cn
// 这个文件定义了交易(transaction)的结构信息

package types

import (
	"hyperchain-alpha/encrypt"
	"strconv"
	"time"
	"encoding/hex"
	"crypto/dsa"
)

type Transaction struct {
	From      string            //从发起账户公钥hash之后的值
								//Publickey string //携带公钥
	To        string            //送达账户公钥hash之后的值
	Value     int               // 交易值
	TimeStamp int64             //时间戳
	Signature encrypt.Signature //数字签名
								//Signedhash string //整体签名
}

type Transactions []Transaction

func (tx Transactions) Len() int {
	return len(tx)
}

func (tx Transactions) Swap(i,j int){
	tx[i],tx[j] = tx[j],tx[i]
}

func (tx Transactions) Less(i,j int) bool {
	return tx[j].TimeStamp < tx[i].TimeStamp
}


func NewTransaction(from string,to string,value int) *Transaction{
	return &Transaction{
		From: from,
		To: to,
		Value: value,
		TimeStamp: time.Now().Unix(),
	}
}

func (tx *Transaction) WithSignTransaction(priKey dsa.PrivateKey) *Transaction {

	signature,_ := encrypt.Sign(priKey,[]byte(tx.From + tx.To + strconv.Itoa(tx.Value) + strconv.FormatInt(tx.TimeStamp, 10)))

	tx.Signature = signature

	return tx

}

// 验证交易.from是否存在,余额是否足够,判断getBalance还有tx pool中这个from的交易，进行加减
func (tx *Transaction) VerifyTransaction(balance Balance,txPoolsTrans Transactions) bool {
	self := tx

	return self.isValid(balance,txPoolsTrans)
}

func (tx *Transaction) Hash() string{
	self := tx
	return string(encrypt.GetHash([]byte(self.From + self.To + strconv.Itoa(self.Value)+ strconv.FormatInt(self.TimeStamp, 10))))
}

// 检查余额
func (tx *Transaction) isValid(balance Balance,txPoolsTrans []Transaction) bool{

	self := tx
	fund := balance.Value

	from := self.From
	amount := self.Value  // balance和交易池中的资金是否足够

	for _,t := range txPoolsTrans{

		if (t.From == from) {
			fund = fund - t.Value
		}

		if (t.To == from) {
			fund = fund + t.Value
		}
	}


	return amount <= fund
}

func (tx Transaction) String()string{
	return hex.EncodeToString([]byte(tx.Hash()))
}