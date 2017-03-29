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

type TransactionValue_Opcode int32

const (
	TransactionValue_NORMAL   TransactionValue_Opcode = 0
	TransactionValue_UPDATE   TransactionValue_Opcode = 1
	TransactionValue_FREEZE   TransactionValue_Opcode = 2
	TransactionValue_UNFREEZE TransactionValue_Opcode = 3
	TransactionValue_ARCHIEVE TransactionValue_Opcode = 100
)

var TransactionValue_Opcode_name = map[int32]string{
	0:   "NORMAL",
	1:   "UPDATE",
	2:   "FREEZE",
	3:   "UNFREEZE",
	100: "ARCHIEVE",
}
var TransactionValue_Opcode_value = map[string]int32{
	"NORMAL":   0,
	"UPDATE":   1,
	"FREEZE":   2,
	"UNFREEZE": 3,
	"ARCHIEVE": 100,
}

func (x TransactionValue_Opcode) String() string {
	return proto.EnumName(TransactionValue_Opcode_name, int32(x))
}
func (TransactionValue_Opcode) EnumDescriptor() ([]byte, []int) { return fileDescriptor7, []int{0, 0} }

type TransactionValue struct {
	Price              int64                   `protobuf:"varint,1,opt,name=price" json:"price,omitempty"`
	GasLimit           int64                   `protobuf:"varint,2,opt,name=gasLimit" json:"gasLimit,omitempty"`
	Amount             int64                   `protobuf:"varint,3,opt,name=amount" json:"amount,omitempty"`
	Payload            []byte                  `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
	EncryptedAmount    []byte                  `protobuf:"bytes,5,opt,name=encryptedAmount,proto3" json:"encryptedAmount,omitempty"`
	HomomorphicAmount  []byte                  `protobuf:"bytes,6,opt,name=homomorphicAmount,proto3" json:"homomorphicAmount,omitempty"`
	HomomorphicBalance []byte                  `protobuf:"bytes,7,opt,name=homomorphicBalance,proto3" json:"homomorphicBalance,omitempty"`
	Op                 TransactionValue_Opcode `protobuf:"varint,8,opt,name=op,enum=types.TransactionValue_Opcode" json:"op,omitempty"`
}

func (m *TransactionValue) Reset()                    { *m = TransactionValue{} }
func (m *TransactionValue) String() string            { return proto.CompactTextString(m) }
func (*TransactionValue) ProtoMessage()               {}
func (*TransactionValue) Descriptor() ([]byte, []int) { return fileDescriptor7, []int{0} }

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

func (m *TransactionValue) GetOp() TransactionValue_Opcode {
	if m != nil {
		return m.Op
	}
	return TransactionValue_NORMAL
}

func init() {
	proto.RegisterType((*TransactionValue)(nil), "types.TransactionValue")
	proto.RegisterEnum("types.TransactionValue_Opcode", TransactionValue_Opcode_name, TransactionValue_Opcode_value)
}

func init() { proto.RegisterFile("transaction_value.proto", fileDescriptor7) }

var fileDescriptor7 = []byte{
	// 283 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x6c, 0x91, 0xcf, 0x4b, 0xfb, 0x30,
	0x14, 0xc0, 0xbf, 0xed, 0xbe, 0xeb, 0xc6, 0x63, 0x68, 0x7d, 0x88, 0x06, 0x0f, 0x52, 0x76, 0xea,
	0x41, 0x72, 0xd0, 0xbf, 0xa0, 0x6a, 0x64, 0xc2, 0xdc, 0x24, 0x6c, 0x3b, 0x78, 0x91, 0x98, 0x06,
	0x57, 0x68, 0x9b, 0xd0, 0x66, 0x42, 0xff, 0x73, 0x8f, 0xd2, 0xb4, 0xfe, 0x60, 0x7a, 0x7b, 0x9f,
	0xf7, 0xf9, 0xe4, 0xf2, 0x02, 0xa7, 0xb6, 0x12, 0x65, 0x2d, 0xa4, 0xcd, 0x74, 0xf9, 0xfc, 0x26,
	0xf2, 0x9d, 0xa2, 0xa6, 0xd2, 0x56, 0xe3, 0xd0, 0x36, 0x46, 0xd5, 0xd3, 0x77, 0x1f, 0xc2, 0xd5,
	0x77, 0xb2, 0x69, 0x0b, 0x3c, 0x86, 0xa1, 0xa9, 0x32, 0xa9, 0x88, 0x17, 0x79, 0xf1, 0x80, 0x77,
	0x80, 0x67, 0x30, 0x7e, 0x15, 0xf5, 0x3c, 0x2b, 0x32, 0x4b, 0x7c, 0x27, 0xbe, 0x18, 0x4f, 0x20,
	0x10, 0x85, 0xde, 0x95, 0x96, 0x0c, 0x9c, 0xe9, 0x09, 0x09, 0x8c, 0x8c, 0x68, 0x72, 0x2d, 0x52,
	0xf2, 0x3f, 0xf2, 0xe2, 0x09, 0xff, 0x44, 0x8c, 0xe1, 0x50, 0x95, 0xb2, 0x6a, 0x8c, 0x55, 0x69,
	0xd2, 0x3d, 0x1d, 0xba, 0x62, 0x7f, 0x8d, 0x17, 0x70, 0xb4, 0xd5, 0x85, 0x2e, 0x74, 0x65, 0xb6,
	0x99, 0xec, 0xdb, 0xc0, 0xb5, 0xbf, 0x05, 0x52, 0xc0, 0x1f, 0xcb, 0x6b, 0x91, 0x8b, 0x52, 0x2a,
	0x32, 0x72, 0xf9, 0x1f, 0x06, 0x29, 0xf8, 0xda, 0x90, 0x71, 0xe4, 0xc5, 0x07, 0x97, 0xe7, 0xd4,
	0x1d, 0x85, 0xee, 0x1f, 0x84, 0x2e, 0x8d, 0xd4, 0xa9, 0xe2, 0xbe, 0x36, 0xd3, 0x19, 0x04, 0x1d,
	0x21, 0x40, 0xb0, 0x58, 0xf2, 0x87, 0x64, 0x1e, 0xfe, 0x6b, 0xe7, 0xf5, 0xe3, 0x6d, 0xb2, 0x62,
	0xa1, 0xd7, 0xce, 0x77, 0x9c, 0xb1, 0x27, 0x16, 0xfa, 0x38, 0x81, 0xf1, 0x7a, 0xd1, 0xd3, 0xa0,
	0xa5, 0x84, 0xdf, 0xcc, 0xee, 0xd9, 0x86, 0x85, 0xe9, 0x4b, 0xe0, 0x3e, 0xe2, 0xea, 0x23, 0x00,
	0x00, 0xff, 0xff, 0xcd, 0xf3, 0x55, 0xdc, 0xa3, 0x01, 0x00, 0x00,
}
