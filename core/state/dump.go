//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
)

type User struct {
	Balance          string            `json:"balance"`
	Nonce            uint64            `json:"nonce"`
	Root             string            `json:"root"`
	CodeHash         string            `json:"codeHash"`
	Code             string            `json:"code"`
	Storage          map[string]string `json:"storage"`
	DeployedContract []string          `json:"contracts"`
	Creator          string            `json:"creator"`
	Status           string            `json:"status"`
	CreateTime       uint64            `json:"createTime"`
}

type World struct {
	Root  string          `json:"root"`
	Users map[string]User `json:"accounts"`
}

func (self *StateDB) RawDump() World {
	world := World{
		Users: make(map[string]User),
	}

	it := self.db.NewIterator([]byte(accountIdentifier))
	for it.Next() {
		address, ok := SplitCompositeAccountKey(it.Key())
		if ok == false {
			continue
		}
		var account Account
		err := Unmarshal(it.Value(), &account)
		if err != nil {
			continue
		}
		code, _ := self.db.Get(CompositeCodeHash(address, account.CodeHash))
		// code could by empty
		user := User{
			Balance:          account.Balance.String(),
			Nonce:            account.Nonce,
			Root:             account.Root.Hex(),
			CodeHash:         common.Bytes2Hex(account.CodeHash),
			Code:             common.Bytes2Hex(code),
			Storage:          make(map[string]string),
			Creator:          account.Creator.Hex(),
			DeployedContract: account.DeployedContracts,
			CreateTime:       account.CreateTime,
		}
		storageIt := self.db.NewIterator(GetStorageKeyPrefix(address))
		for storageIt.Next() {
			storageKey, _ := SplitCompositeStorageKey(address, storageIt.Key())
			self.logger.Debugf("dump key %s value %s", common.Bytes2Hex(storageKey), common.Bytes2Hex(storageIt.Value()))
			user.Storage[common.Bytes2Hex(storageKey)] = common.Bytes2Hex(storageIt.Value())
		}
		if account.Status == STATEOBJECT_STATUS_NORMAL {
			user.Status = "normal"
		} else if account.Status == STATEOBJECT_STATUS_FROZON {
			user.Status = "frozen"
		}
		world.Users[common.Bytes2Hex(address)] = user
	}
	return world
}

func (self *StateDB) Dump() []byte {
	json, err := json.MarshalIndent(self.RawDump(), "", "    ")
	if err != nil {
		fmt.Println("dump err", err)
	}
	return json
}
