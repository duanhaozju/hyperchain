package pbft

import (
	"time"

	pb "hyperchain/protos"
)

func qMsgHelper(prePrepare *PrePrepare, id uint64) *pb.Message {
	timestamp := time.Now().Unix()
	qMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Timestamp: 	timestamp,
		Payload:	&Message{Payload: &Message_PrePrepare{PrePrepare: prePrepare}},
		Id:		id,
	}
	return qMsg
}


func pMsgHelper(prepare *Prepare, id uint64) *pb.Message {
	timestamp := time.Now().Unix()
	pMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Timestamp: 	timestamp,
		Payload:	&Message{Payload: &Message_Prepare{Prepare: prepare}},
		Id:		id,
	}
	return pMsg
}

func cMsgHelper(commit *Commit, id uint64) *pb.Message {
	timestamp := time.Now().Unix()
	cMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Timestamp: 	timestamp,
		Payload:	&Message{Payload: &Message_Commit{Commit: commit}},
		Id:		id,
	}
	return cMsg
}
