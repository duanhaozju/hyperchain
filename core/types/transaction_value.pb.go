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
	Update             bool   `protobuf:"varint,8,opt,name=update" json:"update,omitempty"`
}

func (m *TransactionValue) Reset()                    { *m = TransactionValue{} }
func (m *TransactionValue) String() string            { return proto.CompactTextString(m) }
func (*TransactionValue) ProtoMessage()               {}
func (*TransactionValue) Descriptor() ([]byte, []int) { return fileDescriptor6, []int{0} }

func (m *TransactionValue) GetPrice() int64 {
	if m != nil {
		return m.Price
	}
	return 0
}

func (m *TransactionValue) GetGasLimit() int64 {
	if m != nil {
		return m.GasLimit
	}
	return 0
}

func (m *TransactionValue) GetAmount() int64 {
	if m != nil {
		return m.Amount
	}
	return 0
}

func (m *TransactionValue) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *TransactionValue) GetEncryptedAmount() []byte {
	if m != nil {
		return m.EncryptedAmount
	}
	return nil
}

func (m *TransactionValue) GetHomomorphicAmount() []byte {
	if m != nil {
		return m.HomomorphicAmount
	}
	return nil
}

func (m *TransactionValue) GetHomomorphicBalance() []byte {
	if m != nil {
		return m.HomomorphicBalance
	}
	return nil
}

func (m *TransactionValue) GetUpdate() bool {
	if m != nil {
		return m.Update
	}
	return false
}

func init() {
	proto.RegisterType((*TransactionValue)(nil), "types.TransactionValue")
}

func init() { proto.RegisterFile("transaction_value.proto", fileDescriptor6) }

var fileDescriptor6 = []byte{
	// 216 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x90, 0xcf, 0x4a, 0xc4, 0x30,
	0x10, 0x87, 0xc9, 0xae, 0xed, 0x96, 0x41, 0x50, 0x07, 0xd1, 0xe0, 0xa9, 0x78, 0xca, 0x41, 0xf6,
	0xe2, 0x13, 0xe8, 0xd9, 0x53, 0x11, 0xaf, 0x32, 0xa6, 0xc1, 0x0d, 0x34, 0x99, 0x90, 0x4e, 0x85,
	0xbe, 0x88, 0xcf, 0x2b, 0x9b, 0xad, 0x7f, 0x50, 0x8f, 0xdf, 0xef, 0xfb, 0x02, 0x61, 0xe0, 0x52,
	0x32, 0xc5, 0x91, 0xac, 0x78, 0x8e, 0xcf, 0x6f, 0x34, 0x4c, 0x6e, 0x9b, 0x32, 0x0b, 0x63, 0x25,
	0x73, 0x72, 0xe3, 0xf5, 0xfb, 0x0a, 0x4e, 0x1f, 0xbf, 0x93, 0xa7, 0x7d, 0x81, 0xe7, 0x50, 0xa5,
	0xec, 0xad, 0xd3, 0xaa, 0x55, 0x66, 0xdd, 0x1d, 0x00, 0xaf, 0xa0, 0x79, 0xa5, 0xf1, 0xc1, 0x07,
	0x2f, 0x7a, 0x55, 0xc4, 0x17, 0xe3, 0x05, 0xd4, 0x14, 0x78, 0x8a, 0xa2, 0xd7, 0xc5, 0x2c, 0x84,
	0x1a, 0x36, 0x89, 0xe6, 0x81, 0xa9, 0xd7, 0x47, 0xad, 0x32, 0xc7, 0xdd, 0x27, 0xa2, 0x81, 0x13,
	0x17, 0x6d, 0x9e, 0x93, 0xb8, 0xfe, 0xee, 0xf0, 0xb4, 0x2a, 0xc5, 0xef, 0x19, 0x6f, 0xe0, 0x6c,
	0xc7, 0x81, 0x03, 0xe7, 0xb4, 0xf3, 0x76, 0x69, 0xeb, 0xd2, 0xfe, 0x15, 0xb8, 0x05, 0xfc, 0x31,
	0xde, 0xd3, 0x40, 0xd1, 0x3a, 0xbd, 0x29, 0xf9, 0x3f, 0x66, 0xff, 0xf3, 0x29, 0xf5, 0x24, 0x4e,
	0x37, 0xad, 0x32, 0x4d, 0xb7, 0xd0, 0x4b, 0x5d, 0xce, 0x74, 0xfb, 0x11, 0x00, 0x00, 0xff, 0xff,
	0x74, 0x3d, 0xc5, 0xa8, 0x41, 0x01, 0x00, 0x00,
}
