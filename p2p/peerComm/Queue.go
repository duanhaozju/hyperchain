package peerComm

import "errors"

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

//指定大小进行初始化
func NewQueueBySize(size int) *Queue{
	queueSize = size
	return &Queue{
		front:0,
		rear:size -1,
		currentCount:0,
		elements:make([]interface{},size),
	}
}


//用默认大小进行初始化
func NewQueue()*Queue{
	return NewQueueBySize(defaultQueueSize)
}

func ProbeNext(i int) int{
	return (i+1)%queueSize
}
//清空队列
func (queue *Queue)Clear(){
	queue.front = 0;
	queue.rear = queueSize -1;
	queue.currentCount = 0;
}
//是否为空队列
func(queue *Queue) IsEmpty() bool{
	if ProbeNext(queue.rear) == queue.front{
		return true
	}
	return false
}

//队列是否满
func (queue *Queue) IsFull()bool{
	if ProbeNext(ProbeNext(queue.rear)) == queue.front{
		return true
	}else{
		return false
	}
}

// 入队
func(queue *Queue)Enqueue(ele interface{}) error{
	if queue.IsFull() {
		return errors.New("the queue is full")
	}
	queue.rear = ProbeNext(queue.rear)
	queue.elements[queue.rear] = ele
	queue.currentCount = queue.currentCount + 1
	return nil
}
// 出队
func(queue *Queue) Dequeue()(interface{},error){
	if queue.IsEmpty(){
		return nil,errors.New("the queue is empty")
	}
	tmp := queue.front
	queue.front = ProbeNext(queue.front)
	queue.currentCount = queue.currentCount - 1
	return queue.elements[tmp],nil
}