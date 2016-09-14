// Code generated by protoc-gen-go.
// source: chain.proto
// DO NOT EDIT!

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Chain struct {
	LatestBlockHash  []byte `protobuf:"bytes,1,opt,name=latestBlockHash,proto3" json:"latestBlockHash,omitempty"`
	ParentBlockHash  []byte `protobuf:"bytes,2,opt,name=parentBlockHash,proto3" json:"parentBlockHash,omitempty"`
	Height           uint64 `protobuf:"varint,3,opt,name=height" json:"height,omitempty"`
	RequiredBlockNum uint64 `protobuf:"varint,4,opt,name=requiredBlockNum" json:"requiredBlockNum,omitempty"`
	RequireBlockHash []byte `protobuf:"bytes,5,opt,name=requireBlockHash,proto3" json:"requireBlockHash,omitempty"`
	RecoveryNum      uint64 `protobuf:"varint,6,opt,name=recoveryNum" json:"recoveryNum,omitempty"`
}

func (m *Chain) Reset()                    { *m = Chain{} }
func (m *Chain) String() string            { return proto.CompactTextString(m) }
func (*Chain) ProtoMessage()               {}
func (*Chain) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func init() {
	proto.RegisterType((*Chain)(nil), "types.Chain")
}

func init() { proto.RegisterFile("chain.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 169 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0xe2, 0x4e, 0xce, 0x48, 0xcc,
	0xcc, 0xd3, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17, 0x62, 0x2d, 0xa9, 0x2c, 0x48, 0x2d, 0x56, 0x7a,
	0xcb, 0xc8, 0xc5, 0xea, 0x0c, 0x12, 0x16, 0xd2, 0xe0, 0xe2, 0xcf, 0x49, 0x2c, 0x49, 0x2d, 0x2e,
	0x71, 0xca, 0xc9, 0x4f, 0xce, 0xf6, 0x48, 0x2c, 0xce, 0x90, 0x60, 0x54, 0x60, 0xd4, 0xe0, 0x09,
	0x42, 0x17, 0x06, 0xa9, 0x2c, 0x48, 0x2c, 0x4a, 0xcd, 0x43, 0x52, 0xc9, 0x04, 0x51, 0x89, 0x26,
	0x2c, 0x24, 0xc6, 0xc5, 0x96, 0x91, 0x9a, 0x99, 0x9e, 0x51, 0x22, 0xc1, 0x0c, 0x54, 0xc0, 0x12,
	0x04, 0xe5, 0x09, 0x69, 0x71, 0x09, 0x14, 0xa5, 0x16, 0x96, 0x66, 0x16, 0xa5, 0xa6, 0x80, 0x15,
	0xfb, 0x95, 0xe6, 0x4a, 0xb0, 0x80, 0x55, 0x60, 0x88, 0x23, 0xa9, 0x45, 0x58, 0xc7, 0x0a, 0xb6,
	0x0e, 0x43, 0x5c, 0x48, 0x81, 0x8b, 0xbb, 0x28, 0x35, 0x39, 0xbf, 0x2c, 0xb5, 0xa8, 0x12, 0x64,
	0x24, 0x1b, 0xd8, 0x48, 0x64, 0xa1, 0x24, 0x36, 0xb0, 0xef, 0x8d, 0x01, 0x01, 0x00, 0x00, 0xff,
	0xff, 0x2e, 0xd0, 0x6b, 0x67, 0x0c, 0x01, 0x00, 0x00,
}
