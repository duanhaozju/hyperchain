package peerComm

import "errors"
// this file define a normal queue
// this queue struct is used for peer pool
// but hasn't used because of queue can't optimize the performance
// last modify: 2016-09-05

const(
	defaultQueueSize = 10
)

var queueSize int
type Queue struct{
	front int
	rear int
	currentCount int
	elements []interface{}
}

//NewQueueBySize init the queue by clearly size
func NewQueueBySize(size int) *Queue{
	queueSize = size
	return &Queue{
		front:0,
		rear:size -1,
		currentCount:0,
		elements:make([]interface{},size),
	}
}


//NewQueue return a default size queue instance
func NewQueue()*Queue{
	return NewQueueBySize(defaultQueueSize)
}

func ProbeNext(i int) int{
	return (i+1)%queueSize
}
//Clear clear the queue
func (queue *Queue)Clear(){
	queue.front = 0;
	queue.rear = queueSize -1;
	queue.currentCount = 0;
}
//IsEmpty judge the queue is empty or not
func(queue *Queue) IsEmpty() bool{
	if ProbeNext(queue.rear) == queue.front{
		return true
	}
	return false
}

//IsFull judge the queue is full or not
func (queue *Queue) IsFull()bool{
	if ProbeNext(ProbeNext(queue.rear)) == queue.front{
		return true
	}else{
		return false
	}
}

// Enqueue enqueue
func(queue *Queue)Enqueue(ele interface{}) error{
	if queue.IsFull() {
		return errors.New("the queue is full")
	}
	queue.rear = ProbeNext(queue.rear)
	queue.elements[queue.rear] = ele
	queue.currentCount = queue.currentCount + 1
	return nil
}
// Dequeue dequeue
func(queue *Queue) Dequeue()(interface{},error){
	if queue.IsEmpty(){
		return nil,errors.New("the queue is empty")
	}
	tmp := queue.front
	queue.front = ProbeNext(queue.front)
	queue.currentCount = queue.currentCount - 1
	return queue.elements[tmp],nil
}