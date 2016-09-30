

/*
Package protos is a generated protocol buffer package.

It is generated from these files:
	ca.proto

It has these top-level messages:
	CAStatus
	Empty
	Identity
	Token
	Hash
	PublicKey
	PrivateKey
	Signature
	Registrar
	RegisterUserReq
	ReadUserSetReq
	User
	UserSet
	ECertCreateReq
	ECertCreateResp
	ECertReadReq
	ECertRevokeReq
	ECertCRLReq
	TCertCreateReq
	TCertCreateResp
	TCertCreateSetReq
	TCertAttribute
	TCertCreateSetResp
	TCertReadSetsReq
	TCertRevokeReq
	TCertRevokeSetReq
	TCertCRLReq
	TLSCertCreateReq
	TLSCertCreateResp
	TLSCertReadReq
	TLSCertRevokeReq
	Cert
	TCert
	CertSet
	CertSets
	CertPair
	ACAAttrReq
	ACAAttrResp
	ACAFetchAttrReq
	ACAFetchAttrResp
	FetchAttrsResult
	ACAAttribute
*/
package protos

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"

import google_protobuf "google/protobuf"
//import google_protobuf "github.com/golang/protobuf/ptypes/timestamp"

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

// Public/private keys.
type CryptoType int32

const (
	CryptoType_ECDSA CryptoType = 0
	CryptoType_RSA   CryptoType = 1
	CryptoType_DSA   CryptoType = 2
)

var CryptoType_name = map[int32]string{
	0: "ECDSA",
	1: "RSA",
	2: "DSA",
}
var CryptoType_value = map[string]int32{
	"ECDSA": 0,
	"RSA":   1,
	"DSA":   2,
}

func (x CryptoType) String() string {
	return proto.EnumName(CryptoType_name, int32(x))
}
func (CryptoType) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// User registration.
//
type Role int32

const (
	Role_NONE      Role = 0
	Role_CLIENT    Role = 1
	Role_PEER      Role = 2
	Role_VALIDATOR Role = 4
	Role_AUDITOR   Role = 8
	Role_ALL       Role = 65535
)

var Role_name = map[int32]string{
	0:     "NONE",
	1:     "CLIENT",
	2:     "PEER",
	4:     "VALIDATOR",
	8:     "AUDITOR",
	65535: "ALL",
}
var Role_value = map[string]int32{
	"NONE":      0,
	"CLIENT":    1,
	"PEER":      2,
	"VALIDATOR": 4,
	"AUDITOR":   8,
	"ALL":       65535,
}

func (x Role) String() string {
	return proto.EnumName(Role_name, int32(x))
}
func (Role) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

type CAStatus_StatusCode int32

const (
	CAStatus_OK            CAStatus_StatusCode = 0
	CAStatus_UNKNOWN_ERROR CAStatus_StatusCode = 1
)

var CAStatus_StatusCode_name = map[int32]string{
	0: "OK",
	1: "UNKNOWN_ERROR",
}
var CAStatus_StatusCode_value = map[string]int32{
	"OK":            0,
	"UNKNOWN_ERROR": 1,
}

func (x CAStatus_StatusCode) String() string {
	return proto.EnumName(CAStatus_StatusCode_name, int32(x))
}
func (CAStatus_StatusCode) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type ACAAttrResp_StatusCode int32

const (
	// Processed OK and all attributes included.
	ACAAttrResp_FULL_SUCCESSFUL ACAAttrResp_StatusCode = 0
	// Processed OK  but some attributes included.
	ACAAttrResp_PARTIAL_SUCCESSFUL ACAAttrResp_StatusCode = 1
	// Processed OK  but no attributes included.
	ACAAttrResp_NO_ATTRIBUTES_FOUND ACAAttrResp_StatusCode = 8
	ACAAttrResp_FAILURE_MINVAL      ACAAttrResp_StatusCode = 100
	ACAAttrResp_FAILURE             ACAAttrResp_StatusCode = 100
	ACAAttrResp_BAD_REQUEST         ACAAttrResp_StatusCode = 200
	// Missing parameters
	ACAAttrResp_FAIL_NIL_TS         ACAAttrResp_StatusCode = 201
	ACAAttrResp_FAIL_NIL_ID         ACAAttrResp_StatusCode = 202
	ACAAttrResp_FAIL_NIL_ECERT      ACAAttrResp_StatusCode = 203
	ACAAttrResp_FAIL_NIL_SIGNATURE  ACAAttrResp_StatusCode = 204
	ACAAttrResp_FAIL_NIL_ATTRIBUTES ACAAttrResp_StatusCode = 205
	ACAAttrResp_FAILURE_MAXVAL      ACAAttrResp_StatusCode = 205
)

var ACAAttrResp_StatusCode_name = map[int32]string{
	0:   "FULL_SUCCESSFUL",
	1:   "PARTIAL_SUCCESSFUL",
	8:   "NO_ATTRIBUTES_FOUND",
	100: "FAILURE_MINVAL",
	// Duplicate value: 100: "FAILURE",
	200: "BAD_REQUEST",
	201: "FAIL_NIL_TS",
	202: "FAIL_NIL_ID",
	203: "FAIL_NIL_ECERT",
	204: "FAIL_NIL_SIGNATURE",
	205: "FAIL_NIL_ATTRIBUTES",
	// Duplicate value: 205: "FAILURE_MAXVAL",
}
var ACAAttrResp_StatusCode_value = map[string]int32{
	"FULL_SUCCESSFUL":     0,
	"PARTIAL_SUCCESSFUL":  1,
	"NO_ATTRIBUTES_FOUND": 8,
	"FAILURE_MINVAL":      100,
	"FAILURE":             100,
	"BAD_REQUEST":         200,
	"FAIL_NIL_TS":         201,
	"FAIL_NIL_ID":         202,
	"FAIL_NIL_ECERT":      203,
	"FAIL_NIL_SIGNATURE":  204,
	"FAIL_NIL_ATTRIBUTES": 205,
	"FAILURE_MAXVAL":      205,
}

func (x ACAAttrResp_StatusCode) String() string {
	return proto.EnumName(ACAAttrResp_StatusCode_name, int32(x))
}
func (ACAAttrResp_StatusCode) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{37, 0} }

type ACAFetchAttrResp_StatusCode int32

const (
	// Processed OK
	ACAFetchAttrResp_SUCCESS ACAFetchAttrResp_StatusCode = 0
	// Processed with errors.
	ACAFetchAttrResp_FAILURE ACAFetchAttrResp_StatusCode = 100
)

var ACAFetchAttrResp_StatusCode_name = map[int32]string{
	0:   "SUCCESS",
	100: "FAILURE",
}
var ACAFetchAttrResp_StatusCode_value = map[string]int32{
	"SUCCESS": 0,
	"FAILURE": 100,
}

func (x ACAFetchAttrResp_StatusCode) String() string {
	return proto.EnumName(ACAFetchAttrResp_StatusCode_name, int32(x))
}
func (ACAFetchAttrResp_StatusCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{39, 0}
}

type FetchAttrsResult_StatusCode int32

const (
	// Processed OK
	FetchAttrsResult_SUCCESS FetchAttrsResult_StatusCode = 0
	// Processed with errors
	FetchAttrsResult_FAILURE FetchAttrsResult_StatusCode = 100
)

var FetchAttrsResult_StatusCode_name = map[int32]string{
	0:   "SUCCESS",
	100: "FAILURE",
}
var FetchAttrsResult_StatusCode_value = map[string]int32{
	"SUCCESS": 0,
	"FAILURE": 100,
}

func (x FetchAttrsResult_StatusCode) String() string {
	return proto.EnumName(FetchAttrsResult_StatusCode_name, int32(x))
}
func (FetchAttrsResult_StatusCode) EnumDescriptor() ([]byte, []int) {
	return fileDescriptor0, []int{40, 0}
}

// Status codes shared by both CAs.
//
type CAStatus struct {
	Status CAStatus_StatusCode `protobuf:"varint,1,opt,name=status,enum=protos.CAStatus_StatusCode" json:"status,omitempty"`
}

func (m *CAStatus) Reset()                    { *m = CAStatus{} }
func (m *CAStatus) String() string            { return proto.CompactTextString(m) }
func (*CAStatus) ProtoMessage()               {}
func (*CAStatus) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

// Empty message.
type Empty struct {
}

func (m *Empty) Reset()                    { *m = Empty{} }
func (m *Empty) String() string            { return proto.CompactTextString(m) }
func (*Empty) ProtoMessage()               {}
func (*Empty) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

// Uniquely identifies a user towards either CA.
type Identity struct {
	Id string `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *Identity) Reset()                    { *m = Identity{} }
func (m *Identity) String() string            { return proto.CompactTextString(m) }
func (*Identity) ProtoMessage()               {}
func (*Identity) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

type Token struct {
	Tok []byte `protobuf:"bytes,1,opt,name=tok,proto3" json:"tok,omitempty"`
}

func (m *Token) Reset()                    { *m = Token{} }
func (m *Token) String() string            { return proto.CompactTextString(m) }
func (*Token) ProtoMessage()               {}
func (*Token) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type Hash struct {
	Hash []byte `protobuf:"bytes,1,opt,name=hash,proto3" json:"hash,omitempty"`
}

func (m *Hash) Reset()                    { *m = Hash{} }
func (m *Hash) String() string            { return proto.CompactTextString(m) }
func (*Hash) ProtoMessage()               {}
func (*Hash) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type PublicKey struct {
	Type CryptoType `protobuf:"varint,1,opt,name=type,enum=protos.CryptoType" json:"type,omitempty"`
	Key  []byte     `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *PublicKey) Reset()                    { *m = PublicKey{} }
func (m *PublicKey) String() string            { return proto.CompactTextString(m) }
func (*PublicKey) ProtoMessage()               {}
func (*PublicKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type PrivateKey struct {
	Type CryptoType `protobuf:"varint,1,opt,name=type,enum=protos.CryptoType" json:"type,omitempty"`
	Key  []byte     `protobuf:"bytes,2,opt,name=key,proto3" json:"key,omitempty"`
}

func (m *PrivateKey) Reset()                    { *m = PrivateKey{} }
func (m *PrivateKey) String() string            { return proto.CompactTextString(m) }
func (*PrivateKey) ProtoMessage()               {}
func (*PrivateKey) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

// Signature.
//
type Signature struct {
	Type CryptoType `protobuf:"varint,1,opt,name=type,enum=protos.CryptoType" json:"type,omitempty"`
	R    []byte     `protobuf:"bytes,2,opt,name=r,proto3" json:"r,omitempty"`
	S    []byte     `protobuf:"bytes,3,opt,name=s,proto3" json:"s,omitempty"`
}

func (m *Signature) Reset()                    { *m = Signature{} }
func (m *Signature) String() string            { return proto.CompactTextString(m) }
func (*Signature) ProtoMessage()               {}
func (*Signature) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

type Registrar struct {
	Id            *Identity `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Roles         []string  `protobuf:"bytes,2,rep,name=roles" json:"roles,omitempty"`
	DelegateRoles []string  `protobuf:"bytes,3,rep,name=delegateRoles" json:"delegateRoles,omitempty"`
}

func (m *Registrar) Reset()                    { *m = Registrar{} }
func (m *Registrar) String() string            { return proto.CompactTextString(m) }
func (*Registrar) ProtoMessage()               {}
func (*Registrar) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *Registrar) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

type RegisterUserReq struct {
	Id          *Identity  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Role        Role       `protobuf:"varint,2,opt,name=role,enum=protos.Role" json:"role,omitempty"`
	Affiliation string     `protobuf:"bytes,4,opt,name=affiliation" json:"affiliation,omitempty"`
	Registrar   *Registrar `protobuf:"bytes,5,opt,name=registrar" json:"registrar,omitempty"`
	Sig         *Signature `protobuf:"bytes,6,opt,name=sig" json:"sig,omitempty"`
}

func (m *RegisterUserReq) Reset()                    { *m = RegisterUserReq{} }
func (m *RegisterUserReq) String() string            { return proto.CompactTextString(m) }
func (*RegisterUserReq) ProtoMessage()               {}
func (*RegisterUserReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *RegisterUserReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *RegisterUserReq) GetRegistrar() *Registrar {
	if m != nil {
		return m.Registrar
	}
	return nil
}

func (m *RegisterUserReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type ReadUserSetReq struct {
	Req  *Identity  `protobuf:"bytes,1,opt,name=req" json:"req,omitempty"`
	Role Role       `protobuf:"varint,2,opt,name=role,enum=protos.Role" json:"role,omitempty"`
	Sig  *Signature `protobuf:"bytes,3,opt,name=sig" json:"sig,omitempty"`
}

func (m *ReadUserSetReq) Reset()                    { *m = ReadUserSetReq{} }
func (m *ReadUserSetReq) String() string            { return proto.CompactTextString(m) }
func (*ReadUserSetReq) ProtoMessage()               {}
func (*ReadUserSetReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

func (m *ReadUserSetReq) GetReq() *Identity {
	if m != nil {
		return m.Req
	}
	return nil
}

func (m *ReadUserSetReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type User struct {
	Id   *Identity `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Role Role      `protobuf:"varint,2,opt,name=role,enum=protos.Role" json:"role,omitempty"`
}

func (m *User) Reset()                    { *m = User{} }
func (m *User) String() string            { return proto.CompactTextString(m) }
func (*User) ProtoMessage()               {}
func (*User) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

func (m *User) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

type UserSet struct {
	Users []*User `protobuf:"bytes,1,rep,name=users" json:"users,omitempty"`
}

func (m *UserSet) Reset()                    { *m = UserSet{} }
func (m *UserSet) String() string            { return proto.CompactTextString(m) }
func (*UserSet) ProtoMessage()               {}
func (*UserSet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

func (m *UserSet) GetUsers() []*User {
	if m != nil {
		return m.Users
	}
	return nil
}

// Certificate requests.
//
type ECertCreateReq struct {
	Ts   *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	Id   *Identity                  `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Tok  *Token                     `protobuf:"bytes,3,opt,name=tok" json:"tok,omitempty"`
	Sign *PublicKey                 `protobuf:"bytes,4,opt,name=sign" json:"sign,omitempty"`
	Enc  *PublicKey                 `protobuf:"bytes,5,opt,name=enc" json:"enc,omitempty"`
	Sig  *Signature                 `protobuf:"bytes,6,opt,name=sig" json:"sig,omitempty"`
}

func (m *ECertCreateReq) Reset()                    { *m = ECertCreateReq{} }
func (m *ECertCreateReq) String() string            { return proto.CompactTextString(m) }
func (*ECertCreateReq) ProtoMessage()               {}
func (*ECertCreateReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

func (m *ECertCreateReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *ECertCreateReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *ECertCreateReq) GetTok() *Token {
	if m != nil {
		return m.Tok
	}
	return nil
}

func (m *ECertCreateReq) GetSign() *PublicKey {
	if m != nil {
		return m.Sign
	}
	return nil
}

func (m *ECertCreateReq) GetEnc() *PublicKey {
	if m != nil {
		return m.Enc
	}
	return nil
}

func (m *ECertCreateReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type ECertCreateResp struct {
	Certs       *CertPair         `protobuf:"bytes,1,opt,name=certs" json:"certs,omitempty"`
	Chain       *Token            `protobuf:"bytes,2,opt,name=chain" json:"chain,omitempty"`
	Pkchain     []byte            `protobuf:"bytes,5,opt,name=pkchain,proto3" json:"pkchain,omitempty"`
	Tok         *Token            `protobuf:"bytes,3,opt,name=tok" json:"tok,omitempty"`
	FetchResult *FetchAttrsResult `protobuf:"bytes,4,opt,name=fetchResult" json:"fetchResult,omitempty"`
}

func (m *ECertCreateResp) Reset()                    { *m = ECertCreateResp{} }
func (m *ECertCreateResp) String() string            { return proto.CompactTextString(m) }
func (*ECertCreateResp) ProtoMessage()               {}
func (*ECertCreateResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *ECertCreateResp) GetCerts() *CertPair {
	if m != nil {
		return m.Certs
	}
	return nil
}

func (m *ECertCreateResp) GetChain() *Token {
	if m != nil {
		return m.Chain
	}
	return nil
}

func (m *ECertCreateResp) GetTok() *Token {
	if m != nil {
		return m.Tok
	}
	return nil
}

func (m *ECertCreateResp) GetFetchResult() *FetchAttrsResult {
	if m != nil {
		return m.FetchResult
	}
	return nil
}

type ECertReadReq struct {
	Id *Identity `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *ECertReadReq) Reset()                    { *m = ECertReadReq{} }
func (m *ECertReadReq) String() string            { return proto.CompactTextString(m) }
func (*ECertReadReq) ProtoMessage()               {}
func (*ECertReadReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

func (m *ECertReadReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

type ECertRevokeReq struct {
	Id   *Identity  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Cert *Cert      `protobuf:"bytes,2,opt,name=cert" json:"cert,omitempty"`
	Sig  *Signature `protobuf:"bytes,3,opt,name=sig" json:"sig,omitempty"`
}

func (m *ECertRevokeReq) Reset()                    { *m = ECertRevokeReq{} }
func (m *ECertRevokeReq) String() string            { return proto.CompactTextString(m) }
func (*ECertRevokeReq) ProtoMessage()               {}
func (*ECertRevokeReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *ECertRevokeReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *ECertRevokeReq) GetCert() *Cert {
	if m != nil {
		return m.Cert
	}
	return nil
}

func (m *ECertRevokeReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type ECertCRLReq struct {
	Id  *Identity  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Sig *Signature `protobuf:"bytes,2,opt,name=sig" json:"sig,omitempty"`
}

func (m *ECertCRLReq) Reset()                    { *m = ECertCRLReq{} }
func (m *ECertCRLReq) String() string            { return proto.CompactTextString(m) }
func (*ECertCRLReq) ProtoMessage()               {}
func (*ECertCRLReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func (m *ECertCRLReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *ECertCRLReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TCertCreateReq struct {
	Ts  *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	Id  *Identity                  `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Pub *PublicKey                 `protobuf:"bytes,3,opt,name=pub" json:"pub,omitempty"`
	Sig *Signature                 `protobuf:"bytes,4,opt,name=sig" json:"sig,omitempty"`
}

func (m *TCertCreateReq) Reset()                    { *m = TCertCreateReq{} }
func (m *TCertCreateReq) String() string            { return proto.CompactTextString(m) }
func (*TCertCreateReq) ProtoMessage()               {}
func (*TCertCreateReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

func (m *TCertCreateReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *TCertCreateReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TCertCreateReq) GetPub() *PublicKey {
	if m != nil {
		return m.Pub
	}
	return nil
}

func (m *TCertCreateReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TCertCreateResp struct {
	Cert *Cert `protobuf:"bytes,1,opt,name=cert" json:"cert,omitempty"`
}

func (m *TCertCreateResp) Reset()                    { *m = TCertCreateResp{} }
func (m *TCertCreateResp) String() string            { return proto.CompactTextString(m) }
func (*TCertCreateResp) ProtoMessage()               {}
func (*TCertCreateResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

func (m *TCertCreateResp) GetCert() *Cert {
	if m != nil {
		return m.Cert
	}
	return nil
}

type TCertCreateSetReq struct {
	Ts         *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	Id         *Identity                  `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Num        uint32                     `protobuf:"varint,3,opt,name=num" json:"num,omitempty"`
	Attributes []*TCertAttribute          `protobuf:"bytes,4,rep,name=attributes" json:"attributes,omitempty"`
	Sig        *Signature                 `protobuf:"bytes,5,opt,name=sig" json:"sig,omitempty"`
}

func (m *TCertCreateSetReq) Reset()                    { *m = TCertCreateSetReq{} }
func (m *TCertCreateSetReq) String() string            { return proto.CompactTextString(m) }
func (*TCertCreateSetReq) ProtoMessage()               {}
func (*TCertCreateSetReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

func (m *TCertCreateSetReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *TCertCreateSetReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TCertCreateSetReq) GetAttributes() []*TCertAttribute {
	if m != nil {
		return m.Attributes
	}
	return nil
}

func (m *TCertCreateSetReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TCertAttribute struct {
	AttributeName string `protobuf:"bytes,1,opt,name=attributeName" json:"attributeName,omitempty"`
}

func (m *TCertAttribute) Reset()                    { *m = TCertAttribute{} }
func (m *TCertAttribute) String() string            { return proto.CompactTextString(m) }
func (*TCertAttribute) ProtoMessage()               {}
func (*TCertAttribute) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

type TCertCreateSetResp struct {
	Certs *CertSet `protobuf:"bytes,1,opt,name=certs" json:"certs,omitempty"`
}

func (m *TCertCreateSetResp) Reset()                    { *m = TCertCreateSetResp{} }
func (m *TCertCreateSetResp) String() string            { return proto.CompactTextString(m) }
func (*TCertCreateSetResp) ProtoMessage()               {}
func (*TCertCreateSetResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

func (m *TCertCreateSetResp) GetCerts() *CertSet {
	if m != nil {
		return m.Certs
	}
	return nil
}

type TCertReadSetsReq struct {
	Begin *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=begin" json:"begin,omitempty"`
	End   *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=end" json:"end,omitempty"`
	Req   *Identity                  `protobuf:"bytes,3,opt,name=req" json:"req,omitempty"`
	Role  Role                       `protobuf:"varint,4,opt,name=role,enum=protos.Role" json:"role,omitempty"`
	Sig   *Signature                 `protobuf:"bytes,5,opt,name=sig" json:"sig,omitempty"`
}

func (m *TCertReadSetsReq) Reset()                    { *m = TCertReadSetsReq{} }
func (m *TCertReadSetsReq) String() string            { return proto.CompactTextString(m) }
func (*TCertReadSetsReq) ProtoMessage()               {}
func (*TCertReadSetsReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

func (m *TCertReadSetsReq) GetBegin() *google_protobuf.Timestamp {
	if m != nil {
		return m.Begin
	}
	return nil
}

func (m *TCertReadSetsReq) GetEnd() *google_protobuf.Timestamp {
	if m != nil {
		return m.End
	}
	return nil
}

func (m *TCertReadSetsReq) GetReq() *Identity {
	if m != nil {
		return m.Req
	}
	return nil
}

func (m *TCertReadSetsReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TCertRevokeReq struct {
	Id   *Identity  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Cert *Cert      `protobuf:"bytes,2,opt,name=cert" json:"cert,omitempty"`
	Sig  *Signature `protobuf:"bytes,3,opt,name=sig" json:"sig,omitempty"`
}

func (m *TCertRevokeReq) Reset()                    { *m = TCertRevokeReq{} }
func (m *TCertRevokeReq) String() string            { return proto.CompactTextString(m) }
func (*TCertRevokeReq) ProtoMessage()               {}
func (*TCertRevokeReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{24} }

func (m *TCertRevokeReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TCertRevokeReq) GetCert() *Cert {
	if m != nil {
		return m.Cert
	}
	return nil
}

func (m *TCertRevokeReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TCertRevokeSetReq struct {
	Id  *Identity                  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Ts  *google_protobuf.Timestamp `protobuf:"bytes,2,opt,name=ts" json:"ts,omitempty"`
	Sig *Signature                 `protobuf:"bytes,3,opt,name=sig" json:"sig,omitempty"`
}

func (m *TCertRevokeSetReq) Reset()                    { *m = TCertRevokeSetReq{} }
func (m *TCertRevokeSetReq) String() string            { return proto.CompactTextString(m) }
func (*TCertRevokeSetReq) ProtoMessage()               {}
func (*TCertRevokeSetReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{25} }

func (m *TCertRevokeSetReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TCertRevokeSetReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *TCertRevokeSetReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TCertCRLReq struct {
	Id  *Identity  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Sig *Signature `protobuf:"bytes,2,opt,name=sig" json:"sig,omitempty"`
}

func (m *TCertCRLReq) Reset()                    { *m = TCertCRLReq{} }
func (m *TCertCRLReq) String() string            { return proto.CompactTextString(m) }
func (*TCertCRLReq) ProtoMessage()               {}
func (*TCertCRLReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{26} }

func (m *TCertCRLReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TCertCRLReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TLSCertCreateReq struct {
	Ts  *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	Id  *Identity                  `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Pub *PublicKey                 `protobuf:"bytes,3,opt,name=pub" json:"pub,omitempty"`
	Sig *Signature                 `protobuf:"bytes,4,opt,name=sig" json:"sig,omitempty"`
}

func (m *TLSCertCreateReq) Reset()                    { *m = TLSCertCreateReq{} }
func (m *TLSCertCreateReq) String() string            { return proto.CompactTextString(m) }
func (*TLSCertCreateReq) ProtoMessage()               {}
func (*TLSCertCreateReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{27} }

func (m *TLSCertCreateReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *TLSCertCreateReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TLSCertCreateReq) GetPub() *PublicKey {
	if m != nil {
		return m.Pub
	}
	return nil
}

func (m *TLSCertCreateReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

type TLSCertCreateResp struct {
	Cert     *Cert `protobuf:"bytes,1,opt,name=cert" json:"cert,omitempty"`
	RootCert *Cert `protobuf:"bytes,2,opt,name=rootCert" json:"rootCert,omitempty"`
}

func (m *TLSCertCreateResp) Reset()                    { *m = TLSCertCreateResp{} }
func (m *TLSCertCreateResp) String() string            { return proto.CompactTextString(m) }
func (*TLSCertCreateResp) ProtoMessage()               {}
func (*TLSCertCreateResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{28} }

func (m *TLSCertCreateResp) GetCert() *Cert {
	if m != nil {
		return m.Cert
	}
	return nil
}

func (m *TLSCertCreateResp) GetRootCert() *Cert {
	if m != nil {
		return m.RootCert
	}
	return nil
}

type TLSCertReadReq struct {
	Id *Identity `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
}

func (m *TLSCertReadReq) Reset()                    { *m = TLSCertReadReq{} }
func (m *TLSCertReadReq) String() string            { return proto.CompactTextString(m) }
func (*TLSCertReadReq) ProtoMessage()               {}
func (*TLSCertReadReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{29} }

func (m *TLSCertReadReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

type TLSCertRevokeReq struct {
	Id   *Identity  `protobuf:"bytes,1,opt,name=id" json:"id,omitempty"`
	Cert *Cert      `protobuf:"bytes,2,opt,name=cert" json:"cert,omitempty"`
	Sig  *Signature `protobuf:"bytes,3,opt,name=sig" json:"sig,omitempty"`
}

func (m *TLSCertRevokeReq) Reset()                    { *m = TLSCertRevokeReq{} }
func (m *TLSCertRevokeReq) String() string            { return proto.CompactTextString(m) }
func (*TLSCertRevokeReq) ProtoMessage()               {}
func (*TLSCertRevokeReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{30} }

func (m *TLSCertRevokeReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *TLSCertRevokeReq) GetCert() *Cert {
	if m != nil {
		return m.Cert
	}
	return nil
}

func (m *TLSCertRevokeReq) GetSig() *Signature {
	if m != nil {
		return m.Sig
	}
	return nil
}

// Certificate issued by either the ECA or TCA.
//
type Cert struct {
	Cert []byte `protobuf:"bytes,1,opt,name=cert,proto3" json:"cert,omitempty"`
}

func (m *Cert) Reset()                    { *m = Cert{} }
func (m *Cert) String() string            { return proto.CompactTextString(m) }
func (*Cert) ProtoMessage()               {}
func (*Cert) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{31} }

// TCert
//
type TCert struct {
	Cert  []byte `protobuf:"bytes,1,opt,name=cert,proto3" json:"cert,omitempty"`
	Prek0 []byte `protobuf:"bytes,2,opt,name=prek0,proto3" json:"prek0,omitempty"`
}

func (m *TCert) Reset()                    { *m = TCert{} }
func (m *TCert) String() string            { return proto.CompactTextString(m) }
func (*TCert) ProtoMessage()               {}
func (*TCert) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{32} }

type CertSet struct {
	Ts    *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	Id    *Identity                  `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	Key   []byte                     `protobuf:"bytes,3,opt,name=key,proto3" json:"key,omitempty"`
	Certs []*TCert                   `protobuf:"bytes,4,rep,name=certs" json:"certs,omitempty"`
}

func (m *CertSet) Reset()                    { *m = CertSet{} }
func (m *CertSet) String() string            { return proto.CompactTextString(m) }
func (*CertSet) ProtoMessage()               {}
func (*CertSet) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{33} }

func (m *CertSet) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *CertSet) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *CertSet) GetCerts() []*TCert {
	if m != nil {
		return m.Certs
	}
	return nil
}

type CertSets struct {
	Sets []*CertSet `protobuf:"bytes,1,rep,name=sets" json:"sets,omitempty"`
}

func (m *CertSets) Reset()                    { *m = CertSets{} }
func (m *CertSets) String() string            { return proto.CompactTextString(m) }
func (*CertSets) ProtoMessage()               {}
func (*CertSets) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{34} }

func (m *CertSets) GetSets() []*CertSet {
	if m != nil {
		return m.Sets
	}
	return nil
}

type CertPair struct {
	Sign []byte `protobuf:"bytes,1,opt,name=sign,proto3" json:"sign,omitempty"`
	Enc  []byte `protobuf:"bytes,2,opt,name=enc,proto3" json:"enc,omitempty"`
}

func (m *CertPair) Reset()                    { *m = CertPair{} }
func (m *CertPair) String() string            { return proto.CompactTextString(m) }
func (*CertPair) ProtoMessage()               {}
func (*CertPair) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{35} }

// ACAAttrReq is sent to request an ACert (attributes certificate) to the Attribute Certificate Authority (ACA).
type ACAAttrReq struct {
	// Request time
	Ts *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	// User identity
	Id *Identity `protobuf:"bytes,2,opt,name=id" json:"id,omitempty"`
	// Enrollment certificate
	ECert *Cert `protobuf:"bytes,3,opt,name=eCert" json:"eCert,omitempty"`
	// Collection of requested attributes including the attribute name and its respective value hash.
	Attributes []*TCertAttribute `protobuf:"bytes,4,rep,name=attributes" json:"attributes,omitempty"`
	// The request is signed by the TCA.
	Signature *Signature `protobuf:"bytes,5,opt,name=signature" json:"signature,omitempty"`
}

func (m *ACAAttrReq) Reset()                    { *m = ACAAttrReq{} }
func (m *ACAAttrReq) String() string            { return proto.CompactTextString(m) }
func (*ACAAttrReq) ProtoMessage()               {}
func (*ACAAttrReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{36} }

func (m *ACAAttrReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *ACAAttrReq) GetId() *Identity {
	if m != nil {
		return m.Id
	}
	return nil
}

func (m *ACAAttrReq) GetECert() *Cert {
	if m != nil {
		return m.ECert
	}
	return nil
}

func (m *ACAAttrReq) GetAttributes() []*TCertAttribute {
	if m != nil {
		return m.Attributes
	}
	return nil
}

func (m *ACAAttrReq) GetSignature() *Signature {
	if m != nil {
		return m.Signature
	}
	return nil
}

// ACAAttrResp is the response of Attribute Certificate Authority (ACA) to the attribute request. Is composed by the following fields:
type ACAAttrResp struct {
	// Indicates the request process status.
	Status ACAAttrResp_StatusCode `protobuf:"varint,1,opt,name=status,enum=protos.ACAAttrResp_StatusCode" json:"status,omitempty"`
	// Attribute certificate. Include all the attributes certificated.
	Cert *Cert `protobuf:"bytes,2,opt,name=cert" json:"cert,omitempty"`
	// The response is signed by the ACA.
	Signature *Signature `protobuf:"bytes,3,opt,name=signature" json:"signature,omitempty"`
}

func (m *ACAAttrResp) Reset()                    { *m = ACAAttrResp{} }
func (m *ACAAttrResp) String() string            { return proto.CompactTextString(m) }
func (*ACAAttrResp) ProtoMessage()               {}
func (*ACAAttrResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{37} }

func (m *ACAAttrResp) GetCert() *Cert {
	if m != nil {
		return m.Cert
	}
	return nil
}

func (m *ACAAttrResp) GetSignature() *Signature {
	if m != nil {
		return m.Signature
	}
	return nil
}

// ACAFetchAttrReq is a request to the Attribute Certificate Authority (ACA) to refresh attributes values from the sources.
type ACAFetchAttrReq struct {
	// Request timestamp
	Ts *google_protobuf.Timestamp `protobuf:"bytes,1,opt,name=ts" json:"ts,omitempty"`
	// Enrollment certificate.
	ECert *Cert `protobuf:"bytes,2,opt,name=eCert" json:"eCert,omitempty"`
	// The request is signed by the ECA.
	Signature *Signature `protobuf:"bytes,3,opt,name=signature" json:"signature,omitempty"`
}

func (m *ACAFetchAttrReq) Reset()                    { *m = ACAFetchAttrReq{} }
func (m *ACAFetchAttrReq) String() string            { return proto.CompactTextString(m) }
func (*ACAFetchAttrReq) ProtoMessage()               {}
func (*ACAFetchAttrReq) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{38} }

func (m *ACAFetchAttrReq) GetTs() *google_protobuf.Timestamp {
	if m != nil {
		return m.Ts
	}
	return nil
}

func (m *ACAFetchAttrReq) GetECert() *Cert {
	if m != nil {
		return m.ECert
	}
	return nil
}

func (m *ACAFetchAttrReq) GetSignature() *Signature {
	if m != nil {
		return m.Signature
	}
	return nil
}

// ACAFetchAttrReq is the answer of the Attribute Certificate Authority (ACA) to the refresh request.
type ACAFetchAttrResp struct {
	// Status of the fetch process.
	Status ACAFetchAttrResp_StatusCode `protobuf:"varint,1,opt,name=status,enum=protos.ACAFetchAttrResp_StatusCode" json:"status,omitempty"`
	// Error message.
	Msg string `protobuf:"bytes,2,opt,name=Msg,json=msg" json:"Msg,omitempty"`
}

func (m *ACAFetchAttrResp) Reset()                    { *m = ACAFetchAttrResp{} }
func (m *ACAFetchAttrResp) String() string            { return proto.CompactTextString(m) }
func (*ACAFetchAttrResp) ProtoMessage()               {}
func (*ACAFetchAttrResp) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{39} }

// FetchAttrsResult is returned within the ECertCreateResp indicating the results of the fetch attributes invoked during enroll.
type FetchAttrsResult struct {
	// Status of the fetch process.
	Status FetchAttrsResult_StatusCode `protobuf:"varint,1,opt,name=status,enum=protos.FetchAttrsResult_StatusCode" json:"status,omitempty"`
	// Error message.
	Msg string `protobuf:"bytes,2,opt,name=Msg,json=msg" json:"Msg,omitempty"`
}

func (m *FetchAttrsResult) Reset()                    { *m = FetchAttrsResult{} }
func (m *FetchAttrsResult) String() string            { return proto.CompactTextString(m) }
func (*FetchAttrsResult) ProtoMessage()               {}
func (*FetchAttrsResult) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{40} }

// ACAAttribute is an instance of an attribute with the time constraints. Is used to marshal attributes to be stored within the certificate extensions.
type ACAAttribute struct {
	// Name of the attribute.
	AttributeName string `protobuf:"bytes,1,opt,name=attributeName" json:"attributeName,omitempty"`
	// Value of the attribute.
	AttributeValue []byte `protobuf:"bytes,2,opt,name=attributeValue,proto3" json:"attributeValue,omitempty"`
	// The timestamp which attribute is valid from.
	ValidFrom *google_protobuf.Timestamp `protobuf:"bytes,3,opt,name=validFrom" json:"validFrom,omitempty"`
	// The timestamp which attribute is valid to.
	ValidTo *google_protobuf.Timestamp `protobuf:"bytes,4,opt,name=validTo" json:"validTo,omitempty"`
}

func (m *ACAAttribute) Reset()                    { *m = ACAAttribute{} }
func (m *ACAAttribute) String() string            { return proto.CompactTextString(m) }
func (*ACAAttribute) ProtoMessage()               {}
func (*ACAAttribute) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{41} }

func (m *ACAAttribute) GetValidFrom() *google_protobuf.Timestamp {
	if m != nil {
		return m.ValidFrom
	}
	return nil
}

func (m *ACAAttribute) GetValidTo() *google_protobuf.Timestamp {
	if m != nil {
		return m.ValidTo
	}
	return nil
}

func init() {
	proto.RegisterType((*CAStatus)(nil), "protos.CAStatus")
	proto.RegisterType((*Empty)(nil), "protos.Empty")
	proto.RegisterType((*Identity)(nil), "protos.Identity")
	proto.RegisterType((*Token)(nil), "protos.Token")
	proto.RegisterType((*Hash)(nil), "protos.Hash")
	proto.RegisterType((*PublicKey)(nil), "protos.PublicKey")
	proto.RegisterType((*PrivateKey)(nil), "protos.PrivateKey")
	proto.RegisterType((*Signature)(nil), "protos.Signature")
	proto.RegisterType((*Registrar)(nil), "protos.Registrar")
	proto.RegisterType((*RegisterUserReq)(nil), "protos.RegisterUserReq")
	proto.RegisterType((*ReadUserSetReq)(nil), "protos.ReadUserSetReq")
	proto.RegisterType((*User)(nil), "protos.User")
	proto.RegisterType((*UserSet)(nil), "protos.UserSet")
	proto.RegisterType((*ECertCreateReq)(nil), "protos.ECertCreateReq")
	proto.RegisterType((*ECertCreateResp)(nil), "protos.ECertCreateResp")
	proto.RegisterType((*ECertReadReq)(nil), "protos.ECertReadReq")
	proto.RegisterType((*ECertRevokeReq)(nil), "protos.ECertRevokeReq")
	proto.RegisterType((*ECertCRLReq)(nil), "protos.ECertCRLReq")
	proto.RegisterType((*TCertCreateReq)(nil), "protos.TCertCreateReq")
	proto.RegisterType((*TCertCreateResp)(nil), "protos.TCertCreateResp")
	proto.RegisterType((*TCertCreateSetReq)(nil), "protos.TCertCreateSetReq")
	proto.RegisterType((*TCertAttribute)(nil), "protos.TCertAttribute")
	proto.RegisterType((*TCertCreateSetResp)(nil), "protos.TCertCreateSetResp")
	proto.RegisterType((*TCertReadSetsReq)(nil), "protos.TCertReadSetsReq")
	proto.RegisterType((*TCertRevokeReq)(nil), "protos.TCertRevokeReq")
	proto.RegisterType((*TCertRevokeSetReq)(nil), "protos.TCertRevokeSetReq")
	proto.RegisterType((*TCertCRLReq)(nil), "protos.TCertCRLReq")
	proto.RegisterType((*TLSCertCreateReq)(nil), "protos.TLSCertCreateReq")
	proto.RegisterType((*TLSCertCreateResp)(nil), "protos.TLSCertCreateResp")
	proto.RegisterType((*TLSCertReadReq)(nil), "protos.TLSCertReadReq")
	proto.RegisterType((*TLSCertRevokeReq)(nil), "protos.TLSCertRevokeReq")
	proto.RegisterType((*Cert)(nil), "protos.Cert")
	proto.RegisterType((*TCert)(nil), "protos.TCert")
	proto.RegisterType((*CertSet)(nil), "protos.CertSet")
	proto.RegisterType((*CertSets)(nil), "protos.CertSets")
	proto.RegisterType((*CertPair)(nil), "protos.CertPair")
	proto.RegisterType((*ACAAttrReq)(nil), "protos.ACAAttrReq")
	proto.RegisterType((*ACAAttrResp)(nil), "protos.ACAAttrResp")
	proto.RegisterType((*ACAFetchAttrReq)(nil), "protos.ACAFetchAttrReq")
	proto.RegisterType((*ACAFetchAttrResp)(nil), "protos.ACAFetchAttrResp")
	proto.RegisterType((*FetchAttrsResult)(nil), "protos.FetchAttrsResult")
	proto.RegisterType((*ACAAttribute)(nil), "protos.ACAAttribute")
	proto.RegisterEnum("protos.CryptoType", CryptoType_name, CryptoType_value)
	proto.RegisterEnum("protos.Role", Role_name, Role_value)
	proto.RegisterEnum("protos.CAStatus_StatusCode", CAStatus_StatusCode_name, CAStatus_StatusCode_value)
	proto.RegisterEnum("protos.ACAAttrResp_StatusCode", ACAAttrResp_StatusCode_name, ACAAttrResp_StatusCode_value)
	proto.RegisterEnum("protos.ACAFetchAttrResp_StatusCode", ACAFetchAttrResp_StatusCode_name, ACAFetchAttrResp_StatusCode_value)
	proto.RegisterEnum("protos.FetchAttrsResult_StatusCode", FetchAttrsResult_StatusCode_name, FetchAttrsResult_StatusCode_value)
}

// Reference imports to suppress errors if they are not otherwise used.
var _ context.Context
var _ grpc.ClientConn

// This is a compile-time assertion to ensure that this generated file
// is compatible with the grpc package it is being compiled against.
const _ = grpc.SupportPackageIsVersion3

// Client API for ECAP service

type ECAPClient interface {
	ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error)
	CreateCertificatePair(ctx context.Context, in *ECertCreateReq, opts ...grpc.CallOption) (*ECertCreateResp, error)
	ReadCertificatePair(ctx context.Context, in *ECertReadReq, opts ...grpc.CallOption) (*CertPair, error)
	ReadCertificateByHash(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Cert, error)
	RevokeCertificatePair(ctx context.Context, in *ECertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error)
}

type eCAPClient struct {
	cc *grpc.ClientConn
}

func NewECAPClient(cc *grpc.ClientConn) ECAPClient {
	return &eCAPClient{cc}
}

func (c *eCAPClient) ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error) {
	out := new(Cert)
	err := grpc.Invoke(ctx, "/protos.ECAP/ReadCACertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAPClient) CreateCertificatePair(ctx context.Context, in *ECertCreateReq, opts ...grpc.CallOption) (*ECertCreateResp, error) {
	out := new(ECertCreateResp)
	err := grpc.Invoke(ctx, "/protos.ECAP/CreateCertificatePair", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAPClient) ReadCertificatePair(ctx context.Context, in *ECertReadReq, opts ...grpc.CallOption) (*CertPair, error) {
	out := new(CertPair)
	err := grpc.Invoke(ctx, "/protos.ECAP/ReadCertificatePair", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAPClient) ReadCertificateByHash(ctx context.Context, in *Hash, opts ...grpc.CallOption) (*Cert, error) {
	out := new(Cert)
	err := grpc.Invoke(ctx, "/protos.ECAP/ReadCertificateByHash", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAPClient) RevokeCertificatePair(ctx context.Context, in *ECertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.ECAP/RevokeCertificatePair", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ECAP service

type ECAPServer interface {
	ReadCACertificate(context.Context, *Empty) (*Cert, error)
	CreateCertificatePair(context.Context, *ECertCreateReq) (*ECertCreateResp, error)
	ReadCertificatePair(context.Context, *ECertReadReq) (*CertPair, error)
	ReadCertificateByHash(context.Context, *Hash) (*Cert, error)
	RevokeCertificatePair(context.Context, *ECertRevokeReq) (*CAStatus, error)
}

func RegisterECAPServer(s *grpc.Server, srv ECAPServer) {
	s.RegisterService(&_ECAP_serviceDesc, srv)
}

func _ECAP_ReadCACertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAPServer).ReadCACertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAP/ReadCACertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAPServer).ReadCACertificate(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAP_CreateCertificatePair_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ECertCreateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAPServer).CreateCertificatePair(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAP/CreateCertificatePair",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAPServer).CreateCertificatePair(ctx, req.(*ECertCreateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAP_ReadCertificatePair_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ECertReadReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAPServer).ReadCertificatePair(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAP/ReadCertificatePair",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAPServer).ReadCertificatePair(ctx, req.(*ECertReadReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAP_ReadCertificateByHash_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Hash)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAPServer).ReadCertificateByHash(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAP/ReadCertificateByHash",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAPServer).ReadCertificateByHash(ctx, req.(*Hash))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAP_RevokeCertificatePair_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ECertRevokeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAPServer).RevokeCertificatePair(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAP/RevokeCertificatePair",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAPServer).RevokeCertificatePair(ctx, req.(*ECertRevokeReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _ECAP_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.ECAP",
	HandlerType: (*ECAPServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadCACertificate",
			Handler:    _ECAP_ReadCACertificate_Handler,
		},
		{
			MethodName: "CreateCertificatePair",
			Handler:    _ECAP_CreateCertificatePair_Handler,
		},
		{
			MethodName: "ReadCertificatePair",
			Handler:    _ECAP_ReadCertificatePair_Handler,
		},
		{
			MethodName: "ReadCertificateByHash",
			Handler:    _ECAP_ReadCertificateByHash_Handler,
		},
		{
			MethodName: "RevokeCertificatePair",
			Handler:    _ECAP_RevokeCertificatePair_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for ECAA service

type ECAAClient interface {
	RegisterUser(ctx context.Context, in *RegisterUserReq, opts ...grpc.CallOption) (*Token, error)
	ReadUserSet(ctx context.Context, in *ReadUserSetReq, opts ...grpc.CallOption) (*UserSet, error)
	RevokeCertificate(ctx context.Context, in *ECertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error)
	PublishCRL(ctx context.Context, in *ECertCRLReq, opts ...grpc.CallOption) (*CAStatus, error)
}

type eCAAClient struct {
	cc *grpc.ClientConn
}

func NewECAAClient(cc *grpc.ClientConn) ECAAClient {
	return &eCAAClient{cc}
}

func (c *eCAAClient) RegisterUser(ctx context.Context, in *RegisterUserReq, opts ...grpc.CallOption) (*Token, error) {
	out := new(Token)
	err := grpc.Invoke(ctx, "/protos.ECAA/RegisterUser", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAAClient) ReadUserSet(ctx context.Context, in *ReadUserSetReq, opts ...grpc.CallOption) (*UserSet, error) {
	out := new(UserSet)
	err := grpc.Invoke(ctx, "/protos.ECAA/ReadUserSet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAAClient) RevokeCertificate(ctx context.Context, in *ECertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.ECAA/RevokeCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *eCAAClient) PublishCRL(ctx context.Context, in *ECertCRLReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.ECAA/PublishCRL", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ECAA service

type ECAAServer interface {
	RegisterUser(context.Context, *RegisterUserReq) (*Token, error)
	ReadUserSet(context.Context, *ReadUserSetReq) (*UserSet, error)
	RevokeCertificate(context.Context, *ECertRevokeReq) (*CAStatus, error)
	PublishCRL(context.Context, *ECertCRLReq) (*CAStatus, error)
}

func RegisterECAAServer(s *grpc.Server, srv ECAAServer) {
	s.RegisterService(&_ECAA_serviceDesc, srv)
}

func _ECAA_RegisterUser_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(RegisterUserReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAAServer).RegisterUser(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAA/RegisterUser",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAAServer).RegisterUser(ctx, req.(*RegisterUserReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAA_ReadUserSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ReadUserSetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAAServer).ReadUserSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAA/ReadUserSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAAServer).ReadUserSet(ctx, req.(*ReadUserSetReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAA_RevokeCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ECertRevokeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAAServer).RevokeCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAA/RevokeCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAAServer).RevokeCertificate(ctx, req.(*ECertRevokeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ECAA_PublishCRL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ECertCRLReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ECAAServer).PublishCRL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ECAA/PublishCRL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ECAAServer).PublishCRL(ctx, req.(*ECertCRLReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _ECAA_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.ECAA",
	HandlerType: (*ECAAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RegisterUser",
			Handler:    _ECAA_RegisterUser_Handler,
		},
		{
			MethodName: "ReadUserSet",
			Handler:    _ECAA_ReadUserSet_Handler,
		},
		{
			MethodName: "RevokeCertificate",
			Handler:    _ECAA_RevokeCertificate_Handler,
		},
		{
			MethodName: "PublishCRL",
			Handler:    _ECAA_PublishCRL_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for TCAP service

type TCAPClient interface {
	ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error)
	CreateCertificateSet(ctx context.Context, in *TCertCreateSetReq, opts ...grpc.CallOption) (*TCertCreateSetResp, error)
	RevokeCertificate(ctx context.Context, in *TCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error)
	RevokeCertificateSet(ctx context.Context, in *TCertRevokeSetReq, opts ...grpc.CallOption) (*CAStatus, error)
}

type tCAPClient struct {
	cc *grpc.ClientConn
}

func NewTCAPClient(cc *grpc.ClientConn) TCAPClient {
	return &tCAPClient{cc}
}

func (c *tCAPClient) ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error) {
	out := new(Cert)
	err := grpc.Invoke(ctx, "/protos.TCAP/ReadCACertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tCAPClient) CreateCertificateSet(ctx context.Context, in *TCertCreateSetReq, opts ...grpc.CallOption) (*TCertCreateSetResp, error) {
	out := new(TCertCreateSetResp)
	err := grpc.Invoke(ctx, "/protos.TCAP/CreateCertificateSet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tCAPClient) RevokeCertificate(ctx context.Context, in *TCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TCAP/RevokeCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tCAPClient) RevokeCertificateSet(ctx context.Context, in *TCertRevokeSetReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TCAP/RevokeCertificateSet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TCAP service

type TCAPServer interface {
	ReadCACertificate(context.Context, *Empty) (*Cert, error)
	CreateCertificateSet(context.Context, *TCertCreateSetReq) (*TCertCreateSetResp, error)
	RevokeCertificate(context.Context, *TCertRevokeReq) (*CAStatus, error)
	RevokeCertificateSet(context.Context, *TCertRevokeSetReq) (*CAStatus, error)
}

func RegisterTCAPServer(s *grpc.Server, srv TCAPServer) {
	s.RegisterService(&_TCAP_serviceDesc, srv)
}

func _TCAP_ReadCACertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAPServer).ReadCACertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAP/ReadCACertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAPServer).ReadCACertificate(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _TCAP_CreateCertificateSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TCertCreateSetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAPServer).CreateCertificateSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAP/CreateCertificateSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAPServer).CreateCertificateSet(ctx, req.(*TCertCreateSetReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _TCAP_RevokeCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TCertRevokeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAPServer).RevokeCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAP/RevokeCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAPServer).RevokeCertificate(ctx, req.(*TCertRevokeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _TCAP_RevokeCertificateSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TCertRevokeSetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAPServer).RevokeCertificateSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAP/RevokeCertificateSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAPServer).RevokeCertificateSet(ctx, req.(*TCertRevokeSetReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _TCAP_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.TCAP",
	HandlerType: (*TCAPServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadCACertificate",
			Handler:    _TCAP_ReadCACertificate_Handler,
		},
		{
			MethodName: "CreateCertificateSet",
			Handler:    _TCAP_CreateCertificateSet_Handler,
		},
		{
			MethodName: "RevokeCertificate",
			Handler:    _TCAP_RevokeCertificate_Handler,
		},
		{
			MethodName: "RevokeCertificateSet",
			Handler:    _TCAP_RevokeCertificateSet_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for TCAA service

type TCAAClient interface {
	RevokeCertificate(ctx context.Context, in *TCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error)
	RevokeCertificateSet(ctx context.Context, in *TCertRevokeSetReq, opts ...grpc.CallOption) (*CAStatus, error)
	PublishCRL(ctx context.Context, in *TCertCRLReq, opts ...grpc.CallOption) (*CAStatus, error)
}

type tCAAClient struct {
	cc *grpc.ClientConn
}

func NewTCAAClient(cc *grpc.ClientConn) TCAAClient {
	return &tCAAClient{cc}
}

func (c *tCAAClient) RevokeCertificate(ctx context.Context, in *TCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TCAA/RevokeCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tCAAClient) RevokeCertificateSet(ctx context.Context, in *TCertRevokeSetReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TCAA/RevokeCertificateSet", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tCAAClient) PublishCRL(ctx context.Context, in *TCertCRLReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TCAA/PublishCRL", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TCAA service

type TCAAServer interface {
	RevokeCertificate(context.Context, *TCertRevokeReq) (*CAStatus, error)
	RevokeCertificateSet(context.Context, *TCertRevokeSetReq) (*CAStatus, error)
	PublishCRL(context.Context, *TCertCRLReq) (*CAStatus, error)
}

func RegisterTCAAServer(s *grpc.Server, srv TCAAServer) {
	s.RegisterService(&_TCAA_serviceDesc, srv)
}

func _TCAA_RevokeCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TCertRevokeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAAServer).RevokeCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAA/RevokeCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAAServer).RevokeCertificate(ctx, req.(*TCertRevokeReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _TCAA_RevokeCertificateSet_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TCertRevokeSetReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAAServer).RevokeCertificateSet(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAA/RevokeCertificateSet",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAAServer).RevokeCertificateSet(ctx, req.(*TCertRevokeSetReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _TCAA_PublishCRL_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TCertCRLReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TCAAServer).PublishCRL(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TCAA/PublishCRL",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TCAAServer).PublishCRL(ctx, req.(*TCertCRLReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _TCAA_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.TCAA",
	HandlerType: (*TCAAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RevokeCertificate",
			Handler:    _TCAA_RevokeCertificate_Handler,
		},
		{
			MethodName: "RevokeCertificateSet",
			Handler:    _TCAA_RevokeCertificateSet_Handler,
		},
		{
			MethodName: "PublishCRL",
			Handler:    _TCAA_PublishCRL_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for TLSCAP service

type TLSCAPClient interface {
	ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error)
	CreateCertificate(ctx context.Context, in *TLSCertCreateReq, opts ...grpc.CallOption) (*TLSCertCreateResp, error)
	ReadCertificate(ctx context.Context, in *TLSCertReadReq, opts ...grpc.CallOption) (*Cert, error)
	RevokeCertificate(ctx context.Context, in *TLSCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error)
}

type tLSCAPClient struct {
	cc *grpc.ClientConn
}

func NewTLSCAPClient(cc *grpc.ClientConn) TLSCAPClient {
	return &tLSCAPClient{cc}
}

func (c *tLSCAPClient) ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error) {
	out := new(Cert)
	err := grpc.Invoke(ctx, "/protos.TLSCAP/ReadCACertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tLSCAPClient) CreateCertificate(ctx context.Context, in *TLSCertCreateReq, opts ...grpc.CallOption) (*TLSCertCreateResp, error) {
	out := new(TLSCertCreateResp)
	err := grpc.Invoke(ctx, "/protos.TLSCAP/CreateCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tLSCAPClient) ReadCertificate(ctx context.Context, in *TLSCertReadReq, opts ...grpc.CallOption) (*Cert, error) {
	out := new(Cert)
	err := grpc.Invoke(ctx, "/protos.TLSCAP/ReadCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *tLSCAPClient) RevokeCertificate(ctx context.Context, in *TLSCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TLSCAP/RevokeCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TLSCAP service

type TLSCAPServer interface {
	ReadCACertificate(context.Context, *Empty) (*Cert, error)
	CreateCertificate(context.Context, *TLSCertCreateReq) (*TLSCertCreateResp, error)
	ReadCertificate(context.Context, *TLSCertReadReq) (*Cert, error)
	RevokeCertificate(context.Context, *TLSCertRevokeReq) (*CAStatus, error)
}

func RegisterTLSCAPServer(s *grpc.Server, srv TLSCAPServer) {
	s.RegisterService(&_TLSCAP_serviceDesc, srv)
}

func _TLSCAP_ReadCACertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TLSCAPServer).ReadCACertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TLSCAP/ReadCACertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TLSCAPServer).ReadCACertificate(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _TLSCAP_CreateCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TLSCertCreateReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TLSCAPServer).CreateCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TLSCAP/CreateCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TLSCAPServer).CreateCertificate(ctx, req.(*TLSCertCreateReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _TLSCAP_ReadCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TLSCertReadReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TLSCAPServer).ReadCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TLSCAP/ReadCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TLSCAPServer).ReadCertificate(ctx, req.(*TLSCertReadReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _TLSCAP_RevokeCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TLSCertRevokeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TLSCAPServer).RevokeCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TLSCAP/RevokeCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TLSCAPServer).RevokeCertificate(ctx, req.(*TLSCertRevokeReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _TLSCAP_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.TLSCAP",
	HandlerType: (*TLSCAPServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadCACertificate",
			Handler:    _TLSCAP_ReadCACertificate_Handler,
		},
		{
			MethodName: "CreateCertificate",
			Handler:    _TLSCAP_CreateCertificate_Handler,
		},
		{
			MethodName: "ReadCertificate",
			Handler:    _TLSCAP_ReadCertificate_Handler,
		},
		{
			MethodName: "RevokeCertificate",
			Handler:    _TLSCAP_RevokeCertificate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for TLSCAA service

type TLSCAAClient interface {
	RevokeCertificate(ctx context.Context, in *TLSCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error)
}

type tLSCAAClient struct {
	cc *grpc.ClientConn
}

func NewTLSCAAClient(cc *grpc.ClientConn) TLSCAAClient {
	return &tLSCAAClient{cc}
}

func (c *tLSCAAClient) RevokeCertificate(ctx context.Context, in *TLSCertRevokeReq, opts ...grpc.CallOption) (*CAStatus, error) {
	out := new(CAStatus)
	err := grpc.Invoke(ctx, "/protos.TLSCAA/RevokeCertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for TLSCAA service

type TLSCAAServer interface {
	RevokeCertificate(context.Context, *TLSCertRevokeReq) (*CAStatus, error)
}

func RegisterTLSCAAServer(s *grpc.Server, srv TLSCAAServer) {
	s.RegisterService(&_TLSCAA_serviceDesc, srv)
}

func _TLSCAA_RevokeCertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(TLSCertRevokeReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(TLSCAAServer).RevokeCertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.TLSCAA/RevokeCertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(TLSCAAServer).RevokeCertificate(ctx, req.(*TLSCertRevokeReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _TLSCAA_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.TLSCAA",
	HandlerType: (*TLSCAAServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "RevokeCertificate",
			Handler:    _TLSCAA_RevokeCertificate_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

// Client API for ACAP service

type ACAPClient interface {
	ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error)
	RequestAttributes(ctx context.Context, in *ACAAttrReq, opts ...grpc.CallOption) (*ACAAttrResp, error)
	FetchAttributes(ctx context.Context, in *ACAFetchAttrReq, opts ...grpc.CallOption) (*ACAFetchAttrResp, error)
}

type aCAPClient struct {
	cc *grpc.ClientConn
}

func NewACAPClient(cc *grpc.ClientConn) ACAPClient {
	return &aCAPClient{cc}
}

func (c *aCAPClient) ReadCACertificate(ctx context.Context, in *Empty, opts ...grpc.CallOption) (*Cert, error) {
	out := new(Cert)
	err := grpc.Invoke(ctx, "/protos.ACAP/ReadCACertificate", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aCAPClient) RequestAttributes(ctx context.Context, in *ACAAttrReq, opts ...grpc.CallOption) (*ACAAttrResp, error) {
	out := new(ACAAttrResp)
	err := grpc.Invoke(ctx, "/protos.ACAP/RequestAttributes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

func (c *aCAPClient) FetchAttributes(ctx context.Context, in *ACAFetchAttrReq, opts ...grpc.CallOption) (*ACAFetchAttrResp, error) {
	out := new(ACAFetchAttrResp)
	err := grpc.Invoke(ctx, "/protos.ACAP/FetchAttributes", in, out, c.cc, opts...)
	if err != nil {
		return nil, err
	}
	return out, nil
}

// Server API for ACAP service

type ACAPServer interface {
	ReadCACertificate(context.Context, *Empty) (*Cert, error)
	RequestAttributes(context.Context, *ACAAttrReq) (*ACAAttrResp, error)
	FetchAttributes(context.Context, *ACAFetchAttrReq) (*ACAFetchAttrResp, error)
}

func RegisterACAPServer(s *grpc.Server, srv ACAPServer) {
	s.RegisterService(&_ACAP_serviceDesc, srv)
}

func _ACAP_ReadCACertificate_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(Empty)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACAPServer).ReadCACertificate(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ACAP/ReadCACertificate",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACAPServer).ReadCACertificate(ctx, req.(*Empty))
	}
	return interceptor(ctx, in, info, handler)
}

func _ACAP_RequestAttributes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ACAAttrReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACAPServer).RequestAttributes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ACAP/RequestAttributes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACAPServer).RequestAttributes(ctx, req.(*ACAAttrReq))
	}
	return interceptor(ctx, in, info, handler)
}

func _ACAP_FetchAttributes_Handler(srv interface{}, ctx context.Context, dec func(interface{}) error, interceptor grpc.UnaryServerInterceptor) (interface{}, error) {
	in := new(ACAFetchAttrReq)
	if err := dec(in); err != nil {
		return nil, err
	}
	if interceptor == nil {
		return srv.(ACAPServer).FetchAttributes(ctx, in)
	}
	info := &grpc.UnaryServerInfo{
		Server:     srv,
		FullMethod: "/protos.ACAP/FetchAttributes",
	}
	handler := func(ctx context.Context, req interface{}) (interface{}, error) {
		return srv.(ACAPServer).FetchAttributes(ctx, req.(*ACAFetchAttrReq))
	}
	return interceptor(ctx, in, info, handler)
}

var _ACAP_serviceDesc = grpc.ServiceDesc{
	ServiceName: "protos.ACAP",
	HandlerType: (*ACAPServer)(nil),
	Methods: []grpc.MethodDesc{
		{
			MethodName: "ReadCACertificate",
			Handler:    _ACAP_ReadCACertificate_Handler,
		},
		{
			MethodName: "RequestAttributes",
			Handler:    _ACAP_RequestAttributes_Handler,
		},
		{
			MethodName: "FetchAttributes",
			Handler:    _ACAP_FetchAttributes_Handler,
		},
	},
	Streams:  []grpc.StreamDesc{},
	Metadata: fileDescriptor0,
}

func init() { proto.RegisterFile("ca.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1818 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xd4, 0x59, 0x4f, 0x6f, 0xdb, 0xc6,
	0x12, 0x7f, 0x14, 0x25, 0x5b, 0x1a, 0xc9, 0x32, 0xbd, 0x76, 0x62, 0x45, 0x0f, 0x78, 0x09, 0x98,
	0x97, 0xbc, 0xbc, 0xa0, 0xb5, 0x53, 0xbb, 0x48, 0x8b, 0xa6, 0x41, 0xc1, 0xc8, 0x74, 0xa3, 0x46,
	0x96, 0x5c, 0x8a, 0x4a, 0x7b, 0x13, 0x64, 0x9b, 0xb6, 0x09, 0xcb, 0x92, 0x4c, 0x52, 0x01, 0x84,
	0xdc, 0x0b, 0xb4, 0x05, 0x7a, 0xe8, 0x77, 0x28, 0x50, 0xf4, 0x1b, 0x14, 0x28, 0xd0, 0x6b, 0xff,
	0x24, 0x40, 0x2f, 0x3d, 0xf6, 0x5c, 0xa0, 0xa7, 0x7e, 0x82, 0xa6, 0xb3, 0xcb, 0x25, 0x45, 0x52,
	0x7f, 0xcc, 0x38, 0x2e, 0xd2, 0x1e, 0x0c, 0x2d, 0x77, 0x66, 0x77, 0x67, 0x7e, 0x33, 0x3b, 0x7f,
	0xd6, 0x90, 0xde, 0x6d, 0xad, 0xf4, 0xac, 0xae, 0xd3, 0x25, 0x33, 0xec, 0xc7, 0x2e, 0x5e, 0x3e,
	0xe8, 0x76, 0x0f, 0xda, 0xc6, 0x2a, 0xfb, 0xdc, 0xe9, 0xef, 0xaf, 0x3a, 0xe6, 0xb1, 0x61, 0x3b,
	0xad, 0xe3, 0x9e, 0xcb, 0x28, 0x1f, 0x42, 0xba, 0xa4, 0xd4, 0x9d, 0x96, 0xd3, 0xb7, 0xc9, 0x3a,
	0xcc, 0xd8, 0x6c, 0x54, 0x10, 0xae, 0x08, 0x37, 0xf2, 0x6b, 0xff, 0x76, 0x79, 0xec, 0x15, 0x8f,
	0x63, 0xc5, 0xfd, 0x29, 0x75, 0xf7, 0x0c, 0x8d, 0xb3, 0xca, 0xff, 0x03, 0x18, 0xce, 0x92, 0x19,
	0x48, 0xd4, 0x1e, 0x48, 0xff, 0x22, 0x0b, 0x30, 0xd7, 0xa8, 0x3e, 0xa8, 0xd6, 0x3e, 0xa8, 0x36,
	0x55, 0x4d, 0xab, 0x69, 0x92, 0x20, 0xcf, 0x42, 0x4a, 0x3d, 0xee, 0x39, 0x03, 0xb9, 0x08, 0xe9,
	0xf2, 0x9e, 0xd1, 0x71, 0x4c, 0x67, 0x40, 0xf2, 0x90, 0x30, 0xf7, 0xd8, 0x71, 0x19, 0x0d, 0x47,
	0xf2, 0x25, 0x48, 0xe9, 0xdd, 0x23, 0xa3, 0x43, 0x24, 0x10, 0x9d, 0xee, 0x11, 0xa3, 0xe4, 0x34,
	0x3a, 0xc4, 0x65, 0xc9, 0xfb, 0x2d, 0xfb, 0x90, 0x10, 0x48, 0x1e, 0xe2, 0x2f, 0x27, 0xb1, 0xb1,
	0xac, 0x42, 0x66, 0xbb, 0xbf, 0xd3, 0x36, 0x77, 0x1f, 0x18, 0x03, 0x72, 0x1d, 0x92, 0xce, 0xa0,
	0x67, 0x70, 0x25, 0x88, 0xaf, 0x84, 0x35, 0xe8, 0x39, 0x5d, 0x1d, 0x29, 0x1a, 0xa3, 0xd3, 0x23,
	0x8e, 0x8c, 0x41, 0x21, 0xe1, 0x1e, 0x81, 0x43, 0x79, 0x13, 0x60, 0xdb, 0x32, 0x1f, 0xb5, 0x1c,
	0xe3, 0xc5, 0xf6, 0xa9, 0x41, 0xa6, 0x6e, 0x1e, 0x74, 0x10, 0x15, 0xcb, 0x88, 0xbd, 0x4d, 0x0e,
	0x04, 0x8b, 0x6f, 0x22, 0x58, 0xf4, 0xcb, 0x2e, 0x88, 0xee, 0x97, 0x2d, 0x9b, 0x90, 0xd1, 0x8c,
	0x03, 0xd3, 0x76, 0xac, 0x96, 0x45, 0xae, 0xf8, 0x98, 0x65, 0xd7, 0x24, 0x6f, 0x3b, 0x0f, 0x51,
	0x8a, 0x22, 0x59, 0x82, 0x94, 0xd5, 0x6d, 0x1b, 0x36, 0x6e, 0x27, 0x22, 0xb0, 0xee, 0x07, 0xf9,
	0x2f, 0xcc, 0xed, 0x19, 0x6d, 0xe3, 0x00, 0xd5, 0xd3, 0x18, 0x55, 0x64, 0xd4, 0xf0, 0xa4, 0xfc,
	0x54, 0x80, 0x79, 0xf7, 0x2c, 0xc3, 0x6a, 0xd8, 0x86, 0xa5, 0x19, 0x27, 0x31, 0x4e, 0xbc, 0x02,
	0x49, 0x7a, 0x08, 0x93, 0x3f, 0xbf, 0x96, 0xf3, 0x78, 0xe8, 0x96, 0x1a, 0xa3, 0x20, 0x47, 0xb6,
	0xb5, 0xbf, 0x6f, 0xb6, 0xcd, 0x96, 0x63, 0x76, 0x3b, 0x85, 0x24, 0x33, 0x79, 0x70, 0x8a, 0xac,
	0x42, 0xc6, 0xf2, 0x94, 0x2c, 0xa4, 0xd8, 0x61, 0x0b, 0xfe, 0x46, 0x1e, 0x41, 0x1b, 0xf2, 0x90,
	0xab, 0x20, 0xda, 0xe6, 0x41, 0x61, 0x26, 0xcc, 0xea, 0x23, 0xaf, 0x51, 0xaa, 0xfc, 0x18, 0xf2,
	0x9a, 0xd1, 0xda, 0xa3, 0xaa, 0xd4, 0x0d, 0x87, 0x6a, 0x23, 0x83, 0x68, 0x19, 0x27, 0x13, 0xd5,
	0xa1, 0xc4, 0x18, 0xfa, 0xf0, 0xc3, 0xc5, 0xa9, 0x87, 0xbf, 0x07, 0x49, 0x7a, 0xf0, 0x79, 0x00,
	0x28, 0xbf, 0x0a, 0xb3, 0x5c, 0x09, 0xd4, 0x20, 0xd5, 0xc7, 0x21, 0xbd, 0xa7, 0x22, 0xee, 0xe8,
	0x73, 0x33, 0x7b, 0xb9, 0x24, 0xf9, 0x77, 0x01, 0xf2, 0x6a, 0xc9, 0xb0, 0x9c, 0x92, 0x65, 0x50,
	0xe3, 0xa2, 0x52, 0x37, 0x21, 0xe1, 0xd8, 0x5c, 0x8a, 0xe2, 0x8a, 0x1b, 0x19, 0x56, 0xbc, 0xc8,
	0xb0, 0xa2, 0x7b, 0x91, 0x41, 0x43, 0x2e, 0x2e, 0x71, 0x62, 0x8a, 0xc4, 0x97, 0xdd, 0x1b, 0xea,
	0x02, 0x30, 0xe7, 0xb1, 0xb0, 0xdb, 0xcb, 0x2e, 0x2c, 0xb9, 0x06, 0x49, 0xc4, 0xc0, 0x35, 0x75,
	0x00, 0x22, 0xff, 0xa2, 0x6a, 0x8c, 0x4c, 0x81, 0x34, 0x3a, 0xbb, 0x51, 0x83, 0x0f, 0xb9, 0x28,
	0x35, 0x9e, 0xa9, 0x7f, 0x46, 0xd7, 0x0d, 0xa9, 0x6c, 0xf7, 0xf0, 0xf6, 0xa5, 0x76, 0x71, 0xc6,
	0x8e, 0x82, 0x4f, 0xd9, 0xb6, 0x5b, 0x26, 0xc2, 0xc5, 0xc8, 0x78, 0x40, 0x6a, 0xf7, 0xb0, 0x65,
	0x76, 0xb8, 0xca, 0x11, 0x7d, 0x5c, 0x1a, 0x29, 0xc0, 0x6c, 0xef, 0xc8, 0x65, 0x4b, 0xb1, 0xab,
	0xe9, 0x7d, 0x9e, 0x0e, 0xc6, 0x5b, 0x90, 0xdd, 0x37, 0x9c, 0xdd, 0x43, 0x14, 0xaa, 0xdf, 0x76,
	0x38, 0x26, 0x05, 0x8f, 0x71, 0x93, 0x92, 0x14, 0xc7, 0xb1, 0x6c, 0x97, 0xae, 0x05, 0x99, 0xe5,
	0x5b, 0x90, 0x63, 0x6a, 0x51, 0x3f, 0x8e, 0x75, 0x1d, 0xe5, 0x01, 0xb7, 0xbd, 0x66, 0x3c, 0x42,
	0x11, 0x62, 0x5f, 0x61, 0x0a, 0x05, 0x07, 0x20, 0x17, 0x04, 0x4a, 0x63, 0x94, 0x78, 0x2e, 0xaf,
	0x43, 0xd6, 0xb5, 0x81, 0x56, 0x89, 0x77, 0x2e, 0xdf, 0x35, 0x31, 0x75, 0xd7, 0x2f, 0xd1, 0x9b,
	0xf5, 0xbf, 0xd2, 0x9b, 0x51, 0x8a, 0x5e, 0x7f, 0x27, 0xaa, 0x5b, 0xc0, 0x0b, 0x91, 0xea, 0x89,
	0x9a, 0x9c, 0x2a, 0xea, 0x3a, 0xcc, 0xeb, 0x11, 0x27, 0xf4, 0xa0, 0x15, 0x26, 0x41, 0x2b, 0xff,
	0x24, 0xc0, 0x42, 0x60, 0x15, 0x8f, 0x54, 0xe7, 0xab, 0x22, 0xe6, 0xa9, 0x4e, 0xff, 0x98, 0xa9,
	0x38, 0xa7, 0xd1, 0x21, 0xb9, 0x0d, 0xd0, 0x42, 0xa7, 0x33, 0x77, 0xfa, 0x0e, 0xa6, 0x83, 0x24,
	0x0b, 0x26, 0x17, 0x7d, 0xe7, 0xa5, 0xe2, 0x28, 0x1e, 0x59, 0x0b, 0x70, 0x7a, 0x38, 0xa4, 0xa6,
	0xe2, 0x70, 0x9b, 0x5b, 0xcc, 0xdf, 0x82, 0x26, 0x20, 0x7f, 0x93, 0x6a, 0xeb, 0xd8, 0xe0, 0x79,
	0x3f, 0x3c, 0x29, 0xdf, 0x01, 0x12, 0x45, 0x02, 0x21, 0xbc, 0x16, 0xbe, 0xc7, 0xf3, 0x41, 0x0c,
	0x29, 0x8f, 0x4b, 0x95, 0x7f, 0x11, 0x40, 0xd2, 0xbd, 0xbb, 0x82, 0xf3, 0x36, 0x85, 0xf1, 0x16,
	0xa4, 0x76, 0x30, 0x69, 0x74, 0x62, 0x20, 0xe9, 0x32, 0x92, 0x57, 0x68, 0x4c, 0xf2, 0xd0, 0x9c,
	0xc6, 0x4f, 0xd9, 0xbc, 0x84, 0x22, 0xc6, 0x49, 0x28, 0xc9, 0xd3, 0x12, 0xca, 0x74, 0x50, 0x07,
	0x1c, 0xd4, 0x97, 0x70, 0xb1, 0x3f, 0xf2, 0x5c, 0xd4, 0x3d, 0x9b, 0xbb, 0xe8, 0xe9, 0xc7, 0xbb,
	0x4e, 0x9c, 0x88, 0xe5, 0xc4, 0x71, 0x23, 0x8c, 0x7e, 0xfe, 0x11, 0xe6, 0x2b, 0xea, 0x39, 0x95,
	0xfa, 0x3f, 0x23, 0xc6, 0x34, 0xd1, 0x14, 0x61, 0x59, 0xe3, 0x44, 0x19, 0x72, 0x03, 0xd2, 0x56,
	0xb7, 0xeb, 0x94, 0x26, 0x79, 0x83, 0x4f, 0x95, 0xd7, 0xd0, 0xcf, 0xdc, 0x03, 0xe2, 0x27, 0x9d,
	0xc7, 0x3e, 0x80, 0x2f, 0xc1, 0x3b, 0xb1, 0x3b, 0xa0, 0x4b, 0x68, 0x77, 0xe0, 0x83, 0x90, 0xe3,
	0xc1, 0xf5, 0x35, 0x6c, 0x2a, 0x26, 0x11, 0x69, 0xad, 0xdc, 0xb3, 0x8c, 0xa3, 0x5b, 0xbc, 0xf4,
	0x76, 0x3f, 0xe4, 0xcf, 0x04, 0x98, 0xe5, 0xa1, 0xe5, 0xfc, 0xa3, 0x30, 0xed, 0x16, 0x44, 0xbf,
	0x5b, 0x60, 0xa5, 0x07, 0x0b, 0x6d, 0x6e, 0x00, 0x9e, 0x0b, 0x05, 0x60, 0x2f, 0xb0, 0xad, 0x62,
	0x9f, 0xe6, 0xca, 0x43, 0x6f, 0x49, 0xd2, 0x36, 0x1c, 0xaf, 0xfa, 0x1b, 0x09, 0x85, 0x8c, 0x88,
	0x45, 0x43, 0xda, 0xab, 0x71, 0xa8, 0xde, 0xac, 0x12, 0xe3, 0x7a, 0xb3, 0xb2, 0x4b, 0x72, 0xcb,
	0x2e, 0xde, 0xb5, 0xe0, 0x50, 0xfe, 0x55, 0x00, 0x50, 0x4a, 0x0a, 0x8d, 0xd7, 0xe7, 0xef, 0xfb,
	0x58, 0xb2, 0x1a, 0xcc, 0xef, 0xc4, 0x31, 0x76, 0x76, 0x49, 0x67, 0x4e, 0x47, 0xd8, 0x38, 0xd8,
	0x9e, 0x37, 0x4c, 0x8e, 0x9f, 0x43, 0x1e, 0xf9, 0x0b, 0x11, 0xb2, 0xbe, 0xa6, 0x78, 0x73, 0x6e,
	0x47, 0x1a, 0xdf, 0xff, 0x78, 0xab, 0x03, 0x4c, 0x63, 0x7a, 0xdf, 0x18, 0xbe, 0x1b, 0x12, 0x4d,
	0x8c, 0x21, 0xda, 0x27, 0x89, 0x50, 0x3f, 0xbd, 0x08, 0xf3, 0x9b, 0x8d, 0x4a, 0xa5, 0x59, 0x6f,
	0x94, 0x4a, 0x6a, 0xbd, 0x8e, 0x63, 0x6c, 0xae, 0x2f, 0x02, 0xd9, 0x56, 0x34, 0xbd, 0xac, 0x84,
	0xe6, 0x05, 0xb2, 0x0c, 0x8b, 0xd5, 0x5a, 0x53, 0xd1, 0x75, 0xad, 0x7c, 0xaf, 0xa1, 0xab, 0xf5,
	0xe6, 0x66, 0xad, 0x51, 0xdd, 0x90, 0xd2, 0x68, 0xff, 0xfc, 0xa6, 0x52, 0xae, 0x34, 0x34, 0xb5,
	0xb9, 0x55, 0xae, 0x3e, 0x54, 0x2a, 0xd2, 0x1e, 0xc9, 0xc2, 0x2c, 0x9f, 0x93, 0xa8, 0x53, 0x66,
	0xef, 0x29, 0x1b, 0x4d, 0x4d, 0x7d, 0xbf, 0xa1, 0xd6, 0x75, 0xe9, 0x3b, 0x81, 0xce, 0x50, 0x72,
	0xb3, 0x8a, 0x7f, 0x7a, 0x5d, 0xfa, 0x3e, 0x3c, 0x53, 0xde, 0x90, 0x7e, 0x10, 0x50, 0xb8, 0xbc,
	0x3f, 0xa3, 0x96, 0x54, 0x4d, 0x97, 0x7e, 0xa4, 0x42, 0x10, 0x7f, 0xb2, 0x5e, 0x7e, 0xb7, 0xaa,
	0xe8, 0xf4, 0x88, 0x27, 0x02, 0x16, 0xcf, 0x8b, 0x3e, 0x61, 0x28, 0xa3, 0xf4, 0xd4, 0xdf, 0x87,
	0x89, 0xa7, 0x7c, 0x48, 0xc5, 0x7b, 0x2a, 0x14, 0x13, 0x92, 0x20, 0x7f, 0x8e, 0x05, 0x3d, 0x9a,
	0xc0, 0xaf, 0x8e, 0x9f, 0xd7, 0x2d, 0x7d, 0xa7, 0x4b, 0x4c, 0x76, 0xba, 0xe7, 0xb6, 0xd0, 0xc7,
	0x98, 0x28, 0xc2, 0x42, 0xa1, 0x07, 0xdd, 0x89, 0x78, 0xd0, 0xd5, 0x80, 0x07, 0x85, 0x38, 0xc7,
	0xb9, 0x11, 0x5e, 0xc5, 0x2d, 0xdb, 0xcd, 0x4f, 0x19, 0x4d, 0x3c, 0xb6, 0x0f, 0xe4, 0xeb, 0x21,
	0x27, 0x40, 0x53, 0x71, 0x3b, 0xa3, 0xf1, 0x83, 0x76, 0x63, 0xb2, 0x44, 0x7b, 0x87, 0xc9, 0xb2,
	0x44, 0x39, 0xcf, 0x57, 0x96, 0x27, 0x02, 0xe4, 0xf8, 0x7d, 0x79, 0x8e, 0x72, 0x0f, 0x1b, 0xb4,
	0xbc, 0x3f, 0xf1, 0xb0, 0xd5, 0xee, 0x1b, 0x3c, 0x24, 0x45, 0x66, 0xc9, 0x9b, 0x90, 0x79, 0xd4,
	0x6a, 0x9b, 0x7b, 0x9b, 0x56, 0xf7, 0x98, 0xdb, 0x69, 0x9a, 0xf9, 0x87, 0xcc, 0xe4, 0x75, 0x98,
	0x65, 0x1f, 0x7a, 0x97, 0x67, 0xd5, 0x69, 0xeb, 0x3c, 0xd6, 0x9b, 0xff, 0x07, 0x18, 0x3e, 0xd1,
	0x90, 0x0c, 0xa4, 0xd4, 0xd2, 0x46, 0x5d, 0x41, 0xa5, 0x67, 0x41, 0xd4, 0x70, 0x20, 0xd0, 0x01,
	0x9d, 0x49, 0xdc, 0xdc, 0x82, 0x24, 0xad, 0xe3, 0x48, 0x1a, 0x92, 0xd5, 0x5a, 0x55, 0x45, 0x1e,
	0x80, 0x99, 0x52, 0xa5, 0xac, 0x56, 0x75, 0x64, 0xc3, 0xd9, 0x6d, 0x55, 0xd5, 0xa4, 0x04, 0x99,
	0x83, 0x0c, 0x3a, 0x77, 0x79, 0x43, 0xd1, 0x6b, 0x9a, 0x94, 0xa4, 0xe8, 0x29, 0x8d, 0x8d, 0x32,
	0xfd, 0x48, 0xe3, 0x01, 0xa2, 0x52, 0xa9, 0x48, 0xcf, 0x9e, 0x89, 0x6b, 0x5f, 0x27, 0x20, 0xa9,
	0x96, 0x94, 0x6d, 0xac, 0x5b, 0x17, 0x68, 0xf6, 0x2d, 0x29, 0xd4, 0x51, 0xcd, 0x7d, 0x73, 0x17,
	0x33, 0x3d, 0xf1, 0xd3, 0x03, 0x7b, 0x4c, 0x2b, 0x86, 0x7c, 0x9a, 0xdc, 0x87, 0x0b, 0x6e, 0x41,
	0x10, 0x58, 0xc1, 0x32, 0x80, 0x1f, 0x46, 0xc3, 0x4f, 0x02, 0xc5, 0xe5, 0xb1, 0xf3, 0xe8, 0xd0,
	0x77, 0x61, 0x91, 0x9d, 0x1d, 0xd9, 0x67, 0x29, 0xc4, 0xcf, 0x6b, 0x83, 0xe2, 0x48, 0x57, 0x4d,
	0xd6, 0xe1, 0x42, 0x64, 0xf9, 0xbd, 0x01, 0x7b, 0xbd, 0xf3, 0xe5, 0xa5, 0x5f, 0x11, 0xe9, 0x15,
	0xba, 0x88, 0x56, 0x0e, 0xd3, 0xa5, 0xf7, 0xab, 0x8b, 0xc0, 0xb9, 0xfc, 0x81, 0x72, 0xed, 0x37,
	0x81, 0x61, 0xa7, 0x60, 0x48, 0xcf, 0x05, 0x5f, 0xb1, 0xc8, 0x72, 0xf8, 0x25, 0xc9, 0x7f, 0xdb,
	0x2a, 0x86, 0x9b, 0x75, 0x5c, 0x97, 0x0d, 0x3c, 0x17, 0x0d, 0x4f, 0x0e, 0xbf, 0x21, 0x15, 0xe7,
	0x83, 0x4f, 0x2e, 0x94, 0xf1, 0x2e, 0xb5, 0x55, 0x44, 0xf6, 0xf8, 0x72, 0x23, 0x5e, 0xc0, 0xea,
	0x40, 0xfb, 0x10, 0xab, 0x5a, 0xb2, 0x18, 0xb6, 0x0a, 0xab, 0x73, 0xc7, 0x28, 0xfb, 0x29, 0x3a,
	0x8a, 0x7e, 0x36, 0x47, 0xd9, 0x82, 0xa5, 0x11, 0x47, 0xa1, 0x6a, 0x5c, 0x0a, 0xa5, 0xdb, 0x60,
	0x33, 0x5a, 0x2c, 0x4e, 0x22, 0x31, 0x6f, 0x99, 0xa6, 0xbd, 0x7e, 0x9a, 0xf6, 0x25, 0x58, 0x1a,
	0x59, 0x3e, 0x2a, 0x4d, 0xb0, 0xef, 0x18, 0x83, 0xc6, 0xb7, 0x02, 0x43, 0x43, 0xf9, 0x3b, 0x08,
	0x33, 0xc9, 0x9e, 0xfa, 0x54, 0x7b, 0xfe, 0x21, 0xc0, 0x0c, 0xad, 0xa0, 0xcf, 0x78, 0xf5, 0x17,
	0x46, 0x2c, 0x4a, 0xfc, 0x07, 0xa6, 0x68, 0x67, 0x53, 0xbc, 0x34, 0x81, 0x82, 0xc6, 0x7c, 0x83,
	0x3e, 0x00, 0x87, 0xee, 0x6e, 0x00, 0xbd, 0x50, 0x53, 0x10, 0x11, 0xe1, 0x9d, 0x71, 0xc0, 0x17,
	0x46, 0x96, 0x4e, 0xbe, 0xbd, 0x65, 0xae, 0xbf, 0xf2, 0xe2, 0x5b, 0x7d, 0x83, 0xde, 0xa0, 0x9c,
	0x0d, 0xc9, 0xb7, 0xe9, 0x8a, 0x93, 0x3e, 0x26, 0x04, 0x65, 0x58, 0x63, 0x92, 0x91, 0x92, 0xf0,
	0xa4, 0xb8, 0x38, 0xa6, 0x4c, 0x24, 0x1b, 0x58, 0xb1, 0x79, 0x79, 0x96, 0xaf, 0x5d, 0x1e, 0x5f,
	0x0c, 0x9c, 0x14, 0x0b, 0x93, 0xaa, 0x84, 0x1d, 0xf7, 0xff, 0x37, 0xeb, 0x7f, 0x06, 0x00, 0x00,
	0xff, 0xff, 0x9f, 0x47, 0x1c, 0x48, 0xd2, 0x19, 0x00, 0x00,
}