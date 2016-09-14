package hpc

import (
	"hyperchain/common"
	"hyperchain/core"
)

type PublicAccountAPI struct {}

type AccountResult struct {
	Account common.Address `json:"account"`
	Balance string         `json:"balance"`
}

func NewPublicAccountAPI() *PublicAccountAPI {
	return &PublicAccountAPI{}
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