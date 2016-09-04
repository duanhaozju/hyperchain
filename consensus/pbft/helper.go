package pbft

import (
	"time"

	pb "hyperchain/protos"

	"github.com/golang/protobuf/proto"
)

// consensusMsgHelper help convert the ConsensusMessage to pb.Message
func consensusMsgHelper(msg *ConsensusMessage, id uint64) *pb.Message {

	msgPayload, err := proto.Marshal(msg)

	if err != nil {
		logger.Errorf("ConsensusMessage Marshal Error", err)
		return nil
	}

	pbMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Payload:	msgPayload,
		Timestamp:	time.Now().UnixNano(),
		Id:		id,
	}

	return pbMsg
}

// pbftMsgHelper help convert the pbftMessage to pb.Message
func pbftMsgHelper(msg *Message, id uint64) *pb.Message {

	consensusMsg := &ConsensusMessage{Payload: &ConsensusMessage_PbftMessage{PbftMessage: msg}}
	pbMsg := consensusMsgHelper(consensusMsg, id)

	return pbMsg
}

// exeBatchHelper help convert the RequestBatch to pb.ExeMessage
func exeBatchHelper(reqBatch *RequestBatch, no uint64) *pb.ExeMessage {

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

	exeMsg := &pb.ExeMessage{
		Batch:		batches,
		Timestamp:	time.Now().UnixNano(),
		No:		no,
	}

	return exeMsg
}