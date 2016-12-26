//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package state

import (
	"encoding/json"
	"fmt"

	"hyperchain/common"
)

type Account struct {
	Balance  string            `json:"balance"`
	Nonce    uint64            `json:"nonce"`
	Root     string            `json:"root"`
	CodeHash string            `json:"codeHash"`
	Code     string            `json:"code"`
	Abi      string            `json:"abi"`
	Storage  map[string]string `json:"storage"`
}

type World struct {
	Root     string             `json:"root"`
	Accounts map[string]Account `json:"accounts"`
}

func (self *StateDB) RawDump() World {
	world := World{
		Root:     common.Bytes2Hex(self.trie.Root()),
		Accounts: make(map[string]Account),
	}

	it := self.trie.Iterator()
	for it.Next() {
		addr := self.trie.GetKey(it.Key)
		stateObject, err := DecodeObject(common.BytesToAddress(addr), self.db, it.Value)
		if err != nil {
			panic(err)
		}

		account := Account{
			Balance:  stateObject.BalanceData.String(),
			Nonce:    stateObject.nonce,
			Root:     common.Bytes2Hex(stateObject.Root()),
			CodeHash: common.Bytes2Hex(stateObject.codeHash),
			Code:     common.Bytes2Hex(stateObject.Code(self.db)),
			Abi:      common.Bytes2Hex(stateObject.ABI()),
			Storage:  make(map[string]string),
		}
		storageIt := stateObject.trie.Iterator()
		for storageIt.Next() {
			account.Storage[common.Bytes2Hex(self.trie.GetKey(storageIt.Key))] = common.Bytes2Hex(storageIt.Value)
		}
		world.Accounts[common.Bytes2Hex(addr)] = account
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

// Debug stuff
func (self *StateObject) CreateOutputForDiff() {
	fmt.Printf("%x %x %x %x\n", self.Address(), self.Root(), self.BalanceData.Bytes(), self.nonce)
	it := self.trie.Iterator()
	for it.Next() {
		fmt.Printf("%x %x\n", it.Key, it.Value)
	}
}
