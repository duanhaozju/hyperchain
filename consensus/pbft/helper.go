package pbft

import (
	"time"

	pb "hyperchain/protos"
	"github.com/golang/protobuf/proto"
)

func batchMsgHelper(msg *BatchMessage, id uint64) *pb.Message {
	msgPayload, _ := proto.Marshal(msg)
	pbMsg := &pb.Message{
		Type:		pb.Message_CONSENSUS,
		Payload:	msgPayload,
		Timestamp:	time.Now().Unix(),
		Id:		id,
	}
	return pbMsg
}

func pbftMsgHelper(msg *Message, id uint64) *pb.Message {
	pbftPayload, _ := proto.Marshal(msg)
	batchMsg := &BatchMessage{Payload: &BatchMessage_PbftMessage{PbftMessage: pbftPayload}}
	pbMsg := batchMsgHelper(batchMsg, id)
	return pbMsg
}

