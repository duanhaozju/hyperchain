package types

import (
	"fmt"
)

type Base struct {
	Key        string
	BitNum     int
	BitIndex   int
	AddressNum int
	Value      string
	IsNewSlot  bool
	IsNewDoneSlot  bool
	IsAddress  bool
}

func (self *Base) GetBitNum() int {
	return self.BitNum
}

func (self *Base) GetIsNewSlot() bool {
	return self.IsNewSlot
}

func (self *Base) GetIsNewDoneSlot() bool {
	return self.IsNewDoneSlot
}

func (self *Base) SetBitIndex(bitIndex int) {
	self.BitIndex = bitIndex
}

func (self *Base) getValue() string {
	return self.Value
}

func (self *Base) SetValue(value string) {
	self.Value = value
}

func (self *Base) GetIsAddress() bool {
	return self.IsAddress
}

func (self *Base) SetIsAddress(isAddress bool) {
	self.IsAddress = isAddress
}

func (self *Base) GetKey() string {
	return self.Key
}

func (self *Base) GetValue() string {
	return self.Value
}

func (self *Base) String() string {
	return fmt.Sprintf("{ Key: %v BitNum: %v BitIndex: %v AddressNum: %v Value: %v IsNewSlot: %v IsNewDoneSlot: %v IsAddress: %v}",
		self.Key, self.BitNum, self.BitIndex, self.AddressNum, self.Value, self.IsNewSlot, self.IsNewDoneSlot,self.IsAddress)
}
