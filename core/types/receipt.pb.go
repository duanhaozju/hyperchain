// Code generated by protoc-gen-go.
// source: receipt.proto
// DO NOT EDIT!

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	receipt.proto

It has these top-level messages:
	Receipt
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

type Receipt_STATUS int32

const (
	Receipt_SUCCESS Receipt_STATUS = 0
	Receipt_FAILED  Receipt_STATUS = 1
)

var Receipt_STATUS_name = map[int32]string{
	0: "SUCCESS",
	1: "FAILED",
}
var Receipt_STATUS_value = map[string]int32{
	"SUCCESS": 0,
	"FAILED":  1,
}

func (x Receipt_STATUS) String() string {
	return proto.EnumName(Receipt_STATUS_name, int32(x))
}
func (Receipt_STATUS) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Receipt struct {
	PostState         []byte         `protobuf:"bytes,1,opt,name=PostState,json=postState,proto3" json:"PostState,omitempty"`
	CumulativeGasUsed int64          `protobuf:"varint,2,opt,name=CumulativeGasUsed,json=cumulativeGasUsed" json:"CumulativeGasUsed,omitempty"`
	TxHash            []byte         `protobuf:"bytes,3,opt,name=TxHash,json=txHash,proto3" json:"TxHash,omitempty"`
	ContractAddress   []byte         `protobuf:"bytes,4,opt,name=ContractAddress,json=contractAddress,proto3" json:"ContractAddress,omitempty"`
	GasUsed           int64          `protobuf:"varint,5,opt,name=GasUsed,json=gasUsed" json:"GasUsed,omitempty"`
	Ret               []byte         `protobuf:"bytes,6,opt,name=Ret,json=ret,proto3" json:"Ret,omitempty"`
	Logs              []byte         `protobuf:"bytes,7,opt,name=Logs,json=logs,proto3" json:"Logs,omitempty"`
	Status            Receipt_STATUS `protobuf:"varint,8,opt,name=Status,json=status,enum=types.Receipt_STATUS" json:"Status,omitempty"`
	Message           []byte         `protobuf:"bytes,9,opt,name=Message,json=message,proto3" json:"Message,omitempty"`
}

func (m *Receipt) Reset()                    { *m = Receipt{} }
func (m *Receipt) String() string            { return proto.CompactTextString(m) }
func (*Receipt) ProtoMessage()               {}
func (*Receipt) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func init() {
	proto.RegisterType((*Receipt)(nil), "types.Receipt")
	proto.RegisterEnum("types.Receipt_STATUS", Receipt_STATUS_name, Receipt_STATUS_value)
}

func init() { proto.RegisterFile("receipt.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 272 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x90, 0x4d, 0x4e, 0xc3, 0x30,
	0x10, 0x85, 0x69, 0x93, 0xda, 0x64, 0xf8, 0x69, 0x3a, 0x12, 0xc8, 0x0b, 0x16, 0xd0, 0x55, 0x17,
	0x90, 0x05, 0x9c, 0x20, 0x0a, 0xe5, 0x47, 0x2a, 0x12, 0x8a, 0x93, 0x03, 0x98, 0xc4, 0x0a, 0x95,
	0x5a, 0x12, 0xc5, 0x13, 0x04, 0xe7, 0xe3, 0x62, 0xa4, 0xd3, 0x76, 0x01, 0x3b, 0xcf, 0x9b, 0xe7,
	0xf7, 0xf9, 0x19, 0x4e, 0x5a, 0x5b, 0xd8, 0x65, 0x43, 0x51, 0xd3, 0xd6, 0x54, 0xe3, 0x88, 0xbe,
	0x1b, 0xeb, 0xa6, 0x3f, 0x43, 0x90, 0xe9, 0x76, 0x81, 0x17, 0x10, 0xbc, 0xd6, 0x8e, 0x34, 0x19,
	0xb2, 0x6a, 0x70, 0x39, 0x98, 0x1d, 0xa7, 0x41, 0xb3, 0x17, 0xf0, 0x1a, 0x26, 0x49, 0xb7, 0xee,
	0x56, 0x86, 0x96, 0x9f, 0xf6, 0xd1, 0xb8, 0xdc, 0xd9, 0x52, 0x0d, 0x7b, 0x97, 0x97, 0x4e, 0x8a,
	0xff, 0x0b, 0x3c, 0x07, 0x91, 0x7d, 0x3d, 0x19, 0xf7, 0xae, 0x3c, 0x0e, 0x12, 0xc4, 0x13, 0xce,
	0x60, 0x9c, 0xd4, 0x1f, 0xd4, 0x9a, 0x82, 0xe2, 0xb2, 0x6c, 0xad, 0x73, 0xca, 0x67, 0xc3, 0xb8,
	0xf8, 0x2b, 0xa3, 0x02, 0xb9, 0xa7, 0x8c, 0x98, 0x22, 0xab, 0x5d, 0x76, 0x08, 0x5e, 0x6a, 0x49,
	0x09, 0xbe, 0xe7, 0xb5, 0x96, 0x10, 0xc1, 0x5f, 0xd4, 0x95, 0x53, 0x92, 0x25, 0x7f, 0xd5, 0x9f,
	0xf1, 0x06, 0xc4, 0xe6, 0xe1, 0x9d, 0x53, 0x87, 0xbd, 0x7a, 0x7a, 0x7b, 0x16, 0x71, 0xe3, 0x68,
	0xd7, 0x36, 0xd2, 0x59, 0x9c, 0xe5, 0x3a, 0x15, 0x8e, 0x4d, 0x1b, 0xdc, 0x4b, 0x8f, 0x35, 0x95,
	0x55, 0x01, 0xa7, 0xc8, 0xf5, 0x76, 0x9c, 0x5e, 0xf5, 0x41, 0xec, 0xc5, 0x23, 0x90, 0x3a, 0x4f,
	0x92, 0xb9, 0xd6, 0xe1, 0x01, 0x02, 0x88, 0x87, 0xf8, 0x79, 0x31, 0xbf, 0x0f, 0x07, 0x6f, 0x82,
	0xff, 0xf4, 0xee, 0x37, 0x00, 0x00, 0xff, 0xff, 0x3d, 0xb2, 0xc7, 0x7f, 0x64, 0x01, 0x00, 0x00,
}
