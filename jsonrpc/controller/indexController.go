package controller
import (
	"net/http"
	"html/template"
	"path"
	"os"
)

type ResData struct{
	Data interface{}
	Code int
}

type data struct{
	Trans []Transaction
	Balances Balance
}

// 处理请求 : GET "/"
func Index(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	indexpath := path.Join(pwd,"./jsonrpc/static/tmpl/index.html")

	var tmpl = template.Must(template.ParseFiles(indexpath))

	transactions := GetAllTransactions()
	balances := GetAllBalances()

	//sort.Sort(types.Transactions(transactions)) // 交易排序


	tmpl.Execute(w,data{
		Trans:transactions,
		Balances: balances,
	})
}




