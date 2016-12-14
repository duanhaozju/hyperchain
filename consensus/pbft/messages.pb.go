// Code generated by protoc-gen-go.
// source: messages.proto
// DO NOT EDIT!

/*
Package pbft is a generated protocol buffer package.

It is generated from these files:
	messages.proto

It has these top-level messages:
	ConsensusMessage
	TransactionBatch
	PrePrepare
	Prepare
	Commit
	BlockInfo
	Checkpoint
	ViewChange
	PQset
	NewView
	FetchRequestBatch
	NegotiateView
	NegotiateViewResponse
	RecoveryInit
	RecoveryResponse
	RecoveryFetchPQC
	RecoveryReturnPQC
	Metadata
*/
package pbft

import proto "github.com/golang/protobuf/proto"
import fmt "fmt"
import math "math"
import types "hyperchain/core/types"

// Reference imports to suppress errors if they are not otherwise used.
var _ = proto.Marshal
var _ = fmt.Errorf
var _ = math.Inf

// This is a compile-time assertion to ensure that this generated file
// is compatible with the proto package it is being compiled against.
// A compilation error at this line likely means your copy of the
// proto package needs to be updated.
const _ = proto.ProtoPackageIsVersion2 // please upgrade the proto package

type ConsensusMessage_Type int32

const (
	ConsensusMessage_TRANSACTION             ConsensusMessage_Type = 0
	ConsensusMessage_TRANSATION_BATCH        ConsensusMessage_Type = 1
	ConsensusMessage_PRE_PREPARE             ConsensusMessage_Type = 2
	ConsensusMessage_PREPARE                 ConsensusMessage_Type = 3
	ConsensusMessage_COMMIT                  ConsensusMessage_Type = 4
	ConsensusMessage_CHECKPOINT              ConsensusMessage_Type = 5
	ConsensusMessage_VIEW_CHANGE             ConsensusMessage_Type = 6
	ConsensusMessage_NEW_VIEW                ConsensusMessage_Type = 7
	ConsensusMessage_FRTCH_REQUEST_BATCH     ConsensusMessage_Type = 8
	ConsensusMessage_RETURN_REQUEST_BATCH    ConsensusMessage_Type = 9
	ConsensusMessage_NEGOTIATE_VIEW          ConsensusMessage_Type = 10
	ConsensusMessage_NEGOTIATE_VIEW_RESPONSE ConsensusMessage_Type = 11
	ConsensusMessage_RECOVERY_INIT           ConsensusMessage_Type = 12
	ConsensusMessage_RECOVERY_RESPONSE       ConsensusMessage_Type = 13
	ConsensusMessage_RECOVERY_FETCH_QPC      ConsensusMessage_Type = 14
	ConsensusMessage_RECOVERY_RETURN_QPC     ConsensusMessage_Type = 15
)

var ConsensusMessage_Type_name = map[int32]string{
	0:  "TRANSACTION",
	1:  "TRANSATION_BATCH",
	2:  "PRE_PREPARE",
	3:  "PREPARE",
	4:  "COMMIT",
	5:  "CHECKPOINT",
	6:  "VIEW_CHANGE",
	7:  "NEW_VIEW",
	8:  "FRTCH_REQUEST_BATCH",
	9:  "RETURN_REQUEST_BATCH",
	10: "NEGOTIATE_VIEW",
	11: "NEGOTIATE_VIEW_RESPONSE",
	12: "RECOVERY_INIT",
	13: "RECOVERY_RESPONSE",
	14: "RECOVERY_FETCH_QPC",
	15: "RECOVERY_RETURN_QPC",
}
var ConsensusMessage_Type_value = map[string]int32{
	"TRANSACTION":             0,
	"TRANSATION_BATCH":        1,
	"PRE_PREPARE":             2,
	"PREPARE":                 3,
	"COMMIT":                  4,
	"CHECKPOINT":              5,
	"VIEW_CHANGE":             6,
	"NEW_VIEW":                7,
	"FRTCH_REQUEST_BATCH":     8,
	"RETURN_REQUEST_BATCH":    9,
	"NEGOTIATE_VIEW":          10,
	"NEGOTIATE_VIEW_RESPONSE": 11,
	"RECOVERY_INIT":           12,
	"RECOVERY_RESPONSE":       13,
	"RECOVERY_FETCH_QPC":      14,
	"RECOVERY_RETURN_QPC":     15,
}

func (x ConsensusMessage_Type) String() string {
	return proto.EnumName(ConsensusMessage_Type_name, int32(x))
}
func (ConsensusMessage_Type) EnumDescriptor() ([]byte, []int) { return fileDescriptor0, []int{0, 0} }

type ConsensusMessage struct {
	Type    ConsensusMessage_Type `protobuf:"varint,1,opt,name=type,enum=pbft.ConsensusMessage_Type" json:"type,omitempty"`
	Payload []byte                `protobuf:"bytes,2,opt,name=payload,proto3" json:"payload,omitempty"`
}

func (m *ConsensusMessage) Reset()                    { *m = ConsensusMessage{} }
func (m *ConsensusMessage) String() string            { return proto.CompactTextString(m) }
func (*ConsensusMessage) ProtoMessage()               {}
func (*ConsensusMessage) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{0} }

type TransactionBatch struct {
	Batch     []*types.Transaction `protobuf:"bytes,1,rep,name=batch" json:"batch,omitempty"`
	Timestamp int64                `protobuf:"varint,2,opt,name=timestamp" json:"timestamp,omitempty"`
}

func (m *TransactionBatch) Reset()                    { *m = TransactionBatch{} }
func (m *TransactionBatch) String() string            { return proto.CompactTextString(m) }
func (*TransactionBatch) ProtoMessage()               {}
func (*TransactionBatch) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{1} }

func (m *TransactionBatch) GetBatch() []*types.Transaction {
	if m != nil {
		return m.Batch
	}
	return nil
}

type PrePrepare struct {
	View             uint64            `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	SequenceNumber   uint64            `protobuf:"varint,2,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	BatchDigest      string            `protobuf:"bytes,3,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	TransactionBatch *TransactionBatch `protobuf:"bytes,4,opt,name=transaction_batch,json=transactionBatch" json:"transaction_batch,omitempty"`
	ReplicaId        uint64            `protobuf:"varint,5,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *PrePrepare) Reset()                    { *m = PrePrepare{} }
func (m *PrePrepare) String() string            { return proto.CompactTextString(m) }
func (*PrePrepare) ProtoMessage()               {}
func (*PrePrepare) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *PrePrepare) GetTransactionBatch() *TransactionBatch {
	if m != nil {
		return m.TransactionBatch
	}
	return nil
}

type Prepare struct {
	View           uint64 `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	SequenceNumber uint64 `protobuf:"varint,2,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	BatchDigest    string `protobuf:"bytes,3,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	ReplicaId      uint64 `protobuf:"varint,4,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *Prepare) Reset()                    { *m = Prepare{} }
func (m *Prepare) String() string            { return proto.CompactTextString(m) }
func (*Prepare) ProtoMessage()               {}
func (*Prepare) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

type Commit struct {
	View           uint64 `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	SequenceNumber uint64 `protobuf:"varint,2,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	BatchDigest    string `protobuf:"bytes,3,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	ReplicaId      uint64 `protobuf:"varint,4,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *Commit) Reset()                    { *m = Commit{} }
func (m *Commit) String() string            { return proto.CompactTextString(m) }
func (*Commit) ProtoMessage()               {}
func (*Commit) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type BlockInfo struct {
	BlockNumber uint64 `protobuf:"varint,1,opt,name=block_number,json=blockNumber" json:"block_number,omitempty"`
	BlockHash   []byte `protobuf:"bytes,2,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
}

func (m *BlockInfo) Reset()                    { *m = BlockInfo{} }
func (m *BlockInfo) String() string            { return proto.CompactTextString(m) }
func (*BlockInfo) ProtoMessage()               {}
func (*BlockInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type Checkpoint struct {
	SequenceNumber uint64 `protobuf:"varint,1,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	ReplicaId      uint64 `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Id             string `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
}

func (m *Checkpoint) Reset()                    { *m = Checkpoint{} }
func (m *Checkpoint) String() string            { return proto.CompactTextString(m) }
func (*Checkpoint) ProtoMessage()               {}
func (*Checkpoint) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type ViewChange struct {
	View      uint64           `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	H         uint64           `protobuf:"varint,2,opt,name=h" json:"h,omitempty"`
	Cset      []*ViewChange_C  `protobuf:"bytes,3,rep,name=cset" json:"cset,omitempty"`
	Pset      []*ViewChange_PQ `protobuf:"bytes,4,rep,name=pset" json:"pset,omitempty"`
	Qset      []*ViewChange_PQ `protobuf:"bytes,5,rep,name=qset" json:"qset,omitempty"`
	ReplicaId uint64           `protobuf:"varint,6,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Signature []byte           `protobuf:"bytes,7,opt,name=signature,proto3" json:"signature,omitempty"`
}

func (m *ViewChange) Reset()                    { *m = ViewChange{} }
func (m *ViewChange) String() string            { return proto.CompactTextString(m) }
func (*ViewChange) ProtoMessage()               {}
func (*ViewChange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

func (m *ViewChange) GetCset() []*ViewChange_C {
	if m != nil {
		return m.Cset
	}
	return nil
}

func (m *ViewChange) GetPset() []*ViewChange_PQ {
	if m != nil {
		return m.Pset
	}
	return nil
}

func (m *ViewChange) GetQset() []*ViewChange_PQ {
	if m != nil {
		return m.Qset
	}
	return nil
}

// This message should go away and become a checkpoint once replica_id is removed
type ViewChange_C struct {
	SequenceNumber uint64 `protobuf:"varint,1,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	Id             string `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
}

func (m *ViewChange_C) Reset()                    { *m = ViewChange_C{} }
func (m *ViewChange_C) String() string            { return proto.CompactTextString(m) }
func (*ViewChange_C) ProtoMessage()               {}
func (*ViewChange_C) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 0} }

type ViewChange_PQ struct {
	SequenceNumber uint64 `protobuf:"varint,1,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	BatchDigest    string `protobuf:"bytes,2,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	View           uint64 `protobuf:"varint,3,opt,name=view" json:"view,omitempty"`
}

func (m *ViewChange_PQ) Reset()                    { *m = ViewChange_PQ{} }
func (m *ViewChange_PQ) String() string            { return proto.CompactTextString(m) }
func (*ViewChange_PQ) ProtoMessage()               {}
func (*ViewChange_PQ) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7, 1} }

type PQset struct {
	Set []*ViewChange_PQ `protobuf:"bytes,1,rep,name=set" json:"set,omitempty"`
}

func (m *PQset) Reset()                    { *m = PQset{} }
func (m *PQset) String() string            { return proto.CompactTextString(m) }
func (*PQset) ProtoMessage()               {}
func (*PQset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

func (m *PQset) GetSet() []*ViewChange_PQ {
	if m != nil {
		return m.Set
	}
	return nil
}

type NewView struct {
	View      uint64            `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	Vset      []*ViewChange     `protobuf:"bytes,2,rep,name=vset" json:"vset,omitempty"`
	Xset      map[uint64]string `protobuf:"bytes,3,rep,name=xset" json:"xset,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	ReplicaId uint64            `protobuf:"varint,4,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *NewView) Reset()                    { *m = NewView{} }
func (m *NewView) String() string            { return proto.CompactTextString(m) }
func (*NewView) ProtoMessage()               {}
func (*NewView) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{9} }

func (m *NewView) GetVset() []*ViewChange {
	if m != nil {
		return m.Vset
	}
	return nil
}

func (m *NewView) GetXset() map[uint64]string {
	if m != nil {
		return m.Xset
	}
	return nil
}

type FetchRequestBatch struct {
	BatchDigest string `protobuf:"bytes,1,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	ReplicaId   uint64 `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *FetchRequestBatch) Reset()                    { *m = FetchRequestBatch{} }
func (m *FetchRequestBatch) String() string            { return proto.CompactTextString(m) }
func (*FetchRequestBatch) ProtoMessage()               {}
func (*FetchRequestBatch) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

type NegotiateView struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *NegotiateView) Reset()                    { *m = NegotiateView{} }
func (m *NegotiateView) String() string            { return proto.CompactTextString(m) }
func (*NegotiateView) ProtoMessage()               {}
func (*NegotiateView) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

type NegotiateViewResponse struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	View      uint64 `protobuf:"varint,2,opt,name=view" json:"view,omitempty"`
}

func (m *NegotiateViewResponse) Reset()                    { *m = NegotiateViewResponse{} }
func (m *NegotiateViewResponse) String() string            { return proto.CompactTextString(m) }
func (*NegotiateViewResponse) ProtoMessage()               {}
func (*NegotiateViewResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type RecoveryInit struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *RecoveryInit) Reset()                    { *m = RecoveryInit{} }
func (m *RecoveryInit) String() string            { return proto.CompactTextString(m) }
func (*RecoveryInit) ProtoMessage()               {}
func (*RecoveryInit) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

type RecoveryResponse struct {
	ReplicaId     uint64            `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Chkpts        map[uint64]string `protobuf:"bytes,2,rep,name=chkpts" json:"chkpts,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	BlockHeight   uint64            `protobuf:"varint,3,opt,name=blockHeight" json:"blockHeight,omitempty"`
	LastBlockHash string            `protobuf:"bytes,4,opt,name=lastBlockHash" json:"lastBlockHash,omitempty"`
}

func (m *RecoveryResponse) Reset()                    { *m = RecoveryResponse{} }
func (m *RecoveryResponse) String() string            { return proto.CompactTextString(m) }
func (*RecoveryResponse) ProtoMessage()               {}
func (*RecoveryResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

func (m *RecoveryResponse) GetChkpts() map[uint64]string {
	if m != nil {
		return m.Chkpts
	}
	return nil
}

type RecoveryFetchPQC struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	H         uint64 `protobuf:"varint,2,opt,name=h" json:"h,omitempty"`
}

func (m *RecoveryFetchPQC) Reset()                    { *m = RecoveryFetchPQC{} }
func (m *RecoveryFetchPQC) String() string            { return proto.CompactTextString(m) }
func (*RecoveryFetchPQC) ProtoMessage()               {}
func (*RecoveryFetchPQC) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

type RecoveryReturnPQC struct {
	ReplicaId uint64        `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	PrepreSet []*PrePrepare `protobuf:"bytes,2,rep,name=prepre_set,json=prepreSet" json:"prepre_set,omitempty"`
	PreSent   []bool        `protobuf:"varint,3,rep,packed,name=pre_sent,json=preSent" json:"pre_sent,omitempty"`
	CmtSent   []bool        `protobuf:"varint,4,rep,packed,name=cmt_sent,json=cmtSent" json:"cmt_sent,omitempty"`
}

func (m *RecoveryReturnPQC) Reset()                    { *m = RecoveryReturnPQC{} }
func (m *RecoveryReturnPQC) String() string            { return proto.CompactTextString(m) }
func (*RecoveryReturnPQC) ProtoMessage()               {}
func (*RecoveryReturnPQC) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

func (m *RecoveryReturnPQC) GetPrepreSet() []*PrePrepare {
	if m != nil {
		return m.PrepreSet
	}
	return nil
}

// consensus metadata
type Metadata struct {
	SeqNo uint64 `protobuf:"varint,1,opt,name=seqNo" json:"seqNo,omitempty"`
}

func (m *Metadata) Reset()                    { *m = Metadata{} }
func (m *Metadata) String() string            { return proto.CompactTextString(m) }
func (*Metadata) ProtoMessage()               {}
func (*Metadata) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func init() {
	proto.RegisterType((*ConsensusMessage)(nil), "pbft.consensus_message")
	proto.RegisterType((*TransactionBatch)(nil), "pbft.transaction_batch")
	proto.RegisterType((*PrePrepare)(nil), "pbft.pre_prepare")
	proto.RegisterType((*Prepare)(nil), "pbft.prepare")
	proto.RegisterType((*Commit)(nil), "pbft.commit")
	proto.RegisterType((*BlockInfo)(nil), "pbft.block_info")
	proto.RegisterType((*Checkpoint)(nil), "pbft.checkpoint")
	proto.RegisterType((*ViewChange)(nil), "pbft.view_change")
	proto.RegisterType((*ViewChange_C)(nil), "pbft.view_change.C")
	proto.RegisterType((*ViewChange_PQ)(nil), "pbft.view_change.PQ")
	proto.RegisterType((*PQset)(nil), "pbft.PQset")
	proto.RegisterType((*NewView)(nil), "pbft.new_view")
	proto.RegisterType((*FetchRequestBatch)(nil), "pbft.fetch_request_batch")
	proto.RegisterType((*NegotiateView)(nil), "pbft.negotiate_view")
	proto.RegisterType((*NegotiateViewResponse)(nil), "pbft.negotiate_view_response")
	proto.RegisterType((*RecoveryInit)(nil), "pbft.recovery_init")
	proto.RegisterType((*RecoveryResponse)(nil), "pbft.recovery_response")
	proto.RegisterType((*RecoveryFetchPQC)(nil), "pbft.recovery_fetchPQC")
	proto.RegisterType((*RecoveryReturnPQC)(nil), "pbft.recovery_returnPQC")
	proto.RegisterType((*Metadata)(nil), "pbft.metadata")
	proto.RegisterEnum("pbft.ConsensusMessage_Type", ConsensusMessage_Type_name, ConsensusMessage_Type_value)
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1053 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xc4, 0x56, 0xcb, 0x72, 0xe3, 0x44,
	0x17, 0xfe, 0x25, 0xcb, 0x89, 0x7d, 0x94, 0x38, 0x72, 0x27, 0x3f, 0x31, 0x61, 0xa6, 0x2a, 0x88,
	0xcb, 0x64, 0x41, 0x29, 0x53, 0x61, 0xc1, 0x75, 0x81, 0xa3, 0xd1, 0x4c, 0x5c, 0x30, 0xb2, 0xdd,
	0xd1, 0xcc, 0x40, 0xb1, 0x50, 0x29, 0x72, 0x27, 0x52, 0x25, 0x96, 0x34, 0x52, 0x27, 0x90, 0x27,
	0x80, 0x0d, 0x4f, 0xc0, 0x03, 0xf0, 0x24, 0xac, 0xd8, 0xf1, 0x28, 0x3c, 0x01, 0x7d, 0x91, 0x1d,
	0x59, 0x4e, 0x4d, 0xc2, 0x06, 0x16, 0xae, 0xd2, 0xf9, 0xce, 0x77, 0x4e, 0x9f, 0x6b, 0xb7, 0xa1,
	0x33, 0x25, 0x45, 0x11, 0x9c, 0x91, 0xc2, 0xca, 0xf2, 0x94, 0xa6, 0x48, 0xcb, 0x4e, 0x4e, 0xe9,
	0xce, 0xa3, 0xe8, 0x3a, 0x23, 0x79, 0x18, 0x05, 0x71, 0xb2, 0x1f, 0xa6, 0x39, 0xd9, 0xa7, 0x4c,
	0x2e, 0xf6, 0x69, 0x1e, 0x24, 0x45, 0x10, 0xd2, 0x38, 0x4d, 0x24, 0xdd, 0xfc, 0xad, 0x01, 0xdd,
	0x30, 0x4d, 0x0a, 0x92, 0x14, 0x97, 0x85, 0x5f, 0xfa, 0x42, 0x8f, 0x41, 0xe3, 0x06, 0x3d, 0x65,
	0x57, 0xd9, 0xeb, 0x1c, 0x3c, 0xb0, 0xb8, 0x4f, 0x6b, 0x89, 0x66, 0x79, 0x8c, 0x83, 0x05, 0x13,
	0xf5, 0x60, 0x35, 0x0b, 0xae, 0x2f, 0xd2, 0x60, 0xd2, 0x53, 0x99, 0xd1, 0x1a, 0x9e, 0x89, 0xe6,
	0xef, 0x2a, 0x68, 0x9c, 0x88, 0x36, 0x40, 0xf7, 0x70, 0xdf, 0x3d, 0xee, 0xdb, 0xde, 0x60, 0xe8,
	0x1a, 0xff, 0x43, 0x5b, 0x60, 0x48, 0x80, 0xcb, 0xfe, 0x61, 0xdf, 0xb3, 0x8f, 0x0c, 0x85, 0xd3,
	0x46, 0xd8, 0xf1, 0xd9, 0x6f, 0xd4, 0xc7, 0x8e, 0xa1, 0x22, 0x1d, 0x56, 0x67, 0x42, 0x03, 0x01,
	0xac, 0xd8, 0xc3, 0xe7, 0xcf, 0x07, 0x9e, 0xa1, 0xa1, 0x0e, 0x80, 0x7d, 0xe4, 0xd8, 0x5f, 0x8f,
	0x86, 0x03, 0xd7, 0x33, 0x9a, 0xdc, 0xf2, 0xe5, 0xc0, 0x79, 0xe5, 0xdb, 0x47, 0x7d, 0xf7, 0x99,
	0x63, 0xac, 0xa0, 0x35, 0x68, 0xb9, 0x4c, 0xe6, 0xa0, 0xb1, 0x8a, 0xb6, 0x61, 0xf3, 0x29, 0x66,
	0x67, 0xf8, 0xd8, 0x19, 0xbf, 0x70, 0x8e, 0xbd, 0xf2, 0xc4, 0x16, 0x8b, 0x7d, 0x0b, 0x3b, 0xde,
	0x0b, 0xec, 0xd6, 0x34, 0x6d, 0x84, 0xa0, 0xe3, 0x3a, 0xcf, 0x86, 0xde, 0xa0, 0xef, 0x39, 0xd2,
	0x0d, 0xa0, 0x77, 0x60, 0x7b, 0x11, 0x63, 0x56, 0xc7, 0xa3, 0xa1, 0x7b, 0xec, 0x18, 0x3a, 0xea,
	0xc2, 0x3a, 0x76, 0xec, 0xe1, 0x4b, 0x07, 0x7f, 0xe7, 0x0f, 0x5c, 0x16, 0xe5, 0x1a, 0xfa, 0x3f,
	0x74, 0xe7, 0xd0, 0x9c, 0xb9, 0x8e, 0xde, 0x02, 0x34, 0x87, 0x9f, 0x3a, 0x3c, 0xac, 0xf1, 0xc8,
	0x36, 0x3a, 0x3c, 0xca, 0x0a, 0x5d, 0x44, 0xc5, 0x15, 0x1b, 0xe6, 0xf7, 0xd0, 0xad, 0xb4, 0xcf,
	0x3f, 0x09, 0x68, 0x18, 0xa1, 0x3d, 0x68, 0x8a, 0x0f, 0xd6, 0xa9, 0xc6, 0x9e, 0x7e, 0x80, 0x2c,
	0xd1, 0x67, 0xcb, 0xbb, 0x21, 0x62, 0x49, 0x40, 0x0f, 0xa0, 0x4d, 0x63, 0xd6, 0x38, 0x1a, 0x4c,
	0x33, 0xd1, 0xa2, 0x06, 0xbe, 0x01, 0xcc, 0x3f, 0x15, 0xd0, 0xb3, 0x9c, 0xf8, 0xec, 0x97, 0x05,
	0x39, 0x61, 0x89, 0x6b, 0x57, 0x31, 0xf9, 0x41, 0x0c, 0x80, 0x86, 0xc5, 0x37, 0x7a, 0x04, 0x1b,
	0x05, 0x79, 0x7d, 0x49, 0x92, 0x90, 0xf8, 0xc9, 0xe5, 0xf4, 0x84, 0xe4, 0xc2, 0x8f, 0x86, 0x3b,
	0x33, 0xd8, 0x15, 0x28, 0x7a, 0x17, 0xd6, 0xc4, 0x99, 0xfe, 0x24, 0x66, 0x73, 0x49, 0x7b, 0x0d,
	0xc6, 0x6a, 0x63, 0x5d, 0x60, 0x4f, 0x04, 0x84, 0x9e, 0xdc, 0x92, 0x4c, 0x4f, 0x63, 0x3c, 0xfd,
	0x60, 0x5b, 0x4e, 0xdb, 0x92, 0x1a, 0x1b, 0x15, 0xe8, 0x50, 0xe4, 0xf4, 0x10, 0x80, 0xc5, 0x7b,
	0x11, 0x87, 0x81, 0x1f, 0x4f, 0x7a, 0x4d, 0x11, 0x4c, 0xbb, 0x44, 0x06, 0x13, 0xf3, 0x67, 0x85,
	0x0d, 0xe5, 0xbf, 0x94, 0xd0, 0x62, 0x28, 0x5a, 0x3d, 0x94, 0x9f, 0x14, 0x58, 0x09, 0xd3, 0xe9,
	0x34, 0xa6, 0xff, 0x75, 0x24, 0x2e, 0xc0, 0xc9, 0x45, 0x1a, 0x9e, 0xfb, 0x71, 0x72, 0x9a, 0x0a,
	0x7f, 0x42, 0x2a, 0x4f, 0x95, 0x41, 0xe9, 0x02, 0x2b, 0x8f, 0x7c, 0x38, 0x33, 0x88, 0x82, 0x22,
	0x2a, 0x97, 0xbb, 0x2d, 0x90, 0x23, 0x06, 0x98, 0x13, 0x80, 0x30, 0x22, 0xe1, 0x79, 0x96, 0xc6,
	0x09, 0xbd, 0x2d, 0x11, 0xe5, 0xd6, 0x44, 0x16, 0xa3, 0x54, 0x6b, 0x51, 0xb2, 0xd5, 0x56, 0x19,
	0x2c, 0xb3, 0x63, 0x5f, 0xe6, 0x2f, 0x0d, 0xd0, 0x79, 0xa5, 0x7c, 0x76, 0xa5, 0x25, 0x67, 0xb7,
	0xb7, 0x73, 0x0d, 0x94, 0xa8, 0xf4, 0xa4, 0x44, 0x2c, 0x12, 0x2d, 0x2c, 0x08, 0xaf, 0x10, 0x5f,
	0x8c, 0x4d, 0x39, 0x54, 0x15, 0x17, 0x96, 0x8d, 0x05, 0x81, 0xad, 0x90, 0x96, 0x71, 0xa2, 0x26,
	0x88, 0x5b, 0xcb, 0xc4, 0xd1, 0x18, 0x0b, 0x06, 0x67, 0xbe, 0xe6, 0xcc, 0xe6, 0x9b, 0x98, 0x9c,
	0x51, 0xcb, 0x6e, 0xa5, 0x9e, 0x1d, 0xdb, 0xc5, 0x22, 0x3e, 0x4b, 0x02, 0x7a, 0x99, 0x93, 0xde,
	0xaa, 0xac, 0xe8, 0x1c, 0xd8, 0xf9, 0x12, 0x14, 0xfb, 0xfe, 0x85, 0xac, 0x55, 0x6a, 0x67, 0x02,
	0xea, 0x68, 0x7c, 0x7f, 0xf3, 0xfa, 0x40, 0xa9, 0xcb, 0x03, 0x35, 0xab, 0x75, 0xe3, 0xa6, 0xd6,
	0xe6, 0x3e, 0x34, 0x47, 0x63, 0x9e, 0xe9, 0x87, 0xd0, 0xe0, 0x25, 0x51, 0xde, 0x50, 0x12, 0x4e,
	0x30, 0xff, 0x50, 0xa0, 0x95, 0x30, 0x58, 0x74, 0xea, 0xb6, 0xee, 0x7d, 0xc0, 0x30, 0xee, 0x49,
	0x15, 0x9e, 0xba, 0x4b, 0x9e, 0xb0, 0x50, 0xa3, 0x8f, 0x40, 0xfb, 0xf1, 0xa6, 0xad, 0x3d, 0x49,
	0x9b, 0x39, 0xb6, 0xbe, 0x65, 0x2a, 0x27, 0xa1, 0xf9, 0x35, 0x16, 0xac, 0x3b, 0x76, 0x61, 0xe7,
	0x13, 0x68, 0xcf, 0x2d, 0x90, 0x01, 0x8d, 0x73, 0x72, 0x5d, 0xc6, 0xc4, 0x3f, 0xd9, 0xfb, 0xd4,
	0xbc, 0x0a, 0x2e, 0x2e, 0x49, 0x59, 0x14, 0x29, 0x7c, 0xae, 0x7e, 0xaa, 0x98, 0xaf, 0x60, 0xf3,
	0x94, 0xf0, 0xaa, 0xe5, 0xbc, 0x9a, 0x05, 0x2d, 0x6f, 0xe3, 0x7a, 0x31, 0x95, 0xbb, 0xb6, 0xb3,
	0x3e, 0xf7, 0xac, 0xae, 0x9d, 0x84, 0x9c, 0xa5, 0x34, 0x0e, 0x28, 0x91, 0xb5, 0x5a, 0x34, 0x50,
	0xea, 0x06, 0xdf, 0xc0, 0xf6, 0xa2, 0x01, 0x0b, 0xa9, 0xc8, 0xf8, 0x4b, 0x7d, 0x87, 0xe5, 0xbc,
	0x09, 0x6a, 0xa5, 0xad, 0x16, 0xac, 0xe7, 0x24, 0x4c, 0xaf, 0x48, 0x7e, 0xcd, 0xee, 0x87, 0x98,
	0xde, 0x75, 0xfa, 0x5f, 0x0a, 0x74, 0xe7, 0x06, 0xf7, 0x3d, 0xf8, 0x0b, 0x76, 0x15, 0x46, 0xe7,
	0x19, 0x2d, 0xca, 0x5e, 0xbf, 0x27, 0x9b, 0xb8, 0xe4, 0xc7, 0xb2, 0x05, 0x4b, 0xf6, 0xb3, 0x34,
	0x41, 0xbb, 0x20, 0x2f, 0xa7, 0x23, 0x12, 0x9f, 0x45, 0xb4, 0x9c, 0xc9, 0x2a, 0x84, 0xde, 0x87,
	0xf5, 0x8b, 0xa0, 0xa0, 0x87, 0xb3, 0x1b, 0x4a, 0xb4, 0xbd, 0x8d, 0x17, 0xc1, 0x9d, 0xcf, 0x40,
	0xaf, 0xb8, 0xff, 0x47, 0xcd, 0xff, 0xaa, 0x92, 0xb3, 0x98, 0x82, 0xd1, 0xd8, 0xbe, 0x2b, 0xe7,
	0x85, 0xbb, 0xc9, 0xfc, 0x55, 0x01, 0x54, 0x49, 0x97, 0x6d, 0x7d, 0x72, 0x0f, 0x1f, 0x8f, 0x01,
	0xf8, 0x6b, 0xc6, 0x5e, 0xe9, 0xa5, 0x3d, 0xa9, 0x3c, 0xdd, 0xb8, 0x2d, 0x49, 0xc7, 0x6c, 0xfc,
	0xdf, 0x86, 0x96, 0xa4, 0x27, 0x72, 0x61, 0x5a, 0x78, 0x55, 0x68, 0x12, 0xa1, 0x0a, 0xa7, 0x54,
	0xaa, 0x34, 0xa9, 0x62, 0x32, 0x57, 0x99, 0xbb, 0xd0, 0x9a, 0x12, 0x1a, 0x4c, 0x02, 0x1a, 0xf0,
	0x2a, 0xb0, 0x0b, 0xc3, 0x4d, 0xcb, 0x68, 0xa4, 0x70, 0xb2, 0x22, 0xfe, 0x3b, 0x7e, 0xfc, 0x77,
	0x00, 0x00, 0x00, 0xff, 0xff, 0x2d, 0x39, 0xaa, 0x43, 0x7c, 0x0a, 0x00, 0x00,
}
