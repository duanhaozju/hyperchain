package controller
import (
	"net/http"
	"html/template"
	"hyperchain/core"
	"path"
	"os"
	"hyperchain/jsonrpc"
)

type ResData struct{
	Data interface{}
	Code int
}

type data struct{
	Trans []jsonrpc.Transaction
	Balances core.BalanceMap
}

// 处理请求 : GET "/"
func Index(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	indexpath := path.Join(pwd,"./jsonrpc/static/tmpl/index.html")

	var tmpl = template.Must(template.ParseFiles(indexpath))

	transactions := jsonrpc.GetAllTransactions()
	balances := jsonrpc.GetAllBalances()

	//sort.Sort(types.Transactions(transactions)) // 交易排序


	tmpl.Execute(w,data{
		Trans:transactions,
		Balances: balances,
	})
}




