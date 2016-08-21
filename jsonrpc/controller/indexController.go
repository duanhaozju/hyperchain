package controller
import (
	"net/http"
	"html/template"
	"hyperchain-alpha/core"
	"path"
	"os"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/core/types"
)

//type Transaction struct{
//	From string
//	To string
//	Value int
//	Time string
//}
//
//type Transacctions []Transaction

type data struct{
	Trans types.Transactions
	Accounts utils.Accounts
}

// 处理请求 : GET "/"
func Index(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	indexpath := path.Join(pwd,"./jsonrpc/static/tmpl/index.html")

	var tmpl = template.Must(template.ParseFiles(indexpath))

	// 得到所有账户
	accounts,_ := utils.GetAccount()

	transactions,_ := core.GetAllTransactionFromLDB()

	//tmpl.Execute(w,transactions)
	tmpl.Execute(w,data{
		Trans:transactions,
		Accounts:accounts,
	})
}




