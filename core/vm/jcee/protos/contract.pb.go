// Code generated by protoc-gen-go.
// source: contract.proto
// DO NOT EDIT!

/*
Package contract is a generated protocol buffer package.

It is generated from these files:
	contract.proto

It has these top-level messages:
	Request
	Response
	RequestContext
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

type Request struct {
	Context *RequestContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	Method  string          `protobuf:"bytes,2,opt,name=method" json:"method,omitempty"`
	Args    [][]byte        `protobuf:"bytes,3,rep,name=args,proto3" json:"args,omitempty"`
}

func (m *Request) Reset()                    { *m = Request{} }
func (m *Request) String() string            { return proto.CompactTextString(m) }
func (*Request) ProtoMessage()               {}
func (*Request) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

func (m *Request) GetContext() *RequestContext {
	if m != nil {
		return m.Context
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
func (*Response) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type RequestContext struct {
	Txid      string `protobuf:"bytes,1,opt,name=txid" json:"txid,omitempty"`
	Namespace string `protobuf:"bytes,2,opt,name=namespace" json:"namespace,omitempty"`
	Cid       string `protobuf:"bytes,3,opt,name=cid" json:"cid,omitempty"`
}

func (m *RequestContext) Reset()                    { *m = RequestContext{} }
func (m *RequestContext) String() string            { return proto.CompactTextString(m) }
func (*RequestContext) ProtoMessage()               {}
func (*RequestContext) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type Key struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	K       []byte         `protobuf:"bytes,2,opt,name=k,proto3" json:"k,omitempty"`
}

func (m *Key) Reset()                    { *m = Key{} }
func (m *Key) String() string            { return proto.CompactTextString(m) }
func (*Key) ProtoMessage()               {}
func (*Key) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

func (m *Key) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
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
func (*BatchKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

func (m *BatchKey) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
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
func (*BathValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type BatchKV struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	Kv      []*KeyValue    `protobuf:"bytes,2,rep,name=kv" json:"kv,omitempty"`
}

func (m *BatchKV) Reset()                    { *m = BatchKV{} }
func (m *BatchKV) String() string            { return proto.CompactTextString(m) }
func (*BatchKV) ProtoMessage()               {}
func (*BatchKV) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

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
func (*Range) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *Range) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
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
func (*Value) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

type KeyValue struct {
	Context *LedgerContext `protobuf:"bytes,1,opt,name=context" json:"context,omitempty"`
	K       []byte         `protobuf:"bytes,2,opt,name=k,proto3" json:"k,omitempty"`
	V       []byte         `protobuf:"bytes,3,opt,name=v,proto3" json:"v,omitempty"`
}

func (m *KeyValue) Reset()                    { *m = KeyValue{} }
func (m *KeyValue) String() string            { return proto.CompactTextString(m) }
func (*KeyValue) ProtoMessage()               {}
func (*KeyValue) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *KeyValue) GetContext() *LedgerContext {
	if m != nil {
		return m.Context
	}
	return nil
}

type LedgerContext struct {
	Txid      string `protobuf:"bytes,1,opt,name=txid" json:"txid,omitempty"`
	Namespace string `protobuf:"bytes,2,opt,name=namespace" json:"namespace,omitempty"`
	Cid       string `protobuf:"bytes,3,opt,name=cid" json:"cid,omitempty"`
}

func (m *LedgerContext) Reset()                    { *m = LedgerContext{} }
func (m *LedgerContext) String() string            { return proto.CompactTextString(m) }
func (*LedgerContext) ProtoMessage()               {}
func (*LedgerContext) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func init() {
	proto.RegisterType((*Request)(nil), "Request")
	proto.RegisterType((*Response)(nil), "Response")
	proto.RegisterType((*RequestContext)(nil), "RequestContext")
	proto.RegisterType((*Key)(nil), "Key")
	proto.RegisterType((*BatchKey)(nil), "BatchKey")
	proto.RegisterType((*BathValue)(nil), "BathValue")
	proto.RegisterType((*BatchKV)(nil), "BatchKV")
	proto.RegisterType((*Range)(nil), "Range")
	proto.RegisterType((*Value)(nil), "Value")
	proto.RegisterType((*KeyValue)(nil), "KeyValue")
	proto.RegisterType((*LedgerContext)(nil), "LedgerContext")
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.

// Client API for Contract service

type ContractClient interface {
	// used tp execute contract method remotely
	Execute(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
	// used to detect the health state of contract
	HeartBeat(ctx context.Context, in *Request, opts ...grpc.CallOption) (*Response, error)
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

// Server API for Contract service

type ContractServer interface {
	// used tp execute contract method remotely
	Execute(context.Context, *Request) (*Response, error)
	// used to detect the health state of contract
	HeartBeat(context.Context, *Request) (*Response, error)
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
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for Ledger service

type LedgerClient interface {
	Get(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Value, error)
	Put(ctx context.Context, in *KeyValue, opts ...grpc.CallOption) (*Response, error)
	Delete(ctx context.Context, in *Key, opts ...grpc.CallOption) (*Response, error)
	BatchRead(ctx context.Context, in *BatchKey, opts ...grpc.CallOption) (*BathValue, error)
	BatchWrite(ctx context.Context, in *BatchKV, opts ...grpc.CallOption) (*Response, error)
	RangeQuery(ctx context.Context, in *Range, opts ...grpc.CallOption) (*BathValue, error)
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

func (c *ledgerClient) RangeQuery(ctx context.Context, in *Range, opts ...grpc.CallOption) (*BathValue, error) {
	out := new(BathValue)
	err := grpc.Invoke(ctx, "/Ledger/RangeQuery", in, out, c.cc, opts...)
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
	RangeQuery(context.Context, *Range) (*BathValue, error)
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

func _Ledger_RangeQuery_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Range)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(LedgerServer).RangeQuery(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/Ledger/RangeQuery",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(LedgerServer).RangeQuery(ctx, req.(*Range))
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
			MethodName: "RangeQuery",
			Handler:    _Ledger_RangeQuery_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("contract.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 525 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xac, 0x54, 0xc1, 0x6e, 0xd3, 0x40,
	0x10, 0x8d, 0xed, 0xc6, 0xb1, 0xa7, 0x69, 0x40, 0xa3, 0x82, 0x42, 0x00, 0x29, 0x5a, 0x40, 0x84,
	0x8b, 0x0f, 0xe1, 0xcc, 0xc5, 0x01, 0x51, 0xa9, 0x50, 0x85, 0x05, 0x15, 0x71, 0x63, 0xb1, 0x47,
	0x71, 0xe4, 0xd4, 0x0e, 0xbb, 0xeb, 0x28, 0xf9, 0x45, 0xbe, 0x0a, 0x79, 0xbd, 0x4e, 0x95, 0x02,
	0x87, 0x56, 0xdc, 0xe6, 0xed, 0x78, 0xdf, 0xbe, 0x79, 0x33, 0x63, 0x18, 0x24, 0x65, 0xa1, 0xa5,
	0x48, 0x74, 0xb4, 0x96, 0xa5, 0x2e, 0xd9, 0x77, 0xe8, 0x71, 0xfa, 0x59, 0x91, 0xd2, 0xf8, 0x0a,
	0x7a, 0x75, 0x92, 0xb6, 0x7a, 0xe8, 0x8c, 0x9d, 0xc9, 0xf1, 0xf4, 0x5e, 0x64, 0x53, 0xb3, 0xe6,
	0x98, 0xb7, 0x79, 0x7c, 0x08, 0xfe, 0x15, 0xe9, 0xac, 0x4c, 0x87, 0xee, 0xd8, 0x99, 0x84, 0xdc,
	0x22, 0x44, 0x38, 0x12, 0x72, 0xa1, 0x86, 0xde, 0xd8, 0x9b, 0xf4, 0xb9, 0x89, 0xd9, 0x05, 0x04,
	0x9c, 0xd4, 0xba, 0x2c, 0x14, 0xe1, 0x00, 0xdc, 0x32, 0x37, 0xec, 0x01, 0x77, 0xcb, 0xbc, 0xe6,
	0x91, 0xa4, 0xaa, 0x95, 0x36, 0x3c, 0x7d, 0x6e, 0x11, 0x8e, 0x20, 0x48, 0xca, 0x94, 0xce, 0x84,
	0xca, 0x86, 0x9e, 0x79, 0x61, 0x8f, 0xd9, 0x17, 0x18, 0x1c, 0xca, 0xaa, 0x5f, 0xd5, 0xdb, 0x65,
	0x6a, 0x78, 0x43, 0x6e, 0x62, 0x7c, 0x02, 0x61, 0x21, 0xae, 0x48, 0xad, 0x45, 0x42, 0x56, 0xe4,
	0xf5, 0x01, 0xde, 0x07, 0x2f, 0x59, 0xa6, 0x96, 0xba, 0x0e, 0xd9, 0x1b, 0xf0, 0xce, 0x69, 0x87,
	0x93, 0x9b, 0x1e, 0x0c, 0xa2, 0x0f, 0x94, 0x2e, 0x48, 0xfe, 0x61, 0x41, 0x1f, 0x9c, 0xdc, 0xaa,
	0x76, 0x72, 0x16, 0x43, 0x10, 0x0b, 0x9d, 0x64, 0x77, 0xe2, 0xf0, 0x1a, 0x8e, 0x19, 0x84, 0xb1,
	0xd0, 0xd9, 0xa5, 0x58, 0x55, 0xc6, 0xa9, 0x7d, 0x45, 0xee, 0x32, 0xc5, 0x21, 0xf4, 0x32, 0xa1,
	0x3e, 0x96, 0xb2, 0xa9, 0x26, 0xe0, 0x2d, 0xac, 0x49, 0x36, 0xd6, 0x70, 0x67, 0xc3, 0x2e, 0xa0,
	0xd7, 0x08, 0xb9, 0xbc, 0x85, 0x8e, 0x47, 0xe0, 0xe6, 0x1b, 0x23, 0xe4, 0x78, 0x1a, 0x46, 0xe7,
	0xb4, 0x33, 0x1a, 0xb8, 0x9b, 0x6f, 0xd8, 0x37, 0xe8, 0x72, 0x51, 0x2c, 0xe8, 0x16, 0x6c, 0xa7,
	0xd0, 0x55, 0x5a, 0xc8, 0xb6, 0xa7, 0x0d, 0xa8, 0x2d, 0xa7, 0xa2, 0xb1, 0xbc, 0xcf, 0xeb, 0x90,
	0xbd, 0x80, 0xee, 0xdf, 0x6b, 0x35, 0x15, 0x59, 0x6b, 0x37, 0x6c, 0x0e, 0x41, 0xab, 0xe8, 0xae,
	0xed, 0x69, 0x3d, 0xb2, 0x8c, 0x9f, 0xe1, 0xe4, 0xe0, 0xd6, 0xff, 0x18, 0xa0, 0xe9, 0x1c, 0x82,
	0x99, 0x5d, 0x2d, 0x1c, 0x43, 0xef, 0xdd, 0x96, 0x92, 0x4a, 0x13, 0x06, 0xed, 0x0e, 0x8d, 0xc2,
	0xa8, 0x5d, 0x03, 0xd6, 0x41, 0x06, 0xe1, 0x19, 0x09, 0xa9, 0x63, 0x12, 0xfa, 0x1f, 0xdf, 0x4c,
	0x7f, 0x39, 0xe0, 0x37, 0x3a, 0xf1, 0x01, 0x78, 0xef, 0x49, 0xe3, 0x51, 0xdd, 0x9b, 0x91, 0x1f,
	0x19, 0x33, 0x58, 0x07, 0x9f, 0x82, 0x37, 0xaf, 0x34, 0x5e, 0xb7, 0xec, 0xf0, 0x91, 0xc7, 0xe0,
	0xbf, 0xa5, 0x15, 0x69, 0xb2, 0x17, 0x0f, 0x92, 0xcf, 0xcd, 0xb4, 0x25, 0x19, 0x27, 0x91, 0x62,
	0x18, 0xb5, 0xd3, 0x3b, 0x82, 0x68, 0x3f, 0x84, 0xac, 0x83, 0xcf, 0x00, 0x4c, 0xe6, 0xab, 0x5c,
	0x9a, 0x62, 0xec, 0x6c, 0xdd, 0x2c, 0x06, 0xcc, 0x8c, 0x7c, 0xaa, 0x48, 0xee, 0xd0, 0x8f, 0x0c,
	0x38, 0x24, 0x8a, 0x5f, 0xc2, 0x69, 0x52, 0x44, 0xd9, 0x6e, 0x4d, 0x32, 0xc9, 0xc4, 0xb2, 0x68,
	0x7e, 0x3f, 0x2a, 0x3e, 0x69, 0x4d, 0x9b, 0xd7, 0x78, 0xde, 0xf9, 0xe1, 0x9b, 0xc4, 0xeb, 0xdf,
	0x01, 0x00, 0x00, 0xff, 0xff, 0x7e, 0xaa, 0x29, 0x9d, 0xa9, 0x04, 0x00, 0x00,
}
