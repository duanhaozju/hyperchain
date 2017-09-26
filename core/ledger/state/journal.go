// Copyright 2016-2017 Hyperchain Corp.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
package state

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"hyperchain/common"
	"hyperchain/hyperdb/db"
	"math/big"
	"sort"
)

// journal type name definition
const (
	CreateObjectChangeType     = "CreateObjectChange"
	ResetObjectChangeType      = "ResetObjectChange"
	SuicideChangeType          = "SuicideChange"
	BalanceChangeType          = "BalanceChange"
	NonceChangeType            = "NonceChange"
	StorageChangeType          = "StorageChange"
	CodeChangeType             = "CodeChange"
	AddLogChangeType           = "AddLogChange"
	StorageHashChangeType      = "StorageHashChange"
	StatusChangeType           = "StatusChange"
	DeployedContractChangeType = "DeployedContractChange"
	SetCreatorChangeType       = "SetCreatorChange"
	SetCreateTimeChangeType    = "SetCreateTimeChange"
)

type (
	// Changes to the account database
	CreateObjectChange struct {
		Account *common.Address `json:"account,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	ResetObjectChange struct {
		Prev *StateObject `json:"prev,omitempty"`
		Type string       `json:"type,omitempty"`
	}
	SuicideChange struct {
		Account     *common.Address `json:"account,omitempty"`
		Prev        bool            `json:"prev,omitempty"` // whether account had already suicided
		Prevbalance *big.Int        `json:"prevbalance,omitempty"`
		PreObject   *StateObject    `json:"preObject,omitempty"`
		Type        string          `json:"type,omitempty"`
	}
	// Changes to individual accounts.
	BalanceChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    *big.Int        `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	NonceChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    uint64          `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	StorageChange struct {
		Account    *common.Address `json:"account,omitempty"`
		Key        common.Hash     `json:"key,omitempty"`
		Prevalue   []byte          `json:"prevalue,omitempty"`
		Exist      bool            `json:"exist,omitempty"`
		LastModify uint64          `json:"brith,omitempty"`
		Type       string          `json:"type,omitempty"`
	}
	CodeChange struct {
		Account  *common.Address `json:"account,omitempty"`
		Prevcode []byte          `json:"prevcode,omitempty"`
		Prevhash []byte          `json:"prevhash,omitempty"`
		Type     string          `json:"type,omitempty"`
	}
	StatusChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    int             `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	DeployedContractChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    *common.Address `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	SetCreatorChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    common.Address  `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	SetCreateTimeChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    uint64          `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	AddLogChange struct {
		Txhash common.Hash `json:"txhash,omitempty"`
		Type   string      `json:"type,omitempty"`
	}
	StorageHashChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    []byte          `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
)

type JournalEntry interface {
	Undo(*StateDB, *JournalCache, db.Batch, bool)
	String() string
	Marshal() ([]byte, error)
	SetType()
	GetType() string
}

type Journal struct {
	JournalList []JournalEntry
}

type MemJournal struct {
	JournalList [][]byte
}

func (self *Journal) Marshal() ([]byte, error) {
	var list [][]byte
	// marshal all journal one by once
	for _, j := range self.JournalList {
		// IMPORTANT to set change type before marshal function called, otherwise unmarshal will crush down
		j.SetType()
		res, err := j.Marshal()
		if err != nil {
			break
		}
		list = append(list, res)
	}
	return json.Marshal(MemJournal{
		JournalList: list,
	})
}

func UnmarshalJournal(data []byte) (*Journal, error) {
	memJournal := &MemJournal{}
	err := json.Unmarshal(data, memJournal)
	if err != nil {
		return nil, err
	}
	list := memJournal.JournalList
	var jos []JournalEntry
	for _, res := range list {
		var jo interface{}
		err = json.Unmarshal(res, &jo)
		if err != nil {
			return nil, err
		}
		m := jo.(map[string]interface{})
		switch m["type"] {
		case CreateObjectChangeType:
			var tmp CreateObjectChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case ResetObjectChangeType:
			var tmp ResetObjectChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case AddLogChangeType:
			var tmp AddLogChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case BalanceChangeType:
			var tmp BalanceChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case NonceChangeType:
			var tmp NonceChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case SuicideChangeType:
			var tmp SuicideChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case StorageChangeType:
			var tmp StorageChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case CodeChangeType:
			var tmp CodeChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case StorageHashChangeType:
			var tmp StorageHashChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case StatusChangeType:
			var tmp StatusChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case DeployedContractChangeType:
			var tmp DeployedContractChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case SetCreatorChangeType:
			var tmp SetCreatorChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case SetCreateTimeChangeType:
			var tmp SetCreateTimeChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		default:
			return nil, errors.New("unmarshal journal failed")
		}
	}

	ret := &Journal{
		JournalList: jos,
	}
	return ret, nil
}

// createObjectChange
func (ch *CreateObjectChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		delete(s.stateObjects, *ch.Account)
		delete(s.stateObjectsDirty, *ch.Account)
	} else {
		obj := cache.Fetch(*ch.Account)
		if obj == nil {
			s.logger.Warningf("missing state object %s, it may be a empty account or lost in database", ch.Account.Hex())
			return
		}
		obj.suicided = true
	}
}
func (ch *CreateObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [createObjectChange] %s\n", ch.Account.Hex())
	return str
}
func (ch *CreateObjectChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *CreateObjectChange) SetType() {
	ch.Type = CreateObjectChangeType
}
func (ch *CreateObjectChange) GetType() string {
	return ch.Type
}

// resetObjectChange
func (ch *ResetObjectChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		s.setStateObject(ch.Prev)
	} else {
		cache.Add(ch.Prev)
	}
}
func (ch *ResetObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [resetObjectChange] %s\n", ch.Prev.String())
	return str
}

func (ch *ResetObjectChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *ResetObjectChange) SetType() {
	ch.Type = ResetObjectChangeType
}
func (ch *ResetObjectChange) GetType() string {
	return ch.Type
}

// suicideChange
func (ch *SuicideChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		// undo contract account
		if ch.Prev == true {
			return
		} else {
			obj := ch.PreObject
			obj.suicided = ch.Prev
			obj.data.Balance = ch.Prevbalance
			s.setStateObject(obj)
		}
	} else {
		if ch.Prev == true {
			return
		} else {
			obj := ch.PreObject
			cache.Add(obj)
		}
	}
}
func (ch *SuicideChange) String() string {
	var str string
	str = fmt.Sprintf("journal [suicideChange] %s  %#v  %d\n", ch.Account.Hex(), ch.Prev, ch.Prevbalance)
	return str
}
func (ch *SuicideChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *SuicideChange) SetType() {
	ch.Type = SuicideChangeType
}
func (ch *SuicideChange) GetType() string {
	return ch.Type
}

// balanceChange
func (ch *BalanceChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			obj.setBalance(ch.Prev)
		}
	} else {
		obj := cache.Fetch(*ch.Account)
		if obj == nil {
			s.logger.Warningf("missing state object %s, it may be a empty account or lost in database", ch.Account.Hex())
			obj = cache.Create(*ch.Account, s)
		}
		obj.data.Balance = ch.Prev
	}
}
func (ch *BalanceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [balanceChange] %s %#v\n", ch.Account.Hex(), ch.Prev)
	return str
}
func (ch *BalanceChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *BalanceChange) SetType() {
	ch.Type = BalanceChangeType
}
func (ch *BalanceChange) GetType() string {
	return ch.Type
}

// nonceChange
func (ch *NonceChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			obj.setNonce(ch.Prev)
		}
	} else {
		obj := cache.Fetch(*ch.Account)
		if obj == nil {
			s.logger.Warningf("missing state object %s, it may be a empty account or lost in database", ch.Account.Hex())
			obj = cache.Create(*ch.Account, s)
		}
		obj.data.Nonce = ch.Prev
	}
}

func (ch *NonceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [nonceChange] %s %d \n", ch.Account.Hex(), ch.Prev)
	return str

}
func (ch *NonceChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *NonceChange) SetType() {
	ch.Type = NonceChangeType
}
func (ch *NonceChange) GetType() string {
	return ch.Type
}

// codeChange
func (ch *CodeChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			obj.setCode(common.BytesToHash(ch.Prevhash), ch.Prevcode)
		}
	} else {
		obj := cache.Fetch(*ch.Account)
		if obj == nil {
			s.logger.Warningf("missing state object %s, it may be a empty account or lost in database", ch.Account.Hex())
			obj = cache.Create(*ch.Account, s)
		}
		batch.Delete(CompositeCodeHash(ch.Account.Bytes(), obj.data.CodeHash))
		obj.data.CodeHash = ch.Prevhash
		batch.Put(CompositeCodeHash(ch.Account.Bytes(), ch.Prevhash), ch.Prevcode)
	}
}
func (ch *CodeChange) String() string {
	var str string
	str = fmt.Sprintf("journal [codeChange] %s codeHash %s code %s\n", ch.Account.Hex(), common.BytesToHash(ch.Prevhash).Hex(), common.Bytes2Hex(ch.Prevcode))
	return str
}
func (ch *CodeChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *CodeChange) SetType() {
	ch.Type = CodeChangeType
}
func (ch *CodeChange) GetType() string {
	return ch.Type
}

// storageChange
func (ch *StorageChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if ch.Exist {
			if obj := s.GetStateObject(*ch.Account); obj != nil {
				obj.setState(ch.Key, ch.Prevalue)
				obj.evictList[ch.Key] = ch.LastModify
			}
		} else {
			if obj := s.GetStateObject(*ch.Account); obj != nil {
				obj.removeState(ch.Key)
			}
		}
	} else {
		obj := cache.Fetch(*ch.Account)
		if obj == nil {
			// should never happen
			s.logger.Warningf("missing state object %s, it should not happen when undo storage change", ch.Account.Hex())
			return
		}
		obj.cachedStorage[ch.Key] = ch.Prevalue
		obj.dirtyStorage[ch.Key] = ch.Prevalue
	}
}
func (ch *StorageChange) String() string {
	var str string
	str = fmt.Sprintf("journal [storageChange] %s previous key %s  previous value %s \n", ch.Account.Hex(), ch.Key.Hex(), common.Bytes2Hex(ch.Prevalue))
	return str
}
func (ch *StorageChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *StorageChange) SetType() {
	ch.Type = StorageChangeType
}
func (ch *StorageChange) GetType() string {
	return ch.Type
}

// addLogChange
func (ch *AddLogChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	logs := s.logs[ch.Txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.Txhash)
	} else {
		s.logs[ch.Txhash] = logs[:len(logs)-1]
	}
}
func (ch *AddLogChange) String() string {
	var str string
	str = fmt.Sprintf("journal [addLogChange] tx hash %s \n", ch.Txhash.Hex())
	return str
}
func (ch *AddLogChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *AddLogChange) SetType() {
	ch.Type = AddLogChangeType
}
func (ch *AddLogChange) GetType() string {
	return ch.Type
}

// StorageHashChange
func (ch *StorageHashChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {

	} else {
		obj := cache.Fetch(*ch.Account)
		if obj == nil {
			// should never happen
			s.logger.Warningf("missing state object %s, it should not happen when undo storage hash change", ch.Account.Hex())
			return
		}
		obj.data.Root = common.BytesToHash(ch.Prev)
	}
}
func (ch *StorageHashChange) String() string {
	var str string
	str = fmt.Sprintf("journal [storagehashChange] address %s prev %s\n", ch.Account.Hex(), common.Bytes2Hex(ch.Prev))
	return str
}
func (ch *StorageHashChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *StorageHashChange) SetType() {
	ch.Type = StorageHashChangeType
}
func (ch *StorageHashChange) GetType() string {
	return ch.Type
}

// StatusChange
func (ch *StatusChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			obj.setStatus(ch.Prev)
		}
	} else {
		if obj := cache.Fetch(*ch.Account); obj != nil {
			obj.setStatus(ch.Prev)
		}
	}
}

func (ch *StatusChange) String() string {
	var str string
	var status string
	if ch.Prev == OBJ_NORMAL {
		status = "normal"
	} else if ch.Prev == OBJ_FROZON {
		status = "frozen"
	} else {
		status = "Undefined"
	}
	str = fmt.Sprintf("journal [StatusChange] address %s prev %s\n", ch.Account.Hex(), status)
	return str
}

func (ch *StatusChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}

func (ch *StatusChange) SetType() {
	ch.Type = StatusChangeType
}
func (ch *StatusChange) GetType() string {
	return ch.Type
}

// DeployedContractChange
func (ch *DeployedContractChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			success := obj.removeDeployedContract(*ch.Prev)
			if !success {
				s.logger.Errorf("miss contract %s in deployed contract list in creator %s", ch.Prev.Hex(), ch.Account.Hex())
			}
		}
	} else {
		obj := cache.Fetch(*ch.Account)
		remove := func(contracts []string, address common.Address) []string {
			if len(contracts) == 0 {
				return contracts
			}
			if idx := sort.SearchStrings(contracts, address.Hex()); idx < len(contracts) && contracts[idx] == address.Hex() {
				contracts = append(contracts[:idx], contracts[idx+1:]...)
				if len(contracts) == 0 {
					contracts = nil
				} else {
					sort.Strings(contracts)
				}
			}
			return contracts
		}
		if obj != nil {
			obj.data.DeployedContracts = remove(obj.data.DeployedContracts, *ch.Prev)
		}
	}
}
func (ch *DeployedContractChange) String() string {
	var str string
	str = fmt.Sprintf("journal [DeployedContractChange] address %s prev %s\n", ch.Account.Hex(), ch.Prev.Hex())
	return str
}
func (ch *DeployedContractChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *DeployedContractChange) SetType() {
	ch.Type = DeployedContractChangeType
}
func (ch *DeployedContractChange) GetType() string {
	return ch.Type
}

// SetCreatorChange
func (ch *SetCreatorChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			obj.setCreator(ch.Prev)
		}
	} else {
		if obj := cache.Fetch(*ch.Account); obj != nil {
			obj.setCreator(ch.Prev)
		}
	}
}

func (ch *SetCreatorChange) String() string {
	var str string
	str = fmt.Sprintf("journal [SetCreatorChange] address %s prev %s\n", ch.Account.Hex(), ch.Prev.Hex())
	return str
}
func (ch *SetCreatorChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *SetCreatorChange) SetType() {
	ch.Type = SetCreatorChangeType
}
func (ch *SetCreatorChange) GetType() string {
	return ch.Type
}

// SetCreateTimeChange
func (ch *SetCreateTimeChange) Undo(s *StateDB, cache *JournalCache, batch db.Batch, writeThrough bool) {
	if !writeThrough {
		if obj := s.GetStateObject(*ch.Account); obj != nil {
			obj.setCreateTime(ch.Prev)
		}
	} else {
		if obj := cache.Fetch(*ch.Account); obj != nil {
			obj.setCreateTime(ch.Prev)
		}
	}
}

func (ch *SetCreateTimeChange) String() string {
	var str string
	str = fmt.Sprintf("journal [SetCreateTimeChange] address %s prev %d\n", ch.Account.Hex(), ch.Prev)
	return str
}
func (ch *SetCreateTimeChange) Marshal() ([]byte, error) {
	return json.Marshal(ch)
}
func (ch *SetCreateTimeChange) SetType() {
	ch.Type = SetCreateTimeChangeType
}
func (ch *SetCreateTimeChange) GetType() string {
	return ch.Type
}
