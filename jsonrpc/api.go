package jsonrpc

import (
	"hyperchain/core/types"
	"fmt"
	"hyperchain/event"
	"github.com/golang/protobuf/proto"
)

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value string `json:"value"`
	Timestamp int64 `json:"timestamp"`
}

type ResData struct{
	Data interface{}
	Code int
}

// 参数是一个json对象
func SendTransaction(args TxArgs) ResData {

	var tx *types.Transaction

	fmt.Println(args)

	// 构造 transaction 实例
	tx = types.NewTransaction([]byte(args.From), []byte(args.To), []byte(args.Value))

	// 判断交易余额是否足够
	if (tx.VerifyTransaction()) {
		// 余额足够
		// 抛 NewTxEvent 事件
		 eventmux:=new(event.TypeMux)
		 eventmux.Post(event.NewTxEvent{Payload: proto.Marshal(*tx)})

		return ResData{
			Data: nil,
			Code: 1,
		}

	} else {
		// 余额不足
		return ResData{
			Data: nil,
			Code: 0,
		}
	}
}

// TODO GetAllTransactions
func GetAllTransactions()  ResData{
	return ResData{}
}

// TODO GetAllBalances
func GetAllBalances() ResData{
	return ResData{}
}


