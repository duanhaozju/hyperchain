// Code generated by protoc-gen-go.
// source: message.proto
// DO NOT EDIT!

/*
Package recovery is a generated protocol buffer package.

It is generated from these files:
	message.proto

It has these top-level messages:
	Message
	CheckPointMessage
*/
package recovery

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

// MsgType message type during data transfer
type Message_MsgType int32

const (
	Message_SYNCCHECKPOINT Message_MsgType = 0
	Message_SYNCBLOCK      Message_MsgType = 1
	Message_RELAYTX        Message_MsgType = 2
	Message_SYNCSINGLE     Message_MsgType = 3
)

var Message_MsgType_name = map[int32]string{
	0: "SYNCCHECKPOINT",
	1: "SYNCBLOCK",
	2: "RELAYTX",
	3: "SYNCSINGLE",
}
var Message_MsgType_value = map[string]int32{
	"SYNCCHECKPOINT": 0,
	"SYNCBLOCK":      1,
	"RELAYTX":        2,
	"SYNCSINGLE":     3,
}

func (x Message_MsgType) String() string {
	return proto.EnumName(Message_MsgType_name, int32(x))
}
func (Message_MsgType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Message struct {
	// From the message source
	MessageType  Message_MsgType `protobuf:"varint,1,opt,name=MessageType,json=messageType,enum=recovery.Message_MsgType" json:"MessageType,omitempty"`
	Payload      []byte          `protobuf:"bytes,2,opt,name=Payload,json=payload,proto3" json:"Payload,omitempty"`
	MsgTimeStamp int64           `protobuf:"varint,15,opt,name=MsgTimeStamp,json=msgTimeStamp" json:"MsgTimeStamp,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type CheckPointMessage struct {
	// From the message source
	CurrentNumber  uint64 `protobuf:"varint,1,opt,name=currentNumber" json:"currentNumber,omitempty"`
	RequiredNumber uint64 `protobuf:"varint,2,opt,name=requiredNumber" json:"requiredNumber,omitempty"`
	PeerId         uint64 `protobuf:"varint,3,opt,name=peerId" json:"peerId,omitempty"`
}

func (m *CheckPointMessage) Reset()                    { *m = CheckPointMessage{} }
func (m *CheckPointMessage) String() string            { return proto.CompactTextString(m) }
func (*CheckPointMessage) ProtoMessage()               {}
func (*CheckPointMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func init() {
	proto.RegisterType((*Message)(nil), "recovery.Message")
	proto.RegisterType((*CheckPointMessage)(nil), "recovery.CheckPointMessage")
	proto.RegisterEnum("recovery.Message_MsgType", Message_MsgType_name, Message_MsgType_value)
}

func init() { proto.RegisterFile("message.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 274 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x90, 0x3b, 0x4f, 0xb4, 0x40,
	0x14, 0x86, 0x3f, 0xe0, 0xcb, 0xa2, 0x87, 0x8b, 0x78, 0x0a, 0x83, 0x9d, 0x21, 0xc6, 0x58, 0x51,
	0x68, 0x69, 0xa5, 0x84, 0x28, 0x59, 0x96, 0x25, 0x40, 0xe1, 0x96, 0x2c, 0x9c, 0xac, 0x44, 0x67,
	0xc1, 0x01, 0x4c, 0xf8, 0x95, 0xfe, 0x25, 0xb9, 0x6d, 0xbc, 0x74, 0x73, 0xde, 0xe7, 0x39, 0x93,
	0x79, 0x07, 0x34, 0x46, 0x75, 0x9d, 0xee, 0xc8, 0xae, 0x78, 0xd9, 0x94, 0x78, 0xc4, 0x29, 0x2b,
	0x3f, 0x88, 0x77, 0xd6, 0xa7, 0x00, 0xf2, 0x6a, 0x62, 0x78, 0x07, 0xca, 0x7c, 0x4c, 0xba, 0x8a,
	0x4c, 0xe1, 0x42, 0xb8, 0xd6, 0x6f, 0xce, 0xed, 0x83, 0x6b, 0xcf, 0xd0, 0x5e, 0xd5, 0xbb, 0x41,
	0x88, 0x14, 0xf6, 0x6d, 0xa3, 0x09, 0x72, 0x98, 0x76, 0x6f, 0x65, 0x9a, 0x9b, 0x62, 0xbf, 0xa8,
	0x46, 0x72, 0x35, 0x8d, 0x68, 0x81, 0x3a, 0x6c, 0x14, 0x8c, 0xe2, 0x26, 0x65, 0x95, 0x79, 0xd2,
	0x63, 0x29, 0x52, 0xd9, 0x8f, 0xcc, 0xf2, 0xfa, 0x57, 0x4c, 0xb7, 0x22, 0x82, 0x1e, 0x6f, 0x02,
	0xc7, 0x79, 0x72, 0x9d, 0x65, 0xb8, 0xf6, 0x82, 0xc4, 0xf8, 0x87, 0x1a, 0x1c, 0x0f, 0xd9, 0x83,
	0xbf, 0x76, 0x96, 0x86, 0x80, 0x0a, 0xc8, 0x91, 0xeb, 0xdf, 0x6f, 0x92, 0x67, 0x43, 0x44, 0x1d,
	0x60, 0x60, 0xb1, 0x17, 0x3c, 0xfa, 0xae, 0x21, 0x59, 0x1d, 0x9c, 0x3a, 0x2f, 0x94, 0xbd, 0x86,
	0x65, 0xb1, 0x6f, 0x0e, 0xd5, 0x2e, 0x41, 0xcb, 0x5a, 0xce, 0x69, 0xdf, 0x04, 0x2d, 0xdb, 0x12,
	0x1f, 0xcb, 0xfd, 0x8f, 0x7e, 0x87, 0x78, 0x05, 0x3a, 0xa7, 0xf7, 0xb6, 0xe0, 0x94, 0xcf, 0x9a,
	0x38, 0x6a, 0x7f, 0x52, 0x3c, 0x83, 0x45, 0x45, 0xc4, 0xbd, 0xdc, 0x94, 0x46, 0x3e, 0x4f, 0xdb,
	0xc5, 0xf8, 0xbb, 0xb7, 0x5f, 0x01, 0x00, 0x00, 0xff, 0xff, 0xfd, 0x41, 0x15, 0xc0, 0x6e, 0x01,
	0x00, 0x00,
}
