// hyperchain alpha version 2016
// Copyright 2014 hyperchain.cn
// 这个文件定义了交易(transaction)的结构信息

package types

import (
	"time"
	"hyperchain-alpha/encrypt"
	"strconv"
)

type Transaction struct {
	From string //从发起账户公钥hash之后的值
	//Publickey string //携带公钥
	To string //送达账户公钥hash之后的值
	Value int // 交易值
	TimeStamp time.Time //时间戳
	Singnature encrypt.Signature //数字签名
	//Signedhash string //整体签名
}

type Transactions []Transaction

func NewTransaction(from string,to string,value int, timestamp time.Time) *Transaction{
	return &Transaction{
		From: from,
		To: to,
		Value: value,
		TimeStamp: timestamp,
	}
}

//TODO 验证交易，发送者是否存在,余额是否足够,判断getBalance还有tx pool中这个from的交易，进行加减
func (t *Transaction)VerifyTransaction()bool{
	//self := t
	return true
}

func (t *Transaction) Hash() string{
	self := t
	return string(encrypt.GetHash([]byte(self.From + self.To + strconv.Itoa(self.Value)+ string(self.TimeStamp.Format("2006-01-02 15:04:05")))))
}