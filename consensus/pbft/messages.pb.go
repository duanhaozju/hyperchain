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
	ReturnRequestBatch
	PrePrepare
	Prepare
	Commit
	BlockInfo
	Checkpoint
	ViewChange
	NewView
	FinishVcReset
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
	AgreeUpdateN
	UpdateN
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
	ConsensusMessage_PRE_PREPARE             ConsensusMessage_Type = 1
	ConsensusMessage_PREPARE                 ConsensusMessage_Type = 2
	ConsensusMessage_COMMIT                  ConsensusMessage_Type = 3
	ConsensusMessage_CHECKPOINT              ConsensusMessage_Type = 4
	ConsensusMessage_VIEW_CHANGE             ConsensusMessage_Type = 5
	ConsensusMessage_NEW_VIEW                ConsensusMessage_Type = 6
	ConsensusMessage_FINISH_VCRESET          ConsensusMessage_Type = 7
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
	ConsensusMessage_AGREE_UPDATE_N          ConsensusMessage_Type = 19
	ConsensusMessage_UPDATE_N                ConsensusMessage_Type = 20
)

var ConsensusMessage_Type_name = map[int32]string{
	0:  "TRANSACTION",
	1:  "PRE_PREPARE",
	2:  "PREPARE",
	3:  "COMMIT",
	4:  "CHECKPOINT",
	5:  "VIEW_CHANGE",
	6:  "NEW_VIEW",
	7:  "FINISH_VCRESET",
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
	19: "AGREE_UPDATE_N",
	20: "UPDATE_N",
}
var ConsensusMessage_Type_value = map[string]int32{
	"TRANSACTION":             0,
	"PRE_PREPARE":             1,
	"PREPARE":                 2,
	"COMMIT":                  3,
	"CHECKPOINT":              4,
	"VIEW_CHANGE":             5,
	"NEW_VIEW":                6,
	"FINISH_VCRESET":          7,
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
	"AGREE_UPDATE_N":          19,
	"UPDATE_N":                20,
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

type ReturnRequestBatch struct {
	Batch  *TransactionBatch `protobuf:"bytes,1,opt,name=batch" json:"batch,omitempty"`
	Digest string            `protobuf:"bytes,2,opt,name=digest" json:"digest,omitempty"`
}

func (m *ReturnRequestBatch) Reset()                    { *m = ReturnRequestBatch{} }
func (m *ReturnRequestBatch) String() string            { return proto.CompactTextString(m) }
func (*ReturnRequestBatch) ProtoMessage()               {}
func (*ReturnRequestBatch) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{2} }

func (m *ReturnRequestBatch) GetBatch() *TransactionBatch {
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
func (*PrePrepare) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{3} }

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
func (*Prepare) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{4} }

type Commit struct {
	View           uint64 `protobuf:"varint,1,opt,name=view" json:"view,omitempty"`
	SequenceNumber uint64 `protobuf:"varint,2,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	BatchDigest    string `protobuf:"bytes,3,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	ReplicaId      uint64 `protobuf:"varint,4,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *Commit) Reset()                    { *m = Commit{} }
func (m *Commit) String() string            { return proto.CompactTextString(m) }
func (*Commit) ProtoMessage()               {}
func (*Commit) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{5} }

type BlockInfo struct {
	BlockNumber uint64 `protobuf:"varint,1,opt,name=block_number,json=blockNumber" json:"block_number,omitempty"`
	BlockHash   []byte `protobuf:"bytes,2,opt,name=block_hash,json=blockHash,proto3" json:"block_hash,omitempty"`
}

func (m *BlockInfo) Reset()                    { *m = BlockInfo{} }
func (m *BlockInfo) String() string            { return proto.CompactTextString(m) }
func (*BlockInfo) ProtoMessage()               {}
func (*BlockInfo) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{6} }

type Checkpoint struct {
	SequenceNumber uint64 `protobuf:"varint,1,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	ReplicaId      uint64 `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Id             string `protobuf:"bytes,3,opt,name=id" json:"id,omitempty"`
}

func (m *Checkpoint) Reset()                    { *m = Checkpoint{} }
func (m *Checkpoint) String() string            { return proto.CompactTextString(m) }
func (*Checkpoint) ProtoMessage()               {}
func (*Checkpoint) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{7} }

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
func (*ViewChange) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8} }

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
func (*ViewChange_C) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8, 0} }

type ViewChange_PQ struct {
	SequenceNumber uint64 `protobuf:"varint,1,opt,name=sequence_number,json=sequenceNumber" json:"sequence_number,omitempty"`
	BatchDigest    string `protobuf:"bytes,2,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	View           uint64 `protobuf:"varint,3,opt,name=view" json:"view,omitempty"`
}

func (m *ViewChange_PQ) Reset()                    { *m = ViewChange_PQ{} }
func (m *ViewChange_PQ) String() string            { return proto.CompactTextString(m) }
func (*ViewChange_PQ) ProtoMessage()               {}
func (*ViewChange_PQ) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{8, 1} }

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

type FinishVcReset struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	View      uint64 `protobuf:"varint,2,opt,name=view" json:"view,omitempty"`
	LowH      uint64 `protobuf:"varint,3,opt,name=low_h,json=lowH" json:"low_h,omitempty"`
}

func (m *FinishVcReset) Reset()                    { *m = FinishVcReset{} }
func (m *FinishVcReset) String() string            { return proto.CompactTextString(m) }
func (*FinishVcReset) ProtoMessage()               {}
func (*FinishVcReset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{10} }

type FetchRequestBatch struct {
	BatchDigest string `protobuf:"bytes,1,opt,name=batch_digest,json=batchDigest" json:"batch_digest,omitempty"`
	ReplicaId   uint64 `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *FetchRequestBatch) Reset()                    { *m = FetchRequestBatch{} }
func (m *FetchRequestBatch) String() string            { return proto.CompactTextString(m) }
func (*FetchRequestBatch) ProtoMessage()               {}
func (*FetchRequestBatch) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{11} }

type NegotiateView struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *NegotiateView) Reset()                    { *m = NegotiateView{} }
func (m *NegotiateView) String() string            { return proto.CompactTextString(m) }
func (*NegotiateView) ProtoMessage()               {}
func (*NegotiateView) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{12} }

type NegotiateViewResponse struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	View      uint64 `protobuf:"varint,2,opt,name=view" json:"view,omitempty"`
	N         uint64 `protobuf:"varint,3,opt,name=n" json:"n,omitempty"`
	Routers   string `protobuf:"bytes,4,opt,name=routers" json:"routers,omitempty"`
}

func (m *NegotiateViewResponse) Reset()                    { *m = NegotiateViewResponse{} }
func (m *NegotiateViewResponse) String() string            { return proto.CompactTextString(m) }
func (*NegotiateViewResponse) ProtoMessage()               {}
func (*NegotiateViewResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{13} }

type RecoveryInit struct {
	ReplicaId uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
}

func (m *RecoveryInit) Reset()                    { *m = RecoveryInit{} }
func (m *RecoveryInit) String() string            { return proto.CompactTextString(m) }
func (*RecoveryInit) ProtoMessage()               {}
func (*RecoveryInit) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{14} }

type RecoveryResponse struct {
	ReplicaId     uint64            `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Chkpts        map[uint64]string `protobuf:"bytes,2,rep,name=chkpts" json:"chkpts,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
	BlockHeight   uint64            `protobuf:"varint,3,opt,name=blockHeight" json:"blockHeight,omitempty"`
	LastBlockHash string            `protobuf:"bytes,4,opt,name=lastBlockHash" json:"lastBlockHash,omitempty"`
}

func (m *RecoveryResponse) Reset()                    { *m = RecoveryResponse{} }
func (m *RecoveryResponse) String() string            { return proto.CompactTextString(m) }
func (*RecoveryResponse) ProtoMessage()               {}
func (*RecoveryResponse) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{15} }

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
func (*RecoveryFetchPQC) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{16} }

type RecoveryReturnPQC struct {
	ReplicaId uint64        `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	PrepreSet []*PrePrepare `protobuf:"bytes,2,rep,name=prepre_set,json=prepreSet" json:"prepre_set,omitempty"`
	PreSent   []bool        `protobuf:"varint,3,rep,packed,name=pre_sent,json=preSent" json:"pre_sent,omitempty"`
	CmtSent   []bool        `protobuf:"varint,4,rep,packed,name=cmt_sent,json=cmtSent" json:"cmt_sent,omitempty"`
}

func (m *RecoveryReturnPQC) Reset()                    { *m = RecoveryReturnPQC{} }
func (m *RecoveryReturnPQC) String() string            { return proto.CompactTextString(m) }
func (*RecoveryReturnPQC) ProtoMessage()               {}
func (*RecoveryReturnPQC) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{17} }

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
func (*Qset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{18} }

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
func (*Pset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{19} }

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
func (*Cset) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{20} }

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
func (*AddNode) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{21} }

type DelNode struct {
	ReplicaId  uint64 `protobuf:"varint,1,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key        string `protobuf:"bytes,2,opt,name=key" json:"key,omitempty"`
	RouterHash string `protobuf:"bytes,3,opt,name=router_hash,json=routerHash" json:"router_hash,omitempty"`
}

func (m *DelNode) Reset()                    { *m = DelNode{} }
func (m *DelNode) String() string            { return proto.CompactTextString(m) }
func (*DelNode) ProtoMessage()               {}
func (*DelNode) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{22} }

type ReadyForN struct {
	ReplicaId uint64 `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key       string `protobuf:"bytes,3,opt,name=key" json:"key,omitempty"`
}

func (m *ReadyForN) Reset()                    { *m = ReadyForN{} }
func (m *ReadyForN) String() string            { return proto.CompactTextString(m) }
func (*ReadyForN) ProtoMessage()               {}
func (*ReadyForN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{23} }

type AgreeUpdateN struct {
	Flag       bool             `protobuf:"varint,1,opt,name=flag" json:"flag,omitempty"`
	ReplicaId  uint64           `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key        string           `protobuf:"bytes,3,opt,name=key" json:"key,omitempty"`
	RouterHash string           `protobuf:"bytes,4,opt,name=routerHash" json:"routerHash,omitempty"`
	N          int64            `protobuf:"varint,5,opt,name=n" json:"n,omitempty"`
	View       uint64           `protobuf:"varint,6,opt,name=view" json:"view,omitempty"`
	H          uint64           `protobuf:"varint,7,opt,name=h" json:"h,omitempty"`
	Cset       []*ViewChange_C  `protobuf:"bytes,8,rep,name=cset" json:"cset,omitempty"`
	Pset       []*ViewChange_PQ `protobuf:"bytes,9,rep,name=pset" json:"pset,omitempty"`
	Qset       []*ViewChange_PQ `protobuf:"bytes,10,rep,name=qset" json:"qset,omitempty"`
}

func (m *AgreeUpdateN) Reset()                    { *m = AgreeUpdateN{} }
func (m *AgreeUpdateN) String() string            { return proto.CompactTextString(m) }
func (*AgreeUpdateN) ProtoMessage()               {}
func (*AgreeUpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{24} }

func (m *AgreeUpdateN) GetCset() []*ViewChange_C {
	if m != nil {
		return m.Cset
	}
	return nil
}

func (m *AgreeUpdateN) GetPset() []*ViewChange_PQ {
	if m != nil {
		return m.Pset
	}
	return nil
}

func (m *AgreeUpdateN) GetQset() []*ViewChange_PQ {
	if m != nil {
		return m.Qset
	}
	return nil
}

type UpdateN struct {
	Flag       bool              `protobuf:"varint,1,opt,name=flag" json:"flag,omitempty"`
	ReplicaId  uint64            `protobuf:"varint,2,opt,name=replica_id,json=replicaId" json:"replica_id,omitempty"`
	Key        string            `protobuf:"bytes,3,opt,name=key" json:"key,omitempty"`
	RouterHash string            `protobuf:"bytes,4,opt,name=routerHash" json:"routerHash,omitempty"`
	N          int64             `protobuf:"varint,5,opt,name=n" json:"n,omitempty"`
	View       uint64            `protobuf:"varint,6,opt,name=view" json:"view,omitempty"`
	Vset       []*ViewChange     `protobuf:"bytes,7,rep,name=vset" json:"vset,omitempty"`
	Xset       map[uint64]string `protobuf:"bytes,8,rep,name=xset" json:"xset,omitempty" protobuf_key:"varint,1,opt,name=key" protobuf_val:"bytes,2,opt,name=value"`
}

func (m *UpdateN) Reset()                    { *m = UpdateN{} }
func (m *UpdateN) String() string            { return proto.CompactTextString(m) }
func (*UpdateN) ProtoMessage()               {}
func (*UpdateN) Descriptor() ([]byte, []int) { return fileDescriptor0, []int{25} }

func (m *UpdateN) GetVset() []*ViewChange {
	if m != nil {
		return m.Vset
	}
	return nil
}

func (m *UpdateN) GetXset() map[uint64]string {
	if m != nil {
		return m.Xset
	}
	return nil
}

func init() {
	proto.RegisterType((*ConsensusMessage)(nil), "pbft.consensus_message")
	proto.RegisterType((*TransactionBatch)(nil), "pbft.transaction_batch")
	proto.RegisterType((*ReturnRequestBatch)(nil), "pbft.return_request_batch")
	proto.RegisterType((*PrePrepare)(nil), "pbft.pre_prepare")
	proto.RegisterType((*Prepare)(nil), "pbft.prepare")
	proto.RegisterType((*Commit)(nil), "pbft.commit")
	proto.RegisterType((*BlockInfo)(nil), "pbft.block_info")
	proto.RegisterType((*Checkpoint)(nil), "pbft.checkpoint")
	proto.RegisterType((*ViewChange)(nil), "pbft.view_change")
	proto.RegisterType((*ViewChange_C)(nil), "pbft.view_change.C")
	proto.RegisterType((*ViewChange_PQ)(nil), "pbft.view_change.PQ")
	proto.RegisterType((*NewView)(nil), "pbft.new_view")
	proto.RegisterType((*FinishVcReset)(nil), "pbft.finish_vcReset")
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
	proto.RegisterType((*AgreeUpdateN)(nil), "pbft.agree_update_n")
	proto.RegisterType((*UpdateN)(nil), "pbft.update_n")
	proto.RegisterEnum("pbft.ConsensusMessage_Type", ConsensusMessage_Type_name, ConsensusMessage_Type_value)
}

func init() { proto.RegisterFile("messages.proto", fileDescriptor0) }

var fileDescriptor0 = []byte{
	// 1333 bytes of a gzipped FileDescriptorProto
	0x1f, 0x8b, 0x08, 0x00, 0x00, 0x09, 0x6e, 0x88, 0x02, 0xff, 0xcc, 0x57, 0x4b, 0x73, 0xdb, 0x54,
	0x14, 0x46, 0xb6, 0xfc, 0x3a, 0x4e, 0x5c, 0xf9, 0x26, 0x34, 0x26, 0xb4, 0x50, 0x54, 0xa0, 0x9d,
	0x01, 0x9c, 0x4e, 0x59, 0xf0, 0x28, 0xc3, 0xe0, 0xca, 0x6a, 0xed, 0x81, 0xca, 0xce, 0xb5, 0xfa,
	0x60, 0x80, 0xd1, 0x28, 0xf2, 0x8d, 0xad, 0x89, 0x23, 0xb9, 0x92, 0x9c, 0x92, 0x5f, 0x00, 0x1b,
	0x36, 0x6c, 0x59, 0x30, 0xc3, 0xdf, 0x61, 0xc7, 0x4f, 0x61, 0xf8, 0x01, 0xdc, 0x87, 0x64, 0xcb,
	0x72, 0xa8, 0xdd, 0x2e, 0x80, 0x85, 0x67, 0x74, 0xce, 0xfd, 0xce, 0xf3, 0x9e, 0xc7, 0x35, 0xd4,
	0x4e, 0x49, 0x18, 0xda, 0x23, 0x12, 0x36, 0xa7, 0x81, 0x1f, 0xf9, 0x48, 0x9e, 0x1e, 0x1d, 0x47,
	0xfb, 0x37, 0xc6, 0xe7, 0x53, 0x12, 0x38, 0x63, 0xdb, 0xf5, 0x0e, 0x1c, 0x3f, 0x20, 0x07, 0x11,
	0xa5, 0xc3, 0x83, 0x28, 0xb0, 0xbd, 0xd0, 0x76, 0x22, 0xd7, 0xf7, 0x04, 0x5c, 0xfd, 0x2b, 0x0f,
	0x75, 0xc7, 0xf7, 0x42, 0xe2, 0x85, 0xb3, 0xd0, 0x8a, 0x75, 0xa1, 0x5b, 0x20, 0x33, 0x81, 0x86,
	0x74, 0x4d, 0xba, 0x59, 0xbb, 0x7d, 0xa5, 0xc9, 0x74, 0x36, 0x57, 0x60, 0x4d, 0x93, 0x62, 0x30,
	0x47, 0xa2, 0x06, 0x94, 0xa6, 0xf6, 0xf9, 0xc4, 0xb7, 0x87, 0x8d, 0x1c, 0x15, 0xda, 0xc2, 0x09,
	0xa9, 0xfe, 0x9c, 0x07, 0x99, 0x01, 0xd1, 0x25, 0xa8, 0x9a, 0xb8, 0x65, 0x0c, 0x5a, 0x9a, 0xd9,
	0xed, 0x19, 0xca, 0x2b, 0x8c, 0xd1, 0xc7, 0xba, 0x45, 0x7f, 0xfd, 0x16, 0xd6, 0x15, 0x09, 0x55,
	0xa1, 0x94, 0x10, 0x39, 0x04, 0x50, 0xd4, 0x7a, 0x0f, 0x1e, 0x74, 0x4d, 0x25, 0x8f, 0x6a, 0x00,
	0x5a, 0x47, 0xd7, 0xbe, 0xec, 0xf7, 0xba, 0x86, 0xa9, 0xc8, 0x4c, 0xf2, 0x51, 0x57, 0x7f, 0x6c,
	0x69, 0x9d, 0x96, 0x71, 0x5f, 0x57, 0x0a, 0x68, 0x0b, 0xca, 0x06, 0xa5, 0x19, 0x53, 0x29, 0x22,
	0x04, 0xb5, 0x7b, 0x5d, 0xa3, 0x3b, 0xe8, 0x58, 0x8f, 0x34, 0xac, 0x0f, 0x74, 0x53, 0x29, 0xa1,
	0x3d, 0xd8, 0xb9, 0x87, 0x4d, 0xad, 0x63, 0x61, 0xfd, 0xf0, 0xa1, 0x3e, 0x30, 0xad, 0xbb, 0x2d,
	0x4a, 0x29, 0x65, 0xea, 0xf9, 0x2e, 0xd6, 0xcd, 0x87, 0xd8, 0xc8, 0x9c, 0x54, 0x98, 0x1a, 0x43,
	0xbf, 0xdf, 0x33, 0xbb, 0x2d, 0x53, 0x17, 0xaa, 0x01, 0xbd, 0x0e, 0x7b, 0xcb, 0x3c, 0x2a, 0x35,
	0xe8, 0xf7, 0x8c, 0x81, 0xae, 0x54, 0x51, 0x1d, 0xb6, 0xb1, 0xae, 0xf5, 0x1e, 0xe9, 0xf8, 0x6b,
	0x8b, 0xda, 0x37, 0x95, 0x2d, 0xf4, 0x2a, 0xd4, 0xe7, 0xac, 0x39, 0x72, 0x1b, 0x5d, 0x06, 0x34,
	0x67, 0xdf, 0xd3, 0x99, 0x5b, 0x87, 0x7d, 0x4d, 0xa9, 0x31, 0x2f, 0x53, 0x70, 0xee, 0x15, 0x3b,
	0xb8, 0xc4, 0x02, 0x6c, 0xb5, 0xdb, 0x96, 0xd1, 0x6b, 0xeb, 0x8a, 0xc2, 0xa8, 0xb6, 0xfe, 0x95,
	0xa0, 0xea, 0x2c, 0x1b, 0x58, 0x6f, 0xb5, 0xa9, 0xa6, 0x1e, 0xb6, 0x0c, 0x05, 0x31, 0xc7, 0x5b,
	0xf7, 0xb1, 0xae, 0x5b, 0x0f, 0xfb, 0x6d, 0xe6, 0xa7, 0xa1, 0xec, 0x30, 0x91, 0x39, 0xb5, 0xab,
	0x7e, 0x03, 0xf5, 0x54, 0x2d, 0x58, 0x47, 0x76, 0xe4, 0x8c, 0xd1, 0x4d, 0x28, 0xf0, 0x0f, 0x7a,
	0xed, 0xf9, 0x9b, 0xd5, 0xdb, 0xa8, 0xc9, 0x8b, 0xa6, 0x69, 0x2e, 0x80, 0x58, 0x00, 0xd0, 0x15,
	0xa8, 0x44, 0x2e, 0xad, 0x82, 0xc8, 0x3e, 0x9d, 0xf2, 0xfb, 0xce, 0xe3, 0x05, 0x43, 0xfd, 0x0e,
	0x76, 0x03, 0x12, 0xcd, 0x02, 0xcf, 0x0a, 0xc8, 0xd3, 0x19, 0x65, 0xc6, 0xfa, 0x3f, 0x58, 0xe8,
	0x97, 0xa8, 0xfe, 0x3d, 0x51, 0x56, 0x2b, 0x7e, 0x24, 0x46, 0x2e, 0x43, 0x71, 0xe8, 0xd2, 0xc2,
	0x8e, 0xb8, 0x85, 0x0a, 0x8e, 0x29, 0xf5, 0x0f, 0x09, 0xaa, 0xd3, 0x80, 0x58, 0xf4, 0x37, 0xb5,
	0x03, 0x42, 0xa3, 0x95, 0xcf, 0x5c, 0xf2, 0x8c, 0x6b, 0x95, 0x31, 0xff, 0x46, 0x37, 0xe0, 0x52,
	0xc8, 0x6c, 0x7b, 0x0e, 0xb1, 0xbc, 0xd9, 0xe9, 0x11, 0x09, 0xb8, 0x12, 0x19, 0xd7, 0x12, 0xb6,
	0xc1, 0xb9, 0xe8, 0x2d, 0xd8, 0xe2, 0xd6, 0xac, 0xd8, 0x54, 0x9e, 0x9b, 0xaa, 0x72, 0x5e, 0x9b,
	0xb3, 0x50, 0xfb, 0x82, 0x5c, 0x35, 0xe4, 0xe7, 0x87, 0xa0, 0xa4, 0x58, 0x77, 0x79, 0x34, 0x57,
	0x01, 0xa8, 0xbf, 0x13, 0xd7, 0xb1, 0x2d, 0x77, 0xd8, 0x28, 0x70, 0x67, 0x2a, 0x31, 0xa7, 0x3b,
	0x54, 0x7f, 0x94, 0x68, 0x03, 0xfd, 0x4b, 0x01, 0x2d, 0xbb, 0x22, 0x67, 0x5d, 0xf9, 0x41, 0x82,
	0xa2, 0xe3, 0x9f, 0x9e, 0xba, 0xd1, 0x7f, 0xed, 0x89, 0x01, 0x70, 0x34, 0xf1, 0x9d, 0x13, 0xcb,
	0xf5, 0x8e, 0x7d, 0xae, 0x8f, 0x53, 0xb1, 0x55, 0xe1, 0x54, 0x95, 0xf3, 0x62, 0x93, 0x57, 0x13,
	0x81, 0xb1, 0x1d, 0x8e, 0xe3, 0x41, 0x54, 0xe1, 0x9c, 0x0e, 0x65, 0xa8, 0x43, 0x00, 0x67, 0x4c,
	0x9c, 0x93, 0xa9, 0xef, 0x7a, 0xd1, 0x45, 0x81, 0x48, 0x17, 0x06, 0xb2, 0xec, 0x65, 0x2e, 0xe3,
	0x25, 0x1d, 0x4e, 0x39, 0xca, 0x16, 0xd1, 0xd1, 0x2f, 0xf5, 0xa7, 0x3c, 0x54, 0x59, 0xa6, 0x2c,
	0x3a, 0x7e, 0xbd, 0xd1, 0xc5, 0xd7, 0xb9, 0x05, 0xd2, 0x38, 0xd6, 0x24, 0x8d, 0xa9, 0x27, 0xb2,
	0x13, 0x12, 0x96, 0x21, 0xd6, 0x77, 0x3b, 0xa2, 0xa8, 0x52, 0x2a, 0x9a, 0x1a, 0xe6, 0x00, 0xda,
	0xa1, 0xf2, 0x94, 0x01, 0x65, 0x0e, 0xdc, 0x5d, 0x05, 0xf6, 0x0f, 0x31, 0x47, 0x30, 0xe4, 0x53,
	0x86, 0x2c, 0x3c, 0x0f, 0xc9, 0x10, 0x99, 0xe8, 0x8a, 0xd9, 0xe8, 0x68, 0xab, 0x87, 0xee, 0xc8,
	0xb3, 0x69, 0x3f, 0x93, 0x46, 0x49, 0x64, 0x74, 0xce, 0xd8, 0xff, 0x0c, 0x24, 0x6d, 0xf3, 0x44,
	0x66, 0x32, 0xb5, 0x3f, 0x84, 0x5c, 0xff, 0x70, 0x73, 0xf1, 0x6c, 0x41, 0xe5, 0x56, 0x0b, 0x2a,
	0xc9, 0x75, 0x7e, 0x91, 0x6b, 0xf5, 0x77, 0x09, 0xca, 0x1e, 0x0d, 0x9c, 0x27, 0xfe, 0xa2, 0xcb,
	0x78, 0x87, 0xf2, 0x58, 0xae, 0x72, 0x3c, 0x57, 0xf5, 0x95, 0x5c, 0x61, 0x7e, 0x8c, 0xde, 0x07,
	0xf9, 0xfb, 0xc5, 0x2d, 0x35, 0x04, 0x2c, 0x51, 0xdc, 0x7c, 0x42, 0x8f, 0x74, 0x2f, 0x0a, 0xce,
	0x31, 0x47, 0xad, 0x29, 0xed, 0xfd, 0x8f, 0xa0, 0x32, 0x97, 0x40, 0x0a, 0xe4, 0x4f, 0xc8, 0x79,
	0xec, 0x13, 0xfb, 0x44, 0xbb, 0x50, 0x38, 0xb3, 0x27, 0x33, 0x12, 0xc7, 0x28, 0x88, 0x4f, 0x73,
	0x1f, 0x4b, 0xea, 0x13, 0xa8, 0x1d, 0xbb, 0x9e, 0x1b, 0x8e, 0xad, 0x33, 0x07, 0x93, 0x55, 0x4b,
	0x52, 0xf6, 0x02, 0x93, 0x88, 0x73, 0xa9, 0x88, 0x77, 0xa0, 0x30, 0xf1, 0x9f, 0x59, 0xe3, 0x24,
	0x4f, 0x94, 0xe8, 0xa8, 0x8f, 0x61, 0xe7, 0x98, 0xb0, 0xf4, 0x2e, 0x4f, 0xed, 0x6c, 0xd6, 0xa5,
	0x75, 0x6d, 0x9c, 0x6d, 0x10, 0xf5, 0x00, 0x6a, 0x1e, 0x19, 0xf9, 0x91, 0x6b, 0x47, 0x44, 0xdc,
	0xc2, 0xf3, 0x5d, 0x56, 0x23, 0xd8, 0x5b, 0x16, 0xa0, 0x2e, 0x85, 0x53, 0xf6, 0xfc, 0x78, 0x99,
	0x60, 0x69, 0xaf, 0x79, 0x71, 0xa0, 0x92, 0xc7, 0x1e, 0x2a, 0x81, 0x3f, 0x8b, 0x48, 0x10, 0xf2,
	0x4b, 0xa9, 0xe0, 0x84, 0x54, 0x9b, 0xb0, 0x1d, 0x10, 0xc7, 0x3f, 0x23, 0xc1, 0x39, 0x1d, 0x38,
	0xee, 0xba, 0xc4, 0xaa, 0x7f, 0x4a, 0x50, 0x9f, 0x0b, 0x6c, 0xea, 0xe0, 0x1d, 0x3a, 0x5b, 0xc7,
	0x27, 0xd3, 0x28, 0x8c, 0xab, 0xed, 0xba, 0x28, 0xa3, 0x15, 0x3d, 0x4d, 0x8d, 0xa3, 0x44, 0x45,
	0xc5, 0x22, 0xe8, 0x1a, 0x88, 0x69, 0xd7, 0x21, 0xee, 0x68, 0x1c, 0xc5, 0x31, 0xa5, 0x59, 0xe8,
	0x6d, 0xd8, 0x9e, 0xd8, 0x61, 0x74, 0x37, 0x19, 0x79, 0x71, 0x8c, 0xcb, 0xcc, 0xfd, 0x4f, 0xa0,
	0x9a, 0x52, 0xff, 0x42, 0xe5, 0xf7, 0x45, 0x2a, 0x66, 0x5e, 0x2d, 0xfd, 0x43, 0x6d, 0x5d, 0xcc,
	0x4b, 0xc3, 0x4e, 0xfd, 0x45, 0x02, 0x94, 0x0a, 0x97, 0xbd, 0x13, 0x36, 0xd0, 0x71, 0x0b, 0x80,
	0xad, 0x47, 0xba, 0xf6, 0x57, 0x3a, 0x35, 0xf5, 0x16, 0xc0, 0x15, 0x01, 0x1a, 0xd0, 0xb6, 0x78,
	0x0d, 0xca, 0x02, 0xee, 0x89, 0x96, 0x2d, 0xe3, 0x12, 0x3f, 0xf1, 0xf8, 0x91, 0x73, 0x1a, 0x89,
	0x23, 0x59, 0x1c, 0x51, 0x9a, 0x1d, 0xa9, 0xef, 0x81, 0x7c, 0xc8, 0x9a, 0xea, 0x3a, 0xe4, 0x99,
	0x21, 0xe9, 0x9f, 0x0c, 0xb1, 0x53, 0x95, 0xce, 0xed, 0x3e, 0x03, 0xbf, 0x99, 0x06, 0x6f, 0xcf,
	0xc1, 0x0b, 0xe0, 0xbb, 0x20, 0x6b, 0x0c, 0xf8, 0x46, 0x1a, 0xb8, 0x95, 0x3c, 0xab, 0xd9, 0xaa,
	0x15, 0xb8, 0x3b, 0x50, 0xb6, 0x87, 0x43, 0xcb, 0xf3, 0x87, 0x6b, 0x0b, 0x29, 0xbe, 0x34, 0x71,
	0x41, 0xec, 0x53, 0xfd, 0x16, 0xca, 0x43, 0x32, 0x79, 0x39, 0x61, 0x1a, 0x42, 0x55, 0xf4, 0x81,
	0x58, 0x9d, 0x62, 0x46, 0x83, 0x60, 0xf1, 0xdd, 0xf9, 0x39, 0x05, 0x10, 0x7b, 0x48, 0x6f, 0xdd,
	0x0f, 0x2c, 0x6f, 0xdd, 0x4e, 0x8c, 0x0d, 0xe4, 0x17, 0xde, 0xfd, 0x96, 0x83, 0x9a, 0x3d, 0x0a,
	0x08, 0xb1, 0x66, 0xd3, 0x21, 0xeb, 0x6b, 0x8f, 0x35, 0xeb, 0xf1, 0xc4, 0x1e, 0x71, 0xf7, 0xca,
	0x98, 0x7f, 0xbf, 0xb0, 0x5e, 0x9a, 0xd2, 0x94, 0x97, 0x71, 0xb9, 0xa7, 0x38, 0xa2, 0xfb, 0x0b,
	0xfc, 0x89, 0x2a, 0x79, 0xf3, 0xf9, 0x50, 0xcc, 0xee, 0xe2, 0x52, 0x76, 0x17, 0x97, 0x37, 0xdd,
	0xc5, 0x95, 0x8d, 0x77, 0x31, 0xac, 0xdb, 0xc5, 0xea, 0xaf, 0x39, 0x28, 0xff, 0xdf, 0xd2, 0x93,
	0x6c, 0xc7, 0xd2, 0x66, 0xdb, 0xb1, 0x9c, 0xde, 0x8e, 0x49, 0x2c, 0xd9, 0xed, 0xf8, 0xd2, 0xeb,
	0xef, 0xa8, 0xc8, 0xff, 0xb6, 0x7e, 0xf8, 0x77, 0x00, 0x00, 0x00, 0xff, 0xff, 0xb2, 0xc7, 0xc6,
	0x64, 0xf7, 0x0e, 0x00, 0x00,
}
