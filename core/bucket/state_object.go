package bucket

import (
	"math/big"
	"hyperchain/common"
)

type StateObject struct {

}

func (self *StateObject) SubBalance(amount *big.Int) {

}
func (self *StateObject) AddBalance(amount *big.Int) {

}
func (self *StateObject) SetBalance(*big.Int) {

}
func (self *StateObject) SetNonce(uint64) {

}
func (self *StateObject) Balance() *big.Int {
	return nil
}
func (self *StateObject) Address() common.Address {
	return common.Address{}
}
func (self *StateObject) ReturnGas(*big.Int, *big.Int) {

}
func (self *StateObject) SetCode([]byte) {

}
func (self *StateObject) ForEachStorage(cb func(key, value common.Hash) bool) {

}
func (self *StateObject) PrintStorages() {

}
func (self *StateObject) Value() *big.Int {
	return nil
}
