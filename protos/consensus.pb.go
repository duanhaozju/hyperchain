// Code generated by protoc-gen-go.
// source: consensus.proto
// DO NOT EDIT!

/*
Package protos is a generated protocol buffer package.

It is generated from these files:
	consensus.proto

It has these top-level messages:
	Message
	ExeMessage
*/
package protos

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

type Message_Type int32

const (
	Message_TRANSACTION Message_Type = 0
	Message_CONSENSUS   Message_Type = 1
)

var Message_Type_name = map[int32]string{
	0: "TRANSACTION",
	1: "CONSENSUS",
}
var Message_Type_value = map[string]int32{
	"TRANSACTION": 0,
	"CONSENSUS":   1,
}

func (x Message_Type) String() string {
	return proto.EnumName(Message_Type_name, int32(x))
}
func (Message_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

// This message type is used between consensus and outside
type Message struct {
	Type      Message_Type `protobuf:"varint,1,opt,name=type,enum=protos.Message_Type" json:"type,omitempty"`
	Timestamp int64        `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
	Payload   []byte       `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
	Id        uint64       `protobuf:"varint,4,opt,name=id" json:"id,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type ExeMessage struct {
	Batch     []*Message `protobuf:"bytes,1,rep,name=batch" json:"batch,omitempty"`
	Timestamp int64      `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
	No        uint64     `protobuf:"varint,3,opt,name=no" json:"no,omitempty"`
}

func (m *ExeMessage) Reset()                    { *m = ExeMessage{} }
func (m *ExeMessage) String() string            { return proto.CompactTextString(m) }
func (*ExeMessage) ProtoMessage()               {}
func (*ExeMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *ExeMessage) GetBatch() []*Message {
	if m != nil {
		return m.Batch
	}
	return nil
}

func init() {
	proto.RegisterType((*Message)(nil), "protos.Message")
	proto.RegisterType((*ExeMessage)(nil), "protos.ExeMessage")
	proto.RegisterEnum("protos.Message_Type", Message_Type_name, Message_Type_value)
}

func init() { proto.RegisterFile("consensus.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 225 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4f, 0xce, 0xcf, 0x2b,
	0x4e, 0xcd, 0x2b, 0x2e, 0x2d, 0xd6, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x03, 0x53, 0xc5,
	0x4a, 0x8b, 0x19, 0xb9, 0xd8, 0x7d, 0x53, 0x8b, 0x8b, 0x13, 0xd3, 0x53, 0x85, 0x34, 0xb8, 0x58,
	0x4a, 0x2a, 0x0b, 0x52, 0x25, 0x18, 0x15, 0x18, 0x35, 0xf8, 0x8c, 0x44, 0x20, 0x2a, 0x8b, 0xf5,
	0xa0, 0xd2, 0x7a, 0x21, 0x40, 0xb9, 0x20, 0xb0, 0x0a, 0x21, 0x19, 0x2e, 0xce, 0x92, 0xcc, 0xdc,
	0xd4, 0xe2, 0x92, 0xc4, 0xdc, 0x02, 0x09, 0x26, 0xa0, 0x72, 0xe6, 0x20, 0x84, 0x80, 0x90, 0x04,
	0x17, 0x7b, 0x41, 0x62, 0x65, 0x4e, 0x7e, 0x62, 0x8a, 0x04, 0x33, 0x50, 0x8e, 0x27, 0x08, 0xc6,
	0x15, 0xe2, 0xe3, 0x62, 0xca, 0x4c, 0x91, 0x60, 0x01, 0x0a, 0xb2, 0x04, 0x01, 0x59, 0x4a, 0x6a,
	0x5c, 0x2c, 0x20, 0x53, 0x85, 0xf8, 0xb9, 0xb8, 0x43, 0x82, 0x1c, 0xfd, 0x82, 0x1d, 0x9d, 0x43,
	0x3c, 0xfd, 0xfd, 0x04, 0x18, 0x84, 0x78, 0xb9, 0x38, 0x9d, 0xfd, 0xfd, 0x82, 0x5d, 0xfd, 0x82,
	0x43, 0x83, 0x05, 0x18, 0x95, 0x12, 0xb9, 0xb8, 0x5c, 0x2b, 0x52, 0x61, 0xee, 0x54, 0xe5, 0x62,
	0x4d, 0x4a, 0x2c, 0x49, 0xce, 0x00, 0x3a, 0x94, 0x59, 0x83, 0xdb, 0x88, 0x1f, 0xcd, 0xa1, 0x41,
	0x10, 0x59, 0x02, 0x8e, 0x04, 0x3a, 0x25, 0x2f, 0x1f, 0xec, 0x3e, 0xa0, 0x53, 0xf2, 0xf2, 0x93,
	0x20, 0x01, 0x62, 0x0c, 0x08, 0x00, 0x00, 0xff, 0xff, 0xa6, 0xcb, 0x73, 0x2d, 0x2a, 0x01, 0x00,
	0x00,
}
