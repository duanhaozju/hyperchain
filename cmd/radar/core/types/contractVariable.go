package types

import (
	"fmt"
	"github.com/hyperchain/hyperchain/common"
)

type SolidityVariable interface {
	Decode() string
	String() string
	GetBitNum() int
	SetBitIndex(bitIndex int)
	GetIsNewSlot() bool
	GetIsNewDoneSlot() bool
	GetIsAddress() bool
	SetIsAddress(isAddress bool)
	SetValue(value string)
	GetKey() string
	GetValue() string
}

type ContractVariable struct {
	Id                 uint
	Name               string
	Key                string
	NowKey             string
	RemainType         string
	Variable           SolidityVariable
	StartAddressOfSlot []byte
	OriginalKeyOfMap   string
	KeyOfMap           string
	Use                bool
	BelongTo           uint
	BelongToFlag       bool
}

func (contractVariable *ContractVariable) SetStartAddressOfSlot(v []byte) {
	contractVariable.StartAddressOfSlot = v
}

func (contractVariable *ContractVariable) String() string {
	return fmt.Sprintf("{ Id: %v Name: %v Key: %v NowKey: %v RemainType:%v Variable: %v StartAddressOfSlot: %v OriginalKeyOfMap: %v KeyOfMap: %v Use: %v BelongTo: %v BelongToFlag: %v}\n",
		contractVariable.Id, contractVariable.Name, contractVariable.Key, contractVariable.NowKey, contractVariable.RemainType,
		contractVariable.Variable.String(), common.Bytes2Hex(contractVariable.StartAddressOfSlot), contractVariable.OriginalKeyOfMap,
		contractVariable.KeyOfMap, contractVariable.Use, contractVariable.BelongTo, contractVariable.BelongToFlag)
}
