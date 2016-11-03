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
	AddNode
	DelNode
	ReadyForN
	UpdateN
	AgreeUpdateN
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
	ConsensusMessage_ADD_NODE                ConsensusMessage_Type = 16
	ConsensusMessage_DEL_NODE                ConsensusMessage_Type = 17
	ConsensusMessage_READY_FOR_N             ConsensusMessage_Type = 18
	ConsensusMessage_UPDATE_N                ConsensusMessage_Type = 19
	ConsensusMessage_AGREE_UPDATE_N          ConsensusMessage_Type = 20
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
	16: "ADD_NODE",
	17: "DEL_NODE",
	18: "READY_FOR_N",
	19: "UPDATE_N",
	20: "AGREE_UPDATE_N",
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
	"ADD_NODE":                16,
	"DEL_NODE":                17,
	"READY_FOR_N":             18,
	"UPDATE_N":                19,
	"AGREE_UPDATE_N":          20,
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
	ReplicaId uint64            `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Chkpts    map[uint64]string `protobuf:"bytes,2,rep,name=chkpts" json:"chkpts,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
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

type AddNode struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
}

func (m *AddNode) Reset()                    { *m = AddNode{} }
func (m *AddNode) String() string            { return proto.CompactTextString(m) }
func (*AddNode) ProtoMessage()               {}
func (*AddNode) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

type DelNode struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
}

func (m *DelNode) Reset()                    { *m = DelNode{} }
func (m *DelNode) String() string            { return proto.CompactTextString(m) }
func (*DelNode) ProtoMessage()               {}
func (*DelNode) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

type ReadyForN struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
}

func (m *ReadyForN) Reset()                    { *m = ReadyForN{} }
func (m *ReadyForN) String() string            { return proto.CompactTextString(m) }
func (*ReadyForN) ProtoMessage()               {}
func (*ReadyForN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

type UpdateN struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	N         uint64 `protobuf:"varint,3,opt,name=n" json:"n,omitempty"`
	SeqNo     uint64 `protobuf:"varint,4,opt,name=seqNo" json:"seqNo,omitempty"`
	View      uint64 `protobuf:"varint,5,opt,name=view" json:"view,omitempty"`
	Flag      bool   `protobuf:"varint,6,opt,name=flag" json:"flag,omitempty"`
}

func (m *UpdateN) Reset()                    { *m = UpdateN{} }
func (m *UpdateN) String() string            { return proto.CompactTextString(m) }
func (*UpdateN) ProtoMessage()               {}
func (*UpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

type AgreeUpdateN struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	N         uint64 `protobuf:"varint,3,opt,name=n" json:"n,omitempty"`
	SeqNo     uint64 `protobuf:"varint,4,opt,name=seqNo" json:"seqNo,omitempty"`
	View      uint64 `protobuf:"varint,5,opt,name=view" json:"view,omitempty"`
	Flag      bool   `protobuf:"varint,6,opt,name=flag" json:"flag,omitempty"`
}

func (m *AgreeUpdateN) Reset()                    { *m = AgreeUpdateN{} }
func (m *AgreeUpdateN) String() string            { return proto.CompactTextString(m) }
func (*AgreeUpdateN) ProtoMessage()               {}
func (*AgreeUpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

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
	proto.RegisterType((*AddNode)(nil), "pbft.add_node")
	proto.RegisterType((*DelNode)(nil), "pbft.del_node")
	proto.RegisterType((*ReadyForN)(nil), "pbft.ready_for_n")
	proto.RegisterType((*UpdateN)(nil), "pbft.update_n")
	proto.RegisterType((*AgreeUpdateN)(nil), "pbft.agree_update_n")
	proto.RegisterEnum("pbft.ConsensusMessage_Type", ConsensusMessage_Type_name, ConsensusMessage_Type_value)
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1144 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xcc, 0x56, 0x4f, 0x73, 0xdb, 0x44,
	0x14, 0x47, 0xb6, 0x9c, 0xd8, 0xcf, 0xa9, 0x2b, 0x6f, 0x02, 0x31, 0xa1, 0x9d, 0x01, 0x31, 0xd0,
	0x1e, 0x18, 0xa7, 0x13, 0x0e, 0xfc, 0x29, 0xc3, 0xe0, 0xca, 0x6a, 0xe2, 0xa1, 0x95, 0xed, 0x8d,
	0xd2, 0xc2, 0x70, 0xd0, 0x28, 0xd2, 0xc6, 0xd6, 0xc4, 0x96, 0x54, 0x49, 0x0e, 0xe4, 0x13, 0xc0,
	0xa5, 0x37, 0x6e, 0x7c, 0x03, 0xbe, 0x0a, 0x37, 0x3e, 0x0c, 0x67, 0xde, 0xee, 0xca, 0x8e, 0xff,
	0x64, 0xea, 0x84, 0x03, 0x70, 0xf0, 0x8c, 0xf6, 0xb7, 0xbf, 0xf7, 0xf6, 0xf7, 0xfe, 0xec, 0x5b,
	0x43, 0x6d, 0xcc, 0xd2, 0xd4, 0x1d, 0xb0, 0xb4, 0x19, 0x27, 0x51, 0x16, 0x11, 0x35, 0x3e, 0x3d,
	0xcb, 0xf6, 0x1e, 0x0c, 0x2f, 0x63, 0x96, 0x78, 0x43, 0x37, 0x08, 0xf7, 0xbd, 0x28, 0x61, 0xfb,
	0x19, 0xae, 0xd3, 0xfd, 0x2c, 0x71, 0xc3, 0xd4, 0xf5, 0xb2, 0x20, 0x0a, 0x25, 0x5d, 0xff, 0xab,
	0x08, 0x75, 0x2f, 0x0a, 0x53, 0x16, 0xa6, 0x93, 0xd4, 0xc9, 0x7d, 0x91, 0x47, 0xa0, 0x72, 0x83,
	0x86, 0xf2, 0xbe, 0xf2, 0xb0, 0x76, 0x70, 0xaf, 0xc9, 0x7d, 0x36, 0x57, 0x68, 0x4d, 0x1b, 0x39,
	0x54, 0x30, 0x49, 0x03, 0x36, 0x63, 0xf7, 0x72, 0x14, 0xb9, 0x7e, 0xa3, 0x80, 0x46, 0x5b, 0x74,
	0xba, 0xd4, 0x7f, 0x2d, 0x82, 0xca, 0x89, 0xe4, 0x2e, 0x54, 0x6d, 0xda, 0xb2, 0x8e, 0x5b, 0x86,
	0xdd, 0xe9, 0x5a, 0xda, 0x5b, 0x64, 0x07, 0x34, 0x09, 0xf0, 0xb5, 0xf3, 0xa4, 0x65, 0x1b, 0x47,
	0x9a, 0xc2, 0x69, 0x3d, 0x6a, 0x3a, 0xf8, 0xeb, 0xb5, 0xa8, 0xa9, 0x15, 0x48, 0x15, 0x36, 0xa7,
	0x8b, 0x22, 0x01, 0xd8, 0x30, 0xba, 0xcf, 0x9f, 0x77, 0x6c, 0x4d, 0x25, 0x35, 0x00, 0xe3, 0xc8,
	0x34, 0xbe, 0xed, 0x75, 0x3b, 0x96, 0xad, 0x95, 0xb8, 0xe5, 0x8b, 0x8e, 0xf9, 0xd2, 0x31, 0x8e,
	0x5a, 0xd6, 0xa1, 0xa9, 0x6d, 0x90, 0x2d, 0x28, 0x5b, 0xb8, 0xe6, 0xa0, 0xb6, 0x49, 0x76, 0x61,
	0xfb, 0x29, 0xc5, 0x33, 0x1c, 0x6a, 0xf6, 0x4f, 0xcc, 0x63, 0x3b, 0x3f, 0xb1, 0x8c, 0xda, 0x77,
	0xa8, 0x69, 0x9f, 0x50, 0x6b, 0x69, 0xa7, 0x42, 0x08, 0xd4, 0x2c, 0xf3, 0xb0, 0x6b, 0x77, 0x5a,
	0xb6, 0x29, 0xdd, 0x00, 0x79, 0x0f, 0x76, 0x17, 0x31, 0xb4, 0x3a, 0xee, 0x75, 0xad, 0x63, 0x53,
	0xab, 0x92, 0x3a, 0xdc, 0xa1, 0xa6, 0xd1, 0x7d, 0x61, 0xd2, 0xef, 0x9d, 0x8e, 0x85, 0x2a, 0xb7,
	0xc8, 0xdb, 0x50, 0x9f, 0x41, 0x33, 0xe6, 0x1d, 0xf2, 0x0e, 0x90, 0x19, 0xfc, 0xd4, 0xe4, 0xb2,
	0xfa, 0x3d, 0x43, 0xab, 0x71, 0x95, 0x73, 0x74, 0xa1, 0x8a, 0x6f, 0xdc, 0xe5, 0xc1, 0xb4, 0xda,
	0x6d, 0xc7, 0xea, 0xb6, 0x4d, 0x4d, 0xe3, 0xab, 0xb6, 0xf9, 0x4c, 0xae, 0xea, 0x3c, 0x72, 0x6a,
	0xb6, 0xda, 0xe8, 0xa9, 0x4b, 0x1d, 0x4b, 0x23, 0x7c, 0xfb, 0xa4, 0xd7, 0xe6, 0x0a, 0x2d, 0x6d,
	0x9b, 0x87, 0xd1, 0x3a, 0xa4, 0xa6, 0xe9, 0xcc, 0xb0, 0x1d, 0xfd, 0x07, 0xa8, 0xcf, 0x75, 0x83,
	0x73, 0xea, 0x66, 0xde, 0x90, 0x3c, 0x84, 0x92, 0xf8, 0xc0, 0xc2, 0x17, 0x1f, 0x56, 0x0f, 0x48,
	0x53, 0xb4, 0x4d, 0xd3, 0xbe, 0x22, 0x52, 0x49, 0x20, 0xf7, 0xa0, 0x92, 0x05, 0xd8, 0x07, 0x99,
	0x3b, 0x8e, 0x45, 0xc5, 0x8b, 0xf4, 0x0a, 0xd0, 0xff, 0x54, 0xa0, 0x1a, 0x27, 0xcc, 0xc1, 0x5f,
	0xec, 0x26, 0x0c, 0x05, 0xa8, 0x17, 0x01, 0xfb, 0x51, 0xf4, 0x93, 0x4a, 0xc5, 0x37, 0x79, 0x00,
	0x77, 0x53, 0xf6, 0x6a, 0xc2, 0x42, 0x8f, 0x39, 0xe1, 0x64, 0x7c, 0xca, 0x12, 0xe1, 0x47, 0xa5,
	0xb5, 0x29, 0x6c, 0x09, 0x94, 0x7c, 0x00, 0x5b, 0xe2, 0x4c, 0xc7, 0x0f, 0xb0, 0xcd, 0xb3, 0x46,
	0x11, 0x59, 0x15, 0x5a, 0x15, 0x58, 0x5b, 0x40, 0xa4, 0x7d, 0x4d, 0x30, 0x0d, 0x15, 0x79, 0xd5,
	0x83, 0x5d, 0xd9, 0xbc, 0x2b, 0xdb, 0x54, 0x9b, 0x83, 0x9e, 0x88, 0x98, 0xee, 0x03, 0xa0, 0xde,
	0x51, 0xe0, 0xb9, 0x4e, 0xe0, 0x37, 0x4a, 0x42, 0x4c, 0x25, 0x47, 0x3a, 0xbe, 0xfe, 0x8b, 0x82,
	0x3d, 0xfe, 0x2f, 0x05, 0xb4, 0x28, 0x45, 0x5d, 0x96, 0xf2, 0xb3, 0x02, 0x1b, 0x5e, 0x34, 0x1e,
	0x07, 0xd9, 0x7f, 0xad, 0xc4, 0x02, 0x38, 0x1d, 0x45, 0xde, 0xb9, 0x13, 0x84, 0x67, 0x91, 0xf0,
	0x27, 0x56, 0xf9, 0xa9, 0x52, 0x54, 0x55, 0x60, 0xf9, 0x91, 0xf7, 0xa7, 0x06, 0x43, 0x37, 0x1d,
	0xe6, 0xb3, 0xa2, 0x22, 0x90, 0x23, 0x04, 0x74, 0x1f, 0xc0, 0x1b, 0x32, 0xef, 0x3c, 0x8e, 0x82,
	0x30, 0xbb, 0x2e, 0x10, 0xe5, 0xda, 0x40, 0x16, 0x55, 0x16, 0x96, 0x54, 0xe2, 0xa4, 0x28, 0x20,
	0x2c, 0xa3, 0xc3, 0x2f, 0xfd, 0x75, 0x11, 0xaa, 0x3c, 0x53, 0x0e, 0x4e, 0xc8, 0x70, 0x70, 0x7d,
	0x39, 0xb7, 0x40, 0x19, 0xe6, 0x9e, 0x94, 0x21, 0x2a, 0x51, 0xbd, 0x94, 0xf1, 0x0c, 0xf1, 0x8b,
	0xb1, 0x2d, 0x9b, 0x6a, 0xce, 0x45, 0xd3, 0xa0, 0x82, 0x80, 0x57, 0x48, 0x8d, 0x39, 0x51, 0x15,
	0xc4, 0x9d, 0x55, 0x62, 0xaf, 0x4f, 0x05, 0x83, 0x33, 0x5f, 0x71, 0x66, 0xe9, 0x4d, 0x4c, 0xce,
	0x58, 0x8a, 0x6e, 0x63, 0x39, 0x3a, 0xbc, 0x8b, 0x69, 0x30, 0x08, 0xdd, 0x6c, 0x92, 0xb0, 0xc6,
	0xa6, 0xcc, 0xe8, 0x0c, 0xd8, 0xfb, 0x0a, 0x14, 0xe3, 0xe6, 0x89, 0x5c, 0xca, 0xd4, 0x9e, 0x0f,
	0x85, 0x5e, 0xff, 0xe6, 0xe6, 0xcb, 0x0d, 0x55, 0x58, 0x6d, 0xa8, 0x69, 0xae, 0x8b, 0x57, 0xb9,
	0xd6, 0xf7, 0xa1, 0xd4, 0xeb, 0xf3, 0x48, 0x3f, 0x86, 0x22, 0x4f, 0x89, 0xf2, 0x86, 0x94, 0x70,
	0x82, 0xfe, 0x87, 0x02, 0xe5, 0x10, 0x61, 0x51, 0xa9, 0xeb, 0xaa, 0xf7, 0x11, 0x62, 0xdc, 0x53,
	0x41, 0x78, 0xaa, 0xaf, 0x78, 0xa2, 0x62, 0x9b, 0x7c, 0x02, 0xea, 0x4f, 0x57, 0x65, 0x6d, 0x48,
	0xda, 0xd4, 0x71, 0xf3, 0x3b, 0xdc, 0x32, 0xc3, 0x2c, 0xb9, 0xa4, 0x82, 0xb5, 0xe6, 0x2e, 0xec,
	0x7d, 0x06, 0x95, 0x99, 0x05, 0xd1, 0xa0, 0x78, 0xce, 0x2e, 0x73, 0x4d, 0xfc, 0x13, 0x9f, 0xbb,
	0xd2, 0x85, 0x3b, 0x9a, 0xb0, 0x3c, 0x29, 0x72, 0xf1, 0x65, 0xe1, 0x73, 0x45, 0x7f, 0x09, 0xdb,
	0x67, 0x8c, 0x67, 0x2d, 0xe1, 0xd9, 0x4c, 0xb3, 0x7c, 0x1a, 0x2f, 0x27, 0x53, 0x59, 0x77, 0x3b,
	0x97, 0xfb, 0x1e, 0xf3, 0x5a, 0x0b, 0xd9, 0x20, 0xca, 0x02, 0x37, 0x63, 0x32, 0x57, 0x8b, 0x06,
	0xca, 0xb2, 0xc1, 0x33, 0xd8, 0x5d, 0x34, 0x40, 0x49, 0x69, 0xcc, 0x1f, 0xfe, 0x35, 0x96, 0xb3,
	0x22, 0x14, 0xe6, 0xca, 0xda, 0x84, 0x3b, 0x09, 0xf3, 0xa2, 0x0b, 0x96, 0x5c, 0xe2, 0x7c, 0x08,
	0xb2, 0x75, 0xa7, 0xff, 0xae, 0x40, 0x7d, 0x66, 0x70, 0xd3, 0x83, 0x1f, 0xe3, 0x28, 0x1c, 0x9e,
	0xc7, 0x59, 0x9a, 0xd7, 0xfa, 0x43, 0x59, 0xc4, 0x15, 0x3f, 0x4d, 0x43, 0xb0, 0x64, 0x3d, 0x73,
	0x93, 0xbd, 0x2f, 0xa0, 0x3a, 0x07, 0xdf, 0xaa, 0x68, 0xdf, 0xcc, 0x69, 0x15, 0xd5, 0xeb, 0xf5,
	0x8d, 0x75, 0x5a, 0x17, 0x66, 0x8a, 0xfe, 0x9b, 0x02, 0x64, 0x4e, 0x26, 0xde, 0xd6, 0xf0, 0x06,
	0x3e, 0x1e, 0x01, 0xf0, 0x57, 0x08, 0x5f, 0xd7, 0x95, 0xfe, 0x9e, 0x7b, 0x72, 0x69, 0x45, 0x92,
	0x8e, 0xb1, 0x6d, 0xdf, 0x85, 0xb2, 0xa4, 0x87, 0xb2, 0xd1, 0xcb, 0x74, 0x53, 0xec, 0x84, 0x62,
	0xcb, 0x1b, 0x67, 0x72, 0x4b, 0x95, 0x5b, 0xb8, 0xe6, 0x5b, 0xfa, 0x63, 0x28, 0xbb, 0xbe, 0xef,
	0x84, 0x91, 0xbf, 0xb6, 0x04, 0x79, 0xda, 0x64, 0x8a, 0xf8, 0x27, 0x37, 0xf6, 0xd9, 0xe8, 0x1f,
	0x1a, 0x7f, 0x0d, 0xd5, 0x84, 0xb9, 0x3e, 0xa6, 0x35, 0x4a, 0x9c, 0xf0, 0xf6, 0xf6, 0xf8, 0x3a,
	0x96, 0x27, 0xb1, 0xcf, 0x5b, 0xf8, 0xf6, 0xd6, 0xbc, 0x46, 0x61, 0x3e, 0x9c, 0x94, 0x90, 0xd7,
	0x1f, 0x47, 0x9c, 0x15, 0xe5, 0xb7, 0x5d, 0x2e, 0x66, 0xcd, 0x5e, 0x9a, 0x9b, 0x38, 0x88, 0x9d,
	0x8d, 0xdc, 0x81, 0x18, 0xcf, 0x65, 0x2a, 0xbe, 0xf5, 0xd7, 0x0a, 0xd4, 0xdc, 0x41, 0xc2, 0x98,
	0xf3, 0xbf, 0xd0, 0x73, 0xba, 0x21, 0xfe, 0xf4, 0x7f, 0xfa, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff,
	0x10, 0x84, 0x77, 0xcc, 0x35, 0x0c, 0x00, 0x00,
}
