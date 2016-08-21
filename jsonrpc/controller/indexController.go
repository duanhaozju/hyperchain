package controller
import (
	"net/http"
	"html/template"
	"hyperchain-alpha/core"
	"path"
	"os"
	"hyperchain-alpha/utils"
	"fmt"
)

//type Transaction struct{
//	From string
//	To string
//	Value int
//	Time string
//}
//
//type Transacctions []Transaction



// 处理请求 : GET "/"
func Index(w http.ResponseWriter, r *http.Request) {
	pwd, _ := os.Getwd()
	indexpath := path.Join(pwd,"./jsonrpc/static/tmpl/index.html")

	var tmpl = template.Must(template.ParseFiles(indexpath))

	account,_ := utils.GetAccount()
	fmt.Println(account)


	transactions,_ := core.GetAllTransactionFromLDB()

	tmpl.Execute(w,transactions)
	//tmpl.Execute(w,Transacctions{
	//	Transaction{
	//		"mxxim",
	//		"sammy",
	//		10,
	//		"20160909",
	//	},
	//	Transaction{
	//		"sammy",
	//		"mxxim",
	//		10,
	//		"20160909",
	//	},
	//})

}




