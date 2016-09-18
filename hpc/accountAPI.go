package hpc

import (
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/crypto"
	"hyperchain/accounts"
)

type PublicAccountAPI struct {}

type AccountResult struct {
	Account common.Address `json:"account"`
	Balance string         `json:"balance"`
}

func NewPublicAccountAPI() *PublicAccountAPI {
	return &PublicAccountAPI{}
}

//New Account according to args from html
func (acot *PublicAccountAPI)NewAccount(password string) common.Address  {
	keydir := "./keystore/"
	encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := accounts.NewAccountManager(keydir,encryption)
	ac,err :=am.NewAccount(password)
	if err !=nil{
		log.Fatal("New Account error,%v",err)
	}

	balanceIns, err := core.GetBalanceIns()
	balanceIns.PutCacheBalance(ac.Address,[]byte("0"))
	balanceIns.PutDBBalance(ac.Address,[]byte("0"))
	return ac.Address
}
// GetAllBalances returns all account's balance in the db,NOT CACHE DB!
func (acot *PublicAccountAPI) GetAccounts() []*AccountResult{
	var acts []*AccountResult

	balanceIns, err := core.GetBalanceIns()

	if err != nil {
		log.Fatalf("GetBalanceIns error, %v", err)
	}

	balMap := balanceIns.GetAllDBBalance()

	for key, value := range balMap {

		var act = &AccountResult{
			Account: key,
			Balance: string(value),
		}

		acts = append(acts, act)
	}

	return acts
}