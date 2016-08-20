package controller
import (
	"net/http"
	"html/template"
	"hyperchain-alpha/core"
	"path"
	"os"
	"hyperchain-alpha/utils"
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

	//TODO 读取keystore文件夹，包括公钥、私钥、用户名，有个用户名与私钥的map对应表 map[string]string
	account,_ := utils.GetAccount()
	//TODO 传入页面模板

	transactions,_ := core.GetAllTransactionFromLDB()

	//tmpl.Execute(w,transactions)

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




