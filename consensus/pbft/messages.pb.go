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
	LastExec  uint64            `protobuf:"varint,3,opt,name=lastExec" json:"lastExec,omitempty"`
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
	N         int64  `protobuf:"varint,3,opt,name=n" json:"n,omitempty"`
	View      uint64 `protobuf:"varint,5,opt,name=view" json:"view,omitempty"`
	SeqNo     uint64 `protobuf:"varint,6,opt,name=seqNo" json:"seqNo,omitempty"`
	Flag      bool   `protobuf:"varint,7,opt,name=flag" json:"flag,omitempty"`
}

func (m *UpdateN) Reset()                    { *m = UpdateN{} }
func (m *UpdateN) String() string            { return proto.CompactTextString(m) }
func (*UpdateN) ProtoMessage()               {}
func (*UpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

type AgreeUpdateN struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	N         int64  `protobuf:"varint,3,opt,name=n" json:"n,omitempty"`
	View      uint64 `protobuf:"varint,5,opt,name=view" json:"view,omitempty"`
	SeqNo     uint64 `protobuf:"varint,6,opt,name=seqNo" json:"seqNo,omitempty"`
	Flag      bool   `protobuf:"varint,7,opt,name=flag" json:"flag,omitempty"`
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
	// 1145 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xcc, 0x56, 0xcd, 0x72, 0xe3, 0x44,
	0x10, 0x46, 0xb6, 0x9c, 0xd8, 0xed, 0xac, 0x57, 0x9e, 0x04, 0x62, 0xc2, 0x6e, 0x15, 0x88, 0x82,
	0xdd, 0x03, 0xe5, 0x6c, 0x85, 0x03, 0x3f, 0x4b, 0x51, 0x78, 0x65, 0x6d, 0xe2, 0x62, 0x57, 0xb6,
	0x27, 0xca, 0x2e, 0x14, 0x07, 0x95, 0x22, 0x4d, 0x6c, 0x55, 0x6c, 0x49, 0x2b, 0xc9, 0x81, 0x3c,
	0x01, 0x5c, 0xf6, 0xc6, 0x8d, 0x37, 0xe0, 0x55, 0xb8, 0xf1, 0x30, 0x9c, 0xe9, 0x99, 0x91, 0x1d,
	0xff, 0xa4, 0xd6, 0x09, 0x07, 0xe0, 0xe0, 0x2a, 0x75, 0xf7, 0xd7, 0x3d, 0x5f, 0xff, 0x4c, 0x8f,
	0xa1, 0x36, 0x66, 0x69, 0xea, 0x0e, 0x58, 0xda, 0x8c, 0x93, 0x28, 0x8b, 0x88, 0x1a, 0x9f, 0x9e,
	0x65, 0x7b, 0x0f, 0x86, 0x97, 0x31, 0x4b, 0xbc, 0xa1, 0x1b, 0x84, 0xfb, 0x5e, 0x94, 0xb0, 0xfd,
	0x0c, 0xe5, 0x74, 0x3f, 0x4b, 0xdc, 0x30, 0x75, 0xbd, 0x2c, 0x88, 0x42, 0x09, 0xd7, 0xff, 0x2a,
	0x42, 0xdd, 0x8b, 0xc2, 0x94, 0x85, 0xe9, 0x24, 0x75, 0xf2, 0x58, 0xe4, 0x11, 0xa8, 0xdc, 0xa1,
	0xa1, 0xbc, 0xaf, 0x3c, 0xac, 0x1d, 0xdc, 0x6b, 0xf2, 0x98, 0xcd, 0x15, 0x58, 0xd3, 0x46, 0x0c,
	0x15, 0x48, 0xd2, 0x80, 0xcd, 0xd8, 0xbd, 0x1c, 0x45, 0xae, 0xdf, 0x28, 0xa0, 0xd3, 0x16, 0x9d,
	0x8a, 0xfa, 0xaf, 0x45, 0x50, 0x39, 0x90, 0xdc, 0x85, 0xaa, 0x4d, 0x5b, 0xd6, 0x71, 0xcb, 0xb0,
	0x3b, 0x5d, 0x4b, 0x7b, 0x8b, 0xec, 0x80, 0x26, 0x15, 0x5c, 0x76, 0x9e, 0xb4, 0x6c, 0xe3, 0x48,
	0x53, 0x38, 0xac, 0x47, 0x4d, 0x07, 0x7f, 0xbd, 0x16, 0x35, 0xb5, 0x02, 0xa9, 0xc2, 0xe6, 0x54,
	0x28, 0x12, 0x80, 0x0d, 0xa3, 0xfb, 0xfc, 0x79, 0xc7, 0xd6, 0x54, 0x52, 0x03, 0x30, 0x8e, 0x4c,
	0xe3, 0xdb, 0x5e, 0xb7, 0x63, 0xd9, 0x5a, 0x89, 0x7b, 0xbe, 0xe8, 0x98, 0x2f, 0x1d, 0xe3, 0xa8,
	0x65, 0x1d, 0x9a, 0xda, 0x06, 0xd9, 0x82, 0xb2, 0x85, 0x32, 0x57, 0x6a, 0x9b, 0x64, 0x17, 0xb6,
	0x9f, 0x52, 0x3c, 0xc3, 0xa1, 0x66, 0xff, 0xc4, 0x3c, 0xb6, 0xf3, 0x13, 0xcb, 0xc8, 0x7d, 0x87,
	0x9a, 0xf6, 0x09, 0xb5, 0x96, 0x2c, 0x15, 0x42, 0xa0, 0x66, 0x99, 0x87, 0x5d, 0xbb, 0xd3, 0xb2,
	0x4d, 0x19, 0x06, 0xc8, 0x7b, 0xb0, 0xbb, 0xa8, 0x43, 0xaf, 0xe3, 0x5e, 0xd7, 0x3a, 0x36, 0xb5,
	0x2a, 0xa9, 0xc3, 0x1d, 0x6a, 0x1a, 0xdd, 0x17, 0x26, 0xfd, 0xde, 0xe9, 0x58, 0xc8, 0x72, 0x8b,
	0xbc, 0x0d, 0xf5, 0x99, 0x6a, 0x86, 0xbc, 0x43, 0xde, 0x01, 0x32, 0x53, 0x3f, 0x35, 0x39, 0xad,
	0x7e, 0xcf, 0xd0, 0x6a, 0x9c, 0xe5, 0x1c, 0x5c, 0xb0, 0xe2, 0x86, 0xbb, 0x3c, 0x99, 0x56, 0xbb,
	0xed, 0x58, 0xdd, 0xb6, 0xa9, 0x69, 0x5c, 0x6a, 0x9b, 0xcf, 0xa4, 0x54, 0xe7, 0x99, 0x53, 0xb3,
	0xd5, 0xc6, 0x48, 0x5d, 0xea, 0x58, 0x1a, 0xe1, 0xe6, 0x93, 0x5e, 0x9b, 0x33, 0xb4, 0xb4, 0x6d,
	0x9e, 0x46, 0xeb, 0x90, 0x9a, 0xa6, 0x33, 0xd3, 0xed, 0xe8, 0x3f, 0x40, 0x7d, 0x6e, 0x1a, 0x9c,
	0x53, 0x37, 0xf3, 0x86, 0xe4, 0x21, 0x94, 0xc4, 0x07, 0x36, 0xbe, 0xf8, 0xb0, 0x7a, 0x40, 0x9a,
	0x62, 0x6c, 0x9a, 0xf6, 0x15, 0x90, 0x4a, 0x00, 0xb9, 0x07, 0x95, 0x2c, 0xc0, 0x39, 0xc8, 0xdc,
	0x71, 0x2c, 0x3a, 0x5e, 0xa4, 0x57, 0x0a, 0xfd, 0x4f, 0x05, 0xaa, 0x71, 0xc2, 0x1c, 0xfc, 0xc5,
	0x6e, 0xc2, 0x90, 0x80, 0x7a, 0x11, 0xb0, 0x1f, 0xc5, 0x3c, 0xa9, 0x54, 0x7c, 0x93, 0x07, 0x70,
	0x37, 0x65, 0xaf, 0x26, 0x2c, 0xf4, 0x98, 0x13, 0x4e, 0xc6, 0xa7, 0x2c, 0x11, 0x71, 0x54, 0x5a,
	0x9b, 0xaa, 0x2d, 0xa1, 0x25, 0x1f, 0xc0, 0x96, 0x38, 0xd3, 0xf1, 0x03, 0x1c, 0xf3, 0xac, 0x51,
	0x44, 0x54, 0x85, 0x56, 0x85, 0xae, 0x2d, 0x54, 0xa4, 0x7d, 0x4d, 0x32, 0x0d, 0x15, 0x71, 0xd5,
	0x83, 0x5d, 0x39, 0xbc, 0x2b, 0x66, 0xaa, 0xcd, 0xa9, 0x9e, 0x88, 0x9c, 0xee, 0x03, 0x20, 0xdf,
	0x51, 0xe0, 0xb9, 0x4e, 0xe0, 0x37, 0x4a, 0x82, 0x4c, 0x25, 0xd7, 0x74, 0x7c, 0xfd, 0x17, 0x05,
	0x67, 0xfc, 0x5f, 0x4a, 0x68, 0x91, 0x8a, 0xba, 0x4c, 0xe5, 0x67, 0x05, 0x36, 0xbc, 0x68, 0x3c,
	0x0e, 0xb2, 0xff, 0x9a, 0x89, 0x05, 0x70, 0x3a, 0x8a, 0xbc, 0x73, 0x27, 0x08, 0xcf, 0x22, 0x11,
	0x4f, 0x48, 0xf9, 0xa9, 0x92, 0x54, 0x55, 0xe8, 0xf2, 0x23, 0xef, 0x4f, 0x1d, 0x86, 0x6e, 0x3a,
	0xcc, 0x77, 0x45, 0x45, 0x68, 0x8e, 0x50, 0xa1, 0xfb, 0x00, 0xde, 0x90, 0x79, 0xe7, 0x71, 0x14,
	0x84, 0xd9, 0x75, 0x89, 0x28, 0xd7, 0x26, 0xb2, 0xc8, 0xb2, 0xb0, 0xc4, 0x12, 0x37, 0x45, 0x01,
	0xd5, 0x32, 0x3b, 0xfc, 0xd2, 0x5f, 0x17, 0xa1, 0xca, 0x2b, 0xe5, 0xe0, 0x86, 0x0c, 0x07, 0xd7,
	0xb7, 0x73, 0x0b, 0x94, 0x61, 0x1e, 0x49, 0x19, 0x22, 0x13, 0xd5, 0x4b, 0x19, 0xaf, 0x10, 0xbf,
	0x18, 0xdb, 0x72, 0xa8, 0xe6, 0x42, 0x34, 0x0d, 0x2a, 0x00, 0x78, 0x85, 0xd4, 0x98, 0x03, 0x55,
	0x01, 0xdc, 0x59, 0x05, 0xf6, 0xfa, 0x54, 0x20, 0x38, 0xf2, 0x15, 0x47, 0x96, 0xde, 0x84, 0xe4,
	0x88, 0xa5, 0xec, 0x36, 0x96, 0xb3, 0xc3, 0xbb, 0x98, 0x06, 0x83, 0xd0, 0xcd, 0x26, 0x09, 0x6b,
	0x6c, 0xca, 0x8a, 0xce, 0x14, 0x7b, 0x5f, 0x81, 0x62, 0xdc, 0xbc, 0x90, 0x4b, 0x95, 0xda, 0xf3,
	0xa1, 0xd0, 0xeb, 0xdf, 0xdc, 0x7d, 0x79, 0xa0, 0x0a, 0xab, 0x03, 0x35, 0xad, 0x75, 0xf1, 0xaa,
	0xd6, 0xfa, 0x3e, 0x94, 0x7a, 0x7d, 0x9e, 0xe9, 0xc7, 0x50, 0xe4, 0x25, 0x51, 0xde, 0x50, 0x12,
	0x0e, 0xd0, 0xff, 0x50, 0xa0, 0x1c, 0xa2, 0x5a, 0x74, 0xea, 0xba, 0xee, 0x7d, 0x84, 0x3a, 0x1e,
	0xa9, 0x20, 0x22, 0xd5, 0x57, 0x22, 0x51, 0x61, 0x26, 0x9f, 0x80, 0xfa, 0xd3, 0x55, 0x5b, 0x1b,
	0x12, 0x36, 0x0d, 0xdc, 0xfc, 0x0e, 0x4d, 0x66, 0x98, 0x25, 0x97, 0x54, 0xa0, 0xd6, 0xdc, 0x85,
	0xbd, 0xcf, 0xa0, 0x32, 0xf3, 0x20, 0x1a, 0x14, 0xcf, 0xd9, 0x65, 0xce, 0x89, 0x7f, 0xe2, 0x73,
	0x57, 0xba, 0x70, 0x47, 0x13, 0x96, 0x17, 0x45, 0x0a, 0x5f, 0x16, 0x3e, 0x57, 0xf4, 0x97, 0xb0,
	0x7d, 0xc6, 0x78, 0xd5, 0x12, 0x5e, 0xcd, 0x34, 0xcb, 0xb7, 0xf1, 0x72, 0x31, 0x95, 0x75, 0xb7,
	0x73, 0x79, 0xee, 0xb1, 0xae, 0xb5, 0x90, 0x0d, 0xa2, 0x2c, 0x70, 0x33, 0x26, 0x6b, 0xb5, 0xe8,
	0xa0, 0x2c, 0x3b, 0x3c, 0x83, 0xdd, 0x45, 0x07, 0xa4, 0x94, 0xc6, 0xfc, 0xe1, 0x5f, 0xe3, 0x39,
	0x6b, 0x42, 0x61, 0xae, 0xad, 0x4d, 0xb8, 0x93, 0x30, 0x2f, 0xba, 0x60, 0xc9, 0x25, 0xee, 0x87,
	0x20, 0x5b, 0x77, 0xfa, 0xef, 0x0a, 0xd4, 0x67, 0x0e, 0x37, 0x3d, 0xf8, 0x31, 0xae, 0xc2, 0xe1,
	0x79, 0x9c, 0xa5, 0x79, 0xaf, 0x3f, 0x94, 0x4d, 0x5c, 0x89, 0xd3, 0x34, 0x04, 0x4a, 0xf6, 0x33,
	0x77, 0xd9, 0xfb, 0x02, 0xaa, 0x73, 0xea, 0x5b, 0x35, 0xed, 0x9b, 0x39, 0xae, 0xa2, 0x7b, 0xbd,
	0xbe, 0xb1, 0x8e, 0xeb, 0xc2, 0x4e, 0xd1, 0x7f, 0x53, 0x80, 0xcc, 0xd1, 0xc4, 0xdb, 0x1a, 0xde,
	0x20, 0xc6, 0x23, 0x00, 0xfe, 0x0a, 0xe1, 0xeb, 0xba, 0x32, 0xdf, 0x73, 0x4f, 0x2e, 0xad, 0x48,
	0xd0, 0x31, 0x8e, 0xed, 0xbb, 0x50, 0x96, 0xf0, 0x50, 0x0e, 0x7a, 0x99, 0x6e, 0x0a, 0x4b, 0x28,
	0x4c, 0xde, 0x38, 0x93, 0x26, 0x55, 0x9a, 0x50, 0xe6, 0x26, 0xfd, 0x31, 0x94, 0x5d, 0xdf, 0x77,
	0xc2, 0xc8, 0x5f, 0xdb, 0x82, 0xbc, 0x6c, 0xb2, 0x44, 0xfc, 0x93, 0x3b, 0xfb, 0x6c, 0xf4, 0x0f,
	0x9d, 0xbf, 0x86, 0x6a, 0xc2, 0x5c, 0x1f, 0xcb, 0x1a, 0x25, 0x4e, 0x78, 0x7b, 0x7f, 0x7c, 0x1d,
	0xcb, 0x93, 0xd8, 0xe7, 0x23, 0x7c, 0x7b, 0x6f, 0xde, 0xa3, 0x50, 0x2c, 0xa7, 0x22, 0x55, 0xc2,
	0xd9, 0x58, 0x97, 0xe6, 0x76, 0x0b, 0xce, 0x04, 0xae, 0x3d, 0x2b, 0xca, 0x37, 0xb1, 0x14, 0x38,
	0xf2, 0x6c, 0xe4, 0x0e, 0xc4, 0x02, 0x2e, 0x53, 0xf1, 0xad, 0xbf, 0x56, 0xa0, 0xe6, 0x0e, 0x12,
	0xc6, 0x9c, 0xff, 0x05, 0x9f, 0xd3, 0x0d, 0xf1, 0xa7, 0xff, 0xd3, 0xbf, 0x03, 0x00, 0x00, 0xff,
	0xff, 0x48, 0x42, 0x6c, 0x40, 0x35, 0x0c, 0x00, 0x00,
}
