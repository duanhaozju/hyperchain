// Code generated by protoc-gen-go. DO NOT EDIT.
// source: service.proto

/*
Package service is a generated protocol buffer package.

It is generated from these files:
	service.proto

It has these top-level messages:
	IMessage
	RegisterMessage
*/
package service

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import (
	context "golang.org/x/net/context"
	grpc "google.golang.org/grpc"
)

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type Type int32

const (
	Type_UNDEFINED Type = 0
	Type_REGISTER  Type = 1
	Type_DISPATCH  Type = 2
	Type_ADMIN     Type = 3
	Type_RESPONSE  Type = 4
)

var Type_name = map[int32]string{
	0: "UNDEFINED",
	1: "REGISTER",
	2: "DISPATCH",
	3: "ADMIN",
	4: "RESPONSE",
}
var Type_value = map[string]int32{
	"UNDEFINED": 0,
	"REGISTER":  1,
	"DISPATCH":  2,
	"ADMIN":     3,
	"RESPONSE":  4,
}

func (x Type) String() string {
	return proto.EnumName(Type_name, int32(x))
}
func (Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type FROM int32

const (
	FROM_APISERVER     FROM = 0
	FROM_CONSENSUS     FROM = 1
	FROM_EXECUTOR      FROM = 2
	FROM_NETWORK       FROM = 3
	FROM_EVENTHUB      FROM = 4
	FROM_ADMINISTRATOR FROM = 5
)

var FROM_name = map[int32]string{
	0: "APISERVER",
	1: "CONSENSUS",
	2: "EXECUTOR",
	3: "NETWORK",
	4: "EVENTHUB",
	5: "ADMINISTRATOR",
}
var FROM_value = map[string]int32{
	"APISERVER":     0,
	"CONSENSUS":     1,
	"EXECUTOR":      2,
	"NETWORK":       3,
	"EVENTHUB":      4,
	"ADMINISTRATOR": 5,
}

func (x FROM) String() string {
	return proto.EnumName(FROM_name, int32(x))
}
func (FROM) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type Event int32

const (
	// 2.Consensus event
	Event_InformPrimaryEvent       Event = 0
	Event_VCResetEvent             Event = 1
	Event_BroadcastConsensusEvent  Event = 3
	Event_TxUniqueCastEvent        Event = 4
	Event_CommitEvent              Event = 5
	Event_ChainSyncReqEvent        Event = 6
	Event_BroadcastNewPeerEvent    Event = 7
	Event_BroadcastDelPeerEvent    Event = 8
	Event_UpdateRoutingTableEvent  Event = 9
	Event_FilterSystemStatusEvent  Event = 10
	Event_ExecutorToP2PEvent       Event = 11
	Event_ExecutorToConsensusEvent Event = 12
)

var Event_name = map[int32]string{
	0:  "InformPrimaryEvent",
	1:  "VCResetEvent",
	3:  "BroadcastConsensusEvent",
	4:  "TxUniqueCastEvent",
	5:  "CommitEvent",
	6:  "ChainSyncReqEvent",
	7:  "BroadcastNewPeerEvent",
	8:  "BroadcastDelPeerEvent",
	9:  "UpdateRoutingTableEvent",
	10: "FilterSystemStatusEvent",
	11: "ExecutorToP2PEvent",
	12: "ExecutorToConsensusEvent",
}
var Event_value = map[string]int32{
	"InformPrimaryEvent":       0,
	"VCResetEvent":             1,
	"BroadcastConsensusEvent":  3,
	"TxUniqueCastEvent":        4,
	"CommitEvent":              5,
	"ChainSyncReqEvent":        6,
	"BroadcastNewPeerEvent":    7,
	"BroadcastDelPeerEvent":    8,
	"UpdateRoutingTableEvent":  9,
	"FilterSystemStatusEvent":  10,
	"ExecutorToP2PEvent":       11,
	"ExecutorToConsensusEvent": 12,
}

func (x Event) String() string {
	return proto.EnumName(Event_name, int32(x))
}
func (Event) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

// IMessage is the wrapper of specific message used for internal component
type IMessage struct {
	Id      uint64 `protobuf:"varint,1,opt,name=id" json:"id,omitempty"`
	Type    Type   `protobuf:"varint,2,opt,name=type,enum=Type" json:"type,omitempty"`
	From    FROM   `protobuf:"varint,3,opt,name=from,enum=FROM" json:"from,omitempty"`
	Event   Event  `protobuf:"varint,4,opt,name=event,enum=Event" json:"event,omitempty"`
	Ok      bool   `protobuf:"varint,5,opt,name=ok" json:"ok,omitempty"`
	Payload []byte `protobuf:"bytes,6,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *IMessage) Reset()                    { *m = IMessage{} }
func (m *IMessage) String() string            { return proto.CompactTextString(m) }
func (*IMessage) ProtoMessage()               {}
func (*IMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *IMessage) GetId() uint64 {
	if m != nil {
		return m.Id
	}
	return 0
}

func (m *IMessage) GetType() Type {
	if m != nil {
		return m.Type
	}
	return Type_UNDEFINED
}

func (m *IMessage) GetFrom() FROM {
	if m != nil {
		return m.From
	}
	return FROM_APISERVER
}

func (m *IMessage) GetEvent() Event {
	if m != nil {
		return m.Event
	}
	return Event_InformPrimaryEvent
}

func (m *IMessage) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

func (m *IMessage) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

// RegisterMessage used by service component to connect to Dispatcher.
type RegisterMessage struct {
	Address   string `protobuf:"bytes,1,opt,name=address" json:"address,omitempty"`
	Port      string `protobuf:"bytes,2,opt,name=port" json:"port,omitempty"`
	Namespace string `protobuf:"bytes,3,opt,name=namespace" json:"namespace,omitempty"`
	Payload   []byte `protobuf:"bytes,4,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *RegisterMessage) Reset()                    { *m = RegisterMessage{} }
func (m *RegisterMessage) String() string            { return proto.CompactTextString(m) }
func (*RegisterMessage) ProtoMessage()               {}
func (*RegisterMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *RegisterMessage) GetAddress() string {
	if m != nil {
		return m.Address
	}
	return ""
}

func (m *RegisterMessage) GetPort() string {
	if m != nil {
		return m.Port
	}
	return ""
}

func (m *RegisterMessage) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *RegisterMessage) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func init() {
	proto.RegisterType((*IMessage)(nil), "IMessage")
	proto.RegisterType((*RegisterMessage)(nil), "RegisterMessage")
	proto.RegisterEnum("Type", Type_name, Type_value)
	proto.RegisterEnum("FROM", FROM_name, FROM_value)
	proto.RegisterEnum("Event", Event_name, Event_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Dispatcher service

type DispatcherClient interface {
	Register(ctx context.Context, opts ...grpc.CallOption) (Dispatcher_RegisterClient, error)
}

type dispatcherClient struct {
	cc *grpc.ClientConn
}

func NewDispatcherClient(cc *grpc.ClientConn) DispatcherClient {
	return &dispatcherClient{cc}
}

func (c *dispatcherClient) Register(ctx context.Context, opts ...grpc.CallOption) (Dispatcher_RegisterClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Dispatcher_serviceDesc.Streams[0], c.cc, "/Dispatcher/Register", opts...)
	if err != nil {
		return nil, err
	}
	x := &dispatcherRegisterClient{stream}
	return x, nil
}

type Dispatcher_RegisterClient interface {
	Send(*IMessage) error
	Recv() (*IMessage, error)
	grpc.ClientStream
}

type dispatcherRegisterClient struct {
	grpc.ClientStream
}

func (x *dispatcherRegisterClient) Send(m *IMessage) error {
	return x.ClientStream.SendMsg(m)
}

func (x *dispatcherRegisterClient) Recv() (*IMessage, error) {
	m := new(IMessage)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Dispatcher service

type DispatcherServer interface {
	Register(Dispatcher_RegisterServer) error
}

func RegisterDispatcherServer(s *grpc.Server, srv DispatcherServer) {
	s.RegisterService(&_Dispatcher_serviceDesc, srv)
}

func _Dispatcher_Register_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(DispatcherServer).Register(&dispatcherRegisterServer{stream})
}

type Dispatcher_RegisterServer interface {
	Send(*IMessage) error
	Recv() (*IMessage, error)
	grpc.ServerStream
}

type dispatcherRegisterServer struct {
	grpc.ServerStream
}

func (x *dispatcherRegisterServer) Send(m *IMessage) error {
	return x.ServerStream.SendMsg(m)
}

func (x *dispatcherRegisterServer) Recv() (*IMessage, error) {
	m := new(IMessage)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Dispatcher_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Dispatcher",
	HandlerType: (*DispatcherServer)(nil),
	Methods:     []grpc.MethodDesc{},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Register",
			Handler:       _Dispatcher_Register_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "service.proto",
}

func init() { proto.RegisterFile("service.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 554 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x5c, 0x93, 0xd1, 0x4e, 0xdb, 0x30,
	0x14, 0x86, 0x9b, 0x92, 0xb4, 0xcd, 0xa1, 0x40, 0xb0, 0xc4, 0x16, 0x36, 0x2e, 0x2a, 0x2e, 0xa6,
	0x8a, 0x8b, 0x6a, 0x62, 0x7b, 0x81, 0x92, 0x9a, 0x91, 0x4d, 0xa4, 0x91, 0x93, 0xb2, 0x5d, 0xce,
	0x34, 0x07, 0x88, 0x68, 0xe2, 0x60, 0xbb, 0x40, 0xdf, 0x64, 0xef, 0xb1, 0x17, 0x9c, 0x9c, 0xd0,
	0x75, 0xec, 0xce, 0xff, 0xf7, 0x9f, 0x9c, 0xf3, 0x1f, 0x2b, 0x86, 0x1d, 0x85, 0xf2, 0x31, 0x9f,
	0xe3, 0xa8, 0x92, 0x42, 0x8b, 0xe3, 0x5f, 0x16, 0xf4, 0xc2, 0x4b, 0x54, 0x8a, 0xdf, 0x22, 0xd9,
	0x85, 0x76, 0x9e, 0xf9, 0xd6, 0xc0, 0x1a, 0xda, 0xac, 0x9d, 0x67, 0xe4, 0x10, 0x6c, 0xbd, 0xaa,
	0xd0, 0x6f, 0x0f, 0xac, 0xe1, 0xee, 0xa9, 0x33, 0x4a, 0x57, 0x15, 0xb2, 0x1a, 0x19, 0xeb, 0x46,
	0x8a, 0xc2, 0xdf, 0x7a, 0xb1, 0xce, 0xd9, 0xf4, 0x92, 0xd5, 0x88, 0x1c, 0x81, 0x83, 0x8f, 0x58,
	0x6a, 0xdf, 0xae, 0xbd, 0xce, 0x88, 0x1a, 0xc5, 0x1a, 0x68, 0x66, 0x88, 0x7b, 0xdf, 0x19, 0x58,
	0xc3, 0x1e, 0x6b, 0x8b, 0x7b, 0xe2, 0x43, 0xb7, 0xe2, 0xab, 0x85, 0xe0, 0x99, 0xdf, 0x19, 0x58,
	0xc3, 0x3e, 0x5b, 0xcb, 0xe3, 0x27, 0xd8, 0x63, 0x78, 0x9b, 0x2b, 0x8d, 0x72, 0x1d, 0xd0, 0x87,
	0x2e, 0xcf, 0x32, 0x89, 0x4a, 0xd5, 0x29, 0x5d, 0xb6, 0x96, 0x84, 0x80, 0x5d, 0x09, 0xa9, 0xeb,
	0xa8, 0x2e, 0xab, 0xcf, 0xe4, 0x08, 0xdc, 0x92, 0x17, 0xa8, 0x2a, 0x3e, 0xc7, 0x3a, 0xa8, 0xcb,
	0x36, 0xe0, 0xdf, 0xc1, 0xf6, 0xab, 0xc1, 0x27, 0x5f, 0xc1, 0x36, 0x9b, 0x92, 0x1d, 0x70, 0x67,
	0xd1, 0x84, 0x9e, 0x87, 0x11, 0x9d, 0x78, 0x2d, 0xd2, 0x87, 0x1e, 0xa3, 0x5f, 0xc2, 0x24, 0xa5,
	0xcc, 0xb3, 0x8c, 0x9a, 0x84, 0x49, 0x3c, 0x4e, 0x83, 0x0b, 0xaf, 0x4d, 0x5c, 0x70, 0xc6, 0x93,
	0xcb, 0x30, 0xf2, 0xb6, 0x9a, 0xb2, 0x24, 0x9e, 0x46, 0x09, 0xf5, 0xec, 0x93, 0x9f, 0x60, 0x9b,
	0xab, 0x31, 0xbd, 0xc6, 0x71, 0x98, 0x50, 0x76, 0x45, 0x99, 0xd7, 0x32, 0x32, 0x30, 0x15, 0x51,
	0x32, 0x4b, 0x9a, 0x66, 0xf4, 0x07, 0x0d, 0x66, 0xe9, 0x94, 0x79, 0x6d, 0xb2, 0x0d, 0xdd, 0x88,
	0xa6, 0xdf, 0xa7, 0xec, 0x5b, 0xd3, 0x8e, 0x5e, 0xd1, 0x28, 0xbd, 0x98, 0x9d, 0x79, 0x36, 0xd9,
	0x87, 0x9d, 0x7a, 0x4e, 0x98, 0xa4, 0x6c, 0x6c, 0xaa, 0x9d, 0x93, 0xdf, 0x6d, 0x70, 0xea, 0x1b,
	0x26, 0x6f, 0x80, 0x84, 0xe5, 0x8d, 0x90, 0x45, 0x2c, 0xf3, 0x82, 0xcb, 0x55, 0x4d, 0xbd, 0x16,
	0xf1, 0xa0, 0x7f, 0x15, 0x30, 0x54, 0xa8, 0x1b, 0x62, 0x91, 0xf7, 0xf0, 0xf6, 0x4c, 0x0a, 0x9e,
	0xcd, 0xb9, 0xd2, 0x81, 0x28, 0x15, 0x96, 0x6a, 0xa9, 0x1a, 0x73, 0x8b, 0x1c, 0xc0, 0x7e, 0xfa,
	0x3c, 0x2b, 0xf3, 0x87, 0x25, 0x06, 0x5c, 0xbd, 0x7c, 0x63, 0x93, 0x3d, 0xd8, 0x0e, 0x44, 0x51,
	0xe4, 0x2f, 0xc0, 0x31, 0x75, 0xc1, 0x1d, 0xcf, 0xcb, 0x64, 0x55, 0xce, 0x19, 0x3e, 0x34, 0xb8,
	0x43, 0x0e, 0xe1, 0xe0, 0x6f, 0xef, 0x08, 0x9f, 0x62, 0x44, 0xd9, 0x58, 0xdd, 0x57, 0xd6, 0x04,
	0x17, 0x1b, 0xab, 0x67, 0x12, 0xcd, 0xaa, 0x8c, 0x6b, 0x64, 0x62, 0xa9, 0xf3, 0xf2, 0x36, 0xe5,
	0xd7, 0x0b, 0x6c, 0x4c, 0xd7, 0x98, 0xe7, 0xf9, 0x42, 0xa3, 0x4c, 0x56, 0x4a, 0x63, 0x91, 0x68,
	0xae, 0xd7, 0x71, 0xc1, 0x6c, 0x4d, 0x9f, 0x71, 0xbe, 0xd4, 0x42, 0xa6, 0x22, 0x3e, 0x8d, 0x1b,
	0xbe, 0x4d, 0x8e, 0xc0, 0xdf, 0xf0, 0xff, 0x96, 0xec, 0x9f, 0x7e, 0x06, 0x98, 0xe4, 0xaa, 0xe2,
	0x7a, 0x7e, 0x87, 0x92, 0x7c, 0x80, 0xde, 0xfa, 0x57, 0x23, 0xee, 0x68, 0xfd, 0x1e, 0xde, 0x6d,
	0x8e, 0xc7, 0xad, 0xa1, 0xf5, 0xd1, 0xba, 0xee, 0xd4, 0x8f, 0xe6, 0xd3, 0x9f, 0x00, 0x00, 0x00,
	0xff, 0xff, 0x84, 0x0f, 0xb3, 0x3f, 0x45, 0x03, 0x00, 0x00,
}
