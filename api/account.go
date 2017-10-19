//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/manager"
)

// This file implements the handler of Account service API which
// can be invoked by client in JSON-RPC request.

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

// NewPublicAccountAPI creates and returns a new Account instance for given namespace name.
func NewPublicAccountAPI(namespace string, eh *manager.EventHub, config *common.Config) *Account {
	return &Account{
		namespace: namespace,
		eh:        eh,
		config:    config,
		log:       common.GetLogger(namespace, "api"),
	}
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
			return "", &common.AccountNotExistError{Address: addr.Hex()}
		}
	}

}
