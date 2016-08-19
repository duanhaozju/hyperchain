package hyper

import "time"

type TxArgs struct{
	From string `json:"from"`
	To string `json:"to"`
	Value string `json:"value"`
	Timestamp time.Time `json:"timestamp"`
}

type TransactionPoolAPI struct{

}

//func newTransaction() t model.Transaction{
//
//}

func (t *TransactionPoolAPI) SendTransaction(args TxArgs) error {



	//newTransaction()
	return nil
}
