package controllers

import (
	"github.com/gorilla/mux"
	"net/http"
	"hyperchain/hpc"
	"hyperchain/common"
	"encoding/json"
	"github.com/op/go-logging"
)

var (
	log        *logging.Logger // package-level logger
)

func init() {
	log = logging.MustGetLogger("controller")
}

// 处理请求：POST "/v1/transactions"
func GetTransactions(w http.ResponseWriter, r *http.Request) {
	var args hpc.IntervalArgs

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	txs, err := PublicTxAPI.GetTransactions(args)
	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(txs, nil))
	}

}

// 处理请求：POST "/v1/transactions"
func SendTransaction(w http.ResponseWriter, r *http.Request) {
	var args hpc.SendTxArgs

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	hash, err := PublicTxAPI.SendTransaction(args)

	if err != nil {
		write(w, NewJSONObject(hash, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(hash, nil))
	}

}

// 处理请求：GET "/v1/transactions/{txHash}"
func GetTransactionByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash := vars["txHash"]

	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	tx, err := PublicTxAPI.GetTransactionByHash(common.HexToHash(txHash))

	if err != nil {
		write(w, NewJSONObject(tx, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(tx, nil))
	}


}

// 处理请求：GET "/v1/transactions/{blockHash}/{index}"
//func GetTransactionByBlockHashAndIndex(w http.ResponseWriter, r *http.Request) {
//	vars := mux.Vars(r)
//	blockHash := vars["blockHash"]
//	index := vars["index"]
//
//	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
//	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)
//
//	tx, err := PublicTxAPI.GetTransactionByBlockHashAndIndex(common.HexToHash(blockHash), )
//
//}

// 处理请求：GET "/v1/transactions/{txHash}/receipt"
func GetReceiptByHash(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	txHash := vars["txHash"]
	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)
	rep, err := PublicTxAPI.GetTransactionReceipt(common.HexToHash(txHash))

	if err != nil {
		write(w, NewJSONObject(rep, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(rep, nil))
	}
}

// 处理请求："/v1/transactions/sign-hash"
func GetSignHash(w http.ResponseWriter, r *http.Request) {
	var args hpc.SendTxArgs

	decoder := json.NewDecoder(r.Body)

	err := decoder.Decode(&args)

	if err != nil {
		write(w, NewJSONObject(nil, &callbackError{err.Error()}))
	}

	PublicTxAPIInterface := hpc.GetApiObjectByNamespace("tx").Service
	PublicTxAPI := PublicTxAPIInterface.(*hpc.PublicTransactionAPI)

	hash, err := PublicTxAPI.GetSignHash(args)

	if err != nil {
		write(w, NewJSONObject(hash, &callbackError{err.Error()}))
	} else {
		write(w, NewJSONObject(hash, nil))
	}
}
