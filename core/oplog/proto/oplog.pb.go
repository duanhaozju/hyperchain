// Code generated by protoc-gen-go. DO NOT EDIT.
// source: oplog.proto

/*
Package oplog is a generated protocol buffer package.

It is generated from these files:
	oplog.proto

It has these top-level messages:
	LogEntry
	CMap
*/
package oplog

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

type LogEntry_Type int32

const (
	LogEntry_TransactionList LogEntry_Type = 0
	LogEntry_RollBack        LogEntry_Type = 1
	LogEntry_StateUpdate     LogEntry_Type = 3
)

var LogEntry_Type_name = map[int32]string{
	0: "TransactionList",
	1: "RollBack",
	3: "StateUpdate",
}
var LogEntry_Type_value = map[string]int32{
	"TransactionList": 0,
	"RollBack":        1,
	"StateUpdate":     3,
}

func (x LogEntry_Type) String() string {
	return proto.EnumName(LogEntry_Type_name, int32(x))
}
func (LogEntry_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type LogEntry struct {
	Lid     uint64        `protobuf:"varint,1,opt,name=lid" json:"lid,omitempty"`
	Type    LogEntry_Type `protobuf:"varint,2,opt,name=type,enum=LogEntry_Type" json:"type,omitempty"`
	Payload []byte        `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *LogEntry) Reset()                    { *m = LogEntry{} }
func (m *LogEntry) String() string            { return proto.CompactTextString(m) }
func (*LogEntry) ProtoMessage()               {}
func (*LogEntry) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *LogEntry) GetLid() uint64 {
	if m != nil {
		return m.Lid
	}
	return 0
}

func (m *LogEntry) GetType() LogEntry_Type {
	if m != nil {
		return m.Type
	}
	return LogEntry_TransactionList
}

func (m *LogEntry) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

type CMap struct {
	Map map[uint64]uint64 `protobuf:"bytes,1,rep,name=Map" json:"Map,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"varint,2,opt,name=value"`
}

func (m *CMap) Reset()                    { *m = CMap{} }
func (m *CMap) String() string            { return proto.CompactTextString(m) }
func (*CMap) ProtoMessage()               {}
func (*CMap) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *CMap) GetMap() map[uint64]uint64 {
	if m != nil {
		return m.Map
	}
	return nil
}

func init() {
	proto.RegisterType((*LogEntry)(nil), "LogEntry")
	proto.RegisterType((*CMap)(nil), "cMap")
	proto.RegisterEnum("LogEntry_Type", LogEntry_Type_name, LogEntry_Type_value)
}

func init() { proto.RegisterFile("oplog.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 233 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x34, 0x90, 0xbf, 0x4e, 0xc3, 0x30,
	0x10, 0xc6, 0x71, 0x63, 0x20, 0xba, 0x54, 0x69, 0x64, 0x18, 0x2c, 0x26, 0x2b, 0x53, 0x26, 0x0f,
	0x45, 0x42, 0xa8, 0x23, 0x12, 0x5b, 0xb3, 0x98, 0xb2, 0x73, 0x24, 0x56, 0x15, 0xc5, 0x8a, 0x4f,
	0xa9, 0x41, 0xf2, 0x8b, 0xf0, 0xbc, 0xc8, 0xa9, 0xb2, 0xdd, 0xf7, 0x67, 0xb8, 0xdf, 0x07, 0x85,
	0x27, 0xe7, 0xcf, 0x9a, 0x66, 0x1f, 0x7c, 0xfd, 0xc7, 0x20, 0x3f, 0xfa, 0xf3, 0xfb, 0x14, 0xe6,
	0x28, 0x2a, 0xc8, 0xdc, 0xd0, 0x4b, 0xa6, 0x58, 0xc3, 0x4d, 0x3a, 0x45, 0x0d, 0x3c, 0x44, 0xb2,
	0x72, 0xa3, 0x58, 0x53, 0xee, 0x4b, 0xbd, 0x56, 0xf5, 0x29, 0x92, 0x35, 0x4b, 0x26, 0x24, 0xdc,
	0x13, 0x46, 0xe7, 0xb1, 0x97, 0x99, 0x62, 0xcd, 0xd6, 0xac, 0xb2, 0x3e, 0x00, 0x4f, 0x3d, 0xf1,
	0x00, 0xbb, 0xd3, 0x8c, 0xd3, 0x05, 0xbb, 0x30, 0xf8, 0xe9, 0x38, 0x5c, 0x42, 0x75, 0x23, 0xb6,
	0x90, 0x1b, 0xef, 0xdc, 0x1b, 0x76, 0x63, 0xc5, 0xc4, 0x0e, 0x8a, 0x8f, 0x80, 0xc1, 0x7e, 0x52,
	0x8f, 0xc1, 0x56, 0x59, 0xfd, 0x05, 0xbc, 0x6b, 0x91, 0x84, 0x82, 0xac, 0x45, 0x92, 0x4c, 0x65,
	0x4d, 0xb1, 0x2f, 0x75, 0xf2, 0x74, 0x8b, 0xb4, 0x7c, 0x61, 0x52, 0xf4, 0xf4, 0x02, 0xf9, 0x6a,
	0x24, 0x82, 0xd1, 0xc6, 0x95, 0x60, 0xb4, 0x51, 0x3c, 0xc2, 0xed, 0x2f, 0xba, 0x9f, 0x2b, 0x02,
	0x37, 0x57, 0x71, 0xd8, 0xbc, 0xb2, 0xef, 0xbb, 0x65, 0x81, 0xe7, 0xff, 0x00, 0x00, 0x00, 0xff,
	0xff, 0xf3, 0x96, 0xb1, 0x01, 0x10, 0x01, 0x00, 0x00,
}
