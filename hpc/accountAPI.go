package hpc

import (
	"errors"
	"fmt"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/state"
	"hyperchain/hyperdb"
	"hyperchain/manager"
)

type PublicAccountAPI struct {
	pm *manager.ProtocolManager
	db *hyperdb.LDBDatabase
}

type AccountResult struct {
	Account string `json:"account"`
	Balance string `json:"balance"`
}
type UnlockParas struct {
	Address  common.Address `json:"address"`
	Password string `json:"password"`
}

func NewPublicAccountAPI(pm *manager.ProtocolManager, hyperDb *hyperdb.LDBDatabase) *PublicAccountAPI {
	return &PublicAccountAPI{
		pm: pm,
		db: hyperDb,
	}
}

//New Account according to args from html
func (acc *PublicAccountAPI) NewAccount(password string) common.Address {
	//keydir := "./keystore/"
	//encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := acc.pm.AccountManager
	ac, err := am.NewAccount(password)
	if err != nil {
		log.Fatal("New Account error,%v", err)
	}

	/*	balanceIns, err := core.GetBalanceIns()
		balanceIns.PutCacheBalance(ac.Address, []byte("0"))
		balanceIns.PutDBBalance(ac.Address, []byte("0"))*/
	return ac.Address
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
		return false, errors.New("Incorrect address or password!")
	}
	return true, nil
}

// GetAllBalances returns all account's balance in the db,NOT CACHE DB!
func (acc *PublicAccountAPI) GetAccounts() []*AccountResult {
	var acts []*AccountResult
	chain := core.GetChainCopy()

	log.Notice("Current LatestBlockHash:", common.BytesToHash(chain.LatestBlockHash).Hex())
	headBlock, err := getBlockByHash(common.BytesToHash(chain.LatestBlockHash), acc.db)
	if err != nil {
		log.Errorf("%v", err)
		return nil
	}

	stateDB, err := state.New(headBlock.MerkleRoot, acc.db)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return nil
	}
	ctx := stateDB.GetAccounts()

	for k, v := range ctx {
		log.Notice("balance is", v.Balance())
		var act = &AccountResult{
			Account: k,
			//Balance: fmt.Sprintf(`0x%x`, v.Balance()),
			Balance: v.Balance().String(),
		}
		acts = append(acts, act)
	}
	return acts
}

// GetBalance returns account balance for given account address.
func (acc *PublicAccountAPI) GetBalance(addr common.Address) (string, error) {

	headBlock, _ := core.GetBlock(acc.db, core.GetChainCopy().LatestBlockHash)
	stateDB, err := state.New(common.BytesToHash(headBlock.MerkleRoot), acc.db)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
		return "", err
	}

	return fmt.Sprintf(`0x%x`, stateDB.GetStateObject(addr).BalanceData), nil
}
