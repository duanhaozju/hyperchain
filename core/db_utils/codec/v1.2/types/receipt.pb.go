// Code generated by protoc-gen-go.
// source: receipt.proto
// DO NOT EDIT!

/*
Package types is a generated protocol buffer package.

It is generated from these files:
	receipt.proto
	transaction.proto

It has these top-level messages:
	ReceiptV1_2
	TransactionV1_2
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

type ReceiptV1_2_STATUS int32

const (
	ReceiptV1_2_SUCCESS ReceiptV1_2_STATUS = 0
	ReceiptV1_2_FAILED  ReceiptV1_2_STATUS = 1
)

var ReceiptV1_2_STATUS_name = map[int32]string{
	0: "SUCCESS",
	1: "FAILED",
}
var ReceiptV1_2_STATUS_value = map[string]int32{
	"SUCCESS": 0,
	"FAILED":  1,
}

func (x ReceiptV1_2_STATUS) String() string {
	return proto.EnumName(ReceiptV1_2_STATUS_name, int32(x))
}
func (ReceiptV1_2_STATUS) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type ReceiptV1_2 struct {
	Version           []byte             `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	PostState         []byte             `protobuf:"bytes,2,opt,name=PostState,proto3" json:"PostState,omitempty"`
	CumulativeGasUsed int64              `protobuf:"varint,3,opt,name=CumulativeGasUsed" json:"CumulativeGasUsed,omitempty"`
	TxHash            []byte             `protobuf:"bytes,4,opt,name=TxHash,proto3" json:"TxHash,omitempty"`
	ContractAddress   []byte             `protobuf:"bytes,5,opt,name=ContractAddress,proto3" json:"ContractAddress,omitempty"`
	GasUsed           int64              `protobuf:"varint,6,opt,name=GasUsed" json:"GasUsed,omitempty"`
	Ret               []byte             `protobuf:"bytes,7,opt,name=Ret,proto3" json:"Ret,omitempty"`
	Logs              []byte             `protobuf:"bytes,8,opt,name=Logs,proto3" json:"Logs,omitempty"`
	Status            ReceiptV1_2_STATUS `protobuf:"varint,9,opt,name=Status,enum=types.ReceiptV1_2_STATUS" json:"Status,omitempty"`
	Message           []byte             `protobuf:"bytes,10,opt,name=Message,proto3" json:"Message,omitempty"`
}

func (m *ReceiptV1_2) Reset()                    { *m = ReceiptV1_2{} }
func (m *ReceiptV1_2) String() string            { return proto.CompactTextString(m) }
func (*ReceiptV1_2) ProtoMessage()               {}
func (*ReceiptV1_2) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ReceiptV1_2) GetVersion() []byte {
	if m != nil {
		return m.Version
	}
	return nil
}

func (m *ReceiptV1_2) GetPostState() []byte {
	if m != nil {
		return m.PostState
	}
	return nil
}

func (m *ReceiptV1_2) GetCumulativeGasUsed() int64 {
	if m != nil {
		return m.CumulativeGasUsed
	}
	return 0
}

func (m *ReceiptV1_2) GetTxHash() []byte {
	if m != nil {
		return m.TxHash
	}
	return nil
}

func (m *ReceiptV1_2) GetContractAddress() []byte {
	if m != nil {
		return m.ContractAddress
	}
	return nil
}

func (m *ReceiptV1_2) GetGasUsed() int64 {
	if m != nil {
		return m.GasUsed
	}
	return 0
}

func (m *ReceiptV1_2) GetRet() []byte {
	if m != nil {
		return m.Ret
	}
	return nil
}

func (m *ReceiptV1_2) GetLogs() []byte {
	if m != nil {
		return m.Logs
	}
	return nil
}

func (m *ReceiptV1_2) GetStatus() ReceiptV1_2_STATUS {
	if m != nil {
		return m.Status
	}
	return ReceiptV1_2_SUCCESS
}

func (m *ReceiptV1_2) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func init() {
	proto.RegisterType((*ReceiptV1_2)(nil), "types.ReceiptV1_2")
	proto.RegisterEnum("types.ReceiptV1_2_STATUS", ReceiptV1_2_STATUS_name, ReceiptV1_2_STATUS_value)
}

func init() { proto.RegisterFile("receipt.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 275 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x64, 0x50, 0xcb, 0x4a, 0xc3, 0x40,
	0x14, 0x35, 0x7d, 0x4c, 0xec, 0xad, 0x8f, 0x78, 0x17, 0x32, 0x82, 0x8b, 0xd8, 0x55, 0x16, 0x12,
	0x68, 0xfd, 0x82, 0x10, 0xeb, 0x03, 0x2a, 0x48, 0x26, 0x71, 0x2b, 0x63, 0x33, 0xd4, 0x80, 0x76,
	0x42, 0xee, 0xa4, 0xe8, 0x07, 0xfb, 0x1f, 0x92, 0x49, 0x82, 0xa2, 0xbb, 0x7b, 0x1e, 0x33, 0xe7,
	0x70, 0xe0, 0xb0, 0x52, 0x6b, 0x55, 0x94, 0x26, 0x2c, 0x2b, 0x6d, 0x34, 0x8e, 0xcd, 0x67, 0xa9,
	0x68, 0xf6, 0x35, 0x80, 0x69, 0xd2, 0x0a, 0x4f, 0xf3, 0xe7, 0x05, 0x72, 0x70, 0x77, 0xaa, 0xa2,
	0x42, 0x6f, 0xb9, 0xe3, 0x3b, 0xc1, 0x41, 0xd2, 0x43, 0x3c, 0x87, 0xc9, 0xa3, 0x26, 0x23, 0x8c,
	0x34, 0x8a, 0x0f, 0xac, 0xf6, 0x43, 0xe0, 0x25, 0x9c, 0xc4, 0xf5, 0x7b, 0xfd, 0x26, 0x4d, 0xb1,
	0x53, 0xb7, 0x92, 0x32, 0x52, 0x39, 0x1f, 0xfa, 0x4e, 0x30, 0x4c, 0xfe, 0x0b, 0x78, 0x0a, 0x2c,
	0xfd, 0xb8, 0x93, 0xf4, 0xca, 0x47, 0xf6, 0xa3, 0x0e, 0x61, 0x00, 0xc7, 0xb1, 0xde, 0x9a, 0x4a,
	0xae, 0x4d, 0x94, 0xe7, 0x95, 0x22, 0xe2, 0x63, 0x6b, 0xf8, 0x4b, 0x37, 0x3d, 0xfb, 0x14, 0x66,
	0x53, 0x7a, 0x88, 0x1e, 0x0c, 0x13, 0x65, 0xb8, 0x6b, 0xdf, 0x35, 0x27, 0x22, 0x8c, 0x56, 0x7a,
	0x43, 0x7c, 0xdf, 0x52, 0xf6, 0xc6, 0x39, 0xb0, 0xa6, 0x78, 0x4d, 0x7c, 0xe2, 0x3b, 0xc1, 0xd1,
	0xe2, 0x2c, 0xb4, 0x7b, 0x84, 0xbf, 0xb6, 0x08, 0x45, 0x1a, 0xa5, 0x99, 0x48, 0x3a, 0x63, 0x13,
	0xf9, 0xa0, 0x88, 0xe4, 0x46, 0x71, 0x68, 0xa7, 0xe9, 0xe0, 0xec, 0x02, 0x58, 0xeb, 0xc5, 0x29,
	0xb8, 0x22, 0x8b, 0xe3, 0xa5, 0x10, 0xde, 0x1e, 0x02, 0xb0, 0x9b, 0xe8, 0x7e, 0xb5, 0xbc, 0xf6,
	0x9c, 0x17, 0x66, 0x57, 0xbf, 0xfa, 0x0e, 0x00, 0x00, 0xff, 0xff, 0xe4, 0x69, 0x16, 0x48, 0x86,
	0x01, 0x00, 0x00,
}
