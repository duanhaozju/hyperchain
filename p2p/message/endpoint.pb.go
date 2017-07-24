// Code generated by protoc-gen-go. DO NOT EDIT.
// source: endpoint.proto

/*
Package message is a generated protocol buffer package.

It is generated from these files:
	endpoint.proto
	msg.proto
	msg_type.proto

It has these top-level messages:
	Endpoint
	Extend
	Message
*/
package message

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

type Endpoint struct {
	Version  int64   `protobuf:"varint,1,opt,name=Version" json:"Version,omitempty"`
	Hostname []byte  `protobuf:"bytes,2,opt,name=Hostname,proto3" json:"Hostname,omitempty"`
	Field    []byte  `protobuf:"bytes,4,opt,name=Field,proto3" json:"Field,omitempty"`
	UUID     []byte  `protobuf:"bytes,6,opt,name=UUID,proto3" json:"UUID,omitempty"`
	Extend   *Extend `protobuf:"bytes,7,opt,name=Extend" json:"Extend,omitempty"`
}

func (m *Endpoint) Reset()                    { *m = Endpoint{} }
func (m *Endpoint) String() string            { return proto.CompactTextString(m) }
func (*Endpoint) ProtoMessage()               {}
func (*Endpoint) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Endpoint) GetVersion() int64 {
	if m != nil {
		return m.Version
	}
	return 0
}

func (m *Endpoint) GetHostname() []byte {
	if m != nil {
		return m.Hostname
	}
	return nil
}

func (m *Endpoint) GetField() []byte {
	if m != nil {
		return m.Field
	}
	return nil
}

func (m *Endpoint) GetUUID() []byte {
	if m != nil {
		return m.UUID
	}
	return nil
}

func (m *Endpoint) GetExtend() *Extend {
	if m != nil {
		return m.Extend
	}
	return nil
}

type Extend struct {
	IP            []byte `protobuf:"bytes,1,opt,name=IP,proto3" json:"IP,omitempty"`
	Port          int64  `protobuf:"varint,2,opt,name=Port" json:"Port,omitempty"`
	LastTimestamp int64  `protobuf:"varint,3,opt,name=LastTimestamp" json:"LastTimestamp,omitempty"`
	Status        []byte `protobuf:"bytes,4,opt,name=Status,proto3" json:"Status,omitempty"`
	Extends       []byte `protobuf:"bytes,5,opt,name=Extends,proto3" json:"Extends,omitempty"`
}

func (m *Extend) Reset()                    { *m = Extend{} }
func (m *Extend) String() string            { return proto.CompactTextString(m) }
func (*Extend) ProtoMessage()               {}
func (*Extend) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Extend) GetIP() []byte {
	if m != nil {
		return m.IP
	}
	return nil
}

func (m *Extend) GetPort() int64 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *Extend) GetLastTimestamp() int64 {
	if m != nil {
		return m.LastTimestamp
	}
	return 0
}

func (m *Extend) GetStatus() []byte {
	if m != nil {
		return m.Status
	}
	return nil
}

func (m *Extend) GetExtends() []byte {
	if m != nil {
		return m.Extends
	}
	return nil
}

func init() {
	proto.RegisterType((*Endpoint)(nil), "message.Endpoint")
	proto.RegisterType((*Extend)(nil), "message.Extend")
}

func init() { proto.RegisterFile("endpoint.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 232 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x54, 0x90, 0xc1, 0x4a, 0xc4, 0x30,
	0x10, 0x86, 0x49, 0xbb, 0xdb, 0x2e, 0xe3, 0xba, 0xc2, 0x20, 0x12, 0x3c, 0x95, 0x45, 0xb0, 0xa7,
	0x1e, 0xf4, 0x15, 0x5c, 0xb1, 0xe0, 0xa1, 0x44, 0xd7, 0x7b, 0xa4, 0x41, 0x0a, 0x36, 0x29, 0x9d,
	0x11, 0x7c, 0x00, 0xdf, 0xc0, 0x17, 0x96, 0x4e, 0x53, 0x61, 0x6f, 0xf3, 0xfd, 0x5f, 0x48, 0xfe,
	0x0c, 0xec, 0x9c, 0x6f, 0x87, 0xd0, 0x79, 0xae, 0x86, 0x31, 0x70, 0xc0, 0xbc, 0x77, 0x44, 0xf6,
	0xc3, 0xed, 0x7f, 0x15, 0x6c, 0x0e, 0xd1, 0xa1, 0x86, 0xfc, 0xcd, 0x8d, 0xd4, 0x05, 0xaf, 0x55,
	0xa1, 0xca, 0xd4, 0x2c, 0x88, 0xd7, 0xb0, 0x79, 0x0a, 0xc4, 0xde, 0xf6, 0x4e, 0x27, 0x85, 0x2a,
	0xb7, 0xe6, 0x9f, 0xf1, 0x12, 0xd6, 0x8f, 0x9d, 0xfb, 0x6c, 0xf5, 0x4a, 0xc4, 0x0c, 0x88, 0xb0,
	0x3a, 0x1e, 0xeb, 0x07, 0x9d, 0x49, 0x28, 0x33, 0xde, 0x42, 0x76, 0xf8, 0x66, 0xe7, 0x5b, 0x9d,
	0x17, 0xaa, 0x3c, 0xbb, 0xbb, 0xa8, 0x62, 0x8d, 0x6a, 0x8e, 0x4d, 0xd4, 0xfb, 0x1f, 0xb5, 0x9c,
	0xc4, 0x1d, 0x24, 0x75, 0x23, 0x75, 0xb6, 0x26, 0xa9, 0x9b, 0xe9, 0xde, 0x26, 0x8c, 0x2c, 0x2d,
	0x52, 0x23, 0x33, 0xde, 0xc0, 0xf9, 0xb3, 0x25, 0x7e, 0xed, 0x7a, 0x47, 0x6c, 0xfb, 0x41, 0xa7,
	0x22, 0x4f, 0x43, 0xbc, 0x82, 0xec, 0x85, 0x2d, 0x7f, 0x51, 0x2c, 0x1a, 0x69, 0xfa, 0xf5, 0xfc,
	0x16, 0xe9, 0xb5, 0x88, 0x05, 0xdf, 0x33, 0x59, 0xd6, 0xfd, 0x5f, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x17, 0x22, 0x21, 0x0b, 0x3e, 0x01, 0x00, 0x00,
}
