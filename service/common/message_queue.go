package common

import (
	pb "hyperchain/service/common/protos"
)

//MessageQueue queue to store messages.
type MessageQueue interface {
	Put(msg *pb.Message) error
	Get() (*pb.Message, error)
}

//MQImpl message queue implementation.
type MQImpl struct {
	messages chan *pb.Message
}

func newMQImpl(size int) *MQImpl {
	mq := &MQImpl{
		messages: make(chan *pb.Message, size),
	}
	return mq
}

func (mq *MQImpl) Put(msg *pb.Message) error {
	mq.messages <- msg
	return nil
}

func (mq *MQImpl) Get() (*pb.Message, error) {
	msg := <-mq.messages
	return msg, nil
}
