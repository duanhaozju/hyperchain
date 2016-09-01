package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"hyperchain/jsonrpc/api"
)

// TransactionCreate is the handler of "/trans", POST
func TransactionCreate(w http.ResponseWriter, r *http.Request) {

	var res ResData

	// Parse parameter
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

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers","Content-Type")
	w.Header().Set("content-type","application/json")

	b,_ := json.Marshal(res)

	fmt.Fprintf(w,string(b))

	//w.WriteHeader(http.StatusCreated)
	//if err := json.NewEncoder(w).Encode(t); err != nil {
	//	panic(err)
	//}
}
