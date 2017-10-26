package versionAccount

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/core/state"
	"math/big"
	"strings"
)

type Account struct {
	Nonce             uint64         `json:"nonce"`
	Balance           *big.Int       `json:"balance"`
	Root              common.Hash    `json:"merkleRoot"`
	CodeHash          []byte         `json:"codeHash"`
	DeployedContracts []string       `json:"contracts"`
	Creator           common.Address `json:"creator"`
	Status            int            `json:"status"`
	CreateTime        uint64         `json:"createTime"`
}

type AccountView struct {
	Nonce             uint64
	Balance           *big.Int
	Root              string
	CodeHash          string
	DeployedContracts []string
	Creator           string
	Status            string
	CreateTime        uint64
}

type AccountVerboseView struct {
	Nonce             uint64
	Balance           *big.Int
	Root              string
	CodeHash          string
	DeployedContracts []string
	Creator           string
	Status            string
	CreateTime        uint64
	Code              string
	Storage           map[string]string
}

func (account *Account) Encode(address string) string {
	var accounts = make(map[string]AccountView)
	var accountView = AccountView{
		Nonce:             account.Nonce,
		Balance:           account.Balance,
		Root:              account.Root.Hex(),
		CodeHash:          common.Bytes2Hex(account.CodeHash),
		DeployedContracts: account.DeployedContracts,
		Creator:           account.Creator.Hex(),
		CreateTime:        account.CreateTime,
	}
	if account.Status == state.STATEOBJECT_STATUS_NORMAL {
		accountView.Status = "normal"
	} else if account.Status == state.STATEOBJECT_STATUS_FROZON {
		accountView.Status = "frozen"
	}
	accounts[address] = accountView
	res, err := json.MarshalIndent(accounts, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	return string(res)
}

func (account *Account) EncodeVerbose(address string, code string) string {
	var accounts = make(map[string]AccountVerboseView)
	var accountVerboseView = AccountVerboseView{
		Nonce:             account.Nonce,
		Balance:           account.Balance,
		Root:              account.Root.Hex(),
		CodeHash:          common.Bytes2Hex(account.CodeHash),
		DeployedContracts: account.DeployedContracts,
		Creator:           account.Creator.Hex(),
		CreateTime:        account.CreateTime,
		Code:              code,
	}
	if account.Status == state.STATEOBJECT_STATUS_NORMAL {
		accountVerboseView.Status = "normal"
	} else if account.Status == state.STATEOBJECT_STATUS_FROZON {
		accountVerboseView.Status = "frozen"
	}
	accountVerboseView.Storage = make(map[string]string)
	accounts[address] = accountVerboseView
	res, err := json.MarshalIndent(accounts, "", "\t")
	if err != nil {
		fmt.Println(err.Error())
		return ""
	}
	str := string(res)
	return str[:strings.Index(str, "\"Storage\"")+len("\"Storage\": {")]
}
