//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package hyperstate

import (
	"encoding/json"
	"fmt"
	"hyperchain/common"
	"hyperchain/hyperdb"
)

type User struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Storage  map[string]string `json:"storage"`
}

type World struct {
	Root     string             `json:"root"`
	Users map[string]User 	    `json:"accounts"`
}

func (self *StateDB) RawDump() World {
	world := World{
		Root:     self.root.Hex(),
		Users:    make(map[string]User),
	}

	leveldb, ok := self.db.(*hyperdb.LDBDatabase)
	if ok == false {
		return world
	}
	it := leveldb.NewIteratorWithPrefix([]byte(accountIdentifier))
	for it.Next() {
		address, ok := SplitCompositeAccountKey(it.Key())
		if ok == false {
			continue
		}
		var account Account
		err := UnmarshalJSON(it.Value(), &account)
		if err != nil {
			continue
		}
		code, _ := self.db.Get(CompositeCodeHash(address, account.CodeHash))
		// code could by empty
		user := User{
			Balance:  account.Balance.String(),
			Nonce:    account.Nonce,
			Root:     account.Root.Hex(),
			CodeHash: common.Bytes2Hex(account.CodeHash),
			Code:     common.Bytes2Hex(code),
			Storage:  make(map[string]string),
		}
		storageIt := leveldb.NewIteratorWithPrefix(GetStorageKeyPrefix(address))
		for storageIt.Next() {
			storageKey, _ := SplitCompositeStorageKey(address, it.Key())
			user.Storage[common.BytesToHash(storageKey).Hex()] = common.BytesToHash(storageIt.Value()).Hex()
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

