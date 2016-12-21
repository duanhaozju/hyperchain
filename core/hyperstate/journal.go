package hyperstate

import (
	"math/big"
	"hyperchain/common"
	"fmt"
	"encoding/json"
	"github.com/pkg/errors"
)

// journal type name definition
const (
	CreateObjectChangeType = "CreateObjectChange"
	ResetObjectChangeType = "ResetObjectChange"
	SuicideChangeType = "SuicideChange"
	BalanceChangeType = "BalanceChange"
	NonceChangeType = "NonceChange"
	StorageChangeType = "StorageChange"
	CodeChangeType = "CodeChange"
	RefundChangeType = "RefundChange"
	AddLogChangeType = "AddLogChange"
	TouchChangeType = "TouchChange"
)

type JournalEntry interface {
	undo(*StateDB)
	String() string
	Marshal() ([]byte, error)
	SetType()
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
		log.Error("unmarshal memjournal failed")
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
		case "CreateObjectChange":
			var tmp CreateObjectChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "ResetObjectChange":
			var tmp ResetObjectChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "RefundChange":
			var tmp RefundChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "AddLogChange":
			var tmp AddLogChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "BalanceChange":
			var tmp BalanceChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "NonceChange":
			var tmp NonceChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "TouchChange":
			var tmp TouchChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "SuicideChange":
			var tmp SuicideChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "StorageChange":
			var tmp StorageChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		case "CodeChange":
			var tmp CodeChange
			err = json.Unmarshal(res, &tmp)
			if err != nil {
				return nil, err
			}
			jos = append(jos, &tmp)
		default:
			log.Error("unmarshal journal failed")
			return nil, errors.New("unmarshal journal failed")
		}
	}

	ret := &Journal{
		JournalList: jos,
	}
	return ret, nil
}


type (
	// Changes to the account database
	CreateObjectChange struct {
		Account *common.Address `json:"account,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
	ResetObjectChange struct {
		Prev *StateObject       `json:"prev,omitempty"`
		Type string             `json:"type,omitempty"`
	}
	SuicideChange struct {
		Account     *common.Address  `json:"account,omitempty"`
		Prev        bool             `json:"prev,omitempty"`           // whether account had already suicided
		Prevbalance *big.Int         `json:"prevbalance,omitempty"`
		Type        string           `json:"type,omitempty"`
	}

	// Changes to individual accounts.
	BalanceChange struct {
		Account *common.Address     `json:"account,omitempty"`
		Prev    *big.Int            `json:"prev,omitempty"`
		Type    string	            `json:"type,omitempty"`
	}
	NonceChange struct {
		Account *common.Address    `json:"account,omitempty"`
		Prev    uint64	           `json:"prev,omitempty"`
		Type    string             `json:"type,omitempty"`
	}
	StorageChange struct {
		Account       *common.Address `json:"account,omitempty"`
		Key           common.Hash     `json:"key,omitempty"`
		Prevalue      common.Hash     `json:"prevalue,omitempty"`
		Type          string          `json:"type,omitempty"`
	}
	CodeChange struct {
		Account            *common.Address `json:"account,omitempty"`
		Prevcode           []byte          `json:"prevcode,omitempty"`
		Prevhash           []byte          `json:"prevhash,omitempty"`
		Type               string          `json:"type,omitempty"`
	}

	// Changes to other state values.
	RefundChange struct {
		Prev *big.Int         `json:"prev,omitempty"`
		Type string           `json:"type,omitempty"`
	}
	AddLogChange struct {
		Txhash common.Hash   `json:"txhash,omitempty"`
		Type   string        `json:"type,omitempty"`
	}
	TouchChange struct {
		Account *common.Address `json:"account,omitempty"`
		Prev    bool            `json:"prev,omitempty"`
		Type    string          `json:"type,omitempty"`
	}
)
// createObjectChange
func (ch *CreateObjectChange) undo(s *StateDB) {
	delete(s.stateObjects, *ch.Account)
	delete(s.stateObjectsDirty, *ch.Account)
}
func (ch *CreateObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [createObjectChange] %s\n", ch.Account.Hex())
	return str
}
func (ch *CreateObjectChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *CreateObjectChange) SetType() {
	ch.Type = CreateObjectChangeType
}
// resetObjectChange
func (ch *ResetObjectChange) undo(s *StateDB) {
	s.setStateObject(ch.Prev)
}
func (ch *ResetObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [resetObjectChange] %s\n", ch.Prev.String())
	return str
}

func (ch *ResetObjectChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *ResetObjectChange) SetType() {
	ch.Type = ResetObjectChangeType
}
// suicideChange
func (ch *SuicideChange) undo(s *StateDB) {
	obj := s.GetStateObject(*ch.Account)
	if obj != nil {
		obj.suicided = ch.Prev
		obj.setBalance(ch.Prevbalance)
	}
}
func (ch *SuicideChange) String() string {
	var str string
	str = fmt.Sprintf("journal [suicideChange] %s  %#v  %d\n", ch.Account.Hex(), ch.Prev, ch.Prevbalance)
	return str
}
func (ch *SuicideChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *SuicideChange) SetType() {
	ch.Type = SuicideChangeType
}
// touchChange
var ripemd = common.HexToAddress("0000000000000000000000000000000000000003")

func (ch *TouchChange) undo(s *StateDB) {
	if !ch.Prev && *ch.Account != ripemd {
		delete(s.stateObjects, *ch.Account)
		delete(s.stateObjectsDirty, *ch.Account)
	}
}
func (ch *TouchChange) String() string {
	var str string
	str = fmt.Sprintf("journal [touchChange] %s %#v\n", ch.Account.Hex(), ch.Prev)
	return str
}
func (ch *TouchChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *TouchChange) SetType() {
	ch.Type = TouchChangeType
}
// balanceChange
func (ch *BalanceChange) undo(s *StateDB) {
	s.GetStateObject(*ch.Account).setBalance(ch.Prev)
}
func (ch *BalanceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [balanceChange] %s %#v\n", ch.Account.Hex(), ch.Prev)
	return str
}
func (ch *BalanceChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *BalanceChange) SetType() {
	ch.Type = BalanceChangeType
}
// nonceChange
func (ch *NonceChange) undo(s *StateDB) {
	s.GetStateObject(*ch.Account).setNonce(ch.Prev)
}

func (ch *NonceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [nonceChange] %s %d \n", ch.Account.Hex(), ch.Prev)
	return str

}
func (ch *NonceChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *NonceChange) SetType() {
	ch.Type = NonceChangeType
}

// codeChange
func (ch *CodeChange) undo(s *StateDB) {
	s.GetStateObject(*ch.Account).setCode(common.BytesToHash(ch.Prevhash), ch.Prevcode)
}
func (ch *CodeChange) String() string {
	var str string
	str = fmt.Sprintf("journal [codeChange] %s codeHash %s code %s\n", ch.Account.Hex(), common.BytesToHash(ch.Prevhash).Hex(), common.Bytes2Hex(ch.Prevcode))
	return str
}
func (ch *CodeChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *CodeChange) SetType() {
	ch.Type = CodeChangeType
}

// storageChange
func (ch *StorageChange) undo(s *StateDB) {
	s.GetStateObject(*ch.Account).setState(ch.Key, ch.Prevalue)
}
func (ch *StorageChange) String() string {
	var str string
	str = fmt.Sprintf("journal [storageChange] %s previous key %s  previous value %s \n", ch.Account.Hex(), ch.Key.Hex(), ch.Prevalue.Hex())
	return str
}
func (ch *StorageChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *StorageChange) SetType() {
	ch.Type = StorageChangeType
}

// refundChange
func (ch *RefundChange) undo(s *StateDB) {
	s.refund = ch.Prev
}
func (ch *RefundChange) String() string {
	var str string
	str = fmt.Sprintf("journal [refundChange] previous value %s \n", ch.Prev.String())
	return str
}
func (ch *RefundChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *RefundChange) SetType() {
	ch.Type = RefundChangeType
}

// addLogChange
func (ch *AddLogChange) undo(s *StateDB) {
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
func (ch *AddLogChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
func (ch *AddLogChange) SetType() {
	ch.Type = AddLogChangeType
}
