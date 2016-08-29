package controller

import (
	//"encoding/json"
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	//"hyperchain/core"
	//"hyperchain-alpha/hyperchain"
	//"strconv"
	//"strconv"
)

type ResData struct{
	Data interface{}
	Code int
}

/*
func TransactionIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	allTransaction,_ := core.GetAllTransactionFromLDB()
	if err := json.NewEncoder(w).Encode(allTransaction); err != nil {
		panic(err)
	}
}
*/

// 处理请求 : GET "/trans"

func TransacionShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoId := vars["transId"]
	fmt.Fprintln(w, "Translaction show:", todoId)
}


// 处理请求 : POST "/trans"
/*func TransactionCreate(w http.ResponseWriter, r *http.Request) {

	var res ResData

	// 解析url传递的参数，对于POST则解析响应包的主体（request body）
	r.ParseForm()


	val, _ := strconv.Atoi(r.Form["value"][0])


	err := hyperchain.SendTransaction(hyperchain.TxArgs{
		From: r.Form["from"][0],
		To: r.Form["to"][0],
		Value: val,
	})

	if (err != nil) {
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
}*/
