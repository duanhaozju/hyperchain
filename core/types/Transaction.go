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

//需要将签名字符串反序列化 signUndecoded => encrypt.Signature
func NewTransaction(from string,to string,value int,signUndecoded string){

}
//验证交易
func (t *Transaction)Verify()bool{

}

func (t *Transaction) Hash() string{
	self := t
	return encrypt.GetHash([]byte(self.From + self.To + strconv.Itoa(self.Value)+ string(self.TimeStamp)))
}