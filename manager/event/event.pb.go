// Code generated by protoc-gen-go. DO NOT EDIT.
// source: event.proto

/*
Package event is a generated protocol buffer package.

It is generated from these files:
	event.proto

It has these top-level messages:
	AddNamespaceEvent
	DeleteNamespaceEvent
	AdminResponseEvent
	TestEvent
	TransactionBlock
	RollbackEvent
	OpLogAck
	OpLogFetch
	CheckpointAck
*/
package event

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import types "github.com/hyperchain/hyperchain/core/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type AddNamespaceEvent struct {
	Namespace string `protobuf:"bytes,1,opt,name=namespace" json:"namespace,omitempty"`
}

func (m *AddNamespaceEvent) Reset()                    { *m = AddNamespaceEvent{} }
func (m *AddNamespaceEvent) String() string            { return proto.CompactTextString(m) }
func (*AddNamespaceEvent) ProtoMessage()               {}
func (*AddNamespaceEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *AddNamespaceEvent) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

type DeleteNamespaceEvent struct {
	Namespace string `protobuf:"bytes,1,opt,name=namespace" json:"namespace,omitempty"`
}

func (m *DeleteNamespaceEvent) Reset()                    { *m = DeleteNamespaceEvent{} }
func (m *DeleteNamespaceEvent) String() string            { return proto.CompactTextString(m) }
func (*DeleteNamespaceEvent) ProtoMessage()               {}
func (*DeleteNamespaceEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *DeleteNamespaceEvent) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

type AdminResponseEvent struct {
	Ok  bool   `protobuf:"varint,1,opt,name=Ok" json:"Ok,omitempty"`
	Msg string `protobuf:"bytes,2,opt,name=Msg" json:"Msg,omitempty"`
}

func (m *AdminResponseEvent) Reset()                    { *m = AdminResponseEvent{} }
func (m *AdminResponseEvent) String() string            { return proto.CompactTextString(m) }
func (*AdminResponseEvent) ProtoMessage()               {}
func (*AdminResponseEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *AdminResponseEvent) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

func (m *AdminResponseEvent) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type TestEvent struct {
	Id  uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Msg string `protobuf:"bytes,2,opt,name=msg" json:"msg,omitempty"`
}

func (m *TestEvent) Reset()                    { *m = TestEvent{} }
func (m *TestEvent) String() string            { return proto.CompactTextString(m) }
func (*TestEvent) ProtoMessage()               {}
func (*TestEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *TestEvent) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *TestEvent) GetMsg() string {
	if m != nil {
		return m.Msg
	}
	return ""
}

type TransactionBlock struct {
	PreviousHash        string                            `protobuf:"bytes,1,opt,name=PreviousHash" json:"PreviousHash,omitempty"`
	Digest              string                            `protobuf:"bytes,2,opt,name=Digest" json:"Digest,omitempty"`
	Transactions        []*types.Transaction              `protobuf:"bytes,3,rep,name=Transactions" json:"Transactions,omitempty"`
	InvalidTransactions []*types.InvalidTransactionRecord `protobuf:"bytes,4,rep,name=InvalidTransactions" json:"InvalidTransactions,omitempty"`
	SeqNo               uint64                            `protobuf:"varint,5,opt,name=SeqNo" json:"SeqNo,omitempty"`
	View                uint64                            `protobuf:"varint,6,opt,name=View" json:"View,omitempty"`
	IsPrimary           bool                              `protobuf:"varint,7,opt,name=IsPrimary" json:"IsPrimary,omitempty"`
	Timestamp           int64                             `protobuf:"varint,8,opt,name=Timestamp" json:"Timestamp,omitempty"`
}

func (m *TransactionBlock) Reset()                    { *m = TransactionBlock{} }
func (m *TransactionBlock) String() string            { return proto.CompactTextString(m) }
func (*TransactionBlock) ProtoMessage()               {}
func (*TransactionBlock) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *TransactionBlock) GetPreviousHash() string {
	if m != nil {
		return m.PreviousHash
	}
	return ""
}

func (m *TransactionBlock) GetDigest() string {
	if m != nil {
		return m.Digest
	}
	return ""
}

func (m *TransactionBlock) GetTransactions() []*types.Transaction {
	if m != nil {
		return m.Transactions
	}
	return nil
}

func (m *TransactionBlock) GetInvalidTransactions() []*types.InvalidTransactionRecord {
	if m != nil {
		return m.InvalidTransactions
	}
	return nil
}

func (m *TransactionBlock) GetSeqNo() uint64 {
	if m != nil {
		return m.SeqNo
	}
	return 0
}

func (m *TransactionBlock) GetView() uint64 {
	if m != nil {
		return m.View
	}
	return 0
}

func (m *TransactionBlock) GetIsPrimary() bool {
	if m != nil {
		return m.IsPrimary
	}
	return false
}

func (m *TransactionBlock) GetTimestamp() int64 {
	if m != nil {
		return m.Timestamp
	}
	return 0
}

type RollbackEvent struct {
	SeqNo uint64 `protobuf:"varint,1,opt,name=SeqNo" json:"SeqNo,omitempty"`
	Lid   uint64 `protobuf:"varint,2,opt,name=Lid" json:"Lid,omitempty"`
}

func (m *RollbackEvent) Reset()                    { *m = RollbackEvent{} }
func (m *RollbackEvent) String() string            { return proto.CompactTextString(m) }
func (*RollbackEvent) ProtoMessage()               {}
func (*RollbackEvent) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *RollbackEvent) GetSeqNo() uint64 {
	if m != nil {
		return m.SeqNo
	}
	return 0
}

func (m *RollbackEvent) GetLid() uint64 {
	if m != nil {
		return m.Lid
	}
	return 0
}

type OpLogAck struct {
	ExecuteOk bool   `protobuf:"varint,1,opt,name=executeOk" json:"executeOk,omitempty"`
	SeqNo     uint64 `protobuf:"varint,2,opt,name=seqNo" json:"seqNo,omitempty"`
}

func (m *OpLogAck) Reset()                    { *m = OpLogAck{} }
func (m *OpLogAck) String() string            { return proto.CompactTextString(m) }
func (*OpLogAck) ProtoMessage()               {}
func (*OpLogAck) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *OpLogAck) GetExecuteOk() bool {
	if m != nil {
		return m.ExecuteOk
	}
	return false
}

func (m *OpLogAck) GetSeqNo() uint64 {
	if m != nil {
		return m.SeqNo
	}
	return 0
}

type OpLogFetch struct {
	LogID uint64 `protobuf:"varint,1,opt,name=logID" json:"logID,omitempty"`
}

func (m *OpLogFetch) Reset()                    { *m = OpLogFetch{} }
func (m *OpLogFetch) String() string            { return proto.CompactTextString(m) }
func (*OpLogFetch) ProtoMessage()               {}
func (*OpLogFetch) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *OpLogFetch) GetLogID() uint64 {
	if m != nil {
		return m.LogID
	}
	return 0
}

type CheckpointAck struct {
	IsStableCkpt bool   `protobuf:"varint,1,opt,name=isStableCkpt" json:"isStableCkpt,omitempty"`
	Cid          uint64 `protobuf:"varint,2,opt,name=cid" json:"cid,omitempty"`
	Payload      []byte `protobuf:"bytes,3,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *CheckpointAck) Reset()                    { *m = CheckpointAck{} }
func (m *CheckpointAck) String() string            { return proto.CompactTextString(m) }
func (*CheckpointAck) ProtoMessage()               {}
func (*CheckpointAck) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *CheckpointAck) GetIsStableCkpt() bool {
	if m != nil {
		return m.IsStableCkpt
	}
	return false
}

func (m *CheckpointAck) GetCid() uint64 {
	if m != nil {
		return m.Cid
	}
	return 0
}

func (m *CheckpointAck) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func init() {
	proto.RegisterType((*AddNamespaceEvent)(nil), "event.AddNamespaceEvent")
	proto.RegisterType((*DeleteNamespaceEvent)(nil), "event.DeleteNamespaceEvent")
	proto.RegisterType((*AdminResponseEvent)(nil), "event.AdminResponseEvent")
	proto.RegisterType((*TestEvent)(nil), "event.TestEvent")
	proto.RegisterType((*TransactionBlock)(nil), "event.TransactionBlock")
	proto.RegisterType((*RollbackEvent)(nil), "event.RollbackEvent")
	proto.RegisterType((*OpLogAck)(nil), "event.OpLogAck")
	proto.RegisterType((*OpLogFetch)(nil), "event.OpLogFetch")
	proto.RegisterType((*CheckpointAck)(nil), "event.CheckpointAck")
}

func init() { proto.RegisterFile("event.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 477 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x00, 0x00, 0x00, 0x02, 0xff, 0x94, 0x53, 0x5d, 0x8f, 0xd2, 0x40,
	0x14, 0x0d, 0x2d, 0xb0, 0x70, 0x97, 0x35, 0x38, 0x6e, 0x4c, 0x63, 0x4c, 0x24, 0x7d, 0xe2, 0x45,
	0x88, 0x1f, 0x59, 0x9f, 0x34, 0xc1, 0x45, 0x23, 0xc9, 0xba, 0xac, 0xb3, 0xc4, 0x57, 0x33, 0x4c,
	0x6f, 0xda, 0x49, 0x3f, 0x66, 0xec, 0x0c, 0x28, 0xbf, 0xce, 0xbf, 0x66, 0x66, 0x5a, 0x28, 0x44,
	0x5f, 0xf6, 0xed, 0x9e, 0x33, 0xe7, 0x9c, 0x7b, 0x7b, 0x6f, 0x0a, 0xe7, 0xb8, 0xc5, 0xc2, 0x4c,
	0x54, 0x29, 0x8d, 0x24, 0x1d, 0x07, 0x9e, 0xbd, 0x8f, 0x85, 0x49, 0x36, 0xeb, 0x09, 0x97, 0xf9,
	0x34, 0xd9, 0x29, 0x2c, 0x79, 0xc2, 0x44, 0x71, 0x5c, 0x72, 0x59, 0xe2, 0xd4, 0xec, 0x14, 0xea,
	0xa9, 0x29, 0x59, 0xa1, 0x19, 0x37, 0x42, 0x16, 0x55, 0x4a, 0xf8, 0x0a, 0x1e, 0xcf, 0xa2, 0xe8,
	0x96, 0xe5, 0xa8, 0x15, 0xe3, 0xf8, 0xc9, 0x66, 0x92, 0xe7, 0xd0, 0x2f, 0xf6, 0x4c, 0xd0, 0x1a,
	0xb5, 0xc6, 0x7d, 0xda, 0x10, 0xe1, 0x5b, 0xb8, 0x9c, 0x63, 0x86, 0x06, 0x1f, 0xe4, 0xba, 0x02,
	0x32, 0x8b, 0x72, 0x51, 0x50, 0xd4, 0x4a, 0x16, 0xba, 0xf6, 0x3c, 0x02, 0x6f, 0x99, 0x3a, 0x71,
	0x8f, 0x7a, 0xcb, 0x94, 0x0c, 0xc1, 0xff, 0xaa, 0xe3, 0xc0, 0x73, 0x6e, 0x5b, 0x86, 0x2f, 0xa1,
	0xbf, 0x42, 0x6d, 0x0e, 0x72, 0x11, 0x39, 0x79, 0x9b, 0x7a, 0x22, 0xb2, 0xf2, 0xbc, 0x91, 0xe7,
	0x3a, 0x0e, 0xff, 0x78, 0x30, 0x5c, 0x35, 0x5f, 0xf9, 0x31, 0x93, 0x3c, 0x25, 0x21, 0x0c, 0xee,
	0x4a, 0xdc, 0x0a, 0xb9, 0xd1, 0x5f, 0x98, 0x4e, 0xea, 0xe1, 0x4e, 0x38, 0xf2, 0x14, 0xba, 0x73,
	0x11, 0xa3, 0x36, 0x75, 0x5a, 0x8d, 0xc8, 0x15, 0x0c, 0x8e, 0xf2, 0x74, 0xe0, 0x8f, 0xfc, 0xf1,
	0xf9, 0x6b, 0x32, 0x71, 0x0b, 0x9d, 0x1c, 0x3d, 0xd1, 0x13, 0x1d, 0xf9, 0x06, 0x4f, 0x16, 0xc5,
	0x96, 0x65, 0x22, 0x3a, 0xb1, 0xb7, 0x9d, 0xfd, 0x45, 0x6d, 0xff, 0x57, 0x41, 0x91, 0xcb, 0x32,
	0xa2, 0xff, 0xf3, 0x92, 0x4b, 0xe8, 0xdc, 0xe3, 0xcf, 0x5b, 0x19, 0x74, 0xdc, 0x02, 0x2a, 0x40,
	0x08, 0xb4, 0xbf, 0x0b, 0xfc, 0x15, 0x74, 0x1d, 0xe9, 0x6a, 0x7b, 0x8a, 0x85, 0xbe, 0x2b, 0x45,
	0xce, 0xca, 0x5d, 0x70, 0xe6, 0xb6, 0xdb, 0x10, 0xf6, 0x75, 0x25, 0x72, 0xd4, 0x86, 0xe5, 0x2a,
	0xe8, 0x8d, 0x5a, 0x63, 0x9f, 0x36, 0x44, 0xf8, 0x0e, 0x2e, 0xa8, 0xcc, 0xb2, 0x35, 0xe3, 0x69,
	0xb5, 0xf4, 0x43, 0xdb, 0xd6, 0x71, 0xdb, 0x21, 0xf8, 0x37, 0x22, 0x72, 0xcb, 0x6a, 0x53, 0x5b,
	0x86, 0x1f, 0xa0, 0xb7, 0x54, 0x37, 0x32, 0x9e, 0xf1, 0xd4, 0xb6, 0xc0, 0xdf, 0xc8, 0x37, 0x06,
	0x0f, 0xe7, 0x6d, 0x08, 0x9b, 0xa8, 0x5d, 0x62, 0xe5, 0xae, 0x40, 0x18, 0x02, 0x38, 0xff, 0x67,
	0x34, 0x3c, 0xb1, 0x9a, 0x4c, 0xc6, 0x8b, 0xf9, 0xbe, 0xab, 0x03, 0xe1, 0x0f, 0xb8, 0xb8, 0x4e,
	0x90, 0xa7, 0x4a, 0x8a, 0xc2, 0xcc, 0xaa, 0xd3, 0x0a, 0x7d, 0x6f, 0xd8, 0x3a, 0xc3, 0xeb, 0x54,
	0x99, 0xba, 0xd7, 0x09, 0x67, 0x47, 0xe5, 0xcd, 0xa8, 0x5c, 0x44, 0x24, 0x80, 0x33, 0xc5, 0x76,
	0x99, 0x64, 0x51, 0xe0, 0x8f, 0x5a, 0xe3, 0x01, 0xdd, 0xc3, 0x75, 0xd7, 0xfd, 0x16, 0x6f, 0xfe,
	0x06, 0x00, 0x00, 0xff, 0xff, 0x33, 0x8a, 0x33, 0x67, 0x6b, 0x03, 0x00, 0x00,
}
