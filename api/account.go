//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/manager"
	"github.com/op/go-logging"
)

type Account struct {
	eh        *manager.EventHub
	namespace string
	config    *common.Config
	log       *logging.Logger
}

type AccountResult struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}
type UnlockParas struct {
	Address  common.Address `json:"address"`
	Password string         `json:"password"`
}

func NewPublicAccountAPI(namespace string, eh *manager.EventHub, config *common.Config) *Account {
	return &Account{
		namespace: namespace,
		eh:        eh,
		config:    config,
		log:	   common.GetLogger(namespace, "api"),
	}
}

// NewAccount creates a new account under the given password.
func (acc *Account) NewAccount(password string) (common.Address, error) {
	am := acc.eh.GetAccountManager()
	ac, err := am.NewAccount(password)
	if err != nil {
		acc.log.Errorf("New Account error, %v", err)
		return common.Address{}, &common.CallbackError{Message: err.Error()}
	}
	return ac.Address, nil
}

// UnlockAccount unlocks account for given account address and password, if success, return true.
func (acc *Account) UnlockAccount(args UnlockParas) (bool, error) {

	am := acc.eh.GetAccountManager()

	s := args.Address.Hex()
	if len(s) > 1 {
		if s[0:2] == "0x" {
			s = s[2:]
		}
		if len(s)%2 == 1 {
			s = "0" + s
		}
	}
	ac := accounts.Account{Address: args.Address, File: am.KeyStore.JoinPath(s)}
	err := am.Unlock(ac, args.Password)
	if err != nil {
		return false, &common.InvalidParamsError{Message: "incorrect address or password!"}
	}
	return true, nil
}

// GetAccounts returns all account's balance in the ledger.
func (acc *Account) GetAccounts() ([]*AccountResult, error) {
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
func (acc *Account) GetBalance(addr common.Address) (string, error) {
	if stateDB, err := NewStateDb(acc.config, acc.namespace); err != nil {
		return "", &common.CallbackError{Message: err.Error()}
	} else {
		if stateobject := stateDB.GetAccount(addr); stateobject != nil {
			return fmt.Sprintf(`0x%x`, stateobject.Balance()), nil
		} else {
			return "", &common.LeveldbNotFoundError{Message: "stateobject, the account may not exist"}
		}
	}

}
