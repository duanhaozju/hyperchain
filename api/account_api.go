//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hpc

import (
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/manager"
	"hyperchain/hyperdb/db"
)

type PublicAccountAPI struct {
	pm     *manager.EventHub
	namespace string
	db     db.Database
	config *common.Config
}

type AccountResult struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}
type UnlockParas struct {
	Address  common.Address `json:"address"`
	Password string         `json:"password"`
}

func NewPublicAccountAPI(namespace string, pm *manager.EventHub, hyperDb db.Database, config *common.Config) *PublicAccountAPI {
	return &PublicAccountAPI{
		namespace: namespace,
		pm:     pm,
		db:     hyperDb,
		config: config,
	}
}

//New Account according to args from html
func (acc *PublicAccountAPI) NewAccount(password string) (common.Address, error) {
	//keydir := "./keystore/"
	//encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := acc.pm.AccountManager
	ac, err := am.NewAccount(password)
	if err != nil {
		log.Errorf("New Account error,%v", err)
		return common.Address{}, &CallbackError{err.Error()}
	}

	/*	balanceIns, err :=types.go.GetBalanceIns()
		balanceIns.PutCacheBalance(ac.Address, []byte("0"))
		balanceIns.PutDBBalance(ac.Address, []byte("0"))*/
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
		return false, &InvalidParamsError{"incorrect address or password!"}
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
			return "", &LeveldbNotFoundError{"stateobject, the account may not exist"}
		}
	} else {
		return "", &LeveldbNotFoundError{"statedb"}
	}
}
