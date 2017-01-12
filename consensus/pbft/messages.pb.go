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
	Qset
	Pset
	Cset
	AddNode
	DelNode
	ReadyForN
	UpdateN
	AgreeUpdateN
	FinishVcReset
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
	ConsensusMessage_FINISH_VCRESET          ConsensusMessage_Type = 21
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
	21: "FINISH_VCRESET",
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
	"FINISH_VCRESET":          21,
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

type Qset struct {
	Set []*PrePrepare `protobuf:"bytes,1,rep,name=set" json:"set,omitempty"`
}

func (m *Qset) Reset()                    { *m = Qset{} }
func (m *Qset) String() string            { return proto.CompactTextString(m) }
func (*Qset) ProtoMessage()               {}
func (*Qset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

func (m *Qset) GetSet() []*PrePrepare {
	if m != nil {
		return m.Set
	}
	return nil
}

type Pset struct {
	Set []*Prepare `protobuf:"bytes,1,rep,name=set" json:"set,omitempty"`
}

func (m *Pset) Reset()                    { *m = Pset{} }
func (m *Pset) String() string            { return proto.CompactTextString(m) }
func (*Pset) ProtoMessage()               {}
func (*Pset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

func (m *Pset) GetSet() []*Prepare {
	if m != nil {
		return m.Set
	}
	return nil
}

type Cset struct {
	Set []*Commit `protobuf:"bytes,1,rep,name=set" json:"set,omitempty"`
}

func (m *Cset) Reset()                    { *m = Cset{} }
func (m *Cset) String() string            { return proto.CompactTextString(m) }
func (*Cset) ProtoMessage()               {}
func (*Cset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

func (m *Cset) GetSet() []*Commit {
	if m != nil {
		return m.Set
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
func (*AddNode) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

type DelNode struct {
	ReplicaId  uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key        string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	RouterHash string `protobuf:"bytes,3,opt,name=router_hash,json=routerHash" json:"router_hash,omitempty"`
}

func (m *DelNode) Reset()                    { *m = DelNode{} }
func (m *DelNode) String() string            { return proto.CompactTextString(m) }
func (*DelNode) ProtoMessage()               {}
func (*DelNode) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

type ReadyForN struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
}

func (m *ReadyForN) Reset()                    { *m = ReadyForN{} }
func (m *ReadyForN) String() string            { return proto.CompactTextString(m) }
func (*ReadyForN) ProtoMessage()               {}
func (*ReadyForN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

type UpdateN struct {
	ReplicaId  uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key        string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	RouterHash string `protobuf:"bytes,3,opt,name=routerHash" json:"routerHash,omitempty"`
	N          int64  `protobuf:"varint,4,opt,name=n" json:"n,omitempty"`
	View       uint64 `protobuf:"varint,5,opt,name=view" json:"view,omitempty"`
	SeqNo      uint64 `protobuf:"varint,6,opt,name=seqNo" json:"seqNo,omitempty"`
	Flag       bool   `protobuf:"varint,7,opt,name=flag" json:"flag,omitempty"`
}

func (m *UpdateN) Reset()                    { *m = UpdateN{} }
func (m *UpdateN) String() string            { return proto.CompactTextString(m) }
func (*UpdateN) ProtoMessage()               {}
func (*UpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

type AgreeUpdateN struct {
	ReplicaId  uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key        string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	RouterHash string `protobuf:"bytes,3,opt,name=routerHash" json:"routerHash,omitempty"`
	N          int64  `protobuf:"varint,4,opt,name=n" json:"n,omitempty"`
	View       uint64 `protobuf:"varint,5,opt,name=view" json:"view,omitempty"`
	SeqNo      uint64 `protobuf:"varint,6,opt,name=seqNo" json:"seqNo,omitempty"`
	Flag       bool   `protobuf:"varint,7,opt,name=flag" json:"flag,omitempty"`
}

func (m *AgreeUpdateN) Reset()                    { *m = AgreeUpdateN{} }
func (m *AgreeUpdateN) String() string            { return proto.CompactTextString(m) }
func (*AgreeUpdateN) ProtoMessage()               {}
func (*AgreeUpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{24} }

type FinishVcReset struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	View      uint64 `protobuf:"varint,2,opt,name=view" json:"view,omitempty"`
	LowH      uint64 `protobuf:"varint,3,opt,name=low_h,json=lowH" json:"low_h,omitempty"`
}

func (m *FinishVcReset) Reset()                    { *m = FinishVcReset{} }
func (m *FinishVcReset) String() string            { return proto.CompactTextString(m) }
func (*FinishVcReset) ProtoMessage()               {}
func (*FinishVcReset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{25} }

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
	proto.RegisterType((*Qset)(nil), "pbft.Qset")
	proto.RegisterType((*Pset)(nil), "pbft.Pset")
	proto.RegisterType((*Cset)(nil), "pbft.Cset")
	proto.RegisterType((*AddNode)(nil), "pbft.add_node")
	proto.RegisterType((*DelNode)(nil), "pbft.del_node")
	proto.RegisterType((*ReadyForN)(nil), "pbft.ready_for_n")
	proto.RegisterType((*UpdateN)(nil), "pbft.update_n")
	proto.RegisterType((*AgreeUpdateN)(nil), "pbft.agree_update_n")
	proto.RegisterType((*FinishVcReset)(nil), "pbft.finish_vcReset")
	proto.RegisterEnum("pbft.ConsensusMessage_Type", ConsensusMessage_Type_name, ConsensusMessage_Type_value)
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1287 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xd4, 0x57, 0xef, 0x6e, 0xdb, 0x54,
	0x14, 0xc7, 0x89, 0xd3, 0x26, 0x27, 0x6d, 0xe6, 0xdc, 0x76, 0x34, 0x94, 0x0d, 0x86, 0x07, 0xdb,
	0x24, 0x50, 0x3a, 0x8d, 0x0f, 0xfc, 0x19, 0x42, 0x64, 0x8e, 0xb7, 0x44, 0x6c, 0x4e, 0x7a, 0x9b,
	0xfd, 0x41, 0x20, 0x59, 0xae, 0x73, 0x1b, 0x5b, 0x4d, 0x6d, 0xcf, 0x76, 0x3a, 0xfa, 0x04, 0x20,
	0x24, 0x9e, 0x80, 0x07, 0x80, 0x2f, 0x3c, 0x09, 0xdf, 0x78, 0x14, 0x9e, 0x80, 0x73, 0xef, 0x75,
	0x52, 0x27, 0xa9, 0xd6, 0x96, 0x0f, 0x20, 0x3e, 0x54, 0xf2, 0x39, 0xf7, 0x77, 0xce, 0xfd, 0x9d,
	0xbf, 0x37, 0x85, 0xda, 0x11, 0x4b, 0x12, 0x67, 0xc4, 0x92, 0x66, 0x14, 0x87, 0x69, 0x48, 0xd4,
	0x68, 0xff, 0x20, 0xdd, 0xbe, 0xed, 0x9d, 0x44, 0x2c, 0x76, 0x3d, 0xc7, 0x0f, 0x76, 0xdc, 0x30,
	0x66, 0x3b, 0x29, 0xca, 0xc9, 0x4e, 0x1a, 0x3b, 0x41, 0xe2, 0xb8, 0xa9, 0x1f, 0x06, 0x12, 0xae,
	0xff, 0xa4, 0x42, 0xdd, 0x0d, 0x83, 0x84, 0x05, 0xc9, 0x24, 0xb1, 0x33, 0x5f, 0xe4, 0x2e, 0xa8,
	0xdc, 0xa0, 0xa1, 0xdc, 0x50, 0xee, 0xd4, 0xee, 0x5d, 0x6b, 0x72, 0x9f, 0xcd, 0x25, 0x58, 0x73,
	0x80, 0x18, 0x2a, 0x90, 0xa4, 0x01, 0xab, 0x91, 0x73, 0x32, 0x0e, 0x9d, 0x61, 0xa3, 0x80, 0x46,
	0x6b, 0x74, 0x2a, 0xea, 0xbf, 0x15, 0x41, 0xe5, 0x40, 0x72, 0x05, 0xaa, 0x03, 0xda, 0xb2, 0xf6,
	0x5a, 0xc6, 0xa0, 0xdb, 0xb3, 0xb4, 0x37, 0xc8, 0x26, 0x68, 0x52, 0xc1, 0x65, 0xfb, 0x41, 0x6b,
	0x60, 0x74, 0x34, 0x85, 0xc3, 0xfa, 0xd4, 0xb4, 0xf1, 0xaf, 0xdf, 0xa2, 0xa6, 0x56, 0x20, 0x55,
	0x58, 0x9d, 0x0a, 0x45, 0x02, 0xb0, 0x62, 0xf4, 0x9e, 0x3c, 0xe9, 0x0e, 0x34, 0x95, 0xd4, 0x00,
	0x8c, 0x8e, 0x69, 0x7c, 0xdd, 0xef, 0x75, 0xad, 0x81, 0x56, 0xe2, 0x96, 0xcf, 0xba, 0xe6, 0x73,
	0xdb, 0xe8, 0xb4, 0xac, 0x47, 0xa6, 0xb6, 0x42, 0xd6, 0xa0, 0x6c, 0xa1, 0xcc, 0x95, 0xda, 0x2a,
	0xd9, 0x82, 0x8d, 0x87, 0x14, 0xef, 0xb0, 0xa9, 0xb9, 0xfb, 0xd4, 0xdc, 0x1b, 0x64, 0x37, 0x96,
	0x91, 0xfb, 0x26, 0x35, 0x07, 0x4f, 0xa9, 0xb5, 0x70, 0x52, 0x21, 0x04, 0x6a, 0x96, 0xf9, 0xa8,
	0x37, 0xe8, 0xb6, 0x06, 0xa6, 0x74, 0x03, 0xe4, 0x6d, 0xd8, 0x9a, 0xd7, 0xa1, 0xd5, 0x5e, 0xbf,
	0x67, 0xed, 0x99, 0x5a, 0x95, 0xd4, 0x61, 0x9d, 0x9a, 0x46, 0xef, 0x99, 0x49, 0xbf, 0xb1, 0xbb,
	0x16, 0xb2, 0x5c, 0x23, 0x57, 0xa1, 0x3e, 0x53, 0xcd, 0x90, 0xeb, 0xe4, 0x4d, 0x20, 0x33, 0xf5,
	0x43, 0x93, 0xd3, 0xda, 0xed, 0x1b, 0x5a, 0x8d, 0xb3, 0xcc, 0xc1, 0x05, 0x2b, 0x7e, 0x70, 0x85,
	0x07, 0xd3, 0x6a, 0xb7, 0x6d, 0xab, 0xd7, 0x36, 0x35, 0x8d, 0x4b, 0x6d, 0xf3, 0xb1, 0x94, 0xea,
	0x3c, 0x72, 0x6a, 0xb6, 0xda, 0xe8, 0xa9, 0x47, 0x6d, 0x4b, 0x23, 0xfc, 0xf8, 0x69, 0xbf, 0xcd,
	0x19, 0x5a, 0xda, 0x06, 0x0f, 0xa3, 0xf5, 0x88, 0x9a, 0xa6, 0x3d, 0xd3, 0x6d, 0x72, 0xdd, 0x43,
	0x64, 0xb8, 0xd7, 0xb1, 0x9f, 0x19, 0x48, 0xcb, 0x1c, 0x68, 0x57, 0xf5, 0x6f, 0xa1, 0x9e, 0xeb,
	0x10, 0x7b, 0xdf, 0x49, 0x5d, 0x8f, 0xdc, 0x81, 0x92, 0xf8, 0xc0, 0x66, 0x28, 0xde, 0xa9, 0xde,
	0x23, 0x4d, 0xd1, 0x4a, 0xcd, 0xc1, 0x29, 0x90, 0x4a, 0x00, 0xb9, 0x06, 0x95, 0xd4, 0xc7, 0xde,
	0x48, 0x9d, 0xa3, 0x48, 0x74, 0x41, 0x91, 0x9e, 0x2a, 0xf4, 0x3f, 0x15, 0xa8, 0x46, 0x31, 0xb3,
	0xf1, 0x2f, 0x72, 0x62, 0x86, 0x04, 0xd4, 0x63, 0x9f, 0xbd, 0x12, 0x3d, 0xa6, 0x52, 0xf1, 0x4d,
	0x6e, 0xc3, 0x95, 0x84, 0xbd, 0x9c, 0xb0, 0xc0, 0x65, 0x76, 0x30, 0x39, 0xda, 0x67, 0xb1, 0xf0,
	0xa3, 0xd2, 0xda, 0x54, 0x6d, 0x09, 0x2d, 0x79, 0x0f, 0xd6, 0xc4, 0x9d, 0xf6, 0xd0, 0xc7, 0xd6,
	0x4f, 0x1b, 0x45, 0x44, 0x55, 0x68, 0x55, 0xe8, 0xda, 0x42, 0x45, 0xda, 0x67, 0x04, 0xd3, 0x50,
	0x11, 0x57, 0xbd, 0xb7, 0x25, 0x1b, 0x7a, 0xe9, 0x98, 0x6a, 0x39, 0xd5, 0x03, 0x11, 0xd3, 0x75,
	0x00, 0xe4, 0x3b, 0xf6, 0x5d, 0xc7, 0xf6, 0x87, 0x8d, 0x92, 0x20, 0x53, 0xc9, 0x34, 0xdd, 0xa1,
	0xfe, 0xa3, 0x82, 0x7d, 0xff, 0x2f, 0x05, 0x34, 0x4f, 0x45, 0x5d, 0xa4, 0xf2, 0x83, 0x02, 0x2b,
	0x6e, 0x78, 0x74, 0xe4, 0xa7, 0xff, 0x35, 0x13, 0x0b, 0x60, 0x7f, 0x1c, 0xba, 0x87, 0xb6, 0x1f,
	0x1c, 0x84, 0xc2, 0x9f, 0x90, 0xb2, 0x5b, 0x25, 0xa9, 0xaa, 0xd0, 0x65, 0x57, 0x5e, 0x9f, 0x1a,
	0x78, 0x4e, 0xe2, 0x65, 0xfb, 0xa3, 0x22, 0x34, 0x1d, 0x54, 0xe8, 0x43, 0x00, 0xd7, 0x63, 0xee,
	0x61, 0x14, 0xfa, 0x41, 0x7a, 0x56, 0x20, 0xca, 0x99, 0x81, 0xcc, 0xb3, 0x2c, 0x2c, 0xb0, 0xc4,
	0xed, 0x51, 0x40, 0xb5, 0x8c, 0x0e, 0xbf, 0xf4, 0x9f, 0x8b, 0x50, 0xe5, 0x99, 0xb2, 0x71, 0x6b,
	0x06, 0xa3, 0xb3, 0xcb, 0xb9, 0x06, 0x8a, 0x97, 0x79, 0x52, 0x3c, 0x64, 0xa2, 0xba, 0x09, 0xe3,
	0x19, 0xe2, 0x83, 0xb1, 0x21, 0x9b, 0x2a, 0xe7, 0xa2, 0x69, 0x50, 0x01, 0xc0, 0x11, 0x52, 0x23,
	0x0e, 0x54, 0x05, 0x70, 0x73, 0x19, 0xd8, 0xdf, 0xa5, 0x02, 0xc1, 0x91, 0x2f, 0x39, 0xb2, 0xf4,
	0x3a, 0x24, 0x47, 0x2c, 0x44, 0xb7, 0xb2, 0x18, 0x1d, 0xce, 0x62, 0xe2, 0x8f, 0x02, 0x27, 0x9d,
	0xc4, 0xac, 0xb1, 0x2a, 0x33, 0x3a, 0x53, 0x6c, 0x7f, 0x01, 0x8a, 0x71, 0xf1, 0x44, 0x2e, 0x64,
	0x6a, 0x7b, 0x08, 0x85, 0xfe, 0xee, 0xc5, 0xcd, 0x17, 0x1b, 0xaa, 0xb0, 0xdc, 0x50, 0xd3, 0x5c,
	0x17, 0x4f, 0x73, 0xad, 0xef, 0x40, 0xa9, 0xbf, 0xcb, 0x23, 0xbd, 0x05, 0x45, 0x9e, 0x12, 0xe5,
	0x35, 0x29, 0xe1, 0x00, 0xfd, 0x0f, 0x05, 0xca, 0x01, 0xaa, 0x45, 0xa5, 0xce, 0xaa, 0xde, 0x07,
	0xa8, 0xe3, 0x9e, 0x0a, 0xc2, 0x53, 0x7d, 0xc9, 0x13, 0x15, 0xc7, 0xe4, 0x23, 0x50, 0xbf, 0x3f,
	0x2d, 0x6b, 0x43, 0xc2, 0xa6, 0x8e, 0x9b, 0x2f, 0xf0, 0xc8, 0x0c, 0xd2, 0xf8, 0x84, 0x0a, 0xd4,
	0x39, 0xb3, 0xb0, 0xfd, 0x09, 0x54, 0x66, 0x16, 0x44, 0x83, 0xe2, 0x21, 0x3b, 0xc9, 0x38, 0xf1,
	0x4f, 0x7c, 0x02, 0x4b, 0xc7, 0xce, 0x78, 0xc2, 0xb2, 0xa4, 0x48, 0xe1, 0xf3, 0xc2, 0xa7, 0x8a,
	0xfe, 0x1c, 0x36, 0x0e, 0x18, 0xcf, 0x5a, 0xcc, 0xb3, 0x99, 0xa4, 0xd9, 0x36, 0x5e, 0x4c, 0xa6,
	0x72, 0xde, 0x74, 0x2e, 0xf6, 0x3d, 0xe6, 0xb5, 0x16, 0xb0, 0x51, 0x98, 0xfa, 0x4e, 0xca, 0x64,
	0xae, 0xe6, 0x0d, 0x94, 0x45, 0x83, 0xc7, 0xb0, 0x35, 0x6f, 0x80, 0x94, 0x92, 0x88, 0xff, 0x18,
	0x38, 0xc7, 0x72, 0x56, 0x84, 0x42, 0xae, 0xac, 0x4d, 0x58, 0x8f, 0x99, 0x1b, 0x1e, 0xb3, 0xf8,
	0x04, 0xf7, 0x83, 0x9f, 0x9e, 0x77, 0xfb, 0x5f, 0x0a, 0xd4, 0x67, 0x06, 0x17, 0xbd, 0xf8, 0x3e,
	0xae, 0x42, 0xef, 0x30, 0x4a, 0x93, 0xac, 0xd6, 0x37, 0x65, 0x11, 0x97, 0xfc, 0x34, 0x0d, 0x81,
	0x92, 0xf5, 0xcc, 0x4c, 0xc8, 0x0d, 0x90, 0xcb, 0xa9, 0xc3, 0xfc, 0x91, 0x97, 0x66, 0x3d, 0x99,
	0x57, 0x91, 0xf7, 0x61, 0x7d, 0xec, 0x24, 0xe9, 0x83, 0xe9, 0x86, 0x12, 0x65, 0xaf, 0xd0, 0x79,
	0xe5, 0xf6, 0x67, 0x50, 0xcd, 0xb9, 0xbf, 0x54, 0xf1, 0xbf, 0xca, 0xc5, 0x2c, 0xba, 0xa0, 0xbf,
	0x6b, 0x9c, 0x17, 0xf3, 0xdc, 0x6e, 0xd2, 0x7f, 0x51, 0x80, 0xe4, 0xc2, 0xc5, 0xa9, 0x0f, 0x2e,
	0xe0, 0xe3, 0x2e, 0x00, 0x7f, 0xcd, 0xf0, 0x95, 0x5e, 0x9a, 0x93, 0xdc, 0xd3, 0x4d, 0x2b, 0x12,
	0xb4, 0x87, 0xed, 0xff, 0x16, 0x94, 0x25, 0x3c, 0x90, 0x03, 0x53, 0xa6, 0xab, 0xe2, 0x24, 0x10,
	0x47, 0xee, 0x51, 0x2a, 0x8f, 0x54, 0x79, 0x84, 0x32, 0x3f, 0xd2, 0x3f, 0x04, 0x55, 0x8c, 0xf6,
	0xcd, 0xfc, 0x68, 0x9f, 0x71, 0x91, 0x98, 0x6b, 0x5c, 0xb3, 0x7d, 0x0e, 0x7e, 0x37, 0x0f, 0x5e,
	0x9f, 0x81, 0x4f, 0x81, 0xb7, 0x40, 0x35, 0x38, 0xf0, 0x9d, 0x3c, 0x70, 0x6d, 0xfa, 0xe3, 0x95,
	0xbf, 0x8c, 0x12, 0x77, 0x1f, 0xca, 0xce, 0x70, 0x68, 0x07, 0xe1, 0xf0, 0xdc, 0x46, 0xca, 0x8a,
	0x26, 0x0b, 0xc4, 0x3f, 0xf5, 0xef, 0xa0, 0x3c, 0x64, 0xe3, 0x7f, 0x66, 0x8c, 0x21, 0x54, 0xe3,
	0x70, 0x92, 0xb2, 0x58, 0xbe, 0x74, 0x72, 0xa5, 0x82, 0x54, 0x89, 0xa7, 0xee, 0x4b, 0x04, 0x30,
	0x67, 0x88, 0x55, 0x0f, 0x63, 0x3b, 0xb8, 0x3c, 0xbb, 0x5f, 0x71, 0x07, 0x4e, 0xa2, 0x21, 0x9f,
	0xd4, 0xcb, 0x5b, 0x63, 0xe2, 0x72, 0x5c, 0x96, 0xd9, 0xf1, 0x16, 0x0b, 0x44, 0xaf, 0x17, 0xa9,
	0x12, 0xcc, 0xa6, 0xbb, 0x94, 0x5b, 0xb1, 0xd8, 0xd2, 0xb8, 0xfd, 0xad, 0x30, 0x7b, 0x90, 0xa4,
	0xc0, 0x91, 0x07, 0x63, 0x67, 0x24, 0xde, 0xa1, 0x32, 0x15, 0xdf, 0xfa, 0xef, 0x0a, 0xd4, 0x9c,
	0x51, 0xcc, 0x98, 0xfd, 0xff, 0xe0, 0xfb, 0x02, 0x6a, 0x07, 0xb8, 0xaf, 0x12, 0xcf, 0x3e, 0x76,
	0x29, 0x5b, 0xde, 0xfc, 0x17, 0x59, 0x7e, 0x64, 0x03, 0x4a, 0xe3, 0xf0, 0x95, 0xed, 0x4d, 0x1f,
	0x3a, 0x14, 0x3a, 0xfb, 0x2b, 0xe2, 0x3f, 0xb1, 0x8f, 0xff, 0x0e, 0x00, 0x00, 0xff, 0xff, 0xfd,
	0x2b, 0xbe, 0x28, 0xca, 0x0d, 0x00, 0x00,
}
