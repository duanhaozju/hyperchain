package controller
import (
	"net/http"
	"html/template"
	"path"
	"os"
	"hyperchain/jsonrpc/api"
)


type ResData struct{
	Data interface{}
	Code int
}

type data struct{
	Trans    []api.TransactionShow
	Balances api.BalanceShow
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
	balances := api.GetAllBalances()

	//sort.Sort(types.Transactions(transactions)) // sort transactions

	tmpl.Execute(w,data{
		Trans:transactions,
		Balances: balances,
	})
}




