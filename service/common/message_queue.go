package common

//MessageQueue queue to store messages.
type MessageQueue interface{
	Put(msg interface{}) error
	Get() (interface{}, error)
}

//MQImpl message queue implementation.
type MQImpl struct {
	messages chan interface{}
}

func newMQImpl(size int) *MQImpl {
	mq := &MQImpl{
		messages:make(chan interface{}, size),
	}
	return mq
}

func (mq *MQImpl)Put(msg interface{}) error  {
	mq.messages <- msg
	return nil
}

func (mq *MQImpl)Get() (interface{}, error)  {
	msg := mq.messages
	return msg, nil
}


