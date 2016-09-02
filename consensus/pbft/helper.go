package pbft

import (
	"time"

	pb "hyperchain/protos"
	"github.com/golang/protobuf/proto"
)

func consensusMsgHelper(msg *ConsensusMessage, id uint64) *pb.Message {

	msgPayload, _ := proto.Marshal(msg)
	pbMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Payload:	msgPayload,
		Timestamp:	time.Now().UnixNano(),
		Id:		id,
	}

	return pbMsg
}

func pbftMsgHelper(msg *Message, id uint64) *pb.Message {

	consensusMsg := &ConsensusMessage{Payload: &ConsensusMessage_PbftMessage{PbftMessage: msg}}
	pbMsg := consensusMsgHelper(consensusMsg, id)

	return pbMsg
}

func exeBatchHelper(reqBatch *RequestBatch) *pb.ExeMessage {

	batches := []*pb.Message{}
	requests := reqBatch.Batch
	for i := 0; i < len(requests); i++ {
		batch := &pb.Message{
			Timestamp:	requests[i].Timestamp,
			Payload:	requests[i].Payload,
			Id:		requests[i].ReplicaId,
		}
		batches = append(batches, batch)
	}
	exeMsg := &pb.ExeMessage{Batch: batches}
	return exeMsg

}