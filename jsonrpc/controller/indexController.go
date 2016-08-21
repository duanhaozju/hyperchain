package controller
import (
	"net/http"
	"html/template"
	"hyperchain-alpha/core"
	"path"
	"os"
	"hyperchain-alpha/utils"
	"hyperchain-alpha/core/types"
	"hyperchain-alpha/encrypt"
	"log"
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

	godAccounts := utils.GetGodAccount()
	godpubkey := godAccounts[0]["god"].PubKey
	transactions,_ := core.GetAllTransactionFromLDB()


	for i,tx := range transactions{
		from := tx.From
		to := tx.To
		for _,account := range accounts{
			for name,keypair := range account{
				if(encrypt.EncodePublicKey(&godpubkey) == from){
					transactions[i].From = "god"
				}else if(encrypt.EncodePublicKey(&keypair.PubKey)== from){
					transactions[i].From = name
				}
				if(encrypt.EncodePublicKey(&keypair.PubKey) == to){
					transactions[i].To = name
				}
			}
		}
	}

	//tmpl.Execute(w,transactions)
	tmpl.Execute(w,data{
		Trans:transactions,
		Accounts:accounts,
	})
}




