// Code generated by protoc-gen-go.
// source: receipt.proto
// DO NOT EDIT!

package version1_2

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type Receipt_STATUS int32

const (
	Receipt_SUCCESS Receipt_STATUS = 0
	Receipt_FAILED  Receipt_STATUS = 1
)

var Receipt_STATUS_name = map[int32]string{
	0: "SUCCESS",
	1: "FAILED",
}
var Receipt_STATUS_value = map[string]int32{
	"SUCCESS": 0,
	"FAILED":  1,
}

func (x Receipt_STATUS) String() string {
	return proto.EnumName(Receipt_STATUS_name, int32(x))
}
func (Receipt_STATUS) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{0, 0} }

type Receipt struct {
	Version           []byte         `protobuf:"bytes,1,opt,name=version,proto3" json:"version,omitempty"`
	PostState         []byte         `protobuf:"bytes,2,opt,name=PostState,proto3" json:"PostState,omitempty"`
	CumulativeGasUsed int64          `protobuf:"varint,3,opt,name=CumulativeGasUsed" json:"CumulativeGasUsed,omitempty"`
	TxHash            []byte         `protobuf:"bytes,4,opt,name=TxHash,proto3" json:"TxHash,omitempty"`
	ContractAddress   []byte         `protobuf:"bytes,5,opt,name=ContractAddress,proto3" json:"ContractAddress,omitempty"`
	GasUsed           int64          `protobuf:"varint,6,opt,name=GasUsed" json:"GasUsed,omitempty"`
	Ret               []byte         `protobuf:"bytes,7,opt,name=Ret,proto3" json:"Ret,omitempty"`
	Logs              []byte         `protobuf:"bytes,8,opt,name=Logs,proto3" json:"Logs,omitempty"`
	Status            Receipt_STATUS `protobuf:"varint,9,opt,name=Status,enum=version1_2.Receipt_STATUS" json:"Status,omitempty"`
	Message           []byte         `protobuf:"bytes,10,opt,name=Message,proto3" json:"Message,omitempty"`
}

func (m *Receipt) Reset()                    { *m = Receipt{} }
func (m *Receipt) String() string            { return proto.CompactTextString(m) }
func (*Receipt) ProtoMessage()               {}
func (*Receipt) Descriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func (m *Receipt) GetVersion() []byte {
	if m != nil {
		return m.Version
	}
	return nil
}

func (m *Receipt) GetPostState() []byte {
	if m != nil {
		return m.PostState
	}
	return nil
}

func (m *Receipt) GetCumulativeGasUsed() int64 {
	if m != nil {
		return m.CumulativeGasUsed
	}
	return 0
}

func (m *Receipt) GetTxHash() []byte {
	if m != nil {
		return m.TxHash
	}
	return nil
}

func (m *Receipt) GetContractAddress() []byte {
	if m != nil {
		return m.ContractAddress
	}
	return nil
}

func (m *Receipt) GetGasUsed() int64 {
	if m != nil {
		return m.GasUsed
	}
	return 0
}

func (m *Receipt) GetRet() []byte {
	if m != nil {
		return m.Ret
	}
	return nil
}

func (m *Receipt) GetLogs() []byte {
	if m != nil {
		return m.Logs
	}
	return nil
}

func (m *Receipt) GetStatus() Receipt_STATUS {
	if m != nil {
		return m.Status
	}
	return Receipt_SUCCESS
}

func (m *Receipt) GetMessage() []byte {
	if m != nil {
		return m.Message
	}
	return nil
}

func init() {
	proto.RegisterType((*Receipt)(nil), "version1_2.Receipt")
	proto.RegisterEnum("version1_2.Receipt_STATUS", Receipt_STATUS_name, Receipt_STATUS_value)
}

func init() { proto.RegisterFile("receipt.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 272 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x64, 0x50, 0xcd, 0x4a, 0xf3, 0x40,
	0x14, 0xfd, 0xd2, 0xf4, 0x9b, 0xd8, 0xeb, 0x5f, 0xbc, 0x0b, 0x19, 0xc4, 0x45, 0xec, 0x2a, 0x0b,
	0x09, 0x18, 0x9f, 0x20, 0xc4, 0xfa, 0x03, 0x15, 0x24, 0x93, 0xac, 0x65, 0x6c, 0x86, 0x1a, 0xd0,
	0x4e, 0xc9, 0x9d, 0x14, 0x9f, 0xd7, 0x27, 0x91, 0x4c, 0x26, 0x14, 0x74, 0x77, 0xcf, 0xcf, 0xcc,
	0x39, 0x1c, 0x38, 0x6e, 0xd5, 0x4a, 0x35, 0x5b, 0x93, 0x6c, 0x5b, 0x6d, 0x34, 0xc2, 0x4e, 0xb5,
	0xd4, 0xe8, 0xcd, 0xcd, 0x6b, 0x3a, 0xff, 0x9e, 0x40, 0x50, 0x0c, 0x2a, 0x72, 0x08, 0x9c, 0xc2,
	0xbd, 0xc8, 0x8b, 0x8f, 0x8a, 0x11, 0xe2, 0x25, 0xcc, 0x5e, 0x34, 0x19, 0x61, 0xa4, 0x51, 0x7c,
	0x62, 0xb5, 0x3d, 0x81, 0xd7, 0x70, 0x96, 0x77, 0x9f, 0xdd, 0x87, 0x34, 0xcd, 0x4e, 0x3d, 0x48,
	0xaa, 0x48, 0xd5, 0xdc, 0x8f, 0xbc, 0xd8, 0x2f, 0xfe, 0x0a, 0x78, 0x0e, 0xac, 0xfc, 0x7a, 0x94,
	0xf4, 0xce, 0xa7, 0xf6, 0x23, 0x87, 0x30, 0x86, 0xd3, 0x5c, 0x6f, 0x4c, 0x2b, 0x57, 0x26, 0xab,
	0xeb, 0x56, 0x11, 0xf1, 0xff, 0xd6, 0xf0, 0x9b, 0xee, 0x7b, 0x8e, 0x29, 0xcc, 0xa6, 0x8c, 0x10,
	0x43, 0xf0, 0x0b, 0x65, 0x78, 0x60, 0xdf, 0xf5, 0x27, 0x22, 0x4c, 0x97, 0x7a, 0x4d, 0xfc, 0xc0,
	0x52, 0xf6, 0xc6, 0x14, 0x58, 0x5f, 0xbc, 0x23, 0x3e, 0x8b, 0xbc, 0xf8, 0x24, 0xbd, 0x48, 0xf6,
	0x83, 0x24, 0x6e, 0x8c, 0x44, 0x94, 0x59, 0x59, 0x89, 0xc2, 0x39, 0xfb, 0xcc, 0x67, 0x45, 0x24,
	0xd7, 0x8a, 0xc3, 0xb0, 0x8d, 0x83, 0xf3, 0x2b, 0x60, 0x83, 0x17, 0x0f, 0x21, 0x10, 0x55, 0x9e,
	0x2f, 0x84, 0x08, 0xff, 0x21, 0x00, 0xbb, 0xcf, 0x9e, 0x96, 0x8b, 0xbb, 0xd0, 0x7b, 0x63, 0x76,
	0xf7, 0xdb, 0x9f, 0x00, 0x00, 0x00, 0xff, 0xff, 0x73, 0xf4, 0xca, 0xf1, 0x88, 0x01, 0x00, 0x00,
}