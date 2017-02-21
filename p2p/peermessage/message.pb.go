// Code generated by protoc-gen-go.
// source: message.proto
// DO NOT EDIT!

/*
Package peermessage is a generated protocol buffer package.

It is generated from these files:
	message.proto

It has these top-level messages:
	Message
	Routers
	PeerAddress
	Signature
*/
package peermessage

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

// MsgType message type during data transfer
type Message_MsgType int32

const (
	Message_HELLO                  Message_MsgType = 0
	Message_HELLO_RESPONSE         Message_MsgType = 1
	Message_HELLOREVERSE           Message_MsgType = 2
	Message_HELLOREVERSE_RESPONSE  Message_MsgType = 3
	Message_INTRODUCE              Message_MsgType = 4
	Message_INTRODUCE_RESPONSE     Message_MsgType = 5
	Message_ATTEND                 Message_MsgType = 6
	Message_ATTEND_RESPONSE        Message_MsgType = 7
	Message_ATTEND_NOTIFY          Message_MsgType = 8
	Message_ATTEND_NOTIFY_RESPONSE Message_MsgType = 9
	Message_RESPONSE               Message_MsgType = 10
	Message_CONSUS                 Message_MsgType = 11
	Message_KEEPALIVE              Message_MsgType = 12
	Message_SYNCMSG                Message_MsgType = 13
	Message_PENDING                Message_MsgType = 14
	Message_RECONNECT              Message_MsgType = 15
	Message_RECONNECT_RESPONSE     Message_MsgType = 16
)

var Message_MsgType_name = map[int32]string{
	0:  "HELLO",
	1:  "HELLO_RESPONSE",
	2:  "HELLOREVERSE",
	3:  "HELLOREVERSE_RESPONSE",
	4:  "INTRODUCE",
	5:  "INTRODUCE_RESPONSE",
	6:  "ATTEND",
	7:  "ATTEND_RESPONSE",
	8:  "ATTEND_NOTIFY",
	9:  "ATTEND_NOTIFY_RESPONSE",
	10: "RESPONSE",
	11: "CONSUS",
	12: "KEEPALIVE",
	13: "SYNCMSG",
	14: "PENDING",
	15: "RECONNECT",
	16: "RECONNECT_RESPONSE",
}
var Message_MsgType_value = map[string]int32{
	"HELLO":                  0,
	"HELLO_RESPONSE":         1,
	"HELLOREVERSE":           2,
	"HELLOREVERSE_RESPONSE":  3,
	"INTRODUCE":              4,
	"INTRODUCE_RESPONSE":     5,
	"ATTEND":                 6,
	"ATTEND_RESPONSE":        7,
	"ATTEND_NOTIFY":          8,
	"ATTEND_NOTIFY_RESPONSE": 9,
	"RESPONSE":               10,
	"CONSUS":                 11,
	"KEEPALIVE":              12,
	"SYNCMSG":                13,
	"PENDING":                14,
	"RECONNECT":              15,
	"RECONNECT_RESPONSE":     16,
}

func (x Message_MsgType) String() string {
	return proto.EnumName(Message_MsgType_name, int32(x))
}
func (Message_MsgType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type Message struct {
	// From the message source
	From         *PeerAddress    `protobuf:"bytes,1,opt,name=From" json:"From,omitempty"`
	MessageType  Message_MsgType `protobuf:"varint,2,opt,name=MessageType,enum=peermessage.Message_MsgType" json:"MessageType,omitempty"`
	Payload      []byte          `protobuf:"bytes,3,opt,name=Payload,proto3" json:"Payload,omitempty"`
	Signature    *Signature      `protobuf:"bytes,4,opt,name=Signature" json:"Signature,omitempty"`
	MsgTimeStamp int64           `protobuf:"varint,5,opt,name=MsgTimeStamp" json:"MsgTimeStamp,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Message) GetFrom() *PeerAddress {
	if m != nil {
		return m.From
	}
	return nil
}

func (m *Message) GetMessageType() Message_MsgType {
	if m != nil {
		return m.MessageType
	}
	return Message_HELLO
}

func (m *Message) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

func (m *Message) GetSignature() *Signature {
	if m != nil {
		return m.Signature
	}
	return nil
}

func (m *Message) GetMsgTimeStamp() int64 {
	if m != nil {
		return m.MsgTimeStamp
	}
	return 0
}

// Routers table this is the router table which for transfer
type Routers struct {
	Routers []*PeerAddress `protobuf:"bytes,1,rep,name=routers" json:"routers,omitempty"`
}

func (m *Routers) Reset()                    { *m = Routers{} }
func (m *Routers) String() string            { return proto.CompactTextString(m) }
func (*Routers) ProtoMessage()               {}
func (*Routers) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Routers) GetRouters() []*PeerAddress {
	if m != nil {
		return m.Routers
	}
	return nil
}

// PeerAddress the peer address
type PeerAddress struct {
	IP        string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
	Port      int32  `protobuf:"varint,2,opt,name=Port" json:"Port,omitempty"`
	RPCPort   int32  `protobuf:"varint,3,opt,name=RPCPort" json:"RPCPort,omitempty"`
	Hash      string `protobuf:"bytes,4,opt,name=Hash" json:"Hash,omitempty"`
	ID        int32  `protobuf:"varint,5,opt,name=ID" json:"ID,omitempty"`
	IsPrimary bool   `protobuf:"varint,6,opt,name=IsPrimary" json:"IsPrimary,omitempty"`
}

func (m *PeerAddress) Reset()                    { *m = PeerAddress{} }
func (m *PeerAddress) String() string            { return proto.CompactTextString(m) }
func (*PeerAddress) ProtoMessage()               {}
func (*PeerAddress) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PeerAddress) GetIP() string {
	if m != nil {
		return m.IP
	}
	return ""
}

func (m *PeerAddress) GetPort() int32 {
	if m != nil {
		return m.Port
	}
	return 0
}

func (m *PeerAddress) GetRPCPort() int32 {
	if m != nil {
		return m.RPCPort
	}
	return 0
}

func (m *PeerAddress) GetHash() string {
	if m != nil {
		return m.Hash
	}
	return ""
}

func (m *PeerAddress) GetID() int32 {
	if m != nil {
		return m.ID
	}
	return 0
}

func (m *PeerAddress) GetIsPrimary() bool {
	if m != nil {
		return m.IsPrimary
	}
	return false
}

type Signature struct {
	ECert     []byte `protobuf:"bytes,1,opt,name=ECert,proto3" json:"ECert,omitempty"`
	RCert     []byte `protobuf:"bytes,2,opt,name=RCert,proto3" json:"RCert,omitempty"`
	Signature []byte `protobuf:"bytes,3,opt,name=Signature,proto3" json:"Signature,omitempty"`
}

func (m *Signature) Reset()                    { *m = Signature{} }
func (m *Signature) String() string            { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()               {}
func (*Signature) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Signature) GetECert() []byte {
	if m != nil {
		return m.ECert
	}
	return nil
}

func (m *Signature) GetRCert() []byte {
	if m != nil {
		return m.RCert
	}
	return nil
}

func (m *Signature) GetSignature() []byte {
	if m != nil {
		return m.Signature
	}
	return nil
}

func init() {
	proto.RegisterType((*Message)(nil), "peermessage.Message")
	proto.RegisterType((*Routers)(nil), "peermessage.Routers")
	proto.RegisterType((*PeerAddress)(nil), "peermessage.PeerAddress")
	proto.RegisterType((*Signature)(nil), "peermessage.Signature")
	proto.RegisterEnum("peermessage.Message_MsgType", Message_MsgType_name, Message_MsgType_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Chat service

type ChatClient interface {
	// double arrow data tranfer
	Chat(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error)
}

type chatClient struct {
	cc *grpc.ClientConn
}

func NewChatClient(cc *grpc.ClientConn) ChatClient {
	return &chatClient{cc}
}

func (c *chatClient) Chat(ctx context.Context, in *Message, opts ...grpc.CallOption) (*Message, error) {
	out := new(Message)
	err := grpc.Invoke(ctx, "/peermessage.Chat/Chat", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for Chat service

type ChatServer interface {
	// double arrow data tranfer
	Chat(context.Context, *Message) (*Message, error)
}

func RegisterChatServer(s *grpc.Server, srv ChatServer) {
	s.RegisterService(&_Chat_serviceDesc, srv)
}

func _Chat_Chat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Message)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ChatServer).Chat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/peermessage.Chat/Chat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ChatServer).Chat(ctx, req.(*Message))
	}
	return interceptor(ctx, in, info, handler)
}

var _Chat_serviceDesc = grpc.ServiceDesc{
	ServiceName: "peermessage.Chat",
	HandlerType: (*ChatServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Chat",
			Handler:    _Chat_Chat_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: "message.proto",
}

func init() { proto.RegisterFile("message.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 515 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x93, 0xc1, 0x6e, 0x9b, 0x4e,
	0x10, 0xc6, 0x03, 0x06, 0x13, 0x06, 0xec, 0xec, 0x7f, 0xfe, 0xa9, 0x45, 0xa3, 0x1c, 0x10, 0x27,
	0x1f, 0x2a, 0x1f, 0xdc, 0x1c, 0xdb, 0x4a, 0x16, 0xde, 0x24, 0xa8, 0x36, 0xa0, 0x05, 0x47, 0xca,
	0xa9, 0xa2, 0xf2, 0xca, 0xb1, 0x54, 0x17, 0x6b, 0x21, 0x07, 0x3f, 0x44, 0xdf, 0xa2, 0xcf, 0xd0,
	0xe7, 0xab, 0x76, 0xb1, 0x8d, 0x2d, 0x45, 0x3d, 0x31, 0xdf, 0x6f, 0xbf, 0x99, 0x1d, 0x89, 0x6f,
	0xa1, 0xb7, 0xe1, 0x55, 0x55, 0xac, 0xf8, 0x68, 0x2b, 0xca, 0xba, 0x44, 0x67, 0xcb, 0xb9, 0xd8,
	0xa3, 0xe0, 0xb7, 0x01, 0xd6, 0xbc, 0xa9, 0xf1, 0x03, 0x18, 0xf7, 0xa2, 0xdc, 0x78, 0x9a, 0xaf,
	0x0d, 0x9d, 0xb1, 0x37, 0x3a, 0xf1, 0x8d, 0x52, 0xce, 0xc5, 0x64, 0xb9, 0x14, 0xbc, 0xaa, 0x98,
	0x72, 0xe1, 0x17, 0x70, 0xf6, 0x8d, 0xf9, 0x6e, 0xcb, 0x3d, 0xdd, 0xd7, 0x86, 0xfd, 0xf1, 0xed,
	0x59, 0xd3, 0xfc, 0xf0, 0xad, 0x56, 0xd2, 0xc3, 0x4e, 0x1b, 0xd0, 0x03, 0x2b, 0x2d, 0x76, 0x3f,
	0xca, 0x62, 0xe9, 0x75, 0x7c, 0x6d, 0xe8, 0xb2, 0x83, 0xc4, 0x3b, 0xb0, 0xb3, 0xf5, 0xea, 0x67,
	0x51, 0xbf, 0x0a, 0xee, 0x19, 0x6a, 0x99, 0xc1, 0xd9, 0xdc, 0xe3, 0x29, 0x6b, 0x8d, 0x18, 0x80,
	0x2b, 0xef, 0x59, 0x6f, 0x78, 0x56, 0x17, 0x9b, 0xad, 0x67, 0xfa, 0xda, 0xb0, 0xc3, 0xce, 0x58,
	0xf0, 0x47, 0x07, 0x6b, 0xbf, 0x0c, 0xda, 0x60, 0x3e, 0xd2, 0xd9, 0x2c, 0x21, 0x17, 0x88, 0xd0,
	0x57, 0xe5, 0x37, 0x46, 0xb3, 0x34, 0x89, 0x33, 0x4a, 0x34, 0x24, 0xe0, 0x2a, 0xc6, 0xe8, 0x13,
	0x65, 0x19, 0x25, 0x3a, 0xbe, 0x87, 0x77, 0xa7, 0xa4, 0x35, 0x77, 0xb0, 0x07, 0x76, 0x14, 0xe7,
	0x2c, 0x99, 0x2e, 0x42, 0x4a, 0x0c, 0x1c, 0x00, 0x1e, 0x65, 0x6b, 0x33, 0x11, 0xa0, 0x3b, 0xc9,
	0x73, 0x1a, 0x4f, 0x49, 0x17, 0xff, 0x87, 0xab, 0xa6, 0x6e, 0x0d, 0x16, 0xfe, 0x07, 0xbd, 0x3d,
	0x8c, 0x93, 0x3c, 0xba, 0x7f, 0x26, 0x97, 0x78, 0x03, 0x83, 0x33, 0xd4, 0xda, 0x6d, 0x74, 0xe1,
	0xf2, 0xa8, 0x40, 0x4e, 0x0f, 0x93, 0x38, 0x5b, 0x64, 0xc4, 0x91, 0x0b, 0x7d, 0xa5, 0x34, 0x9d,
	0xcc, 0xa2, 0x27, 0x4a, 0x5c, 0x74, 0xc0, 0xca, 0x9e, 0xe3, 0x70, 0x9e, 0x3d, 0x90, 0x9e, 0x14,
	0x29, 0x8d, 0xa7, 0x51, 0xfc, 0x40, 0xfa, 0xd2, 0xc8, 0x68, 0x98, 0xc4, 0x31, 0x0d, 0x73, 0x72,
	0x25, 0x37, 0x3f, 0xca, 0xf6, 0x26, 0x12, 0x7c, 0x06, 0x8b, 0x95, 0xaf, 0x35, 0x17, 0x15, 0x8e,
	0xc1, 0x12, 0x4d, 0xe9, 0x69, 0x7e, 0xe7, 0x9f, 0x41, 0x39, 0x18, 0x83, 0x5f, 0x1a, 0x38, 0x27,
	0x07, 0xd8, 0x07, 0x3d, 0x4a, 0x55, 0xce, 0x6c, 0xa6, 0x47, 0x29, 0x22, 0x18, 0x69, 0x29, 0x6a,
	0x15, 0x22, 0x93, 0xa9, 0x5a, 0xe6, 0x83, 0xa5, 0xa1, 0xc2, 0x1d, 0x85, 0x0f, 0x52, 0xba, 0x1f,
	0x8b, 0xea, 0x45, 0x45, 0xc3, 0x66, 0xaa, 0x56, 0x13, 0xa7, 0xea, 0x9f, 0x9b, 0x4c, 0x8f, 0xa6,
	0x78, 0x0b, 0x76, 0x54, 0xa5, 0x62, 0xbd, 0x29, 0xc4, 0xce, 0xeb, 0xfa, 0xda, 0xf0, 0x92, 0xb5,
	0x20, 0x58, 0x9c, 0x24, 0x0c, 0xaf, 0xc1, 0xa4, 0x21, 0x17, 0xb5, 0xda, 0xc7, 0x65, 0x8d, 0x90,
	0x94, 0x29, 0xaa, 0x37, 0x54, 0x09, 0x39, 0xb6, 0x8d, 0x66, 0x13, 0xdb, 0x16, 0x8c, 0x3f, 0x81,
	0x11, 0xbe, 0x14, 0x35, 0xde, 0xed, 0xbf, 0xd7, 0x6f, 0xbd, 0x86, 0x9b, 0x37, 0x69, 0x70, 0xf1,
	0xbd, 0xab, 0x9e, 0xe7, 0xc7, 0xbf, 0x01, 0x00, 0x00, 0xff, 0xff, 0xbe, 0x6a, 0xe0, 0x1c, 0xaf,
	0x03, 0x00, 0x00,
}
