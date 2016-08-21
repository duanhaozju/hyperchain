package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	"hyperchain-alpha/core"
	"hyperchain-alpha/hyperchain"
)

func TransactionIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// TODO 取得所有的交易记录
	allTransaction,_ := core.GetAllTransactionFromLDB()
	if err := json.NewEncoder(w).Encode(allTransaction); err != nil {
		panic(err)
	}
}

// 处理请求 : GET "/trans"
//TODO 取得某一条交易记录
func TransacionShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoId := vars["transId"]
	fmt.Fprintln(w, "Translaction show:", todoId)
}

// TODO 增加一条记录
// 处理请求 : POST "/trans"
func TransactionCreate(w http.ResponseWriter, r *http.Request) {

	// 解析url传递的参数，对于POST则解析响应包的主体（request body）
	r.ParseForm()

	hyperchain.SendTransaction(r.Form)


	w.Header().Set("Access-Control-Allow-Origin", "*")//允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers","Content-Type")//header的类型
	w.Header().Set("content-type","application/json")
	fmt.Fprintf(w,"success")

	//w.WriteHeader(http.StatusCreated)
	//if err := json.NewEncoder(w).Encode(t); err != nil {
	//	panic(err)
	//}
}
