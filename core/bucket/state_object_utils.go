package bucket

import (
	"encoding/json"
	"bytes"
	"fmt"
	"math/big"
	"hyperchain/common"
	"hyperchain/hyperdb"
)
type extStateObject struct {
	BalanceData *big.Int
	Root        common.Hash
	CodeHash    []byte
	Nonce       uint64
}

// TODO  should be test
func EncodeObject(stateObject StateObject) ([]byte, error) {
	ext := extStateObject{
		BalanceData: stateObject.BalanceData,
		CodeHash:    stateObject.codeHash,
		Nonce:       stateObject.nonce,
	}
	//self.trie.CommitTo(self.db)
	//self.db.Put(self.codeHash, self.code)
	return json.Marshal(ext)
}
// TODO  should be test
func DecodeObject(address common.Address, db hyperdb.Database, data []byte) (*StateObject, error) {
	var (
		obj = &StateObject{
			address: address,
			db:      db,
			storage: make(Storage),
		}
		ext extStateObject
		err error
	)
	err = json.Unmarshal(data, &ext)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(ext.CodeHash, emptyCodeHash) {
		if obj.code, err = db.Get(ext.CodeHash); err != nil {
			return nil, fmt.Errorf("can't get code for hash %x: %v", ext.CodeHash, err)
		}
	}
	obj.BalanceData = ext.BalanceData
	obj.codeHash = ext.CodeHash
	obj.nonce = ext.Nonce
	return obj, nil
}
