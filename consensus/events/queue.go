//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package events

import "hyperchain/event"

//writeOnlyQueue only allow write
type writeOnlyQueue interface {
	Push(event interface{})
}

type Queue interface {
	Push(event interface{})
	Pop() interface{}
	Close()
	Size() int
}

//msgQueue used to send message to other modules
type msgQueue struct {
	tm *event.TypeMux
}

func (q *msgQueue) Push(event interface{})  {
	q.tm.Post(event)
}

type blockingQueue struct {
	queue chan interface{}
	exit chan bool
}

//Push push element into the blocking queue, this
//method will block until the blocking queue is ready.
func (eq *blockingQueue) Push(ele interface{})  {
	eq.queue <- ele
}

//Pop pop element form the blocking queue, this method
//will block until blocking queue has element.
func (eq *blockingQueue) Pop() interface {}{
	select {
	case ele := <- eq.queue:
		return ele
	case <- eq.exit:
		logger.Infof("Queue is closed!")
		return nil
	}
}

//Close close the blocking queue
func (eq *blockingQueue) Close() {
	eq.exit <- true
	eq.queue = nil //block push and pop
}

//Size return the queue size
func (eq *blockingQueue) Size() int {
	return len(eq.queue)
}

//NewBlockingQueue new blocking queue with fixed queue
func NewBlockingQueue(qsize uint64) *blockingQueue {
	if qsize < 1 {
		logger.Errorf("Queue size is too small, size: %d", qsize)
	}
	bq := &blockingQueue{}
	if qsize == 1 {
		bq.queue = make(chan interface{})
	}else {
		bq.queue = make(chan interface{}, qsize)
	}
	bq.exit = make(chan bool)
	return bq
}

//NewSingleEleQueue new a queue just contains one element one time
func NewSingleEleQueue() *blockingQueue {
	return &blockingQueue{
		queue:make(chan interface{}),
		exit:make(chan bool),
	}
}

//GetQueue new a queue using exiting channel.
func GetQueue(existedQueue chan interface{}) *blockingQueue {
	return &blockingQueue{
		queue:existedQueue,
		exit:make(chan bool),
	}
}
