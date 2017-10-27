// Copyright 2014 The go-ethereum Authors
// This file is part of the go-ethereum library.
//
// The go-ethereum library is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// The go-ethereum library is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with the go-ethereum library. If not, see <http://www.gnu.org/licenses/>.
package state

import (
	"encoding/json"
	"fmt"
	"github.com/hyperchain/hyperchain/common"
)

// User a user friendly account representation.
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

// World represents all accounts collection.
type World struct {
	Root  string          `json:"root"`
	Users map[string]User `json:"accounts"`
}

// RawDump gathers all account info and converts them to a user friendly representation format.
func (self *StateDB) RawDump() World {
	world := World{
		Users: make(map[string]User),
	}

	it := self.db.NewIterator([]byte(accountPrefix))
	for it.Next() {
		address, ok := splitCompositeAccountKey(it.Key())
		if ok == false {
			continue
		}
		var account Account
		err := Unmarshal(it.Value(), &account)
		if err != nil {
			continue
		}
		code, _ := self.db.Get(compositeCodeHash(address, account.CodeHash))
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
		storageIt := self.db.NewIterator(getStorageKeyPrefix(address))
		for storageIt.Next() {
			storageKey, _ := splitCompositeStorageKey(address, storageIt.Key())
			user.Storage[common.Bytes2Hex(storageKey)] = common.Bytes2Hex(storageIt.Value())
		}
		// assign status
		user.Status = "normal"
		if account.Status == OBJ_FROZON {
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
