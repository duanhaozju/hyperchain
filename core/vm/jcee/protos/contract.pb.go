// Code generated by protoc-gen-go. DO NOT EDIT.
// source: contract.proto

/*
Package contract is a generated protocol buffer package.

It is generated from these files:
	contract.proto

It has these top-level messages:
	Message
	Request
	Response
	RequestContext
	Event
	Key
	BatchKey
	BathValue
	BatchKV
	Range
	Value
	KeyValue
	LedgerContext
*/
package contract

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

type Message_Type int32

const (
	Message_UNDEFINED   Message_Type = 0
	Message_REGISTER    Message_Type = 1
	Message_TRANSACTION Message_Type = 2
	Message_GET         Message_Type = 3
	Message_PUT         Message_Type = 4
	Message_DELETE      Message_Type = 5
	Message_BATCH_READ  Message_Type = 6
	Message_BATCH_WRITE Message_Type = 7
	Message_RANGE_QUERY Message_Type = 8
	Message_POST_EVENT  Message_Type = 9
	Message_RESPONSE    Message_Type = 10
	Message_HEART_BEAT  Message_Type = 11
)

var Message_Type_name = map[int32]string{
	0:  "UNDEFINED",
	1:  "REGISTER",
	2:  "TRANSACTION",
	3:  "GET",
	4:  "PUT",
	5:  "DELETE",
	6:  "BATCH_READ",
	7:  "BATCH_WRITE",
	8:  "RANGE_QUERY",
	9:  "POST_EVENT",
	10: "RESPONSE",
	11: "HEART_BEAT",
}
var Message_Type_value = map[string]int32{
	"UNDEFINED":   0,
	"REGISTER":    1,
	"TRANSACTION": 2,
	"GET":         3,
	"PUT":         4,
	"DELETE":      5,
	"BATCH_READ":  6,
	"BATCH_WRITE": 7,
	"RANGE_QUERY": 8,
	"POST_EVENT":  9,
	"RESPONSE":    10,
	"HEART_BEAT":  11,
}

func (x Message_Type) String() string {
	return proto.EnumName(Message_Type_name, int32(x))
}
func (Message_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

// Message is the wrapper of specific message
type Message struct {
	Type    Message_Type `protobuf:"varint,1,opt,name=type,enum=Message_Type" json:"type,omitempty"`
	Payload []byte       `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *Message) Reset()                    { *m = Message{} }
func (m *Message) String() string            { return proto.CompactTextString(m) }
func (*Message) ProtoMessage()               {}
func (*Message) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Message) GetType() Message_Type {
	if m != nil {
		return m.Type
	}
	return Message_UNDEFINED
}

func (m *Message) GetPayload() []byte {
	if m != nil {
		return m.Payload
	}
	return nil
}

type Request struct {
	Context *RequestContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	Method  string          `protobuf:"bytes,2,opt,name=method" json:"method,omitempty"`
	Args    [][]byte        `protobuf:"bytes,3,rep,name=args,proto3" json:"args,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *Request) GetContext() *RequestContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *Request) GetMethod() string {
	if m != nil {
		return m.Method
	}
	return ""
}

func (m *Request) GetArgs() [][]byte {
	if m != nil {
		return m.Args
	}
	return nil
}

type Response struct {
	Ok       bool   `protobuf:"varint,1,opt,name=ok" json:"ok,omitempty"`
	Result   []byte `protobuf:"bytes,2,opt,name=result,proto3" json:"result,omitempty"`
	CodeHash string `protobuf:"bytes,3,opt,name=codeHash" json:"codeHash,omitempty"`
}

func (m *Response) Reset()                    { *m = Response{} }
func (m *Response) String() string            { return proto.CompactTextString(m) }
func (*Response) ProtoMessage()               {}
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *Response) GetOk() bool {
	if m != nil {
		return m.Ok
	}
	return false
}

func (m *Response) GetResult() []byte {
	if m != nil {
		return m.Result
	}
	return nil
}

func (m *Response) GetCodeHash() string {
	if m != nil {
		return m.CodeHash
	}
	return ""
}

type RequestContext struct {
	Txid        string `protobuf:"bytes,1,opt,name=txid" json:"txid,omitempty"`
	Namespace   string `protobuf:"bytes,2,opt,name=namespace" json:"namespace,omitempty"`
	Cid         string `protobuf:"bytes,3,opt,name=cid" json:"cid,omitempty"`
	Invoker     string `protobuf:"bytes,4,opt,name=invoker" json:"invoker,omitempty"`
	BlockNumber uint64 `protobuf:"varint,5,opt,name=blockNumber" json:"blockNumber,omitempty"`
}

func (m *RequestContext) Reset()                    { *m = RequestContext{} }
func (m *RequestContext) String() string            { return proto.CompactTextString(m) }
func (*RequestContext) ProtoMessage()               {}
func (*RequestContext) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *RequestContext) GetTxid() string {
	if m != nil {
		return m.Txid
	}
	return ""
}

func (m *RequestContext) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *RequestContext) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *RequestContext) GetInvoker() string {
	if m != nil {
		return m.Invoker
	}
	return ""
}

func (m *RequestContext) GetBlockNumber() uint64 {
	if m != nil {
		return m.BlockNumber
	}
	return 0
}

type Event struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	Topics  [][]byte       `protobuf:"bytes,2,rep,name=topics,proto3" json:"topics,omitempty"`
	Body    []byte         `protobuf:"bytes,3,opt,name=body,proto3" json:"body,omitempty"`
}

func (m *Event) Reset()                    { *m = Event{} }
func (m *Event) String() string            { return proto.CompactTextString(m) }
func (*Event) ProtoMessage()               {}
func (*Event) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *Event) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *Event) GetTopics() [][]byte {
	if m != nil {
		return m.Topics
	}
	return nil
}

func (m *Event) GetBody() []byte {
	if m != nil {
		return m.Body
	}
	return nil
}

type Key struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	K       []byte         `protobuf:"bytes,2,opt,name=k,proto3" json:"k,omitempty"`
}

func (m *Key) Reset()                    { *m = Key{} }
func (m *Key) String() string            { return proto.CompactTextString(m) }
func (*Key) ProtoMessage()               {}
func (*Key) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

func (m *Key) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *Key) GetK() []byte {
	if m != nil {
		return m.K
	}
	return nil
}

// BatchKey used to fetch data by batch
type BatchKey struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	K       [][]byte       `protobuf:"bytes,2,rep,name=k,proto3" json:"k,omitempty"`
}

func (m *BatchKey) Reset()                    { *m = BatchKey{} }
func (m *BatchKey) String() string            { return proto.CompactTextString(m) }
func (*BatchKey) ProtoMessage()               {}
func (*BatchKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

func (m *BatchKey) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *BatchKey) GetK() [][]byte {
	if m != nil {
		return m.K
	}
	return nil
}

// BathKV bach result
type BathValue struct {
	Id      string   `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	HasMore bool     `protobuf:"varint,2,opt,name=hasMore" json:"hasMore,omitempty"`
	V       [][]byte `protobuf:"bytes,3,rep,name=v,proto3" json:"v,omitempty"`
}

func (m *BathValue) Reset()                    { *m = BathValue{} }
func (m *BathValue) String() string            { return proto.CompactTextString(m) }
func (*BathValue) ProtoMessage()               {}
func (*BathValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *BathValue) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *BathValue) GetHasMore() bool {
	if m != nil {
		return m.HasMore
	}
	return false
}

func (m *BathValue) GetV() [][]byte {
	if m != nil {
		return m.V
	}
	return nil
}

type BatchKV struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	Kv      []*KeyValue    `protobuf:"bytes,2,rep,name=kv" json:"kv,omitempty"`
}

func (m *BatchKV) Reset()                    { *m = BatchKV{} }
func (m *BatchKV) String() string            { return proto.CompactTextString(m) }
func (*BatchKV) ProtoMessage()               {}
func (*BatchKV) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *BatchKV) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *BatchKV) GetKv() []*KeyValue {
	if m != nil {
		return m.Kv
	}
	return nil
}

// Range specifiy query range
type Range struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	Start   []byte         `protobuf:"bytes,2,opt,name=start,proto3" json:"start,omitempty"`
	End     []byte         `protobuf:"bytes,3,opt,name=end,proto3" json:"end,omitempty"`
}

func (m *Range) Reset()                    { *m = Range{} }
func (m *Range) String() string            { return proto.CompactTextString(m) }
func (*Range) ProtoMessage()               {}
func (*Range) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *Range) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *Range) GetStart() []byte {
	if m != nil {
		return m.Start
	}
	return nil
}

func (m *Range) GetEnd() []byte {
	if m != nil {
		return m.End
	}
	return nil
}

type Value struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	V  []byte `protobuf:"bytes,2,opt,name=v,proto3" json:"v,omitempty"`
}

func (m *Value) Reset()                    { *m = Value{} }
func (m *Value) String() string            { return proto.CompactTextString(m) }
func (*Value) ProtoMessage()               {}
func (*Value) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *Value) GetId() string {
	if m != nil {
		return m.Id
	}
	return ""
}

func (m *Value) GetV() []byte {
	if m != nil {
		return m.V
	}
	return nil
}

type KeyValue struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	K       []byte         `protobuf:"bytes,2,opt,name=k,proto3" json:"k,omitempty"`
	V       []byte         `protobuf:"bytes,3,opt,name=v,proto3" json:"v,omitempty"`
}

func (m *KeyValue) Reset()                    { *m = KeyValue{} }
func (m *KeyValue) String() string            { return proto.CompactTextString(m) }
func (*KeyValue) ProtoMessage()               {}
func (*KeyValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *KeyValue) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

func (m *KeyValue) GetK() []byte {
	if m != nil {
		return m.K
	}
	return nil
}

func (m *KeyValue) GetV() []byte {
	if m != nil {
		return m.V
	}
	return nil
}

type LedgerContext struct {
	Txid        string `protobuf:"bytes,1,opt,name=txid" json:"txid,omitempty"`
	Namespace   string `protobuf:"bytes,2,opt,name=namespace" json:"namespace,omitempty"`
	Cid         string `protobuf:"bytes,3,opt,name=cid" json:"cid,omitempty"`
	BlockNumber uint64 `protobuf:"varint,4,opt,name=blockNumber" json:"blockNumber,omitempty"`
}

func (m *LedgerContext) Reset()                    { *m = LedgerContext{} }
func (m *LedgerContext) String() string            { return proto.CompactTextString(m) }
func (*LedgerContext) ProtoMessage()               {}
func (*LedgerContext) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *LedgerContext) GetTxid() string {
	if m != nil {
		return m.Txid
	}
	return ""
}

func (m *LedgerContext) GetNamespace() string {
	if m != nil {
		return m.Namespace
	}
	return ""
}

func (m *LedgerContext) GetCid() string {
	if m != nil {
		return m.Cid
	}
	return ""
}

func (m *LedgerContext) GetBlockNumber() uint64 {
	if m != nil {
		return m.BlockNumber
	}
	return 0
}

func init() {
	proto.RegisterType((*Message)(nil), "Message")
	proto.RegisterType((*Request)(nil), "Request")
	proto.RegisterType((*Response)(nil), "Response")
	proto.RegisterType((*RequestContext)(nil), "RequestContext")
	proto.RegisterType((*Event)(nil), "Event")
	proto.RegisterType((*Key)(nil), "Key")
	proto.RegisterType((*BatchKey)(nil), "BatchKey")
	proto.RegisterType((*BathValue)(nil), "BathValue")
	proto.RegisterType((*BatchKV)(nil), "BatchKV")
	proto.RegisterType((*Range)(nil), "Range")
	proto.RegisterType((*Value)(nil), "Value")
	proto.RegisterType((*KeyValue)(nil), "KeyValue")
	proto.RegisterType((*LedgerContext)(nil), "LedgerContext")
	proto.RegisterEnum("Message_Type", Message_Type_name, Message_Type_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion4

// Client API for Contract service

type ContractClient interface {
	// HeartBeat used to detect the health of server
	HeartBeat(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// Interface that provides support to jvm execution which will estabilish a stream between Server
	// and Client. Message type provide the context necessary for server to respond appropriately.
	Register(ctx context.Context, opts ...grpc.CallOption) (Contract_RegisterClient, error)
}

type contractClient struct {
	cc *grpc.ClientConn
}

func NewContractClient(cc *grpc.ClientConn) ContractClient {
	return &contractClient{cc}
}

func (c *contractClient) HeartBeat(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Contract/HeartBeat", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractClient) Register(ctx context.Context, opts ...grpc.CallOption) (Contract_RegisterClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Contract_serviceDesc.Streams[0], c.cc, "/Contract/Register", opts...)
	if err != nil {
		return nil, err
	}
	x := &contractRegisterClient{stream}
	return x, nil
}

type Contract_RegisterClient interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type contractRegisterClient struct {
	grpc.ClientStream
}

func (x *contractRegisterClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *contractRegisterClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Contract service

type ContractServer interface {
	// HeartBeat used to detect the health of server
	HeartBeat(context.Context, *Request) (*Response, error)
	// Interface that provides support to jvm execution which will estabilish a stream between Server
	// and Client. Message type provide the context necessary for server to respond appropriately.
	Register(Contract_RegisterServer) error
}

func RegisterContractServer(s *grpc.Server, srv ContractServer) {
	s.RegisterService(&_Contract_serviceDesc, srv)
}

func _Contract_HeartBeat_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractServer).HeartBeat(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Contract/HeartBeat",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractServer).HeartBeat(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
}

func _Contract_Register_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ContractServer).Register(&contractRegisterServer{stream})
}

type Contract_RegisterServer interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type contractRegisterServer struct {
	grpc.ServerStream
}

func (x *contractRegisterServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *contractRegisterServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Contract_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Contract",
	HandlerType: (*ContractServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "HeartBeat",
			Handler:    _Contract_HeartBeat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "Register",
			Handler:       _Contract_Register_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "contract.proto",
}

// Client API for Ledger service

type LedgerClient interface {
	Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Value, error)
	Put(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*Response, error)
	Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Response, error)
	BatchRead(ctx context.Context, in *BatchKey, opts ...grpc.CallOption) (*BathValue, error)
	BatchWrite(ctx context.Context, in *BatchKV, opts ...grpc.CallOption) (*Response, error)
	RangeQuery(ctx context.Context, in *Range, opts ...grpc.CallOption) (Ledger_RangeQueryClient, error)
	Post(ctx context.Context, in *Event, opts ...grpc.CallOption) (*Response, error)
	Register(ctx context.Context, opts ...grpc.CallOption) (Ledger_RegisterClient, error)
}

type ledgerClient struct {
	cc *grpc.ClientConn
}

func NewLedgerClient(cc *grpc.ClientConn) LedgerClient {
	return &ledgerClient{cc}
}

func (c *ledgerClient) Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Value, error) {
	out := new(Value)
	err := grpc.Invoke(ctx, "/Ledger/Get", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ledgerClient) Put(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Ledger/Put", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ledgerClient) Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Ledger/Delete", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ledgerClient) BatchRead(ctx context.Context, in *BatchKey, opts ...grpc.CallOption) (*BathValue, error) {
	out := new(BathValue)
	err := grpc.Invoke(ctx, "/Ledger/BatchRead", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ledgerClient) BatchWrite(ctx context.Context, in *BatchKV, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Ledger/BatchWrite", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ledgerClient) RangeQuery(ctx context.Context, in *Range, opts ...grpc.CallOption) (Ledger_RangeQueryClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Ledger_serviceDesc.Streams[0], c.cc, "/Ledger/RangeQuery", opts...)
	if err != nil {
		return nil, err
	}
	x := &ledgerRangeQueryClient{stream}
	if err := x.ClientStream.SendMsg(in); err != nil {
		return nil, err
	}
	if err := x.ClientStream.CloseSend(); err != nil {
		return nil, err
	}
	return x, nil
}

type Ledger_RangeQueryClient interface {
	Recv() (*BathValue, error)
	grpc.ClientStream
}

type ledgerRangeQueryClient struct {
	grpc.ClientStream
}

func (x *ledgerRangeQueryClient) Recv() (*BathValue, error) {
	m := new(BathValue)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *ledgerClient) Post(ctx context.Context, in *Event, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Ledger/Post", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *ledgerClient) Register(ctx context.Context, opts ...grpc.CallOption) (Ledger_RegisterClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Ledger_serviceDesc.Streams[1], c.cc, "/Ledger/Register", opts...)
	if err != nil {
		return nil, err
	}
	x := &ledgerRegisterClient{stream}
	return x, nil
}

type Ledger_RegisterClient interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ClientStream
}

type ledgerRegisterClient struct {
	grpc.ClientStream
}

func (x *ledgerRegisterClient) Send(m *Message) error {
	return x.ClientStream.SendMsg(m)
}

func (x *ledgerRegisterClient) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

// Server API for Ledger service

type LedgerServer interface {
	Get(context.Context, *Key) (*Value, error)
	Put(context.Context, *KeyValue) (*Response, error)
	Delete(context.Context, *Key) (*Response, error)
	BatchRead(context.Context, *BatchKey) (*BathValue, error)
	BatchWrite(context.Context, *BatchKV) (*Response, error)
	RangeQuery(*Range, Ledger_RangeQueryServer) error
	Post(context.Context, *Event) (*Response, error)
	Register(Ledger_RegisterServer) error
}

func RegisterLedgerServer(s *grpc.Server, srv LedgerServer) {
	s.RegisterService(&_Ledger_serviceDesc, srv)
}

func _Ledger_Get_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).Get(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/Get",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).Get(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ledger_Put_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(KeyValue)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).Put(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/Put",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).Put(ctx, req.(*KeyValue))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ledger_Delete_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Key)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).Delete(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/Delete",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).Delete(ctx, req.(*Key))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ledger_BatchRead_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchKey)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).BatchRead(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/BatchRead",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).BatchRead(ctx, req.(*BatchKey))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ledger_BatchWrite_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(BatchKV)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).BatchWrite(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/BatchWrite",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).BatchWrite(ctx, req.(*BatchKV))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ledger_RangeQuery_Handler(srv interface{}, stream grpc.ServerStream) error {
	m := new(Range)
	if err := stream.RecvMsg(m); err != nil {
		return err
	}
	return srv.(LedgerServer).RangeQuery(m, &ledgerRangeQueryServer{stream})
}

type Ledger_RangeQueryServer interface {
	Send(*BathValue) error
	grpc.ServerStream
}

type ledgerRangeQueryServer struct {
	grpc.ServerStream
}

func (x *ledgerRangeQueryServer) Send(m *BathValue) error {
	return x.ServerStream.SendMsg(m)
}

func _Ledger_Post_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Event)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).Post(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/Post",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).Post(ctx, req.(*Event))
	}
	return interceptor(ctx, in, info, handler)
}

func _Ledger_Register_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(LedgerServer).Register(&ledgerRegisterServer{stream})
}

type Ledger_RegisterServer interface {
	Send(*Message) error
	Recv() (*Message, error)
	grpc.ServerStream
}

type ledgerRegisterServer struct {
	grpc.ServerStream
}

func (x *ledgerRegisterServer) Send(m *Message) error {
	return x.ServerStream.SendMsg(m)
}

func (x *ledgerRegisterServer) Recv() (*Message, error) {
	m := new(Message)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

var _Ledger_serviceDesc = grpc.ServiceDesc{
	ServiceName: "Ledger",
	HandlerType: (*LedgerServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "Get",
			Handler:    _Ledger_Get_Handler,
		},
		{
			MethodName: "Put",
			Handler:    _Ledger_Put_Handler,
		},
		{
			MethodName: "Delete",
			Handler:    _Ledger_Delete_Handler,
		},
		{
			MethodName: "BatchRead",
			Handler:    _Ledger_BatchRead_Handler,
		},
		{
			MethodName: "BatchWrite",
			Handler:    _Ledger_BatchWrite_Handler,
		},
		{
			MethodName: "Post",
			Handler:    _Ledger_Post_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "RangeQuery",
			Handler:       _Ledger_RangeQuery_Handler,
			ServerStreams: true,
		},
		{
			StreamName:    "Register",
			Handler:       _Ledger_Register_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
	},
	Metadata: "contract.proto",
}

func init() { proto.RegisterFile("contract.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 801 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x55, 0x4d, 0x6f, 0xdb, 0x46,
	0x10, 0x15, 0x3f, 0x44, 0x91, 0x23, 0x59, 0x21, 0x16, 0x69, 0xa1, 0x2a, 0x2d, 0xa0, 0x6e, 0x5b,
	0x54, 0xbd, 0x10, 0x81, 0x7a, 0xee, 0x41, 0x1f, 0x5b, 0xdb, 0x48, 0x42, 0x2b, 0x2b, 0xda, 0x41,
	0x0e, 0x85, 0xbb, 0xa6, 0x06, 0x92, 0x20, 0x99, 0x54, 0xc9, 0x95, 0x10, 0xfe, 0x8a, 0xf6, 0xb7,
	0xf4, 0xef, 0xf5, 0x52, 0xec, 0x92, 0x74, 0x6a, 0x37, 0x17, 0x1b, 0xb9, 0xcd, 0x9b, 0xe1, 0xbe,
	0x9d, 0x7d, 0xf3, 0x76, 0x09, 0xdd, 0x38, 0x4d, 0x64, 0x26, 0x62, 0x19, 0xec, 0xb3, 0x54, 0xa6,
	0xf4, 0x1f, 0x03, 0x5a, 0x6f, 0x30, 0xcf, 0xc5, 0x0a, 0xc9, 0xb7, 0x60, 0xcb, 0x62, 0x8f, 0x3d,
	0x63, 0x60, 0x0c, 0xbb, 0xa3, 0x93, 0xa0, 0xca, 0x07, 0x51, 0xb1, 0x47, 0xae, 0x4b, 0xa4, 0x07,
	0xad, 0xbd, 0x28, 0x76, 0xa9, 0x58, 0xf6, 0xcc, 0x81, 0x31, 0xec, 0xf0, 0x1a, 0xd2, 0xbf, 0x0d,
	0xb0, 0xd5, 0x87, 0xe4, 0x04, 0xbc, 0xcb, 0x70, 0xc6, 0x7e, 0x3d, 0x0f, 0xd9, 0xcc, 0x6f, 0x90,
	0x0e, 0xb8, 0x9c, 0x9d, 0x9e, 0x2f, 0x22, 0xc6, 0x7d, 0x83, 0x3c, 0x83, 0x76, 0xc4, 0xc7, 0xe1,
	0x62, 0x3c, 0x8d, 0xce, 0x2f, 0x42, 0xdf, 0x24, 0x2d, 0xb0, 0x4e, 0x59, 0xe4, 0x5b, 0x2a, 0x98,
	0x5f, 0x46, 0xbe, 0x4d, 0x00, 0x9c, 0x19, 0x7b, 0xcd, 0x22, 0xe6, 0x37, 0x49, 0x17, 0x60, 0x32,
	0x8e, 0xa6, 0x67, 0xd7, 0x9c, 0x8d, 0x67, 0xbe, 0xa3, 0x96, 0x97, 0xf8, 0x1d, 0x3f, 0x8f, 0x98,
	0xdf, 0x52, 0x09, 0x3e, 0x0e, 0x4f, 0xd9, 0xf5, 0xdb, 0x4b, 0xc6, 0xdf, 0xfb, 0xae, 0x5a, 0x31,
	0xbf, 0x58, 0x44, 0xd7, 0xec, 0x8a, 0x85, 0x91, 0xef, 0x95, 0xdb, 0x2f, 0xe6, 0x17, 0xe1, 0x82,
	0xf9, 0xa0, 0xaa, 0x67, 0x6c, 0xcc, 0xa3, 0xeb, 0x09, 0x1b, 0x47, 0x7e, 0x9b, 0xfe, 0x0e, 0x2d,
	0x8e, 0x7f, 0x1c, 0x30, 0x97, 0xe4, 0x27, 0x68, 0x29, 0x69, 0xf0, 0x83, 0xd4, 0xe7, 0x6f, 0x8f,
	0x9e, 0x05, 0x55, 0x69, 0x5a, 0xa6, 0x79, 0x5d, 0x27, 0x5f, 0x82, 0x73, 0x8b, 0x72, 0x9d, 0x96,
	0x1a, 0x78, 0xbc, 0x42, 0x84, 0x80, 0x2d, 0xb2, 0x55, 0xde, 0xb3, 0x06, 0xd6, 0xb0, 0xc3, 0x75,
	0x4c, 0x43, 0x70, 0x39, 0xe6, 0xfb, 0x34, 0xc9, 0x91, 0x74, 0xc1, 0x4c, 0xb7, 0x9a, 0xdd, 0xe5,
	0x66, 0xba, 0x55, 0x3c, 0x19, 0xe6, 0x87, 0x9d, 0xac, 0xb4, 0xac, 0x10, 0xe9, 0x83, 0x1b, 0xa7,
	0x4b, 0x3c, 0x13, 0xf9, 0xba, 0x67, 0xe9, 0x1d, 0xee, 0x30, 0xfd, 0xcb, 0x80, 0xee, 0xfd, 0xbe,
	0xd4, 0xb6, 0xf2, 0xc3, 0x66, 0xa9, 0x89, 0x3d, 0xae, 0x63, 0xf2, 0x35, 0x78, 0x89, 0xb8, 0xc5,
	0x7c, 0x2f, 0x62, 0xac, 0xba, 0xfc, 0x98, 0x20, 0x3e, 0x58, 0xf1, 0x66, 0x59, 0x71, 0xab, 0x50,
	0xcd, 0x75, 0x93, 0x1c, 0xd3, 0x2d, 0x66, 0x3d, 0x5b, 0x67, 0x6b, 0x48, 0x06, 0xd0, 0xbe, 0xd9,
	0xa5, 0xf1, 0x36, 0x3c, 0xdc, 0xde, 0x60, 0xd6, 0x6b, 0x0e, 0x8c, 0xa1, 0xcd, 0xff, 0x9b, 0xa2,
	0xbf, 0x41, 0x93, 0x1d, 0x31, 0x91, 0x64, 0xf8, 0x50, 0xc2, 0x6e, 0xf0, 0x1a, 0x97, 0x2b, 0xcc,
	0x3e, 0xa5, 0xa0, 0x4c, 0xf7, 0x9b, 0x38, 0xef, 0x99, 0x5a, 0xab, 0x0a, 0xa9, 0xa3, 0xdc, 0xa4,
	0xcb, 0x42, 0x77, 0xd6, 0xe1, 0x3a, 0xa6, 0xbf, 0x80, 0xf5, 0x0a, 0x8b, 0x47, 0x90, 0x77, 0xc0,
	0xd8, 0x56, 0x8a, 0x1a, 0x5b, 0x3a, 0x01, 0x77, 0x22, 0x64, 0xbc, 0x7e, 0x12, 0x87, 0x55, 0x72,
	0x4c, 0xc1, 0x9b, 0x08, 0xb9, 0xbe, 0x12, 0xbb, 0x83, 0x9e, 0xe2, 0x9d, 0xd8, 0x66, 0x29, 0xdd,
	0x5a, 0xe4, 0x6f, 0xd2, 0xac, 0x14, 0xda, 0xe5, 0x35, 0x54, 0x24, 0xc7, 0xca, 0x0c, 0xc6, 0x91,
	0x86, 0xd0, 0x2a, 0x1b, 0xb9, 0x7a, 0x44, 0x1f, 0x5f, 0x81, 0xb9, 0x3d, 0xea, 0x46, 0xda, 0x23,
	0x2f, 0x78, 0x85, 0x85, 0xee, 0x81, 0x9b, 0xdb, 0x23, 0x7d, 0x0f, 0x4d, 0x2e, 0x92, 0x15, 0x3e,
	0x82, 0xed, 0x39, 0x34, 0x73, 0x29, 0xb2, 0xda, 0x6f, 0x25, 0x50, 0x6e, 0xc0, 0x64, 0x59, 0x69,
	0xae, 0x42, 0xfa, 0x03, 0x34, 0x3f, 0x7d, 0x56, 0x7d, 0xa2, 0x4a, 0xda, 0x23, 0x9d, 0x83, 0x5b,
	0x77, 0xf4, 0xd4, 0xf1, 0xd4, 0x1a, 0x55, 0x8c, 0x07, 0x38, 0xb9, 0xb7, 0xea, 0xb3, 0x78, 0xfb,
	0x81, 0x83, 0xed, 0xff, 0x39, 0x78, 0x14, 0x81, 0x3b, 0xad, 0x9e, 0x45, 0x42, 0xc1, 0x3b, 0x43,
	0x91, 0xc9, 0x09, 0x0a, 0x49, 0xdc, 0xfa, 0x0d, 0xe8, 0x7b, 0x41, 0x7d, 0x8d, 0x69, 0x83, 0x7c,
	0xaf, 0x2e, 0xf5, 0x6a, 0x93, 0x4b, 0xcc, 0x88, 0x5b, 0x3f, 0x93, 0xfd, 0xbb, 0x88, 0x36, 0x86,
	0xc6, 0x4b, 0x63, 0xf4, 0xa7, 0x09, 0x4e, 0x79, 0x1a, 0xf2, 0x05, 0x58, 0xa7, 0x28, 0x89, 0xad,
	0x26, 0xd8, 0x77, 0x02, 0x2d, 0x19, 0x6d, 0x90, 0x6f, 0xc0, 0x9a, 0x1f, 0x24, 0xf9, 0x38, 0xd8,
	0xfb, 0xdb, 0xbc, 0x00, 0x67, 0x86, 0x3b, 0x94, 0x58, 0x2d, 0x7c, 0xd0, 0x83, 0xa7, 0xed, 0xc4,
	0x51, 0x2c, 0x89, 0x17, 0xd4, 0x1e, 0xef, 0x43, 0x70, 0x67, 0x55, 0xda, 0x20, 0xdf, 0x01, 0xe8,
	0xca, 0xbb, 0x6c, 0x23, 0x91, 0xb8, 0xd5, 0x67, 0x57, 0x0f, 0xa9, 0x40, 0x3b, 0xe9, 0xed, 0x01,
	0xb3, 0x82, 0x38, 0x81, 0x06, 0xf7, 0x89, 0x5e, 0x1a, 0xe4, 0x05, 0xd8, 0xf3, 0x34, 0x97, 0xc4,
	0x09, 0xf4, 0x6d, 0x7f, 0x82, 0x22, 0x93, 0x1f, 0xe1, 0x79, 0x9c, 0x04, 0xeb, 0x62, 0x8f, 0x59,
	0xbc, 0x16, 0x9b, 0xa4, 0xfc, 0x07, 0xe5, 0x93, 0x93, 0x5a, 0xfd, 0xb9, 0xc2, 0xf3, 0xc6, 0x8d,
	0xa3, 0x0b, 0x3f, 0xff, 0x1b, 0x00, 0x00, 0xff, 0xff, 0x8c, 0x7e, 0xef, 0x74, 0xae, 0x06, 0x00,
	0x00,
}
