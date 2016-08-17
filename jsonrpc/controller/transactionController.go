package controller

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"fmt"
	"net/http"
	"strconv"
	"time"
	"hyperchain-alpha/jsonrpc/model"
	"hyperchain-alpha/p2p"
)

func TransactionIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	// TODO 取得所有的交易记录
	allTransaction,_ := model.GetAllTransaction()
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
	var transaction model.Transaction
	r.ParseForm()

	value,_ := strconv.Atoi(r.Form["value"][0])
	transaction = model.Transaction{
		From: r.Form["from"][0],
		To: r.Form["to"][0],
		Value: value,
		TimeStamp: time.Now(),
	}
	//1. 本地存储数据
	transaction = model.SaveTransction(transaction)

	//2. 远程同步数据
	//遍历节点
	for _,remoteNode := range p2p.GLOBALNODES{
		if remoteNode != p2p.LOCALNODE{
			p2p.SaveTrans(remoteNode,p2p.LOCALNODE,transaction)
		}
	}

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
