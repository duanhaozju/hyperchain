package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	//"strconv"
	//"time"
	"hyperchain-alpha/core"
	//"hyperchain-alpha/core/types"
)

func TransactionIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)

	allTransaction,_ := core.GetAllTransactionFromLDB()
	if err := json.NewEncoder(w).Encode(allTransaction); err != nil {
		panic(err)
	}
}

// 处理请求 : GET "/trans"

func TransacionShow(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	todoId := vars["transId"]
	fmt.Fprintln(w, "Translaction show:", todoId)
}


// 处理请求 : POST "/trans"
func TransactionCreate(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Access-Control-Allow-Origin", "*")//允许访问所有域
	w.Header().Add("Access-Control-Allow-Headers","Content-Type")//header的类型
	w.Header().Set("content-type","application/json")
	fmt.Fprintf(w,"success")

	//w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	//w.WriteHeader(http.StatusCreated)
	//if err := json.NewEncoder(w).Encode(t); err != nil {
	//	panic(err)
	//}
}
