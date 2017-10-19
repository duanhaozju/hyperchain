package api

import (
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
)

type AccountDB struct {
	namespace string
	config    *common.Config
	log       *logging.Logger
}

type AccountResult struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}

func NewPublicAccountExecutorAPI(namespace string, config *common.Config) *AccountDB {
	return &Account{
		namespace: namespace,
		config:    config,
		log:       common.GetLogger(namespace, "api"),
	}
}

// GetAccounts returns all account's balance in the ledger.
func (acc *AccountDB) GetAccounts() ([]*AccountResult, error) {
	var acts []*AccountResult
	stateDB, err := NewStateDb(acc.config, acc.namespace)
	if err != nil {
		acc.log.Errorf("Get stateDB error, %v", err)
		return nil, &common.CallbackError{Message: err.Error()}
	}
	ctx := stateDB.GetAccounts()

	for k, v := range ctx {
		var act = &AccountResult{
			Account: k,
			Balance: v.Balance().String(),
		}
		acts = append(acts, act)
	}
	return acts, nil
}

// GetBalance returns account balance for given account address.
func (acc *AccountDB) GetBalance(addr common.Address) (string, error) {
	if stateDB, err := NewStateDb(acc.config, acc.namespace); err != nil {
		return "", &common.CallbackError{Message: err.Error()}
	} else {
		if stateobject := stateDB.GetAccount(addr); stateobject != nil {
			return fmt.Sprintf(`0x%x`, stateobject.Balance()), nil
		} else {
			return "", &common.AccountNotExistError{Address: addr.Hex()}
		}
	}

}
