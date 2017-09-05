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
	// used tp execute contract method remotely
	Execute(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// used to detect the health state of contract
	HeartBeat(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// used to transfer data between hyperchain and contract, just for test now, will be discard later
	StreamExecute(ctx context.Context, opts ...grpc.CallOption) (Contract_StreamExecuteClient, error)
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

func (c *contractClient) Execute(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Contract/Execute", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractClient) HeartBeat(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error) {
	out := new(Response)
	err := grpc.Invoke(ctx, "/Contract/HeartBeat", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *contractClient) StreamExecute(ctx context.Context, opts ...grpc.CallOption) (Contract_StreamExecuteClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Contract_serviceDesc.Streams[0], c.cc, "/Contract/StreamExecute", opts...)
	if err != nil {
		return nil, err
	}
	x := &contractStreamExecuteClient{stream}
	return x, nil
}

type Contract_StreamExecuteClient interface {
	Send(*Request) error
	Recv() (*Response, error)
	grpc.ClientStream
}

type contractStreamExecuteClient struct {
	grpc.ClientStream
}

func (x *contractStreamExecuteClient) Send(m *Request) error {
	return x.ClientStream.SendMsg(m)
}

func (x *contractStreamExecuteClient) Recv() (*Response, error) {
	m := new(Response)
	if err := x.ClientStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
}

func (c *contractClient) Register(ctx context.Context, opts ...grpc.CallOption) (Contract_RegisterClient, error) {
	stream, err := grpc.NewClientStream(ctx, &_Contract_serviceDesc.Streams[1], c.cc, "/Contract/Register", opts...)
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
	// used tp execute contract method remotely
	Execute(context.Context, *Request) (*Response, error)
	// used to detect the health state of contract
	HeartBeat(context.Context, *Request) (*Response, error)
	// used to transfer data between hyperchain and contract, just for test now, will be discard later
	StreamExecute(Contract_StreamExecuteServer) error
	// Interface that provides support to jvm execution which will estabilish a stream between Server
	// and Client. Message type provide the context necessary for server to respond appropriately.
	Register(Contract_RegisterServer) error
}

func RegisterContractServer(s *grpc.Server, srv ContractServer) {
	s.RegisterService(&_Contract_serviceDesc, srv)
}

func _Contract_Execute_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Request)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ContractServer).Execute(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Contract/Execute",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ContractServer).Execute(ctx, req.(*Request))
	}
	return interceptor(ctx, in, info, handler)
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

func _Contract_StreamExecute_Handler(srv interface{}, stream grpc.ServerStream) error {
	return srv.(ContractServer).StreamExecute(&contractStreamExecuteServer{stream})
}

type Contract_StreamExecuteServer interface {
	Send(*Response) error
	Recv() (*Request, error)
	grpc.ServerStream
}

type contractStreamExecuteServer struct {
	grpc.ServerStream
}

func (x *contractStreamExecuteServer) Send(m *Response) error {
	return x.ServerStream.SendMsg(m)
}

func (x *contractStreamExecuteServer) Recv() (*Request, error) {
	m := new(Request)
	if err := x.ServerStream.RecvMsg(m); err != nil {
		return nil, err
	}
	return m, nil
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
			MethodName: "Execute",
			Handler:    _Contract_Execute_Handler,
		},
		{
			MethodName: "HeartBeat",
			Handler:    _Contract_HeartBeat_Handler,
		},
	},
	Streams: []grpc.StreamDesc{
		{
			StreamName:    "StreamExecute",
			Handler:       _Contract_StreamExecute_Handler,
			ServerStreams: true,
			ClientStreams: true,
		},
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

// Server API for Ledger service

type LedgerServer interface {
	Get(context.Context, *Key) (*Value, error)
	Put(context.Context, *KeyValue) (*Response, error)
	Delete(context.Context, *Key) (*Response, error)
	BatchRead(context.Context, *BatchKey) (*BathValue, error)
	BatchWrite(context.Context, *BatchKV) (*Response, error)
	RangeQuery(*Range, Ledger_RangeQueryServer) error
	Post(context.Context, *Event) (*Response, error)
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
	},
	Metadata: "contract.proto",
}

func init() { proto.RegisterFile("contract.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 828 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x55, 0xdd, 0x8e, 0xdb, 0x44,
	0x14, 0x8e, 0x63, 0xc7, 0xb1, 0x4f, 0x7e, 0x6a, 0x8d, 0x0a, 0x0a, 0x29, 0x48, 0xc1, 0x14, 0x11,
	0xb8, 0xb0, 0xaa, 0x70, 0xcd, 0x45, 0x7e, 0x86, 0xdd, 0x55, 0x5b, 0x6f, 0x3a, 0xf1, 0x6e, 0xd5,
	0x0b, 0xb4, 0xcc, 0x3a, 0x47, 0x49, 0x94, 0xac, 0x1d, 0xec, 0x49, 0xb4, 0x79, 0x0b, 0x9e, 0x81,
	0x47, 0xe0, 0xad, 0x78, 0x06, 0x6e, 0xd0, 0x8c, 0xc7, 0x29, 0x59, 0x0a, 0xd2, 0x56, 0xdc, 0x9d,
	0xef, 0x9c, 0x99, 0x6f, 0xce, 0x7c, 0xf3, 0xf9, 0x18, 0xda, 0x71, 0x9a, 0x88, 0x8c, 0xc7, 0x22,
	0xd8, 0x66, 0xa9, 0x48, 0xfd, 0x3f, 0x0d, 0xa8, 0xbf, 0xc6, 0x3c, 0xe7, 0x0b, 0x24, 0x5f, 0x82,
	0x25, 0x0e, 0x5b, 0xec, 0x18, 0x3d, 0xa3, 0xdf, 0x1e, 0xb4, 0x02, 0x9d, 0x0f, 0xa2, 0xc3, 0x16,
	0x99, 0x2a, 0x91, 0x0e, 0xd4, 0xb7, 0xfc, 0xb0, 0x49, 0xf9, 0xbc, 0x53, 0xed, 0x19, 0xfd, 0x26,
	0x2b, 0xa1, 0xff, 0xbb, 0x01, 0x96, 0x5c, 0x48, 0x5a, 0xe0, 0x5e, 0x85, 0x13, 0xfa, 0xe3, 0x45,
	0x48, 0x27, 0x5e, 0x85, 0x34, 0xc1, 0x61, 0xf4, 0xec, 0x62, 0x16, 0x51, 0xe6, 0x19, 0xe4, 0x09,
	0x34, 0x22, 0x36, 0x0c, 0x67, 0xc3, 0x71, 0x74, 0x71, 0x19, 0x7a, 0x55, 0x52, 0x07, 0xf3, 0x8c,
	0x46, 0x9e, 0x29, 0x83, 0xe9, 0x55, 0xe4, 0x59, 0x04, 0xc0, 0x9e, 0xd0, 0x57, 0x34, 0xa2, 0x5e,
	0x8d, 0xb4, 0x01, 0x46, 0xc3, 0x68, 0x7c, 0x7e, 0xc3, 0xe8, 0x70, 0xe2, 0xd9, 0x72, 0x7b, 0x81,
	0xdf, 0xb2, 0x8b, 0x88, 0x7a, 0x75, 0x99, 0x60, 0xc3, 0xf0, 0x8c, 0xde, 0xbc, 0xb9, 0xa2, 0xec,
	0x9d, 0xe7, 0xc8, 0x1d, 0xd3, 0xcb, 0x59, 0x74, 0x43, 0xaf, 0x69, 0x18, 0x79, 0x6e, 0x71, 0xfc,
	0x6c, 0x7a, 0x19, 0xce, 0xa8, 0x07, 0xb2, 0x7a, 0x4e, 0x87, 0x2c, 0xba, 0x19, 0xd1, 0x61, 0xe4,
	0x35, 0xfc, 0x9f, 0xa1, 0xce, 0xf0, 0x97, 0x1d, 0xe6, 0x82, 0x7c, 0x0b, 0x75, 0x29, 0x0d, 0xde,
	0x0b, 0x75, 0xff, 0xc6, 0xe0, 0x49, 0xa0, 0x4b, 0xe3, 0x22, 0xcd, 0xca, 0x3a, 0xf9, 0x14, 0xec,
	0x3b, 0x14, 0xcb, 0xb4, 0xd0, 0xc0, 0x65, 0x1a, 0x11, 0x02, 0x16, 0xcf, 0x16, 0x79, 0xc7, 0xec,
	0x99, 0xfd, 0x26, 0x53, 0xb1, 0x1f, 0x82, 0xc3, 0x30, 0xdf, 0xa6, 0x49, 0x8e, 0xa4, 0x0d, 0xd5,
	0x74, 0xad, 0xd8, 0x1d, 0x56, 0x4d, 0xd7, 0x92, 0x27, 0xc3, 0x7c, 0xb7, 0x11, 0x5a, 0x4b, 0x8d,
	0x48, 0x17, 0x9c, 0x38, 0x9d, 0xe3, 0x39, 0xcf, 0x97, 0x1d, 0x53, 0x9d, 0x70, 0xc4, 0xfe, 0xaf,
	0x06, 0xb4, 0x4f, 0xfb, 0x92, 0xc7, 0x8a, 0xfb, 0xd5, 0x5c, 0x11, 0xbb, 0x4c, 0xc5, 0xe4, 0x73,
	0x70, 0x13, 0x7e, 0x87, 0xf9, 0x96, 0xc7, 0xa8, 0xbb, 0x7c, 0x9f, 0x20, 0x1e, 0x98, 0xf1, 0x6a,
	0xae, 0xb9, 0x65, 0x28, 0xdf, 0x75, 0x95, 0xec, 0xd3, 0x35, 0x66, 0x1d, 0x4b, 0x65, 0x4b, 0x48,
	0x7a, 0xd0, 0xb8, 0xdd, 0xa4, 0xf1, 0x3a, 0xdc, 0xdd, 0xdd, 0x62, 0xd6, 0xa9, 0xf5, 0x8c, 0xbe,
	0xc5, 0xfe, 0x9e, 0xf2, 0x7f, 0x82, 0x1a, 0xdd, 0x63, 0x22, 0x48, 0xff, 0xa1, 0x84, 0xed, 0xe0,
	0x15, 0xce, 0x17, 0x98, 0x7d, 0x48, 0x41, 0x91, 0x6e, 0x57, 0x71, 0xde, 0xa9, 0x2a, 0xad, 0x34,
	0x92, 0x57, 0xb9, 0x4d, 0xe7, 0x07, 0xd5, 0x59, 0x93, 0xa9, 0xd8, 0xff, 0x01, 0xcc, 0x97, 0x78,
	0x78, 0x04, 0x79, 0x13, 0x8c, 0xb5, 0x56, 0xd4, 0x58, 0xfb, 0x23, 0x70, 0x46, 0x5c, 0xc4, 0xcb,
	0x8f, 0xe2, 0x30, 0x0b, 0x8e, 0x31, 0xb8, 0x23, 0x2e, 0x96, 0xd7, 0x7c, 0xb3, 0x53, 0xaf, 0x78,
	0x14, 0xbb, 0x5a, 0x48, 0xb7, 0xe4, 0xf9, 0xeb, 0x34, 0x2b, 0x84, 0x76, 0x58, 0x09, 0x25, 0xc9,
	0x5e, 0x9b, 0xc1, 0xd8, 0xfb, 0x21, 0xd4, 0x8b, 0x46, 0xae, 0x1f, 0xd1, 0xc7, 0x67, 0x50, 0x5d,
	0xef, 0x55, 0x23, 0x8d, 0x81, 0x1b, 0xbc, 0xc4, 0x83, 0xea, 0x81, 0x55, 0xd7, 0x7b, 0xff, 0x1d,
	0xd4, 0x18, 0x4f, 0x16, 0xf8, 0x08, 0xb6, 0xa7, 0x50, 0xcb, 0x05, 0xcf, 0x4a, 0xbf, 0x15, 0x40,
	0xba, 0x01, 0x93, 0xb9, 0xd6, 0x5c, 0x86, 0xfe, 0xd7, 0x50, 0xfb, 0xf0, 0x5d, 0xd5, 0x8d, 0xb4,
	0xb4, 0x7b, 0x7f, 0x0a, 0x4e, 0xd9, 0xd1, 0xc7, 0x3e, 0x4f, 0xa9, 0x91, 0x66, 0xdc, 0x41, 0xeb,
	0x64, 0xd7, 0xff, 0xe2, 0xed, 0x07, 0x0e, 0xb6, 0xfe, 0xe1, 0xe0, 0xc1, 0x6f, 0x06, 0x38, 0x63,
	0x3d, 0x17, 0x49, 0x0f, 0xea, 0xf4, 0x1e, 0xe3, 0x9d, 0x40, 0xe2, 0x94, 0x23, 0xa0, 0xeb, 0x06,
	0xe5, 0x57, 0xec, 0x57, 0x88, 0x0f, 0xee, 0x39, 0xf2, 0x4c, 0x8c, 0x90, 0x8b, 0x7f, 0x5b, 0xf3,
	0x1d, 0xb4, 0x66, 0x22, 0x43, 0x7e, 0xf7, 0xdf, 0x5c, 0x7d, 0xe3, 0x85, 0x41, 0x9e, 0xcb, 0x19,
	0xb1, 0x58, 0xe5, 0x02, 0x33, 0xe2, 0x94, 0x53, 0xb7, 0x7b, 0x8c, 0x8a, 0x55, 0x83, 0x3f, 0x0c,
	0xb0, 0x0b, 0x71, 0xc8, 0x27, 0x60, 0x9e, 0xa1, 0x20, 0x96, 0x34, 0x44, 0xd7, 0x0e, 0xd4, 0x0b,
	0xf8, 0x15, 0xf2, 0x05, 0x98, 0xd3, 0x9d, 0x20, 0xef, 0x7d, 0x72, 0xda, 0xd2, 0x33, 0xb0, 0x27,
	0xb8, 0x41, 0x81, 0x7a, 0xe3, 0x49, 0xf1, 0xb9, 0xb2, 0x78, 0xbc, 0x64, 0xc8, 0xe7, 0xc4, 0x0d,
	0xca, 0x4f, 0xa6, 0x0b, 0xc1, 0xd1, 0xf9, 0x7e, 0x85, 0x7c, 0x05, 0xa0, 0x2a, 0x6f, 0xb3, 0x95,
	0xba, 0x92, 0x36, 0xf4, 0x43, 0x2a, 0x50, 0xc6, 0x7c, 0xb3, 0xc3, 0xec, 0x40, 0xec, 0x40, 0x81,
	0x53, 0xa2, 0x17, 0x06, 0x79, 0x06, 0xd6, 0x34, 0xcd, 0x05, 0xb1, 0x03, 0x35, 0x3c, 0x4e, 0x28,
	0x46, 0xdf, 0xc0, 0xd3, 0x38, 0x09, 0x96, 0x87, 0x2d, 0x66, 0xf1, 0x92, 0xaf, 0x92, 0xe2, 0x67,
	0x95, 0x8f, 0x5a, 0xe5, 0x2b, 0x4d, 0x25, 0x9e, 0x56, 0x6e, 0x6d, 0x55, 0xf8, 0xfe, 0xaf, 0x00,
	0x00, 0x00, 0xff, 0xff, 0x3b, 0x7f, 0x1e, 0x2f, 0xd7, 0x06, 0x00, 0x00,
}
