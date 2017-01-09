// Code generated by protoc-gen-go.
// source: consensus.proto
// DO NOT EDIT!

/*
Package protos is a generated protocol buffer package.

It is generated from these files:
	consensus.proto

It has these top-level messages:
	Message
	UpdateStateMessage
	StateUpdatedMessage
	RemoveCache
	VcResetDone
	BlockchainInfo
	NullRequestMessage
	ValidatedTxs
	NewNodeMessage
	AddNodeMessage
	DelNodeMessage
*/
package protos

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import types "hyperchain/core/types"

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
	Message_TRANSACTION    Message_Type = 0
	Message_CONSENSUS      Message_Type = 1
	Message_STATE_UPDATED  Message_Type = 2
	Message_NULL_REQUEST   Message_Type = 3
	Message_NEGOTIATE_VIEW Message_Type = 4
)

var Message_Type_name = map[int32]string{
	0: "TRANSACTION",
	1: "CONSENSUS",
	2: "STATE_UPDATED",
	3: "NULL_REQUEST",
	4: "NEGOTIATE_VIEW",
}
var Message_Type_value = map[string]int32{
	"TRANSACTION":    0,
	"CONSENSUS":      1,
	"STATE_UPDATED":  2,
	"NULL_REQUEST":   3,
	"NEGOTIATE_VIEW": 4,
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

type UpdateStateMessage struct {
	Id       uint64   `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	SeqNo    uint64   `protobuf:"varint,2,opt,name=seqNo" json:"seqNo,omitempty"`
	TargetId []byte   `protobuf:"bytes,3,opt,name=targetId,proto3" json:"targetId,omitempty"`
	Replicas []uint64 `protobuf:"varint,4,rep,packed,name=replicas" json:"replicas,omitempty"`
}

func (m *UpdateStateMessage) Reset()                    { *m = UpdateStateMessage{} }
func (m *UpdateStateMessage) String() string            { return proto.CompactTextString(m) }
func (*UpdateStateMessage) ProtoMessage()               {}
func (*UpdateStateMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type StateUpdatedMessage struct {
	SeqNo uint64 `protobuf:"varint,1,opt,name=seqNo" json:"seqNo,omitempty"`
}

func (m *StateUpdatedMessage) Reset()                    { *m = StateUpdatedMessage{} }
func (m *StateUpdatedMessage) String() string            { return proto.CompactTextString(m) }
func (*StateUpdatedMessage) ProtoMessage()               {}
func (*StateUpdatedMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type RemoveCache struct {
	Vid uint64 `protobuf:"varint,1,opt,name=vid" json:"vid,omitempty"`
}

func (m *RemoveCache) Reset()                    { *m = RemoveCache{} }
func (m *RemoveCache) String() string            { return proto.CompactTextString(m) }
func (*RemoveCache) ProtoMessage()               {}
func (*RemoveCache) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type VcResetDone struct {
	SeqNo uint64 `protobuf:"varint,1,opt,name=seqNo" json:"seqNo,omitempty"`
}

func (m *VcResetDone) Reset()                    { *m = VcResetDone{} }
func (m *VcResetDone) String() string            { return proto.CompactTextString(m) }
func (*VcResetDone) ProtoMessage()               {}
func (*VcResetDone) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

// BlockchainInfo
type BlockchainInfo struct {
	Height            uint64 `protobuf:"varint,1,opt,name=height" json:"height,omitempty"`
	CurrentBlockHash  []byte `protobuf:"bytes,2,opt,name=currentBlockHash,proto3" json:"currentBlockHash,omitempty"`
	PreviousBlockHash []byte `protobuf:"bytes,3,opt,name=previousBlockHash,proto3" json:"previousBlockHash,omitempty"`
}

func (m *BlockchainInfo) Reset()                    { *m = BlockchainInfo{} }
func (m *BlockchainInfo) String() string            { return proto.CompactTextString(m) }
func (*BlockchainInfo) ProtoMessage()               {}
func (*BlockchainInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type NullRequestMessage struct {
	SeqNo uint64 `protobuf:"varint,1,opt,name=seqNo" json:"seqNo,omitempty"`
}

func (m *NullRequestMessage) Reset()                    { *m = NullRequestMessage{} }
func (m *NullRequestMessage) String() string            { return proto.CompactTextString(m) }
func (*NullRequestMessage) ProtoMessage()               {}
func (*NullRequestMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type ValidatedTxs struct {
	Transactions []*types.Transaction `protobuf:"bytes,1,rep,name=transactions" json:"transactions,omitempty"`
	Hash         string               `protobuf:"bytes,2,opt,name=hash" json:"hash,omitempty"`
	SeqNo        uint64               `protobuf:"varint,3,opt,name=seqNo" json:"seqNo,omitempty"`
	View         uint64               `protobuf:"varint,4,opt,name=view" json:"view,omitempty"`
	Timestamp    int64                `protobuf:"varint,5,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *ValidatedTxs) Reset()                    { *m = ValidatedTxs{} }
func (m *ValidatedTxs) String() string            { return proto.CompactTextString(m) }
func (*ValidatedTxs) ProtoMessage()               {}
func (*ValidatedTxs) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *ValidatedTxs) GetTransactions() []*types.Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

type NewNodeMessage struct {
	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *NewNodeMessage) Reset()                    { *m = NewNodeMessage{} }
func (m *NewNodeMessage) String() string            { return proto.CompactTextString(m) }
func (*NewNodeMessage) ProtoMessage()               {}
func (*NewNodeMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type AddNodeMessage struct {
	Payload []byte `protobuf:"bytes,1,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *AddNodeMessage) Reset()                    { *m = AddNodeMessage{} }
func (m *AddNodeMessage) String() string            { return proto.CompactTextString(m) }
func (*AddNodeMessage) ProtoMessage()               {}
func (*AddNodeMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

type DelNodeMessage struct {
	DelPayload []byte `protobuf:"bytes,1,opt,name=delPayload,proto3" json:"delPayload,omitempty"`
	RouterHash string `protobuf:"bytes,2,opt,name=routerHash" json:"routerHash,omitempty"`
	Id         uint64 `protobuf:"varint,3,opt,name=id" json:"id,omitempty"`
}

func (m *DelNodeMessage) Reset()                    { *m = DelNodeMessage{} }
func (m *DelNodeMessage) String() string            { return proto.CompactTextString(m) }
func (*DelNodeMessage) ProtoMessage()               {}
func (*DelNodeMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func init() {
	proto.RegisterType((*Message)(nil), "protos.Message")
	proto.RegisterType((*UpdateStateMessage)(nil), "protos.UpdateStateMessage")
	proto.RegisterType((*StateUpdatedMessage)(nil), "protos.StateUpdatedMessage")
	proto.RegisterType((*RemoveCache)(nil), "protos.RemoveCache")
	proto.RegisterType((*VcResetDone)(nil), "protos.VcResetDone")
	proto.RegisterType((*BlockchainInfo)(nil), "protos.BlockchainInfo")
	proto.RegisterType((*NullRequestMessage)(nil), "protos.NullRequestMessage")
	proto.RegisterType((*ValidatedTxs)(nil), "protos.ValidatedTxs")
	proto.RegisterType((*NewNodeMessage)(nil), "protos.NewNodeMessage")
	proto.RegisterType((*AddNodeMessage)(nil), "protos.AddNodeMessage")
	proto.RegisterType((*DelNodeMessage)(nil), "protos.DelNodeMessage")
	proto.RegisterEnum("protos.Message_Type", Message_Type_name, Message_Type_value)
}

func init() { proto.RegisterFile("consensus.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 564 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x8c, 0x94, 0xdf, 0x6e, 0xd3, 0x30,
	0x14, 0xc6, 0xc9, 0x92, 0x6d, 0xec, 0xb4, 0xeb, 0x32, 0x33, 0xa1, 0x6a, 0x42, 0x80, 0xc2, 0x05,
	0xd3, 0x40, 0x9d, 0x34, 0x24, 0xee, 0xcb, 0x1a, 0xb1, 0x4a, 0x23, 0x1b, 0x4e, 0x3a, 0x2e, 0x87,
	0x49, 0x0e, 0x4b, 0x44, 0x1a, 0x67, 0xb6, 0xd3, 0xb1, 0x5b, 0x1e, 0x85, 0xe7, 0xe2, 0x61, 0x70,
	0x92, 0xa6, 0x4d, 0x99, 0x40, 0xdc, 0x34, 0xe7, 0xcf, 0xcf, 0x3e, 0xfe, 0xec, 0x4f, 0x85, 0x9d,
	0x90, 0x67, 0x12, 0x33, 0x59, 0xc8, 0x41, 0x2e, 0xb8, 0xe2, 0x64, 0xa3, 0xfa, 0xc8, 0xfd, 0x97,
	0xf1, 0x5d, 0x8e, 0x22, 0x8c, 0x59, 0x92, 0x1d, 0x85, 0x5c, 0xe0, 0x91, 0xd2, 0xb9, 0x3c, 0x52,
	0x82, 0x65, 0x92, 0x85, 0x2a, 0xe1, 0x59, 0xbd, 0xc0, 0xf9, 0x65, 0xc0, 0xe6, 0x07, 0x94, 0x92,
	0x5d, 0x23, 0x39, 0x00, 0xab, 0xc4, 0xfa, 0xc6, 0x73, 0xe3, 0xa0, 0x77, 0xbc, 0x57, 0x13, 0x72,
	0x30, 0x6f, 0x0f, 0x02, 0xdd, 0xa3, 0x15, 0x41, 0x9e, 0xc0, 0x96, 0x4a, 0xa6, 0x28, 0x15, 0x9b,
	0xe6, 0xfd, 0x35, 0x8d, 0x9b, 0x74, 0x59, 0x20, 0x7d, 0xd8, 0xcc, 0xd9, 0x5d, 0xca, 0x59, 0xd4,
	0x37, 0x75, 0xaf, 0x4b, 0x9b, 0x94, 0xf4, 0x60, 0x2d, 0x89, 0xfa, 0x96, 0x2e, 0x5a, 0x54, 0x47,
	0xce, 0x15, 0x58, 0xe5, 0xae, 0x64, 0x07, 0x3a, 0x01, 0x1d, 0x7a, 0xfe, 0xf0, 0x24, 0x18, 0x9f,
	0x7b, 0xf6, 0x03, 0xb2, 0x0d, 0x5b, 0x27, 0xe7, 0x9e, 0xef, 0x7a, 0xfe, 0xc4, 0xb7, 0x0d, 0xb2,
	0x0b, 0xdb, 0x7e, 0x30, 0x0c, 0xdc, 0xab, 0xc9, 0xc5, 0x48, 0x7f, 0x46, 0xf6, 0x1a, 0xb1, 0xa1,
	0xeb, 0x4d, 0xce, 0xce, 0xae, 0xa8, 0xfb, 0x71, 0xe2, 0xfa, 0x81, 0x6d, 0x12, 0x02, 0x3d, 0xcf,
	0x7d, 0x7f, 0x1e, 0x8c, 0x4b, 0xf0, 0x72, 0xec, 0x7e, 0xb2, 0x2d, 0x47, 0x00, 0x99, 0xe4, 0x11,
	0x53, 0xe8, 0x2b, 0xfd, 0xd3, 0x08, 0xad, 0x8f, 0x61, 0x34, 0xc7, 0x20, 0x7b, 0xb0, 0x2e, 0xf1,
	0xc6, 0xe3, 0x95, 0x14, 0x8b, 0xd6, 0x09, 0xd9, 0x87, 0x87, 0x8a, 0x89, 0x6b, 0x54, 0xe3, 0x46,
	0xc7, 0x22, 0x2f, 0x7b, 0x02, 0xf3, 0x34, 0x09, 0x99, 0xd4, 0x72, 0x4c, 0xbd, 0x68, 0x91, 0x3b,
	0xaf, 0xe0, 0x51, 0x35, 0xad, 0x1e, 0x1c, 0x35, 0x43, 0x17, 0x43, 0x8c, 0xd6, 0x10, 0xe7, 0x19,
	0x74, 0x28, 0x4e, 0xf9, 0x0c, 0x4f, 0x58, 0x18, 0xa3, 0x56, 0x65, 0xce, 0x16, 0x47, 0x2b, 0x43,
	0xe7, 0x05, 0x74, 0x2e, 0x43, 0x8a, 0x12, 0xd5, 0x88, 0x67, 0x7f, 0xdb, 0xe5, 0x87, 0x01, 0xbd,
	0x77, 0x29, 0x0f, 0xbf, 0x55, 0x2f, 0x3e, 0xce, 0xbe, 0x72, 0xf2, 0x18, 0x36, 0x62, 0x4c, 0xae,
	0x63, 0x35, 0x27, 0xe7, 0x19, 0x39, 0x04, 0x3b, 0x2c, 0x84, 0xc0, 0x4c, 0x55, 0x0b, 0x4e, 0x99,
	0x8c, 0x2b, 0xd9, 0x5d, 0x7a, 0xaf, 0x4e, 0x5e, 0xc3, 0x6e, 0x2e, 0x70, 0x96, 0xf0, 0x42, 0x2e,
	0xe1, 0xfa, 0x2a, 0xee, 0x37, 0x9c, 0x43, 0x20, 0x5e, 0x91, 0xa6, 0x14, 0x6f, 0x0a, 0x6d, 0x84,
	0x7f, 0xcb, 0xfe, 0x69, 0x40, 0xf7, 0x92, 0xa5, 0x49, 0x75, 0x43, 0xc1, 0x77, 0x49, 0xde, 0x42,
	0xb7, 0x65, 0x4e, 0xa9, 0x69, 0xf3, 0xa0, 0x73, 0x4c, 0x06, 0x95, 0x6f, 0x07, 0xc1, 0xb2, 0x45,
	0x57, 0x38, 0xfd, 0xe8, 0x56, 0xdc, 0x48, 0xd8, 0xa2, 0x55, 0xbc, 0x1c, 0x69, 0xb6, 0x9f, 0x53,
	0x93, 0xb3, 0x04, 0x6f, 0xe7, 0xee, 0xab, 0xe2, 0x55, 0x1f, 0xaf, 0xff, 0xe1, 0x63, 0x2d, 0xa8,
	0xe7, 0xe1, 0xad, 0xc7, 0xa3, 0x85, 0x71, 0x5a, 0xce, 0x36, 0x56, 0x9c, 0x5d, 0xb2, 0xc3, 0x28,
	0xfa, 0x3f, 0xf6, 0x33, 0xf4, 0x46, 0x98, 0xb6, 0xd9, 0xa7, 0x00, 0x11, 0xa6, 0x17, 0x2b, 0x78,
	0xab, 0x52, 0xf6, 0x05, 0x2f, 0x14, 0x8a, 0xd3, 0xa5, 0xd6, 0x56, 0x65, 0x6e, 0x68, 0xb3, 0x31,
	0xf4, 0x97, 0xfa, 0x6f, 0xe0, 0xcd, 0xef, 0x00, 0x00, 0x00, 0xff, 0xff, 0x94, 0x22, 0xef, 0xb5,
	0x20, 0x04, 0x00, 0x00,
}
