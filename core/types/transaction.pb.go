// Code generated by protoc-gen-go.
// source: transaction.proto
// DO NOT EDIT!

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type InvalidTransactionRecord_ErrType int32

const (
	InvalidTransactionRecord_OUTOFBALANCE           InvalidTransactionRecord_ErrType = 0
	InvalidTransactionRecord_SIGFAILED              InvalidTransactionRecord_ErrType = 1
	InvalidTransactionRecord_INVOKE_CONTRACT_FAILED InvalidTransactionRecord_ErrType = 2
	InvalidTransactionRecord_DEPLOY_CONTRACT_FAILED InvalidTransactionRecord_ErrType = 3
)

var InvalidTransactionRecord_ErrType_name = map[int32]string{
	0: "OUTOFBALANCE",
	1: "SIGFAILED",
	2: "INVOKE_CONTRACT_FAILED",
	3: "DEPLOY_CONTRACT_FAILED",
}
var InvalidTransactionRecord_ErrType_value = map[string]int32{
	"OUTOFBALANCE":           0,
	"SIGFAILED":              1,
	"INVOKE_CONTRACT_FAILED": 2,
	"DEPLOY_CONTRACT_FAILED": 3,
}

func (x InvalidTransactionRecord_ErrType) String() string {
	return proto.EnumName(InvalidTransactionRecord_ErrType_name, int32(x))
}
func (InvalidTransactionRecord_ErrType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor5, []int{1, 0}
}

type Transaction struct {
	Version         []byte `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	From            []byte `protobuf:"bytes,2,opt,name=from,proto3" json:"from,omitempty"`
	To              []byte `protobuf:"bytes,3,opt,name=to,proto3" json:"to,omitempty"`
	Value           []byte `protobuf:"bytes,4,opt,name=value,proto3" json:"value,omitempty"`
	Timestamp       int64  `protobuf:"varint,5,opt,name=timestamp" json:"timestamp,omitempty"`
	Signature       []byte `protobuf:"bytes,6,opt,name=signature,proto3" json:"signature,omitempty"`
	Id              uint64 `protobuf:"varint,7,opt,name=id" json:"id,omitempty"`
	TransactionHash []byte `protobuf:"bytes,8,opt,name=transactionHash,proto3" json:"transactionHash,omitempty"`
}

func (m *Transaction) Reset()                    { *m = Transaction{} }
func (m *Transaction) String() string            { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()               {}
func (*Transaction) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{0} }

func (m *Transaction) GetVersion() []byte {
	if m != nil {
		return m.Version
	}
	return nil
}

func (m *Transaction) GetFrom() []byte {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *Transaction) GetTo() []byte {
	if m != nil {
		return m.To
	}
	return nil
}

func (m *Transaction) GetValue() []byte {
	if m != nil {
		return m.Value
	}
	return nil
}

func (m *Transaction) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Transaction) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *Transaction) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}


type InvalidTransactionRecord struct {
	Tx      *Transaction                     `protobuf:"bytes,1,opt,name=tx" json:"tx,omitempty"`
	ErrType InvalidTransactionRecord_ErrType `protobuf:"varint,2,opt,name=errType,enum=types.InvalidTransactionRecord_ErrType" json:"errType,omitempty"`
	ErrMsg  []byte                           `protobuf:"bytes,3,opt,name=errMsg,proto3" json:"errMsg,omitempty"`
}

func (m *InvalidTransactionRecord) Reset()                    { *m = InvalidTransactionRecord{} }
func (m *InvalidTransactionRecord) String() string            { return proto.CompactTextString(m) }
func (*InvalidTransactionRecord) ProtoMessage()               {}
func (*InvalidTransactionRecord) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{1} }

func (m *InvalidTransactionRecord) GetTx() *Transaction {
	if m != nil {
		return m.Tx
	}
	return nil
}

func (m *InvalidTransactionRecord) GetErrType() InvalidTransactionRecord_ErrType {
	if m != nil {
		return m.ErrType
	}
	return InvalidTransactionRecord_OUTOFBALANCE
}

func (m *InvalidTransactionRecord) GetErrMsg() []byte {
	if m != nil {
		return m.ErrMsg
	}
	return nil
}

type InvalidTransactionRecords struct {
	Records []*InvalidTransactionRecord `protobuf:"bytes,1,rep,name=records" json:"records,omitempty"`
}

func (m *InvalidTransactionRecords) Reset()                    { *m = InvalidTransactionRecords{} }
func (m *InvalidTransactionRecords) String() string            { return proto.CompactTextString(m) }
func (*InvalidTransactionRecords) ProtoMessage()               {}
func (*InvalidTransactionRecords) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{2} }

func (m *InvalidTransactionRecords) GetRecords() []*InvalidTransactionRecord {
	if m != nil {
		return m.Records
	}
	return nil
}

type TransactionMeta struct {
	BlockIndex uint64 `protobuf:"varint,1,opt,name=BlockIndex" json:"BlockIndex,omitempty"`
	Index      int64  `protobuf:"varint,2,opt,name=Index" json:"Index,omitempty"`
}

func (m *TransactionMeta) Reset()                    { *m = TransactionMeta{} }
func (m *TransactionMeta) String() string            { return proto.CompactTextString(m) }
func (*TransactionMeta) ProtoMessage()               {}
func (*TransactionMeta) Descriptor() ([]byte, []int) { return fileDescriptor5, []int{3} }

func (m *TransactionMeta) GetBlockIndex() uint64 {
	if m != nil {
		return m.BlockIndex
	}
	return 0
}

func (m *TransactionMeta) GetIndex() int64 {
	if m != nil {
		return m.Index
	}
	return 0
}

func init() {
	proto.RegisterType((*Transaction)(nil), "types.Transaction")
	proto.RegisterType((*InvalidTransactionRecord)(nil), "types.InvalidTransactionRecord")
	proto.RegisterType((*InvalidTransactionRecords)(nil), "types.InvalidTransactionRecords")
	proto.RegisterType((*TransactionMeta)(nil), "types.TransactionMeta")
	proto.RegisterEnum("types.InvalidTransactionRecord_ErrType", InvalidTransactionRecord_ErrType_name, InvalidTransactionRecord_ErrType_value)
}

func init() { proto.RegisterFile("transaction.proto", fileDescriptor5) }

var fileDescriptor5 = []byte{
	// 392 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x52, 0x4f, 0x6f, 0xd3, 0x30,
	0x14, 0xc7, 0xee, 0x9f, 0xb0, 0xd7, 0xb1, 0x95, 0x27, 0x34, 0x19, 0x84, 0x20, 0xca, 0x85, 0x9c,
	0x72, 0x28, 0x27, 0x8e, 0x59, 0x97, 0x8d, 0x88, 0xae, 0x41, 0x26, 0x4c, 0xe2, 0x34, 0x79, 0x8d,
	0x19, 0x11, 0x6d, 0x1c, 0xd9, 0x5e, 0xb5, 0x7d, 0x52, 0x3e, 0x0d, 0x12, 0x8a, 0x93, 0xaa, 0x11,
	0x68, 0xe2, 0xf6, 0x7e, 0x7f, 0x9e, 0xed, 0xdf, 0xf3, 0x83, 0xe7, 0x56, 0x8b, 0xca, 0x88, 0x95,
	0x2d, 0x55, 0x15, 0xd5, 0x5a, 0x59, 0x85, 0x23, 0xfb, 0x50, 0x4b, 0x13, 0xfc, 0x22, 0x30, 0xc9,
	0xf7, 0x22, 0x32, 0xf0, 0xb6, 0x52, 0x9b, 0x52, 0x55, 0x8c, 0xf8, 0x24, 0x3c, 0xe4, 0x3b, 0x88,
	0x08, 0xc3, 0xef, 0x5a, 0x6d, 0x18, 0x75, 0xb4, 0xab, 0xf1, 0x08, 0xa8, 0x55, 0x6c, 0xe0, 0x18,
	0x6a, 0x15, 0xbe, 0x80, 0xd1, 0x56, 0xac, 0xef, 0x24, 0x1b, 0x3a, 0xaa, 0x05, 0xf8, 0x1a, 0x0e,
	0x6c, 0xb9, 0x91, 0xc6, 0x8a, 0x4d, 0xcd, 0x46, 0x3e, 0x09, 0x07, 0x7c, 0x4f, 0x34, 0xaa, 0x29,
	0x6f, 0x2b, 0x61, 0xef, 0xb4, 0x64, 0x63, 0xd7, 0xb7, 0x27, 0x9a, 0x1b, 0xca, 0x82, 0x79, 0x3e,
	0x09, 0x87, 0x9c, 0x96, 0x05, 0x86, 0x70, 0xdc, 0xcb, 0xf2, 0x51, 0x98, 0x1f, 0xec, 0xa9, 0xeb,
	0xf9, 0x9b, 0x0e, 0x7e, 0x13, 0x60, 0x69, 0xb5, 0x15, 0xeb, 0xb2, 0xe8, 0x05, 0xe4, 0x72, 0xa5,
	0x74, 0x81, 0x01, 0x50, 0x7b, 0xef, 0x12, 0x4e, 0x66, 0x18, 0xb9, 0x51, 0x44, 0x7d, 0x17, 0xb5,
	0xf7, 0x18, 0x83, 0x27, 0xb5, 0xce, 0x1f, 0x6a, 0xe9, 0x32, 0x1f, 0xcd, 0xde, 0x75, 0xc6, 0xc7,
	0x4e, 0x8d, 0x92, 0xd6, 0xce, 0x77, 0x7d, 0x78, 0x02, 0x63, 0xa9, 0xf5, 0xa5, 0xb9, 0xed, 0x66,
	0xd4, 0xa1, 0xe0, 0x06, 0xbc, 0xce, 0x8b, 0x53, 0x38, 0xcc, 0xbe, 0xe6, 0xd9, 0xf9, 0x69, 0xbc,
	0x88, 0x97, 0xf3, 0x64, 0xfa, 0x04, 0x9f, 0xc1, 0xc1, 0x97, 0xf4, 0xe2, 0x3c, 0x4e, 0x17, 0xc9,
	0xd9, 0x94, 0xe0, 0x2b, 0x38, 0x49, 0x97, 0x57, 0xd9, 0xa7, 0xe4, 0x7a, 0x9e, 0x2d, 0x73, 0x1e,
	0xcf, 0xf3, 0xeb, 0x4e, 0xa3, 0x8d, 0x76, 0x96, 0x7c, 0x5e, 0x64, 0xdf, 0xfe, 0xd1, 0x06, 0xc1,
	0x15, 0xbc, 0x7c, 0xec, 0xa1, 0x06, 0x3f, 0x80, 0xa7, 0xdb, 0x92, 0x11, 0x7f, 0x10, 0x4e, 0x66,
	0x6f, 0xff, 0x93, 0x8d, 0xef, 0xfc, 0xc1, 0x05, 0x1c, 0xf7, 0xd4, 0x4b, 0x69, 0x05, 0xbe, 0x01,
	0x38, 0x5d, 0xab, 0xd5, 0xcf, 0xb4, 0x2a, 0x64, 0x3b, 0xd5, 0x21, 0xef, 0x31, 0xcd, 0x5a, 0xb4,
	0x12, 0x75, 0x9f, 0xdf, 0x82, 0x9b, 0xb1, 0x5b, 0xc4, 0xf7, 0x7f, 0x02, 0x00, 0x00, 0xff, 0xff,
	0x62, 0x61, 0x5b, 0x70, 0x9d, 0x02, 0x00, 0x00,
}
