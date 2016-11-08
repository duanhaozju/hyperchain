// Code generated by protoc-gen-go.
// source: transaction.proto
// DO NOT EDIT!

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	transaction.proto

It has these top-level messages:
	Transaction
	InvalidTransactionRecord
	InvalidTransactionRecords
	TransactionMeta
*/
package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

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
	return fileDescriptor0, []int{1, 0}
}

type Transaction struct {
	From            []byte `protobuf:"bytes,1,opt,name=from,proto3" json:"from,omitempty"`
	To              []byte `protobuf:"bytes,2,opt,name=to,proto3" json:"to,omitempty"`
	Value           []byte `protobuf:"bytes,3,opt,name=value,proto3" json:"value,omitempty"`
	Timestamp       int64  `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
	Signature       []byte `protobuf:"bytes,5,opt,name=signature,proto3" json:"signature,omitempty"`
	Id              uint64 `protobuf:"varint,6,opt,name=id" json:"id,omitempty"`
	TransactionHash []byte `protobuf:"bytes,7,opt,name=transactionHash,proto3" json:"transactionHash,omitempty"`
}

func (m *Transaction) Reset()                    { *m = Transaction{} }
func (m *Transaction) String() string            { return proto.CompactTextString(m) }
func (*Transaction) ProtoMessage()               {}
func (*Transaction) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type InvalidTransactionRecord struct {
	Tx      *Transaction                     `protobuf:"bytes,1,opt,name=tx" json:"tx,omitempty"`
	ErrType InvalidTransactionRecord_ErrType `protobuf:"varint,2,opt,name=errType,enum=types.InvalidTransactionRecord_ErrType" json:"errType,omitempty"`
	ErrMsg  []byte                           `protobuf:"bytes,3,opt,name=errMsg,proto3" json:"errMsg,omitempty"`
}

func (m *InvalidTransactionRecord) Reset()                    { *m = InvalidTransactionRecord{} }
func (m *InvalidTransactionRecord) String() string            { return proto.CompactTextString(m) }
func (*InvalidTransactionRecord) ProtoMessage()               {}
func (*InvalidTransactionRecord) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *InvalidTransactionRecord) GetTx() *Transaction {
	if m != nil {
		return m.Tx
	}
	return nil
}

type InvalidTransactionRecords struct {
	Records []*InvalidTransactionRecord `protobuf:"bytes,1,rep,name=records" json:"records,omitempty"`
}

func (m *InvalidTransactionRecords) Reset()                    { *m = InvalidTransactionRecords{} }
func (m *InvalidTransactionRecords) String() string            { return proto.CompactTextString(m) }
func (*InvalidTransactionRecords) ProtoMessage()               {}
func (*InvalidTransactionRecords) Descriptor() ([]byte, []int) { return fileDescriptor3, []int{2} }

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
func (*TransactionMeta) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func init() {
	proto.RegisterType((*Transaction)(nil), "types.Transaction")
	proto.RegisterType((*InvalidTransactionRecord)(nil), "types.InvalidTransactionRecord")
	proto.RegisterType((*InvalidTransactionRecords)(nil), "types.InvalidTransactionRecords")
	proto.RegisterType((*TransactionMeta)(nil), "types.TransactionMeta")
	proto.RegisterEnum("types.InvalidTransactionRecord_ErrType", InvalidTransactionRecord_ErrType_name, InvalidTransactionRecord_ErrType_value)
}

func init() { proto.RegisterFile("transaction.proto", fileDescriptor3) }

var fileDescriptor3 = []byte{
	// 382 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x84, 0x92, 0xc1, 0x6f, 0xd3, 0x30,
	0x14, 0xc6, 0xb1, 0x93, 0x36, 0xda, 0xeb, 0xd8, 0xca, 0x13, 0x9a, 0x0c, 0x42, 0x10, 0xe5, 0x42,
	0x4e, 0x39, 0x94, 0x13, 0xc7, 0xac, 0xcb, 0x46, 0x44, 0xd7, 0x20, 0x13, 0x26, 0x71, 0x9a, 0xbc,
	0xc6, 0x8c, 0x88, 0x36, 0x8e, 0x1c, 0x6f, 0xda, 0xfe, 0x38, 0xfe, 0x34, 0x24, 0x14, 0x27, 0x55,
	0x23, 0x50, 0xb5, 0x9b, 0xdf, 0xf7, 0xfb, 0x5e, 0xe2, 0xef, 0xf9, 0xc1, 0x0b, 0xa3, 0x45, 0xd5,
	0x88, 0x95, 0x29, 0x55, 0x15, 0xd5, 0x5a, 0x19, 0x85, 0x23, 0xf3, 0x58, 0xcb, 0x26, 0xf8, 0x4d,
	0x60, 0x92, 0xef, 0x20, 0x22, 0xb8, 0x3f, 0xb4, 0xda, 0x30, 0xe2, 0x93, 0xf0, 0x90, 0xdb, 0x33,
	0x1e, 0x01, 0x35, 0x8a, 0x51, 0xab, 0x50, 0xa3, 0xf0, 0x25, 0x8c, 0xee, 0xc5, 0xfa, 0x4e, 0x32,
	0xc7, 0x4a, 0x5d, 0x81, 0x6f, 0xe0, 0xc0, 0x94, 0x1b, 0xd9, 0x18, 0xb1, 0xa9, 0x99, 0xeb, 0x93,
	0xd0, 0xe1, 0x3b, 0xa1, 0xa5, 0x4d, 0x79, 0x5b, 0x09, 0x73, 0xa7, 0x25, 0x1b, 0xd9, 0xbe, 0x9d,
	0xd0, 0xfe, 0xa1, 0x2c, 0xd8, 0xd8, 0x27, 0xa1, 0xcb, 0x69, 0x59, 0x60, 0x08, 0xc7, 0x83, 0x1b,
	0x7f, 0x12, 0xcd, 0x4f, 0xe6, 0xd9, 0x9e, 0x7f, 0xe5, 0xe0, 0x0f, 0x01, 0x96, 0x56, 0xf7, 0x62,
	0x5d, 0x16, 0x83, 0x18, 0x5c, 0xae, 0x94, 0x2e, 0x30, 0x00, 0x6a, 0x1e, 0x6c, 0x94, 0xc9, 0x0c,
	0x23, 0x1b, 0x38, 0x1a, 0xba, 0xa8, 0x79, 0xc0, 0x18, 0x3c, 0xa9, 0x75, 0xfe, 0x58, 0x4b, 0x9b,
	0xf0, 0x68, 0xf6, 0xbe, 0x37, 0xee, 0xfb, 0x6a, 0x94, 0x74, 0x76, 0xbe, 0xed, 0xc3, 0x13, 0x18,
	0x4b, 0xad, 0x2f, 0x9b, 0xdb, 0x7e, 0x20, 0x7d, 0x15, 0xdc, 0x80, 0xd7, 0x7b, 0x71, 0x0a, 0x87,
	0xd9, 0xb7, 0x3c, 0x3b, 0x3f, 0x8d, 0x17, 0xf1, 0x72, 0x9e, 0x4c, 0x9f, 0xe1, 0x73, 0x38, 0xf8,
	0x9a, 0x5e, 0x9c, 0xc7, 0xe9, 0x22, 0x39, 0x9b, 0x12, 0x7c, 0x0d, 0x27, 0xe9, 0xf2, 0x2a, 0xfb,
	0x9c, 0x5c, 0xcf, 0xb3, 0x65, 0xce, 0xe3, 0x79, 0x7e, 0xdd, 0x33, 0xda, 0xb2, 0xb3, 0xe4, 0xcb,
	0x22, 0xfb, 0xfe, 0x1f, 0x73, 0x82, 0x2b, 0x78, 0xb5, 0xef, 0xa2, 0x0d, 0x7e, 0x04, 0x4f, 0x77,
	0x47, 0x46, 0x7c, 0x27, 0x9c, 0xcc, 0xde, 0x3d, 0x91, 0x8d, 0x6f, 0xfd, 0xc1, 0x05, 0x1c, 0x0f,
	0xe8, 0xa5, 0x34, 0x02, 0xdf, 0x02, 0x9c, 0xae, 0xd5, 0xea, 0x57, 0x5a, 0x15, 0xb2, 0x9b, 0xaa,
	0xcb, 0x07, 0x4a, 0xbb, 0x16, 0x1d, 0xa2, 0xf6, 0xf1, 0xbb, 0xe2, 0x66, 0x6c, 0xd7, 0xed, 0xc3,
	0xdf, 0x00, 0x00, 0x00, 0xff, 0xff, 0x6c, 0x5f, 0xae, 0x32, 0x83, 0x02, 0x00, 0x00,
}
