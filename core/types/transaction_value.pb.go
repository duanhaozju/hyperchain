// Code generated by protoc-gen-go.
// source: transaction_value.proto
// DO NOT EDIT!

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TransactionValue struct {
	Price              int64  `protobuf:"varint,1,opt,name=price" json:"price,omitempty"`
	GasLimit           int64  `protobuf:"varint,2,opt,name=gasLimit" json:"gasLimit,omitempty"`
	Amount             int64  `protobuf:"varint,3,opt,name=amount" json:"amount,omitempty"`
	Payload            []byte `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
	EncryptedAmount    []byte `protobuf:"bytes,5,opt,name=encryptedAmount,proto3" json:"encryptedAmount,omitempty"`
	HomomorphicAmount  []byte `protobuf:"bytes,6,opt,name=homomorphicAmount,proto3" json:"homomorphicAmount,omitempty"`
	HomomorphicBalance []byte `protobuf:"bytes,7,opt,name=homomorphicBalance,proto3" json:"homomorphicBalance,omitempty"`
}

func (m *TransactionValue) Reset()                    { *m = TransactionValue{} }
func (m *TransactionValue) String() string            { return proto.CompactTextString(m) }
func (*TransactionValue) ProtoMessage()               {}
func (*TransactionValue) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{0} }

func init() {
	proto.RegisterType((*TransactionValue)(nil), "types.TransactionValue")
}

func init() { proto.RegisterFile("transaction_value.proto", fileDescriptor6) }

var fileDescriptor6 = []byte{
	// 202 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x12, 0x2f, 0x29, 0x4a, 0xcc,
	0x2b, 0x4e, 0x4c, 0x2e, 0xc9, 0xcc, 0xcf, 0x8b, 0x2f, 0x4b, 0xcc, 0x29, 0x4d, 0xd5, 0x2b, 0x28,
	0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2d, 0xa9, 0x2c, 0x48, 0x2d, 0x56, 0xfa, 0xc7, 0xc8, 0x25, 0x10,
	0x82, 0x50, 0x12, 0x06, 0x52, 0x21, 0x24, 0xc2, 0xc5, 0x5a, 0x50, 0x94, 0x99, 0x9c, 0x2a, 0xc1,
	0xa8, 0xc0, 0xa8, 0xc1, 0x1c, 0x04, 0xe1, 0x08, 0x49, 0x71, 0x71, 0xa4, 0x27, 0x16, 0xfb, 0x64,
	0xe6, 0x66, 0x96, 0x48, 0x30, 0x81, 0x25, 0xe0, 0x7c, 0x21, 0x31, 0x2e, 0xb6, 0xc4, 0xdc, 0xfc,
	0xd2, 0xbc, 0x12, 0x09, 0x66, 0xb0, 0x0c, 0x94, 0x27, 0x24, 0xc1, 0xc5, 0x5e, 0x90, 0x58, 0x99,
	0x93, 0x9f, 0x98, 0x22, 0xc1, 0x02, 0x94, 0xe0, 0x09, 0x82, 0x71, 0x85, 0x34, 0xb8, 0xf8, 0x53,
	0xf3, 0x92, 0x8b, 0x2a, 0x0b, 0x4a, 0x52, 0x53, 0x1c, 0x21, 0x5a, 0x59, 0xc1, 0x2a, 0xd0, 0x85,
	0x85, 0x74, 0xb8, 0x04, 0x33, 0xf2, 0x73, 0x81, 0xb0, 0xa8, 0x20, 0x23, 0x33, 0x19, 0xaa, 0x96,
	0x0d, 0xac, 0x16, 0x53, 0x42, 0x48, 0x8f, 0x4b, 0x08, 0x49, 0xd0, 0x29, 0x31, 0x27, 0x31, 0x0f,
	0xe8, 0x11, 0x76, 0xb0, 0x72, 0x2c, 0x32, 0x49, 0x6c, 0xe0, 0xe0, 0x30, 0x06, 0x04, 0x00, 0x00,
	0xff, 0xff, 0x94, 0x3c, 0xc9, 0x47, 0x29, 0x01, 0x00, 0x00,
}
