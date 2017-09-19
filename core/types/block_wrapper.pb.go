// Code generated by protoc-gen-go.
// source: block_wrapper.proto
// DO NOT EDIT!

package types

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type BlockWrapper struct {
	BlockVersion []byte `protobuf:"bytes,1,opt,name=blockVersion,proto3" json:"blockVersion,omitempty"`
	Block        []byte `protobuf:"bytes,2,opt,name=block,proto3" json:"block,omitempty"`
}

func (m *BlockWrapper) Reset()                    { *m = BlockWrapper{} }
func (m *BlockWrapper) String() string            { return proto.CompactTextString(m) }
func (*BlockWrapper) ProtoMessage()               {}
func (*BlockWrapper) Descriptor() ([]byte, []int) { return fileDescriptor1, []int{0} }

func (m *BlockWrapper) GetBlockVersion() []byte {
	if m != nil {
		return m.BlockVersion
	}
	return nil
}

func (m *BlockWrapper) GetBlock() []byte {
	if m != nil {
		return m.Block
	}
	return nil
}

func init() {
	proto.RegisterType((*BlockWrapper)(nil), "types.BlockWrapper")
}

func init() { proto.RegisterFile("block_wrapper.proto", fileDescriptor1) }

var fileDescriptor1 = []byte{
	// 100 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xe2, 0x12, 0x4e, 0xca, 0xc9, 0x4f,
	0xce, 0x8e, 0x2f, 0x2f, 0x4a, 0x2c, 0x28, 0x48, 0x2d, 0xd2, 0x2b, 0x28, 0xca, 0x2f, 0xc9, 0x17,
	0x62, 0x2d, 0xa9, 0x2c, 0x48, 0x2d, 0x56, 0xf2, 0xe0, 0xe2, 0x71, 0x02, 0xc9, 0x86, 0x43, 0x24,
	0x85, 0x94, 0xb8, 0x78, 0xc0, 0xaa, 0xc3, 0x52, 0x8b, 0x8a, 0x33, 0xf3, 0xf3, 0x24, 0x18, 0x15,
	0x18, 0x35, 0x78, 0x82, 0x50, 0xc4, 0x84, 0x44, 0xb8, 0x58, 0xc1, 0x7c, 0x09, 0x26, 0xb0, 0x24,
	0x84, 0x93, 0xc4, 0x06, 0x36, 0xd7, 0x18, 0x10, 0x00, 0x00, 0xff, 0xff, 0x35, 0xac, 0x91, 0xd9,
	0x6e, 0x00, 0x00, 0x00,
}