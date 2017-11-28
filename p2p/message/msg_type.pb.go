// Code generated by protoc-gen-go. DO NOT EDIT.
// source: msg_type.proto

package message

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

type MsgType int32

const (
	// //////////////////////////
	// HTS issues
	MsgType_CLIENTHELLO  MsgType = 0
	MsgType_SERVERHELLO  MsgType = 1
	MsgType_SERVERREJECT MsgType = 2
	MsgType_CLIENTACCEPT MsgType = 3
	MsgType_CLIENTREJECT MsgType = 4
	MsgType_SERVERDONE   MsgType = 5
	// /////////////////////////
	// HTS update session key
	MsgType_CLIENTUPDATE MsgType = 6
	// hypernet issues
	MsgType_KEEPALIVE MsgType = 7
	MsgType_PENDING   MsgType = 8
	// ////////////////////////
	// CONSENSUS issues
	// general hello
	MsgType_HELLO MsgType = 9
	// new node attend
	MsgType_ATTEND MsgType = 10
	// reconnect
	MsgType_RECONNECT MsgType = 11
	// consensus importtant session
	MsgType_SESSION MsgType = 12
	// generally response
	MsgType_RESPONSE MsgType = 13
	// NVP issues
	MsgType_NVPATTEND MsgType = 14
	MsgType_NVPDELETE MsgType = 15
	MsgType_NVPEXIT   MsgType = 16
	MsgType_VPDELETE  MsgType = 17
)

var MsgType_name = map[int32]string{
	0:  "CLIENTHELLO",
	1:  "SERVERHELLO",
	2:  "SERVERREJECT",
	3:  "CLIENTACCEPT",
	4:  "CLIENTREJECT",
	5:  "SERVERDONE",
	6:  "CLIENTUPDATE",
	7:  "KEEPALIVE",
	8:  "PENDING",
	9:  "HELLO",
	10: "ATTEND",
	11: "RECONNECT",
	12: "SESSION",
	13: "RESPONSE",
	14: "NVPATTEND",
	15: "NVPDELETE",
	16: "NVPEXIT",
	17: "VPDELETE",
}
var MsgType_value = map[string]int32{
	"CLIENTHELLO":  0,
	"SERVERHELLO":  1,
	"SERVERREJECT": 2,
	"CLIENTACCEPT": 3,
	"CLIENTREJECT": 4,
	"SERVERDONE":   5,
	"CLIENTUPDATE": 6,
	"KEEPALIVE":    7,
	"PENDING":      8,
	"HELLO":        9,
	"ATTEND":       10,
	"RECONNECT":    11,
	"SESSION":      12,
	"RESPONSE":     13,
	"NVPATTEND":    14,
	"NVPDELETE":    15,
	"NVPEXIT":      16,
	"VPDELETE":     17,
}

func (x MsgType) String() string {
	return proto.EnumName(MsgType_name, int32(x))
}
func (MsgType) EnumDescriptor() ([]byte, []int) { return fileDescriptor2, []int{0} }

func init() {
	proto.RegisterEnum("message.MsgType", MsgType_name, MsgType_value)
}

func init() { proto.RegisterFile("msg_type.proto", fileDescriptor2) }

var fileDescriptor2 = []byte{
	// 237 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0x44, 0x90, 0x41, 0x4e, 0xc3, 0x30,
	0x10, 0x45, 0xa1, 0xd0, 0xa4, 0x99, 0xa4, 0xe9, 0xe0, 0x63, 0xb0, 0x60, 0xc3, 0x09, 0xa2, 0xe4,
	0x0b, 0x02, 0x61, 0x6c, 0xd9, 0x26, 0x62, 0x87, 0x40, 0x8a, 0xb2, 0xaa, 0x1a, 0x91, 0x6e, 0x7a,
	0x14, 0x6e, 0x8b, 0x8c, 0xa3, 0x76, 0xe9, 0xe7, 0xf7, 0xf4, 0xa5, 0xa1, 0x72, 0x3f, 0x8f, 0x9f,
	0xc7, 0xd3, 0x34, 0x3c, 0x4c, 0x3f, 0x87, 0xe3, 0x41, 0xa5, 0xfb, 0x61, 0x9e, 0xbf, 0xc6, 0xe1,
	0xfe, 0x77, 0x45, 0xe9, 0xdb, 0x3c, 0xfa, 0xd3, 0x34, 0xa8, 0x1d, 0xe5, 0x75, 0xd7, 0x42, 0xfc,
	0x33, 0xba, 0x4e, 0xf3, 0x55, 0x00, 0x0e, 0xb6, 0x87, 0x8d, 0xe0, 0x5a, 0x31, 0x15, 0x11, 0x58,
	0xbc, 0xa0, 0xf6, 0xbc, 0x0a, 0x24, 0x36, 0x55, 0x5d, 0xc3, 0x78, 0xbe, 0xb9, 0x90, 0xc5, 0xb9,
	0x55, 0x25, 0x51, 0xac, 0x1a, 0x2d, 0xe0, 0xf5, 0xc5, 0x78, 0x37, 0x4d, 0xe5, 0xc1, 0x89, 0xda,
	0x52, 0xf6, 0x0a, 0x98, 0xaa, 0x6b, 0x7b, 0x70, 0xaa, 0x72, 0x4a, 0x0d, 0xa4, 0x69, 0xe5, 0x89,
	0x37, 0x2a, 0xa3, 0x75, 0x9c, 0xcf, 0x14, 0x51, 0x52, 0x79, 0x0f, 0x69, 0x98, 0x42, 0x62, 0x51,
	0x6b, 0x91, 0xb0, 0x91, 0x87, 0xc4, 0xc1, 0xb9, 0x56, 0x0b, 0x17, 0xaa, 0xa0, 0x8d, 0x85, 0x33,
	0x5a, 0x1c, 0x78, 0x1b, 0x4c, 0xe9, 0xcd, 0x12, 0x96, 0xcb, 0xb3, 0x41, 0x07, 0x0f, 0xde, 0x85,
	0x50, 0x7a, 0x83, 0x8f, 0xd6, 0x33, 0x87, 0xf0, 0xfc, 0x75, 0xf7, 0x9d, 0xfc, 0xdf, 0xea, 0xf1,
	0x2f, 0x00, 0x00, 0xff, 0xff, 0xc2, 0xf4, 0xef, 0x2b, 0x3d, 0x01, 0x00, 0x00,
}