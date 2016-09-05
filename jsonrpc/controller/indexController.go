package controller
import (
	"net/http"
	"html/template"
	"path"
	"os"
	"hyperchain/jsonrpc/api"
	"encoding/json"
	"io/ioutil"
)


type ResData struct{
	Data interface{}
	Code int
}

type data struct{
	Trans    []api.TransactionShow
	LastestBlock api.LastestBlockShow
}

var Testing bool = false

// Index function is the handler of "/", GET
func Index(w http.ResponseWriter, r *http.Request) {

	var indexpath string

	pwd, _ := os.Getwd()

	if Testing {
		indexpath = path.Join(pwd,"./static/tmpl/index.html")
	} else {
		indexpath = path.Join(pwd,"./jsonrpc/static/tmpl/index.html")
	}


	var tmpl = template.Must(template.ParseFiles(indexpath))


	transactions := api.GetAllTransactions()
	lastestBlock := api.LastestBlock()

	tmpl.Execute(w,data{
		Trans:transactions,
		LastestBlock: lastestBlock,
	})
}

// BalancesGet function is the handler of "/balances", GET
func BalancesGet(w http.ResponseWriter, r *http.Request) {
	balances := api.GetAllBalances()

	data, err := json.Marshal(ResData{
		Data: balances,
		Code: 1,
	})
	if err != nil {
		log.Fatalf("Error: %v",err)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(data)
}

// BlocksGet function is the handler of "/blocks", GET
func BlocksGet(w http.ResponseWriter, r *http.Request) {
	blocks := api.GetAllBlocks()

	data, err := json.Marshal(ResData{
		Data: blocks,
		Code: 1,
	})
	if err != nil {
		log.Fatalf("Error: %v",err)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(data)
}

// TransactionGet function is the handler of "/trans", GET
func TransactionGet(w http.ResponseWriter, r *http.Request) {
	transactions := api.GetAllTransactions()

	data, err := json.Marshal(ResData{
		Data: transactions,
		Code: 1,
	})

	if err != nil {
		log.Fatalf("Error: %v",err)
		return
	}

	w.Header().Set("Content-Type","application/json")
	w.Write(data)
}

func ExecuteTimeQuery(w http.ResponseWriter, r *http.Request) {
	var res ResData
	var p = api.TxArgs{}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Fatalf("Error: %v",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	err = json.Unmarshal(body, &p)
	if err != nil {
		log.Fatalf("Error: %v",err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	time := api.QueryExcuteTime(api.TxArgs{
		From: p.From,
		To: p.To,
	})


	res = ResData{
		Data: time,
		Code:1,
	}

	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Add("Access-Control-Allow-Headers","Content-Type")
	w.Header().Set("Content-Type","application/json")

	b,_ := json.Marshal(res)

	w.Write(b)
}



