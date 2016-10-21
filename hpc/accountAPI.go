package hpc

import (
	"errors"
	"hyperchain/accounts"
	"hyperchain/common"
	"hyperchain/core"
	"hyperchain/core/state"
	"hyperchain/hyperdb"
	"hyperchain/manager"
	"fmt"
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
	Address  string `json:"address"`
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

//Unlock account according to args(address,password)
func (acc *PublicAccountAPI) UnlockAccount(args UnlockParas) error {
	password := string(args.Password)
	address := common.HexToAddress(args.Address)

	//keydir := "./keystore/"
	//encryption := crypto.NewEcdsaEncrypto("ecdsa")
	am := acc.pm.AccountManager

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
func (acc *PublicAccountAPI) GetAccounts() []*AccountResult {
	var acts []*AccountResult
	chain := core.GetChainCopy()

	headBlock, err := getBlockByHash(common.BytesToHash(chain.LatestBlockHash), acc.db)
	if err != nil {
		log.Errorf("%v", err)
	}

	stateDB, err := state.New(headBlock.MerkleRoot, acc.db)
	if err != nil {
		log.Errorf("Get stateDB error, %v", err)
	}
	ctx := stateDB.GetAccounts()

	for k, v := range ctx {
		log.Notice("balance is",v.Balance())
		var act = &AccountResult{
			Account: k,
			//Balance: fmt.Sprintf(`0x%x`, v.Balance()),
			Balance:v.Balance().String(),
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
