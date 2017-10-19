//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package api

import (
	"github.com/op/go-logging"
	"hyperchain/accounts"
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
		return false, &common.InvalidParamsError{Message: "Incorrect address or password!"}
	}
	return true, nil
}


