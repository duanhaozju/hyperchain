// Code generated by protoc-gen-go.
// source: message.proto
// DO NOT EDIT!

/*
Package protos is a generated protocol buffer package.

It is generated from these files:
	message.proto

It has these top-level messages:
	SessionMessage
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

type SessionMessage_Type int32

const (
	SessionMessage_CONSENSUS            SessionMessage_Type = 0
	SessionMessage_FOWARD_TX            SessionMessage_Type = 1
	SessionMessage_SYNC_REPLICA         SessionMessage_Type = 2
	SessionMessage_UNICAST_INVALID      SessionMessage_Type = 3
	SessionMessage_BROADCAST_SINGLE_BLK SessionMessage_Type = 4
	SessionMessage_UNICAST_BLK          SessionMessage_Type = 5
	SessionMessage_SYNC_REQ             SessionMessage_Type = 6
	SessionMessage_SYNC_WORLD_STATE     SessionMessage_Type = 7
	SessionMessage_SEND_WORLD_STATE     SessionMessage_Type = 8
	SessionMessage_ADD_PEER             SessionMessage_Type = 9
	SessionMessage_DEL_PEER             SessionMessage_Type = 10
)

var SessionMessage_Type_name = map[int32]string{
	0:  "CONSENSUS",
	1:  "FOWARD_TX",
	2:  "SYNC_REPLICA",
	3:  "UNICAST_INVALID",
	4:  "BROADCAST_SINGLE_BLK",
	5:  "UNICAST_BLK",
	6:  "SYNC_REQ",
	7:  "SYNC_WORLD_STATE",
	8:  "SEND_WORLD_STATE",
	9:  "ADD_PEER",
	10: "DEL_PEER",
}
var SessionMessage_Type_value = map[string]int32{
	"CONSENSUS":            0,
	"FOWARD_TX":            1,
	"SYNC_REPLICA":         2,
	"UNICAST_INVALID":      3,
	"BROADCAST_SINGLE_BLK": 4,
	"UNICAST_BLK":          5,
	"SYNC_REQ":             6,
	"SYNC_WORLD_STATE":     7,
	"SEND_WORLD_STATE":     8,
	"ADD_PEER":             9,
	"DEL_PEER":             10,
}

func (x SessionMessage_Type) String() string {
	return proto.EnumName(SessionMessage_Type_name, int32(x))
}
func (SessionMessage_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

// This message type is used between consensus and outside
type SessionMessage struct {
	Type    SessionMessage_Type `protobuf:"varint,1,opt,name=type,enum=protos.SessionMessage_Type" json:"type,omitempty"`
	Payload []byte              `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *SessionMessage) Reset()                    { *m = SessionMessage{} }
func (m *SessionMessage) String() string            { return proto.CompactTextString(m) }
func (*SessionMessage) ProtoMessage()               {}
func (*SessionMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *SessionMessage) GetType() SessionMessage_Type {
	if m != nil {
		return m.Type
	}
	return SessionMessage_CONSENSUS
}

func (m *SessionMessage) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func init() {
	proto.RegisterType((*SessionMessage)(nil), "protos.SessionMessage")
	proto.RegisterEnum("protos.SessionMessage_Type", SessionMessage_Type_name, SessionMessage_Type_value)
}

func init() { proto.RegisterFile("message.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 266 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x54, 0x90, 0xbf, 0x4e, 0xf3, 0x30,
	0x14, 0x47, 0xbf, 0xe4, 0x0b, 0xfd, 0x73, 0x49, 0xdb, 0xab, 0x4b, 0x87, 0x48, 0x2c, 0x55, 0xa7,
	0x4e, 0x41, 0x82, 0x27, 0x70, 0x63, 0x83, 0x22, 0x8c, 0x53, 0xec, 0x94, 0xc2, 0x64, 0x05, 0x11,
	0x21, 0x24, 0x20, 0x11, 0xee, 0x92, 0x57, 0xe2, 0x59, 0x78, 0x28, 0x94, 0x84, 0x0e, 0x9d, 0xac,
	0x73, 0x74, 0x7e, 0x1e, 0x2e, 0x4c, 0x3e, 0x4a, 0xe7, 0x8a, 0xd7, 0x32, 0xae, 0xbf, 0xaa, 0x7d,
	0x45, 0x83, 0xee, 0x71, 0xcb, 0x6f, 0x1f, 0xa6, 0xa6, 0x74, 0xee, 0xad, 0xfa, 0xbc, 0xeb, 0x03,
	0xba, 0x80, 0x60, 0xdf, 0xd4, 0x65, 0xe4, 0x2d, 0xbc, 0xd5, 0xf4, 0xf2, 0xbc, 0x1f, 0xb8, 0xf8,
	0xb8, 0x8a, 0xf3, 0xa6, 0x2e, 0x75, 0x17, 0x52, 0x04, 0xc3, 0xba, 0x68, 0xde, 0xab, 0xe2, 0x25,
	0xf2, 0x17, 0xde, 0x2a, 0xd4, 0x07, 0x5c, 0xfe, 0x78, 0x10, 0xb4, 0x21, 0x4d, 0x60, 0x9c, 0x64,
	0xca, 0x08, 0x65, 0xb6, 0x06, 0xff, 0xb5, 0x78, 0x9d, 0xed, 0x98, 0xe6, 0x36, 0x7f, 0x44, 0x8f,
	0x10, 0x42, 0xf3, 0xa4, 0x12, 0xab, 0xc5, 0x46, 0xa6, 0x09, 0x43, 0x9f, 0xce, 0x60, 0xb6, 0x55,
	0x69, 0xc2, 0x4c, 0x6e, 0x53, 0xf5, 0xc0, 0x64, 0xca, 0xf1, 0x3f, 0x45, 0x30, 0x5f, 0xeb, 0x8c,
	0xf1, 0x4e, 0x9b, 0x54, 0xdd, 0x48, 0x61, 0xd7, 0xf2, 0x16, 0x03, 0x9a, 0xc1, 0xe9, 0x21, 0x6f,
	0xc5, 0x09, 0x85, 0x30, 0xfa, 0xfb, 0xf1, 0x1e, 0x07, 0x34, 0x07, 0xec, 0x68, 0x97, 0x69, 0xc9,
	0xad, 0xc9, 0x59, 0x2e, 0x70, 0xd8, 0x59, 0xa1, 0xf8, 0x91, 0x1d, 0xb5, 0x4b, 0xc6, 0xb9, 0xdd,
	0x08, 0xa1, 0x71, 0xdc, 0x12, 0x17, 0xb2, 0x27, 0x78, 0xee, 0x8f, 0x76, 0xf5, 0x1b, 0x00, 0x00,
	0xff, 0xff, 0x88, 0xbc, 0x0e, 0x50, 0x4c, 0x01, 0x00, 0x00,
}
