package hyperstate

import (
	"math/big"
	"hyperchain/common"
	"fmt"
	"encoding/json"
	"github.com/syndtr/goleveldb/leveldb/errors"
)

type JournalEntry interface {
	//undo(*StateDB)
	String() string
	Marshal() ([]byte, error)
}

type Journal struct {
	JournalList []JournalEntry
}

func (self *Journal) Marshal() ([]byte, error) {
	var list [][]byte
	var err error
	for _, j := range self.JournalList {
		res, err := j.Marshal()
		if err != nil {
			break
		}
		list = append(list, res)
	}
	if err != nil {
		return nil, err
	} else {
		return json.Marshal(list)
	}
}

func UnmarshalJournal(data []byte) (*Journal, error) {

	var list [][]byte
	var jos []JournalEntry
	err := json.Unmarshal(data, &list)
	if err != nil {
		return nil, err
	}
	for _, res := range list {
		var jo interface{}
		err = json.Unmarshal(res, &jo)
		if err != nil {
			return nil, err
		}
		m := jo.(map[string]interface{})
		if len(m) == 1 {
			if m["Account"] != nil {
				var tmp CreateObjectChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["StateObj"] != nil {
				var tmp ResetObjectChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["Prev"] != nil {
				var tmp RefundChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["Txhash"] != nil {
				var tmp AddLogChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else {
				return nil, errors.New("Cannot unmarshal")
			}
		} else if len(m) == 2 {
			if m["BalanceAccount"] != nil && m["BalancePrev"] != nil {
				var tmp BalanceChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["NonceAccount"] != nil && m["NoncePrev"] != nil {
				var tmp NonceChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["TouchAccount"] != nil && m["TouchPrev"] != nil {
				var tmp TouchChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else {
				return nil, errors.New("Cannot unmarshal")
			}
		} else if len(m) == 3 {
			if m["Account"] != nil && m["Prev"] != nil && m["Prevbalance"] != nil {
				var tmp SuicideChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["Account"] != nil && m["Key"] != nil && m["Prevalue"] != nil {
				var tmp StorageChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else if m["Account"] != nil && m["Prevcode"] != nil && m["Prevhash"] != nil {
				var tmp CodeChange
				err = json.Unmarshal(res, &tmp)
				if err != nil {
					return nil, err
				}
				jos = append(jos, &tmp)
			} else {
				return nil, errors.New("Cannot unmarshal")
			}
		} else {
			return nil, errors.New("Cannot unmarshal")
		}
	}
	if err != nil {
		return nil, err
	} else {
		ret := &Journal{JournalList: jos}
		return ret, nil
	}
}

type (
	// Changes to the account database
	CreateObjectChange struct {
		Account *common.Address
	}
	ResetObjectChange struct {
		StateObj *StateObject
	}
	SuicideChange struct {
		Account     *common.Address
		Prev        bool // whether account had already suicided
		Prevbalance *big.Int
	}

	// Changes to individual accounts.
	BalanceChange struct {
		BalanceAccount *common.Address
		BalancePrev    *big.Int
	}
	NonceChange struct {
		NonceAccount *common.Address
		NoncePrev    uint64
	}
	StorageChange struct {
		Account       *common.Address
		Key, Prevalue common.Hash
	}
	CodeChange struct {
		Account            *common.Address
		Prevcode, Prevhash []byte
	}

	// Changes to other state values.
	RefundChange struct {
		Prev *big.Int
	}
	AddLogChange struct {
		Txhash common.Hash
	}
	TouchChange struct {
		TouchAccount *common.Address
		TouchPrev    bool
	}
)
// createObjectChange
//func (ch *CreateObjectChange) undo(s *StateDB) {
//	delete(s.stateObjects, *ch.Account)
//	delete(s.stateObjectsDirty, *ch.Account)
//}
func (ch *CreateObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [createObjectChange] %s\n", ch.Account.Hex())
	return str
}
func (ch *CreateObjectChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// resetObjectChange
//func (ch *resetObjectChange) undo(s *StateDB) {
//	s.setStateObject(ch.prev)
//}
func (ch *ResetObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [resetObjectChange] %s\n", ch.StateObj.String())
	return str
}

func (ch *ResetObjectChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
// suicideChange
//func (ch *suicideChange) undo(s *StateDB) {
//	obj := s.GetStateObject(*ch.account)
//	if obj != nil {
//		obj.suicided = ch.prev
//		obj.setBalance(ch.prevbalance)
//	}
//}
func (ch *SuicideChange) String() string {
	var str string
	str = fmt.Sprintf("journal [suicideChange] %s\n", ch.Account.Hex())
	return str
}
func (ch *SuicideChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// touchChange
//var ripemd = common.HexToAddress("0000000000000000000000000000000000000003")
//
//func (ch *TouchChange) undo(s *StateDB) {
//	if !ch.Prev && *ch.Account != ripemd {
//		delete(s.stateObjects, *ch.Account)
//		delete(s.stateObjectsDirty, *ch.Account)
//	}
//}
func (ch *TouchChange) String() string {
	var str string
	str = fmt.Sprintf("journal [touchChange] %s\n", ch.TouchAccount.Hex())
	return str
}
func (ch *TouchChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// balanceChange
//func (ch *balanceChange) undo(s *StateDB) {
//	s.GetStateObject(*ch.account).setBalance(ch.prev)
//}
func (ch *BalanceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [balanceChange] %s previous balance %s \n", ch.BalanceAccount.Hex(), ch.BalancePrev.String())
	return str
}
func (ch *BalanceChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// nonceChange
//func (ch *nonceChange) undo(s *StateDB) {
//	s.GetStateObject(*ch.account).setNonce(ch.prev)
//}

func (ch *NonceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [nonceChange] %s previous nonce %d \n", ch.NonceAccount.Hex(), ch.NoncePrev)
	return str

}
func (ch *NonceChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// codeChange
//func (ch *codeChange) undo(s *StateDB) {
//	s.GetStateObject(*ch.account).setCode(common.BytesToHash(ch.prevhash), ch.prevcode)
//}
func (ch *CodeChange) String() string {
	var str string
	str = fmt.Sprintf("journal [codeChange] %s previous codeHash %s \n", ch.Account.Hex(), common.BytesToHash(ch.Prevhash).Hex())
	return str
}
func (ch *CodeChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

//// storageChange
//func (ch *storageChange) undo(s *StateDB) {
//	s.GetStateObject(*ch.account).setState(ch.key, ch.prevalue)
//}
func (ch *StorageChange) String() string {
	var str string
	str = fmt.Sprintf("journal [storageChange] %s previous key %s  previous value %s \n", ch.Account.Hex(), ch.Key.Hex(), ch.Prevalue.Hex())
	return str
}
func (ch *StorageChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

//// refundChange
//func (ch *RefundChange) undo(s *StateDB) {
//	s.refund = ch.Prev
//}
func (ch *RefundChange) String() string {
	var str string
	str = fmt.Sprintf("journal [refundChange] previous value %s \n", ch.Prev.String())
	return str
}
func (ch *RefundChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// addLogChange
//func (ch *AddLogChange) undo(s *StateDB) {
//	logs := s.logs[ch.Txhash]
//	if len(logs) == 1 {
//		delete(s.logs, ch.Txhash)
//	} else {
//		s.logs[ch.Txhash] = logs[:len(logs)-1]
//	}
//}
func (ch *AddLogChange) String() string {
	var str string
	str = fmt.Sprintf("journal [addLogChange] tx hash %s \n", ch.Txhash.Hex())
	return str
}
func (ch *AddLogChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
