// hyperchain alpha version 2016
// Copyright 2014 hyperchain.cn
// 这个文件定义了交易(transaction)的结构信息

package types

import "time"

type Transaction struct {
	from string //从发起账户公钥hash之后的值
	publickey string //携带公钥
	to string //直接公钥
	value int // 交易值
	timeStamp time.Time //时间戳
	signedhash string //整体签名
}