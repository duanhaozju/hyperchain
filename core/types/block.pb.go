// Code generated by protoc-gen-go.
// source: block.proto
// DO NOT EDIT!

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	block.proto
	chain.proto
	transaction.proto
	transaction_value.proto

It has these top-level messages:
	Block
	Chain
	Transaction
	TransactionValue
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

type Block struct {
	ParentHash   []byte         `protobuf:"bytes,1,opt,name=parentHash,proto3" json:"parentHash,omitempty"`
	BlockHash    []byte         `protobuf:"bytes,2,opt,name=blockHash,proto3" json:"blockHash,omitempty"`
	Transactions []*Transaction `protobuf:"bytes,3,rep,name=transactions" json:"transactions,omitempty"`
	Timestamp    int64          `protobuf:"varint,4,opt,name=timestamp" json:"timestamp,omitempty"`
	MerkleRoot   []byte         `protobuf:"bytes,5,opt,name=merkleRoot,proto3" json:"merkleRoot,omitempty"`
	Number       uint64         `protobuf:"varint,6,opt,name=number" json:"number,omitempty"`
	WriteTime    int64          `protobuf:"varint,7,opt,name=writeTime" json:"writeTime,omitempty"`
	CommitTime   int64          `protobuf:"varint,8,opt,name=commitTime" json:"commitTime,omitempty"`
}

func (m *Block) Reset()                    { *m = Block{} }
func (m *Block) String() string            { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()               {}
func (*Block) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Block) GetTransactions() []*Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

func init() {
	proto.RegisterType((*Block)(nil), "types.Block")
}

func init() { proto.RegisterFile("block.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 214 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x54, 0x90, 0xbb, 0x6a, 0xc3, 0x30,
	0x14, 0x40, 0xf1, 0xb3, 0xad, 0xec, 0xa5, 0x1a, 0x8a, 0x28, 0xa5, 0x98, 0x4e, 0x9e, 0x3c, 0x34,
	0x90, 0x0f, 0xc8, 0x94, 0x59, 0xf8, 0x07, 0x64, 0x23, 0x88, 0xb0, 0xf5, 0x40, 0xba, 0x21, 0xe4,
	0x33, 0xf2, 0xc7, 0xb1, 0xe4, 0x10, 0x39, 0xa3, 0xce, 0xb9, 0x1c, 0x5d, 0x2e, 0xaa, 0x86, 0x59,
	0x8f, 0x53, 0x67, 0xac, 0x06, 0x8d, 0x0b, 0xb8, 0x1a, 0xee, 0xbe, 0x3f, 0xc1, 0x32, 0xe5, 0xd8,
	0x08, 0x42, 0xab, 0xd5, 0xfc, 0xdd, 0x52, 0x54, 0x1c, 0xfc, 0x24, 0xfe, 0x45, 0xc8, 0x30, 0xcb,
	0x15, 0x1c, 0x99, 0x3b, 0x91, 0xa4, 0x49, 0xda, 0x9a, 0x6e, 0x08, 0xfe, 0x41, 0x1f, 0x21, 0x19,
	0x74, 0x1a, 0x74, 0x04, 0x78, 0x8f, 0xea, 0x4d, 0xdc, 0x91, 0xac, 0xc9, 0xda, 0xea, 0x1f, 0x77,
	0xe1, 0xe3, 0xae, 0x8f, 0x8a, 0xbe, 0xcc, 0xf9, 0x2a, 0x08, 0xc9, 0x1d, 0x30, 0x69, 0x48, 0xbe,
	0x54, 0x33, 0x1a, 0x81, 0xdf, 0x49, 0x72, 0x3b, 0xcd, 0x9c, 0x6a, 0x0d, 0xa4, 0x58, 0x77, 0x8a,
	0x04, 0x7f, 0xa1, 0x52, 0x9d, 0xe5, 0xc0, 0x2d, 0x29, 0x17, 0x97, 0xd3, 0xc7, 0xcb, 0x57, 0x2f,
	0x56, 0x00, 0xef, 0x97, 0x12, 0x79, 0x5b, 0xab, 0x4f, 0xe0, 0xab, 0xa3, 0x96, 0x52, 0x40, 0xd0,
	0xef, 0x41, 0x6f, 0xc8, 0x50, 0x86, 0xd3, 0xec, 0xee, 0x01, 0x00, 0x00, 0xff, 0xff, 0xf6, 0xb8,
	0x58, 0xd4, 0x43, 0x01, 0x00, 0x00,
}
