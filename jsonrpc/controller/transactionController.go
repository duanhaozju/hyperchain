package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"hyperchain/jsonrpc/api"
)

// 处理请求 : POST "/trans"
func TransactionCreate(w http.ResponseWriter, r *http.Request) {

	var res ResData

	// 解析url传递的参数，对于POST则解析响应包的主体（request body）
	r.ParseForm()

	isBalanceEnough := api.SendTransaction(api.TxArgs{
		From: r.Form["from"][0],
		To: r.Form["to"][0],
		Value: r.Form["value"][0],

	})

	if (!isBalanceEnough) {
		res = ResData{
				Data: nil,
				Code:0,
			}
	} else {
		res = ResData{
			Data: nil,
			Code:1,
		}
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")//允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers","Content-Type")//header的类型
	w.Header().Set("content-type","application/json")

	b,_ := json.Marshal(res)

	fmt.Fprintf(w,string(b))

	//w.WriteHeader(http.StatusCreated)
	//if err := json.NewEncoder(w).Encode(t); err != nil {
	//	panic(err)
	//}
}
