package bucket

import (
	"github.com/op/go-logging"

	"hyperchain/common"
	"hyperchain/core/vm"
	"math/big"
)
var (
	log *logging.Logger // package-level logger
)

func init() {
	log = logging.MustGetLogger("bucket")
}

type StateDB struct {

}

func (self *StateDB) GetAccount(common.Address) vm.Account {

}
func (self *StateDB) CreateAccount(common.Address) vm.Account {

}

func (self *StateDB) AddBalance(common.Address, *big.Int) {

}
func (self *StateDB) GetBalance(common.Address) *big.Int {

}

func (self *StateDB) GetNonce(common.Address) uint64 {

}

func (self *StateDB) SetNonce(common.Address, uint64) {

}

func (self *StateDB) GetCode(common.Address) []byte {

}
func (self *StateDB) SetCode(common.Address, []byte) {

}

func (self *StateDB) AddRefund(*big.Int) {

}
func (self *StateDB) GetRefund() *big.Int {

}

func (self *StateDB) GetState(common.Address, common.Hash) common.Hash {

}

func (self *StateDB) SetState(common.Address, common.Hash, common.Hash) {

}

func (self *StateDB) Delete(common.Address) bool {

}
func (self *StateDB) Exist(common.Address) bool {

}
func (self *StateDB) IsDeleted(common.Address) bool {

}

func (self *StateDB) StartRecord(common.Hash, common.Hash, int) {

}
func (self *StateDB) AddLog(log *vm.Log) {

}
func (self *StateDB) GetLogs(hash common.Hash) vm.Logs {

}

func (self *StateDB) Copy() vm.Database {

}
func (self *StateDB) Set(vm.Database) {

}

func (self *StateDB) Commit() (common.Hash, error) {

}