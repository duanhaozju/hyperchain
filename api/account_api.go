//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/manager"
)

type PublicAccountAPI struct {
	pm        *manager.EventHub
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

func NewPublicAccountAPI(namespace string, pm *manager.EventHub, config *common.Config) *PublicAccountAPI {
	return &PublicAccountAPI{
		namespace: namespace,
		pm:        pm,
		config:    config,
	}
}

//New Account according to args from html
func (acc *PublicAccountAPI) NewAccount(password string) (common.Address, error) {
	am := acc.pm.AccountManager
	ac, err := am.NewAccount(password)
	if err != nil {
		log.Errorf("New Account error,%v", err)
		return common.Address{}, &common.CallbackError{err.Error()}
	}
	return ac.Address, nil
}

// UnlockAccount unlocks account according to args(address,password), if success, return true.
func (acc *PublicAccountAPI) UnlockAccount(args UnlockParas) (bool, error) {

	am := acc.pm.AccountManager

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
		return false, &common.InvalidParamsError{"incorrect address or password!"}
	}
	return true, nil
}

// GetAllBalances returns all account's balance in the db,NOT CACHE DB!
func (acc *PublicAccountAPI) GetAccounts() []*AccountResult {
	var acts []*AccountResult
	stateDB, err := NewStateDb(acc.config, acc.namespace)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return nil
	}
	ctx := stateDB.GetAccounts()

	for k, v := range ctx {
		var act = &AccountResult{
			Account: k,
			Balance: v.Balance().String(),
		}
		acts = append(acts, act)
	}
	return acts
}

// GetBalance returns account balance for given account address.
func (acc *PublicAccountAPI) GetBalance(addr common.Address) (string, error) {
	if stateDB, err := NewStateDb(acc.config, acc.namespace); err != nil {
		if stateobject := stateDB.GetAccount(addr); stateobject != nil {
			return fmt.Sprintf(`0x%x`, stateobject.Balance()), nil
		} else {
			return "", &common.LeveldbNotFoundError{"stateobject, the account may not exist"}
		}
	} else {
		return "", &common.LeveldbNotFoundError{"statedb"}
	}
}
