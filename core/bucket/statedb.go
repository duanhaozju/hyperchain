package bucket

import (
	"github.com/op/go-logging"

	"hyperchain/common"
	"hyperchain/core/vm"
	"math/big"
	"hyperchain/hyperdb"
)
var (
	log *logging.Logger // package-level logger
)

func init() {
	log = logging.MustGetLogger("bucket")
}

type StateDB struct {
}

func New(root common.Hash, db hyperdb.Database) (*StateDB, error) {
	return nil, nil
}
func (self *StateDB) GetAccount(common.Address) vm.Account {
	return nil
}
func (self *StateDB) CreateAccount(common.Address) vm.Account {
	return nil
}

func (self *StateDB) AddBalance(common.Address, *big.Int) {

}
func (self *StateDB) GetBalance(common.Address) *big.Int {
	return nil
}

func (self *StateDB) GetNonce(common.Address) uint64 {
	return 0
}

func (self *StateDB) SetNonce(common.Address, uint64) {

}

func (self *StateDB) GetCode(common.Address) []byte {
	return nil
}
func (self *StateDB) SetCode(common.Address, []byte) {

}

func (self *StateDB) AddRefund(*big.Int) {

}
func (self *StateDB) GetRefund() *big.Int {
	return nil
}

func (self *StateDB) GetState(common.Address, common.Hash) common.Hash {
	return common.Hash{}
}

func (self *StateDB) SetState(common.Address, common.Hash, common.Hash) {

}

func (self *StateDB) Delete(common.Address) bool {
	return false
}
func (self *StateDB) Exist(common.Address) bool {
	return false
}
func (self *StateDB) IsDeleted(common.Address) bool {
	return false
}

func (self *StateDB) StartRecord(common.Hash, common.Hash, int) {

}
func (self *StateDB) AddLog(log *vm.Log) {

}
func (self *StateDB) GetLogs(hash common.Hash) vm.Logs {
	return nil
}

func (self *StateDB) Copy() vm.Database {
	return nil
}
func (self *StateDB) Set(vm.Database) {

}

func (self *StateDB) Commit() (common.Hash, error) {
	return common.Hash{}, nil
}