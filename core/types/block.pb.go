// Code generated by protoc-gen-go.
// source: block.proto
// DO NOT EDIT!

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	block.proto
	block_wrapper.proto
	chain.proto
	receipt.proto
	receipt_wrapper.proto
	transaction.proto
	transaction_value.proto
	transaction_wrapper.proto

It has these top-level messages:
	Block
	Blocks
	BlockWrapper
	Chain
	ChainStatus
	Receipt
	ReceiptWrapper
	Transaction
	InvalidTransactionRecord
	InvalidTransactionRecords
	TransactionMeta
	TransactionValue
	TransactionWrapper
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
	Version      []byte         `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	ParentHash   []byte         `protobuf:"bytes,2,opt,name=parentHash,proto3" json:"parentHash,omitempty"`
	BlockHash    []byte         `protobuf:"bytes,3,opt,name=blockHash,proto3" json:"blockHash,omitempty"`
	Transactions []*Transaction `protobuf:"bytes,4,rep,name=transactions" json:"transactions,omitempty"`
	Timestamp    int64          `protobuf:"varint,5,opt,name=timestamp" json:"timestamp,omitempty"`
	MerkleRoot   []byte         `protobuf:"bytes,6,opt,name=merkleRoot,proto3" json:"merkleRoot,omitempty"`
	TxRoot       []byte         `protobuf:"bytes,7,opt,name=txRoot,proto3" json:"txRoot,omitempty"`
	ReceiptRoot  []byte         `protobuf:"bytes,8,opt,name=receiptRoot,proto3" json:"receiptRoot,omitempty"`
	Number       uint64         `protobuf:"varint,9,opt,name=number" json:"number,omitempty"`
	WriteTime    int64          `protobuf:"varint,10,opt,name=writeTime" json:"writeTime,omitempty"`
	CommitTime   int64          `protobuf:"varint,11,opt,name=commitTime" json:"commitTime,omitempty"`
	EvmTime      int64          `protobuf:"varint,12,opt,name=evmTime" json:"evmTime,omitempty"`
}

func (m *Block) Reset()                    { *m = Block{} }
func (m *Block) String() string            { return proto.CompactTextString(m) }
func (*Block) ProtoMessage()               {}
func (*Block) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Block) GetVersion() []byte {
	if m != nil {
		return m.Version
	}
	return nil
}

func (m *Block) GetParentHash() []byte {
	if m != nil {
		return m.ParentHash
	}
	return nil
}

func (m *Block) GetBlockHash() []byte {
	if m != nil {
		return m.BlockHash
	}
	return nil
}

func (m *Block) GetTransactions() []*Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

func (m *Block) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

func (m *Block) GetMerkleRoot() []byte {
	if m != nil {
		return m.MerkleRoot
	}
	return nil
}

func (m *Block) GetTxRoot() []byte {
	if m != nil {
		return m.TxRoot
	}
	return nil
}

func (m *Block) GetReceiptRoot() []byte {
	if m != nil {
		return m.ReceiptRoot
	}
	return nil
}

func (m *Block) GetNumber() uint64 {
	if m != nil {
		return m.Number
	}
	return 0
}

func (m *Block) GetWriteTime() int64 {
	if m != nil {
		return m.WriteTime
	}
	return 0
}

func (m *Block) GetCommitTime() int64 {
	if m != nil {
		return m.CommitTime
	}
	return 0
}

func (m *Block) GetEvmTime() int64 {
	if m != nil {
		return m.EvmTime
	}
	return 0
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
	// 279 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x54, 0xd1, 0xc1, 0x4e, 0x84, 0x30,
	0x10, 0x06, 0xe0, 0xb0, 0x2c, 0xac, 0x3b, 0x70, 0xb1, 0x07, 0xd3, 0x18, 0x63, 0x08, 0x27, 0x0e,
	0x86, 0x83, 0x26, 0x3e, 0x80, 0x27, 0xcf, 0xcd, 0xbe, 0x40, 0x21, 0x93, 0x6c, 0xb3, 0x5b, 0x4a,
	0xda, 0x71, 0xd5, 0xa7, 0xd7, 0x74, 0x60, 0x05, 0x8f, 0xf3, 0x7f, 0x85, 0xfe, 0x99, 0x42, 0xd1,
	0x9d, 0x5d, 0x7f, 0x6a, 0x47, 0xef, 0xc8, 0x89, 0x8c, 0xbe, 0x47, 0x0c, 0xf7, 0xb7, 0xe4, 0xf5,
	0x10, 0x74, 0x4f, 0xc6, 0x0d, 0x93, 0xd4, 0x3f, 0x1b, 0xc8, 0xde, 0xe2, 0x49, 0x21, 0x61, 0x77,
	0x41, 0x1f, 0x8c, 0x1b, 0x64, 0x52, 0x25, 0x4d, 0xa9, 0xae, 0xa3, 0x78, 0x04, 0x18, 0xb5, 0xc7,
	0x81, 0xde, 0x75, 0x38, 0xca, 0x0d, 0xe3, 0x2a, 0x11, 0x0f, 0xb0, 0xe7, 0xcb, 0x98, 0x53, 0xe6,
	0x25, 0x10, 0xaf, 0x50, 0xae, 0xae, 0x0d, 0x72, 0x5b, 0xa5, 0x4d, 0xf1, 0x2c, 0x5a, 0xae, 0xd4,
	0x1e, 0x16, 0x52, 0xff, 0xce, 0xc5, 0xbf, 0x92, 0xb1, 0x18, 0x48, 0xdb, 0x51, 0x66, 0x55, 0xd2,
	0xa4, 0x6a, 0x09, 0x62, 0x27, 0x8b, 0xfe, 0x74, 0x46, 0xe5, 0x1c, 0xc9, 0x7c, 0xea, 0xb4, 0x24,
	0xe2, 0x0e, 0x72, 0xfa, 0x62, 0xdb, 0xb1, 0xcd, 0x93, 0xa8, 0xa0, 0xf0, 0xd8, 0xa3, 0x19, 0x89,
	0xf1, 0x86, 0x71, 0x1d, 0xc5, 0x2f, 0x87, 0x0f, 0xdb, 0xa1, 0x97, 0xfb, 0x2a, 0x69, 0xb6, 0x6a,
	0x9e, 0x62, 0x9f, 0x4f, 0x6f, 0x08, 0x0f, 0xc6, 0xa2, 0x84, 0xa9, 0xcf, 0x5f, 0x10, 0xfb, 0xf4,
	0xce, 0x5a, 0x43, 0xcc, 0x05, 0xf3, 0x2a, 0x89, 0xdb, 0xc5, 0x8b, 0x65, 0x2c, 0x19, 0xaf, 0x63,
	0xfd, 0x04, 0x39, 0x3f, 0x40, 0x10, 0x35, 0x64, 0x9d, 0xa6, 0xfe, 0x28, 0x13, 0x5e, 0x51, 0x39,
	0xaf, 0x88, 0x55, 0x4d, 0xd4, 0xe5, 0xfc, 0x6c, 0x2f, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0x34,
	0x03, 0xde, 0x0e, 0xdf, 0x01, 0x00, 0x00,
}
