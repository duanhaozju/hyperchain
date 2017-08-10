// Code generated by protoc-gen-go.
// source: block_wrapper.proto
// DO NOT EDIT!

/*
Package wrapper is a generated protocol buffer package.

It is generated from these files:
	block_wrapper.proto
	receipt_wrapper.proto
	transaction_wrapper.proto

It has these top-level messages:
	BlockWrapper
	ReceiptWrapper
	TransactionWrapper
*/
package wrapper

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

type BlockWrapper struct {
	BlockVersion []byte `protobuf:"bytes,1,opt,name=blockVersion,proto3" json:"blockVersion,omitempty"`
	Block        []byte `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
}

func (m *BlockWrapper) Reset()                    { *m = BlockWrapper{} }
func (m *BlockWrapper) String() string            { return proto.CompactTextString(m) }
func (*BlockWrapper) ProtoMessage()               {}
func (*BlockWrapper) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *BlockWrapper) GetBlockVersion() []byte {
	if m != nil {
		return m.BlockVersion
	}
	return nil
}

func (m *BlockWrapper) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockWrapper)(nil), "wrapper.BlockWrapper")
}

func init() { proto.RegisterFile("block_wrapper.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 97 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0xca, 0xc9, 0x4f,
	0xce, 0x8e, 0x2f, 0x2f, 0x4a, 0x2c, 0x28, 0x48, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x87, 0x72, 0x95, 0x3c, 0xb8, 0x78, 0x9c, 0x40, 0xf2, 0xe1, 0x10, 0xbe, 0x90, 0x12, 0x17,
	0x0f, 0x58, 0x7d, 0x58, 0x6a, 0x51, 0x71, 0x66, 0x7e, 0x9e, 0x04, 0xa3, 0x02, 0xa3, 0x06, 0x4f,
	0x10, 0x8a, 0x98, 0x90, 0x08, 0x17, 0x2b, 0x98, 0x2f, 0xc1, 0x04, 0x96, 0x84, 0x70, 0x92, 0xd8,
	0xc0, 0x26, 0x1b, 0x03, 0x02, 0x00, 0x00, 0xff, 0xff, 0x49, 0x67, 0xd4, 0x7f, 0x70, 0x00, 0x00,
	0x00,
}
