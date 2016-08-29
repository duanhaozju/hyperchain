package pbft

import (
	"time"

	pb "hyperchain/protos"
	"github.com/golang/protobuf/proto"
)

func qMsgHelper(prePrepare *PrePrepare, id uint64) *pb.Message {
	timestamp := time.Now().Unix()
	msg := &Message{Payload: &Message_PrePrepare{PrePrepare: prePrepare}}
	payload, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("qMsg marshal error")
		return nil
	}
	qMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Timestamp: 	timestamp,
		Payload:	payload,
		Id:		id,
	}
	return qMsg
}


func pMsgHelper(prepare *Prepare, id uint64) *pb.Message {
	timestamp := time.Now().Unix()
	msg := &Message{Payload: &Message_Prepare{Prepare: prepare}}
	payload, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("pMsg marshal error")
		return nil
	}
	pMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Timestamp: 	timestamp,
		Payload:	payload,
		Id:		id,
	}
	return pMsg
}

func cMsgHelper(commit *Commit, id uint64) *pb.Message {
	timestamp := time.Now().Unix()
	msg := &Message{Payload: &Message_Commit{Commit: commit}}
	payload, err := proto.Marshal(msg)
	if err != nil {
		logger.Errorf("cMsg marshal error")
		return nil
	}
	cMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Timestamp: 	timestamp,
		Payload:	payload,
		Id:		id,
	}
	return cMsg
}
