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
	Blocks
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

type Blocks struct {
	Batch []*Block `protobuf:"bytes,1,rep,name=batch" json:"batch,omitempty"`
}

func (m *Blocks) Reset()                    { *m = Blocks{} }
func (m *Blocks) String() string            { return proto.CompactTextString(m) }
func (*Blocks) ProtoMessage()               {}
func (*Blocks) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Blocks) GetBatch() []*Block {
	if m != nil {
		return m.Batch
	}
	return nil
}

func init() {
	proto.RegisterType((*Block)(nil), "types.Block")
	proto.RegisterType((*Blocks)(nil), "types.Blocks")
}

func init() { proto.RegisterFile("block.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 237 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x54, 0x90, 0xcf, 0x4a, 0xc4, 0x30,
	0x10, 0xc6, 0xd9, 0xed, 0xb6, 0xea, 0x6c, 0x2f, 0xe6, 0x20, 0x41, 0x44, 0x96, 0x9e, 0xf6, 0x20,
	0x3d, 0x28, 0xf8, 0x00, 0x9e, 0x3c, 0x87, 0x7d, 0x81, 0x34, 0x04, 0x0c, 0xbb, 0xf9, 0x43, 0x32,
	0x22, 0x3e, 0x86, 0x6f, 0x6c, 0x33, 0x29, 0xa6, 0x1e, 0xf3, 0xfd, 0x86, 0x5f, 0x66, 0x3e, 0xd8,
	0x4f, 0x17, 0xaf, 0xce, 0x63, 0x88, 0x1e, 0x3d, 0x6b, 0xf1, 0x3b, 0xe8, 0x74, 0x7f, 0x8b, 0x51,
	0xba, 0x24, 0x15, 0x1a, 0xef, 0x0a, 0x19, 0x7e, 0xb6, 0xd0, 0xbe, 0xe5, 0x49, 0xf6, 0x08, 0x10,
	0x64, 0xd4, 0x0e, 0xdf, 0x65, 0xfa, 0xe0, 0x9b, 0xc3, 0xe6, 0xd8, 0x8b, 0x55, 0xc2, 0x1e, 0xe0,
	0x86, 0x94, 0x84, 0xb7, 0x84, 0x6b, 0xc0, 0x5e, 0xa1, 0x5f, 0xc9, 0x13, 0x6f, 0x0e, 0xcd, 0x71,
	0xff, 0xcc, 0x46, 0xfa, 0x78, 0x3c, 0x55, 0x24, 0xfe, 0xcd, 0x65, 0x2b, 0x1a, 0xab, 0x13, 0x4a,
	0x1b, 0xf8, 0x6e, 0xb6, 0x36, 0xa2, 0x06, 0x79, 0x27, 0xab, 0xe3, 0xf9, 0xa2, 0x85, 0xf7, 0xc8,
	0xdb, 0xb2, 0x53, 0x4d, 0xd8, 0x1d, 0x74, 0xee, 0xd3, 0x4e, 0x3a, 0xf2, 0x6e, 0x66, 0x3b, 0xb1,
	0xbc, 0xb2, 0xf5, 0x2b, 0x1a, 0xd4, 0xa7, 0xd9, 0xc4, 0xaf, 0x8a, 0xf5, 0x2f, 0xc8, 0x56, 0xe5,
	0xad, 0x35, 0x48, 0xf8, 0x9a, 0xf0, 0x2a, 0x19, 0x9e, 0xa0, 0xa3, 0x4a, 0x12, 0x1b, 0xa0, 0x9d,
	0x24, 0xaa, 0x5c, 0x47, 0x3e, 0xa7, 0x5f, 0xce, 0x21, 0x2a, 0x0a, 0x9a, 0x3a, 0x2a, 0xf2, 0xe5,
	0x37, 0x00, 0x00, 0xff, 0xff, 0xfb, 0xf7, 0x70, 0x73, 0x71, 0x01, 0x00, 0x00,
}
