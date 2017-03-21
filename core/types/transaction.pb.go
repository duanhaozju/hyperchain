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
	InvalidTransactionRecord_INVALID_PERMISSION     InvalidTransactionRecord_ErrType = 4
)

var InvalidTransactionRecord_ErrType_name = map[int32]string{
	0: "OUTOFBALANCE",
	1: "SIGFAILED",
	2: "INVOKE_CONTRACT_FAILED",
	3: "DEPLOY_CONTRACT_FAILED",
	4: "INVALID_PERMISSION",
}
var InvalidTransactionRecord_ErrType_value = map[string]int32{
	"OUTOFBALANCE":           0,
	"SIGFAILED":              1,
	"INVOKE_CONTRACT_FAILED": 2,
	"DEPLOY_CONTRACT_FAILED": 3,
	"INVALID_PERMISSION":     4,
}

func (x InvalidTransactionRecord_ErrType) String() string {
	return proto.EnumName(InvalidTransactionRecord_ErrType_name, int32(x))
}
func (InvalidTransactionRecord_ErrType) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor6, []int{1, 0}
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
	Nonce           int64  `protobuf:"varint,9,opt,name=nonce" json:"nonce,omitempty"`
}

func (m *Transaction) Reset()                    { *m = Transaction{} }
func (m *Transaction) String() string            { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()               {}
func (*Transaction) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{0} }

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

func (m *Transaction) GetTransactionHash() []byte {
	if m != nil {
		return m.TransactionHash
	}
	return nil
}

func (m *Transaction) GetNonce() int64 {
	if m != nil {
		return m.Nonce
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
func (*InvalidTransactionRecord) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{1} }

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
func (*InvalidTransactionRecords) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{2} }

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
func (*TransactionMeta) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{3} }

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

func init() { proto.RegisterFile("transaction.proto", fileDescriptor6) }

var fileDescriptor6 = []byte{
	// 422 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x52, 0x4f, 0x6f, 0xd3, 0x30,
	0x14, 0x27, 0x6e, 0xda, 0xd0, 0xd7, 0xb1, 0x95, 0x27, 0x54, 0x19, 0x84, 0xa0, 0xca, 0x85, 0x9c,
	0x7a, 0x28, 0x27, 0x8e, 0x59, 0x9b, 0x0d, 0x8b, 0x36, 0x99, 0xdc, 0x50, 0x89, 0x53, 0x15, 0x1a,
	0x33, 0x22, 0xda, 0x38, 0x72, 0xbc, 0x6a, 0xe3, 0x7b, 0xf0, 0x31, 0xf9, 0x0e, 0x28, 0x4e, 0xaa,
	0x46, 0xa0, 0x89, 0x9b, 0x7f, 0x7f, 0x5e, 0x9c, 0xdf, 0xef, 0x19, 0x9e, 0x6b, 0x95, 0xe4, 0x65,
	0xb2, 0xd5, 0x99, 0xcc, 0x27, 0x85, 0x92, 0x5a, 0x62, 0x57, 0x3f, 0x14, 0xa2, 0x74, 0x7f, 0x5b,
	0x30, 0x88, 0x4f, 0x22, 0x52, 0x70, 0x0e, 0x42, 0x95, 0x99, 0xcc, 0xa9, 0x35, 0xb6, 0xbc, 0x33,
	0x7e, 0x84, 0x88, 0x60, 0x7f, 0x53, 0x72, 0x4f, 0x89, 0xa1, 0xcd, 0x19, 0xcf, 0x81, 0x68, 0x49,
	0x3b, 0x86, 0x21, 0x5a, 0xe2, 0x0b, 0xe8, 0x1e, 0x92, 0xdd, 0x9d, 0xa0, 0xb6, 0xa1, 0x6a, 0x80,
	0xaf, 0xa1, 0xaf, 0xb3, 0xbd, 0x28, 0x75, 0xb2, 0x2f, 0x68, 0x77, 0x6c, 0x79, 0x1d, 0x7e, 0x22,
	0x2a, 0xb5, 0xcc, 0x6e, 0xf3, 0x44, 0xdf, 0x29, 0x41, 0x7b, 0x66, 0xee, 0x44, 0x54, 0x37, 0x64,
	0x29, 0x75, 0xc6, 0x96, 0x67, 0x73, 0x92, 0xa5, 0xe8, 0xc1, 0x45, 0x2b, 0xcb, 0xc7, 0xa4, 0xfc,
	0x4e, 0x9f, 0x9a, 0x99, 0xbf, 0xe9, 0xea, 0x5f, 0x72, 0x99, 0x6f, 0x05, 0xed, 0x9b, 0x1b, 0x6b,
	0xe0, 0xfe, 0x22, 0x40, 0x59, 0x7e, 0x48, 0x76, 0x59, 0xda, 0x8a, 0xcd, 0xc5, 0x56, 0xaa, 0x14,
	0x5d, 0x20, 0xfa, 0xde, 0xe4, 0x1e, 0x4c, 0x71, 0x62, 0x0a, 0x9a, 0xb4, 0x5d, 0x44, 0xdf, 0xa3,
	0x0f, 0x8e, 0x50, 0x2a, 0x7e, 0x28, 0x84, 0x69, 0xe2, 0x7c, 0xfa, 0xae, 0x31, 0x3e, 0xf6, 0xd5,
	0x49, 0x50, 0xdb, 0xf9, 0x71, 0x0e, 0x47, 0xd0, 0x13, 0x4a, 0x2d, 0xcb, 0xdb, 0xa6, 0xb9, 0x06,
	0xb9, 0x3f, 0xc1, 0x69, 0xbc, 0x38, 0x84, 0xb3, 0xe8, 0x73, 0x1c, 0x5d, 0x5d, 0xfa, 0x0b, 0x3f,
	0x9c, 0x05, 0xc3, 0x27, 0xf8, 0x0c, 0xfa, 0x2b, 0x76, 0x7d, 0xe5, 0xb3, 0x45, 0x30, 0x1f, 0x5a,
	0xf8, 0x0a, 0x46, 0x2c, 0x5c, 0x47, 0x9f, 0x82, 0xcd, 0x2c, 0x0a, 0x63, 0xee, 0xcf, 0xe2, 0x4d,
	0xa3, 0x91, 0x4a, 0x9b, 0x07, 0x37, 0x8b, 0xe8, 0xcb, 0x3f, 0x5a, 0x07, 0x47, 0x80, 0x2c, 0x5c,
	0xfb, 0x0b, 0x36, 0xdf, 0xdc, 0x04, 0x7c, 0xc9, 0x56, 0x2b, 0x16, 0x85, 0x43, 0xdb, 0x5d, 0xc3,
	0xcb, 0xc7, 0x02, 0x94, 0xf8, 0x01, 0x1c, 0x55, 0x1f, 0xa9, 0x35, 0xee, 0x78, 0x83, 0xe9, 0xdb,
	0xff, 0x64, 0xe6, 0x47, 0xbf, 0x7b, 0x0d, 0x17, 0x2d, 0x75, 0x29, 0x74, 0x82, 0x6f, 0x00, 0x2e,
	0x77, 0x72, 0xfb, 0x83, 0xe5, 0xa9, 0xa8, 0xdb, 0xb6, 0x79, 0x8b, 0xa9, 0x16, 0x57, 0x4b, 0xa4,
	0x5e, 0x9c, 0x01, 0x5f, 0x7b, 0xe6, 0xd9, 0xbe, 0xff, 0x13, 0x00, 0x00, 0xff, 0xff, 0x73, 0x21,
	0xd6, 0xd4, 0xcb, 0x02, 0x00, 0x00,
}
