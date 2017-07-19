// Code generated by protoc-gen-go.
// source: msg.proto
// DO NOT EDIT!

/*
Package executor is a generated protocol buffer package.

It is generated from these files:
	msg.proto

It has these top-level messages:
	ChainSyncRequest
*/
package executor

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

type ChainSyncRequest struct {
	CurrentNumber  uint64 `protobuf:"varint,1,opt,name=currentNumber" json:"currentNumber,omitempty"`
	RequiredNumber uint64 `protobuf:"varint,2,opt,name=requiredNumber" json:"requiredNumber,omitempty"`
	PeerId         uint64 `protobuf:"varint,3,opt,name=peerId" json:"peerId,omitempty"`
	PeerHash       string `protobuf:"bytes,4,opt,name=peerHash" json:"peerHash,omitempty"`
}

func (m *ChainSyncRequest) Reset()                    { *m = ChainSyncRequest{} }
func (m *ChainSyncRequest) String() string            { return proto.CompactTextString(m) }
func (*ChainSyncRequest) ProtoMessage()               {}
func (*ChainSyncRequest) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *ChainSyncRequest) GetCurrentNumber() uint64 {
	if m != nil {
		return m.CurrentNumber
	}
	return 0
}

func (m *ChainSyncRequest) GetRequiredNumber() uint64 {
	if m != nil {
		return m.RequiredNumber
	}
	return 0
}

func (m *ChainSyncRequest) GetPeerId() uint64 {
	if m != nil {
		return m.PeerId
	}
	return 0
}

func (m *ChainSyncRequest) GetPeerHash() string {
	if m != nil {
		return m.PeerHash
	}
	return ""
}

func init() {
	proto.RegisterType((*ChainSyncRequest)(nil), "executor.ChainSyncRequest")
}

func init() { proto.RegisterFile("msg.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 155 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0xcd, 0x4d, 0x2d, 0x2e,
	0x4e, 0x4c, 0x4f, 0xd5, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0xe2, 0x48, 0xad, 0x48, 0x4d, 0x2e,
	0x2d, 0xc9, 0x2f, 0x52, 0x9a, 0xc2, 0xc8, 0x25, 0xe0, 0x9c, 0x91, 0x98, 0x99, 0x17, 0x5c, 0x99,
	0x97, 0x1c, 0x94, 0x5a, 0x58, 0x9a, 0x5a, 0x5c, 0x22, 0xa4, 0xc2, 0xc5, 0x9b, 0x5c, 0x5a, 0x54,
	0x94, 0x9a, 0x57, 0xe2, 0x57, 0x9a, 0x9b, 0x94, 0x5a, 0x24, 0xc1, 0xa8, 0xc0, 0xa8, 0xc1, 0x12,
	0x84, 0x2a, 0x28, 0xa4, 0xc6, 0xc5, 0x57, 0x94, 0x5a, 0x58, 0x9a, 0x59, 0x94, 0x9a, 0x02, 0x55,
	0xc6, 0x04, 0x56, 0x86, 0x26, 0x2a, 0x24, 0xc6, 0xc5, 0x56, 0x90, 0x9a, 0x5a, 0xe4, 0x99, 0x22,
	0xc1, 0x0c, 0x96, 0x87, 0xf2, 0x84, 0xa4, 0xb8, 0x38, 0x40, 0x2c, 0x8f, 0xc4, 0xe2, 0x0c, 0x09,
	0x16, 0x05, 0x46, 0x0d, 0xce, 0x20, 0x38, 0x3f, 0x89, 0x0d, 0xec, 0x4e, 0x63, 0x40, 0x00, 0x00,
	0x00, 0xff, 0xff, 0x81, 0x71, 0x42, 0x6e, 0xb8, 0x00, 0x00, 0x00,
}
