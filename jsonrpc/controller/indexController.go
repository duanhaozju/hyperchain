package controller
import (
	"net/http"
	"html/template"
	"hyperchain.cn/app/jsonrpc/model"
)

//type Transaction struct{
//	From string
//	To string
//	Value int
//	Time string
//}
//
//type Transacctions []Transaction

var path string = "./static/tmpl/"

// 处理请求 : GET "/"
func Index(w http.ResponseWriter, r *http.Request) {

	var tmpl = template.Must(template.ParseFiles(path+"index.html"))

	Transacctions,_ := model.GetAllTransaction()

	tmpl.Execute(w,Transacctions)
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




