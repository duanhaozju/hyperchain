package hyperstate

import (
	"math/big"
	"hyperchain/common"
	"fmt"
	"encoding/json"
)

type journalEntry interface {
	undo(*StateDB)
	String() string
	Marshal() ([]byte, error)
}

type journal []journalEntry

func (self *journal) Marshal() []byte {
	return nil
}

func UnmarshalJournal(data []byte, ret *journal) error {
	return nil
}


type (
	// Changes to the account database
	createObjectChange struct {
		account *common.Address
	}
	resetObjectChange struct {
		prev *StateObject
	}
	suicideChange struct {
		account     *common.Address
		prev        bool // whether account had already suicided
		prevbalance *big.Int
	}

	// Changes to individual accounts.
	balanceChange struct {
		account *common.Address
		prev    *big.Int
	}
	nonceChange struct {
		account *common.Address
		prev    uint64
	}
	storageChange struct {
		account       *common.Address
		key, prevalue common.Hash
	}
	codeChange struct {
		account            *common.Address
		prevcode, prevhash []byte
	}

	// Changes to other state values.
	refundChange struct {
		prev *big.Int
	}
	addLogChange struct {
		txhash common.Hash
	}
	touchChange struct {
		account *common.Address
		prev    bool
	}
)
// createObjectChange
func (ch *createObjectChange) undo(s *StateDB) {
	delete(s.stateObjects, *ch.account)
	delete(s.stateObjectsDirty, *ch.account)
}
func (ch *createObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [createObjectChange] %s\n", ch.account.Hex())
	return str
}
func (ch *createObjectChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// resetObjectChange
func (ch *resetObjectChange) undo(s *StateDB) {
	s.setStateObject(ch.prev)
}
func (ch *resetObjectChange) String() string {
	var str string
	str = fmt.Sprintf("journal [resetObjectChange] %s\n", ch.prev.String())
	return str
}

func (ch *resetObjectChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
// suicideChange
func (ch *suicideChange) undo(s *StateDB) {
	obj := s.GetStateObject(*ch.account)
	if obj != nil {
		obj.suicided = ch.prev
		obj.setBalance(ch.prevbalance)
	}
}
func (ch *suicideChange) String() string {
	var str string
	str = fmt.Sprintf("journal [suicideChange] %s\n", ch.account.Hex())
	return str
}
func (ch *suicideChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// touchChange
var ripemd = common.HexToAddress("0000000000000000000000000000000000000003")

func (ch *touchChange) undo(s *StateDB) {
	if !ch.prev && *ch.account != ripemd {
		delete(s.stateObjects, *ch.account)
		delete(s.stateObjectsDirty, *ch.account)
	}
}
func (ch *touchChange) String() string {
	var str string
	str = fmt.Sprintf("journal [touchChange] %s\n", ch.account.Hex())
	return str
}
func (ch *touchChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// balanceChange
func (ch *balanceChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setBalance(ch.prev)
}
func (ch *balanceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [balanceChange] %s previous balance %s \n", ch.account.Hex(), ch.prev.String())
	return str
}
func (ch *balanceChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// nonceChange
func (ch *nonceChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setNonce(ch.prev)
}

func (ch *nonceChange) String() string {
	var str string
	str = fmt.Sprintf("journal [nonceChange] %s previous nonce %d \n", ch.account.Hex(), ch.prev)
	return str

}
func (ch *nonceChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// codeChange
func (ch *codeChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setCode(common.BytesToHash(ch.prevhash), ch.prevcode)
}
func (ch *codeChange) String() string {
	var str string
	str = fmt.Sprintf("journal [codeChange] %s previous codeHash %s \n", ch.account.Hex(), common.BytesToHash(ch.prevhash).Hex())
	return str
}
func (ch *codeChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// storageChange
func (ch *storageChange) undo(s *StateDB) {
	s.GetStateObject(*ch.account).setState(ch.key, ch.prevalue)
}
func (ch *storageChange) String() string {
	var str string
	str = fmt.Sprintf("journal [storageChange] %s previous key %s  previous value %s \n", ch.account.Hex(), ch.key.Hex(), ch.prevalue.Hex())
	return str
}
func (ch *storageChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// refundChange
func (ch *refundChange) undo(s *StateDB) {
	s.refund = ch.prev
}
func (ch *refundChange) String() string {
	var str string
	str = fmt.Sprintf("journal [refundChange] previous value %s \n", ch.prev.String())
	return str
}
func (ch *refundChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}

// addLogChange
func (ch *addLogChange) undo(s *StateDB) {
	logs := s.logs[ch.txhash]
	if len(logs) == 1 {
		delete(s.logs, ch.txhash)
	} else {
		s.logs[ch.txhash] = logs[:len(logs)-1]
	}
}
func (ch *addLogChange) String() string {
	var str string
	str = fmt.Sprintf("journal [addLogChange] tx hash %s \n", ch.txhash.Hex())
	return str
}
func (ch *addLogChange) Marshal()([]byte, error) {
	return json.Marshal(ch)
}
