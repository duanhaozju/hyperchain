package controller
import (
	"net/http"
	"html/template"
	"hyperchain/core"
	"path"
	"os"
	"hyperchain/core/types"
	"sort"
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
	Balances []types.Balance
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

	transactions,_ := hyperchain.GetAllTransactions()
	balances,_ := hyperchain.GetAllAccountBalances()

	transactions = append(transactions,core.GetTransactionsFromTxPool()...)
	sort.Sort(types.Transactions(transactions))

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

	for i,balance := range balances{
		pubKeyHash := balance.AccountPublicKeyHash
		for _,account := range accounts{
			for name,keypair := range account{
				if(encrypt.EncodePublicKey(&godpubkey) == pubKeyHash){
					balances[i].AccountPublicKeyHash = "god"
				}else if(encrypt.EncodePublicKey(&keypair.PubKey)== pubKeyHash){
					balances[i].AccountPublicKeyHash = name
				}
			}
		}
	}

	//tmpl.Execute(w,transactions)
	tmpl.Execute(w,data{
		Trans:transactions,
		Accounts:accounts,
		Balances: balances,
	})
}




