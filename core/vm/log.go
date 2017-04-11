package vm

type Log interface {
	SetAttribute(int, interface{})
	GetAttribute(int) interface{}
}

const (
	LogAttr_Address = iota
	LogAttr_Topics
	LogAttr_Data
	LogAttr_BlockNumber
	LogAttr_BlockHash
	LogAttr_TxHash
	LogAttr_TxIndex
	LogAttr_Index
	LogAttr_Type
)
type Logs []Log
