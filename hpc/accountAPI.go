package hpc

import (
	"errors"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/state"
	"hyperchain/hyperdb"
	"hyperchain/manager"
)

type PublicAccountAPI struct {
	pm *manager.ProtocolManager
}

type AccountResult struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}
type UnlockParas struct {
	Address  string
	Password string
}

func NewPublicAccountAPI(pm *manager.ProtocolManager) *PublicAccountAPI {
	return &PublicAccountAPI{
		pm: pm,
	}
}

//New Account according to args from html
func (acot *PublicAccountAPI) NewAccount(password string) common.Address {
	//keydir := "./keystore/"
	//encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := acot.pm.AccountManager
	ac, err := am.NewAccount(password)
	if err != nil {
		log.Fatal("New Account error,%v", err)
	}

	balanceIns, err := core.GetBalanceIns()
	balanceIns.PutCacheBalance(ac.Address, []byte("0"))
	balanceIns.PutDBBalance(ac.Address, []byte("0"))
	return ac.Address
}

//Unlock account according to args(address,password)
func (acot *PublicAccountAPI) UnlockAccount(args UnlockParas) error {
	password := string(args.Password)
	address := common.HexToAddress(args.Address)

	//keydir := "./keystore/"
	//encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := acot.pm.AccountManager

	s := string(args.Address)
	if len(s) > 1 {
		if s[0:2] == "0x" {
			s = s[2:]
		}
		if len(s)%2 == 1 {
			s = "0" + s
		}
	}
	ac := accounts.Account{Address: address, File: am.KeyStore.JoinPath(s)}
	err := am.Unlock(ac, password)
	if err != nil {
		return errors.New("Incorrect address or password!")
	}
	return nil
}

// GetAllBalances returns all account's balance in the db,NOT CACHE DB!
func (acot *PublicAccountAPI) GetAccounts() []*AccountResult {
	var acts []*AccountResult
	chain := core.GetChainCopy()
	db, err := hyperdb.GetLDBDatabase()
	if err != nil {
		log.Fatalf("Get DB error, %v", err)
	}
	headBlock, _ := core.GetBlock(db, chain.LatestBlockHash)
	stateDB, err := state.New(common.BytesToHash(headBlock.MerkleRoot), db)
	if err != nil {
		log.Fatalf("Get stateDB error, %v", err)
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
