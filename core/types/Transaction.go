// hyperchain alpha version 2016
// Copyright 2014 hyperchain.cn
// 这个文件定义了交易(transaction)的结构信息

package types

import "time"

type Transaction struct {
	From string //从发起账户公钥hash之后的值
	Publickey string //携带公钥
	To string //直接公钥
	Value int // 交易值
	TimeStamp time.Time //时间戳
	Signedhash string //整体签名
}