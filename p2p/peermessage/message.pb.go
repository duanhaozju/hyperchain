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
	Message_HELLO              Message_MsgType = 0
	Message_HELLO_RESPONSE     Message_MsgType = 1
	Message_INTRODUCE          Message_MsgType = 2
	Message_INTRODUCE_RESPONSE Message_MsgType = 3
	Message_ATTEND             Message_MsgType = 4
	Message_ATTEND_RESPNSE     Message_MsgType = 5
	Message_RESPONSE           Message_MsgType = 6
	Message_CONSUS             Message_MsgType = 7
	Message_KEEPALIVE          Message_MsgType = 8
	Message_SYNCMSG            Message_MsgType = 9
	Message_PENDING            Message_MsgType = 10
	Message_RECONNECT          Message_MsgType = 11
	Message_RECONNECT_RESPONSE Message_MsgType = 12
)

var Message_MsgType_name = map[int32]string{
	0:  "HELLO",
	1:  "HELLO_RESPONSE",
	2:  "INTRODUCE",
	3:  "INTRODUCE_RESPONSE",
	4:  "ATTEND",
	5:  "ATTEND_RESPNSE",
	6:  "RESPONSE",
	7:  "CONSUS",
	8:  "KEEPALIVE",
	9:  "SYNCMSG",
	10: "PENDING",
	11: "RECONNECT",
	12: "RECONNECT_RESPONSE",
}
var Message_MsgType_value = map[string]int32{
	"HELLO":              0,
	"HELLO_RESPONSE":     1,
	"INTRODUCE":          2,
	"INTRODUCE_RESPONSE": 3,
	"ATTEND":             4,
	"ATTEND_RESPNSE":     5,
	"RESPONSE":           6,
	"CONSUS":             7,
	"KEEPALIVE":          8,
	"SYNCMSG":            9,
	"PENDING":            10,
	"RECONNECT":          11,
	"RECONNECT_RESPONSE": 12,
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
	MsgTimeStamp int64           `protobuf:"varint,15,opt,name=MsgTimeStamp" json:"MsgTimeStamp,omitempty"`
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
	IP      string `protobuf:"bytes,1,opt,name=IP" json:"IP,omitempty"`
	Port    int32  `protobuf:"varint,2,opt,name=Port" json:"Port,omitempty"`
	RpcPort int32  `protobuf:"varint,3,opt,name=rpcPort" json:"rpcPort,omitempty"`
	Hash    string `protobuf:"bytes,4,opt,name=hash" json:"hash,omitempty"`
	ID      int32  `protobuf:"varint,5,opt,name=ID" json:"ID,omitempty"`
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

func (m *PeerAddress) GetRpcPort() int32 {
	if m != nil {
		return m.RpcPort
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

type Signature struct {
	Ecert     []byte `protobuf:"bytes,1,opt,name=ecert,proto3" json:"ecert,omitempty"`
	Rcert     []byte `protobuf:"bytes,2,opt,name=rcert,proto3" json:"rcert,omitempty"`
	Signature []byte `protobuf:"bytes,3,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *Signature) Reset()                    { *m = Signature{} }
func (m *Signature) String() string            { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()               {}
func (*Signature) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Signature) GetEcert() []byte {
	if m != nil {
		return m.Ecert
	}
	return nil
}

func (m *Signature) GetRcert() []byte {
	if m != nil {
		return m.Rcert
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
const _ = grpc.SupportPackageIsVersion3

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
	// 459 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x7c, 0x93, 0xcf, 0x6a, 0xdb, 0x40,
	0x10, 0xc6, 0xa3, 0x7f, 0x56, 0x34, 0x52, 0x5c, 0x31, 0x84, 0x20, 0x4a, 0x0e, 0x42, 0x27, 0x1f,
	0x8a, 0x0f, 0x6e, 0x8e, 0x6d, 0xc1, 0x48, 0xdb, 0x54, 0xd4, 0x5e, 0x89, 0x95, 0x5c, 0xe8, 0xa9,
	0xa8, 0xf1, 0x62, 0x07, 0xea, 0x4a, 0xac, 0x94, 0x43, 0x9e, 0xb0, 0xcf, 0xd0, 0xb7, 0x29, 0xbb,
	0xb2, 0x65, 0x1b, 0x4c, 0x4f, 0xfa, 0xbe, 0x6f, 0x7f, 0xbb, 0x33, 0x30, 0x23, 0xb8, 0xd9, 0xf1,
	0xb6, 0xad, 0x36, 0x7c, 0xda, 0x88, 0xba, 0xab, 0xd1, 0x6d, 0x38, 0x17, 0xfb, 0x28, 0xfa, 0x63,
	0x80, 0xbd, 0xec, 0x35, 0xbe, 0x03, 0xf3, 0xb3, 0xa8, 0x77, 0x81, 0x16, 0x6a, 0x13, 0x77, 0x16,
	0x4c, 0x4f, 0xb8, 0x69, 0xce, 0xb9, 0x98, 0xaf, 0xd7, 0x82, 0xb7, 0x2d, 0x53, 0x14, 0x7e, 0x02,
	0x77, 0x7f, 0xb1, 0x7c, 0x6d, 0x78, 0xa0, 0x87, 0xda, 0x64, 0x3c, 0xbb, 0x3f, 0xbb, 0xb4, 0x3c,
	0x7c, 0xdb, 0x8d, 0x64, 0xd8, 0xe9, 0x05, 0x0c, 0xc0, 0xce, 0xab, 0xd7, 0x5f, 0x75, 0xb5, 0x0e,
	0x8c, 0x50, 0x9b, 0x78, 0xec, 0x60, 0xf1, 0x01, 0x9c, 0xe2, 0x79, 0xf3, 0xbb, 0xea, 0x5e, 0x04,
	0x0f, 0x4c, 0xd5, 0xcc, 0xdd, 0xd9, 0xbb, 0xc3, 0x29, 0x3b, 0x82, 0x18, 0x81, 0x27, 0xeb, 0x3c,
	0xef, 0x78, 0xd1, 0x55, 0xbb, 0x26, 0x78, 0x13, 0x6a, 0x13, 0x83, 0x9d, 0x65, 0xd1, 0x5f, 0x0d,
	0xec, 0x7d, 0x33, 0xe8, 0x80, 0xf5, 0x85, 0x2c, 0x16, 0x99, 0x7f, 0x85, 0x08, 0x63, 0x25, 0x7f,
	0x30, 0x52, 0xe4, 0x19, 0x2d, 0x88, 0xaf, 0xe1, 0x0d, 0x38, 0x29, 0x2d, 0x59, 0x96, 0xac, 0x62,
	0xe2, 0xeb, 0x78, 0x07, 0x38, 0xd8, 0x23, 0x66, 0x20, 0xc0, 0x68, 0x5e, 0x96, 0x84, 0x26, 0xbe,
	0x29, 0x9f, 0xe9, 0xb5, 0x02, 0xe4, 0xb9, 0x85, 0x1e, 0x5c, 0x0f, 0xf4, 0x48, 0xd2, 0x71, 0x46,
	0x8b, 0x55, 0xe1, 0xdb, 0xb2, 0xc0, 0x57, 0x42, 0xf2, 0xf9, 0x22, 0xfd, 0x46, 0xfc, 0x6b, 0x74,
	0xc1, 0x2e, 0xbe, 0xd3, 0x78, 0x59, 0x3c, 0xfa, 0x8e, 0x34, 0x39, 0xa1, 0x49, 0x4a, 0x1f, 0x7d,
	0x90, 0x20, 0x23, 0x71, 0x46, 0x29, 0x89, 0x4b, 0xdf, 0x95, 0x9d, 0x0c, 0xf6, 0xd8, 0x89, 0x17,
	0x7d, 0x04, 0x9b, 0xd5, 0x2f, 0x1d, 0x17, 0x2d, 0xce, 0xc0, 0x16, 0xbd, 0x0c, 0xb4, 0xd0, 0xf8,
	0xef, 0x2c, 0x0f, 0x60, 0x54, 0x83, 0x7b, 0x92, 0xe3, 0x18, 0xf4, 0x34, 0x57, 0x9b, 0xe0, 0x30,
	0x3d, 0xcd, 0x11, 0xc1, 0xcc, 0x6b, 0xd1, 0xa9, 0x31, 0x5b, 0x4c, 0x69, 0x39, 0x41, 0xd1, 0x3c,
	0xa9, 0xd8, 0x50, 0xf1, 0xc1, 0x4a, 0x7a, 0x5b, 0xb5, 0x5b, 0x35, 0x3c, 0x87, 0x29, 0xad, 0x5e,
	0x4c, 0x02, 0x4b, 0x81, 0x7a, 0x9a, 0x44, 0xab, 0x93, 0x29, 0xe3, 0x2d, 0x58, 0xfc, 0x89, 0x8b,
	0x4e, 0x55, 0xf4, 0x58, 0x6f, 0x64, 0x2a, 0x54, 0xaa, 0xf7, 0xa9, 0x32, 0x78, 0x0f, 0x4e, 0x3b,
	0xac, 0x47, 0xbf, 0x3a, 0xc7, 0x60, 0xf6, 0x01, 0xcc, 0x78, 0x5b, 0x75, 0xf8, 0xb0, 0xff, 0xde,
	0x5e, 0xda, 0xc8, 0xb7, 0x17, 0xd3, 0xe8, 0xea, 0xe7, 0x48, 0xfd, 0x22, 0xef, 0xff, 0x05, 0x00,
	0x00, 0xff, 0xff, 0x8b, 0x16, 0xa2, 0x62, 0x33, 0x03, 0x00, 0x00,
}
