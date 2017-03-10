// Code generated by protoc-gen-go.
// source: transaction_wrapper.proto
// DO NOT EDIT!

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type TransactionWrapper struct {
	TransactionVersion []byte `protobuf:"bytes,1,opt,name=transactionVersion,proto3" json:"transactionVersion,omitempty"`
	Transaction        []byte `protobuf:"bytes,2,opt,name=transaction,proto3" json:"transaction,omitempty"`
}

func (m *TransactionWrapper) Reset()                    { *m = TransactionWrapper{} }
func (m *TransactionWrapper) String() string            { return proto.CompactTextString(m) }
func (*TransactionWrapper) ProtoMessage()               {}
func (*TransactionWrapper) Descriptor() ([]byte, []int) { return fileDescriptor7, []int{0} }

func (m *TransactionWrapper) GetTransactionVersion() []byte {
	if m != nil {
		return m.TransactionVersion
	}
	return nil
}

func (m *TransactionWrapper) GetTransaction() []byte {
	if m != nil {
		return m.Transaction
	}
	return nil
}

func init() {
	proto.RegisterType((*TransactionWrapper)(nil), "types.TransactionWrapper")
}

func init() { proto.RegisterFile("transaction_wrapper.proto", fileDescriptor7) }

var fileDescriptor7 = []byte{
	// 107 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x92, 0x2c, 0x29, 0x4a, 0xcc,
	0x2b, 0x4e, 0x4c, 0x2e, 0xc9, 0xcc, 0xcf, 0x8b, 0x2f, 0x2f, 0x4a, 0x2c, 0x28, 0x48, 0x2d, 0xd2,
	0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2d, 0xa9, 0x2c, 0x48, 0x2d, 0x56, 0x4a, 0xe3, 0x12,
	0x0a, 0x41, 0xa8, 0x09, 0x87, 0x28, 0x11, 0xd2, 0xe3, 0x12, 0x42, 0xd2, 0x19, 0x96, 0x5a, 0x54,
	0x9c, 0x99, 0x9f, 0x27, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x13, 0x84, 0x45, 0x46, 0x48, 0x81, 0x8b,
	0x1b, 0x49, 0x54, 0x82, 0x09, 0xac, 0x10, 0x59, 0x28, 0x89, 0x0d, 0x6c, 0xab, 0x31, 0x20, 0x00,
	0x00, 0xff, 0xff, 0xf6, 0x70, 0x09, 0x93, 0x92, 0x00, 0x00, 0x00,
}
