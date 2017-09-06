//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/manager"
)

type Account struct {
	eh        *manager.EventHub
	namespace string
	config    *common.Config
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
	}
}

//New Account according to args from html
func (acc *Account) NewAccount(password string) (common.Address, error) {
	log := common.GetLogger(acc.namespace, "api")
	am := acc.eh.GetAccountManager()
	ac, err := am.NewAccount(password)
	if err != nil {
		log.Errorf("New Account error,%v", err)
		return common.Address{}, &common.CallbackError{Message: err.Error()}
	}
	return ac.Address, nil
}

// UnlockAccount unlocks account according to args(address,password), if success, return true.
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

// GetAllBalances returns all account's balance in the db,NOT CACHE DB!
func (acc *Account) GetAccounts() ([]*AccountResult, error) {
	log := common.GetLogger(acc.namespace, "api")
	var acts []*AccountResult
	stateDB, err := NewStateDb(acc.config, acc.namespace)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
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
			return "", &common.LeveldbNotFoundError{Message:"stateobject, the account may not exist"}
		}
	}

}
