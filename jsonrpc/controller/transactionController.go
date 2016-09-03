package controller

import (
	"encoding/json"
	"fmt"
	"net/http"
	"hyperchain/jsonrpc/api"
	"github.com/op/go-logging"
	"io/ioutil"
)


var log *logging.Logger // package-level logger
func init() {
	log = logging.MustGetLogger("jsonrpc/api")
}

// TransactionCreate is the handler of "/trans", POST
func TransactionCreate(w http.ResponseWriter, r *http.Request) {

	var res ResData
	var p = api.TxArgs{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Error: "+err+"\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		log.Fatalf("Error: "+err+"\n")
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	isBalanceEnough := api.SendTransaction(api.TxArgs{
		From: p.From,
		To: p.To,
		Value: p.Value,

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
	w.Header().Set("Content-Type","application/json")

	b,_ := json.Marshal(res)

	fmt.Fprintf(w,string(b))
}
