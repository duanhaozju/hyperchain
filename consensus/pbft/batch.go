package pbft

import (
	"time"
	"fmt"
	"reflect"

	"hyperchain-alpha/consensus/events"
)

type batch struct {
	batchTimer       events.Timer
	batchTimerActive bool
	batchTimeout     time.Duration
	c                chan int8
	manager          events.Manager
}

const configPrefix = "CORE_PBFT"

type testEvent struct {}

func (b *batch) ProcessEvent(e events.Event) events.Event{
	fmt.Println("----ProcessEvent:",reflect.TypeOf(e))
	switch e.(type) {
	case *testEvent:
		fmt.Println("lalalla")
		b.c <- 1
	default:
		fmt.Println("default")
	}
	return nil
}

func (b *batch) RecvMsg(e events.Event){
	fmt.Println("RecvMsg")
	b.manager.Queue()<- e
}

func newBatch(id uint64, c chan *Message) *batch{
	fmt.Println("new batch")
	batchObj:=&batch{
		manager:events.NewManagerImpl(),
		c:c,//TODO   for test
	}
	batchObj.manager.SetReceiver(batchObj)
	batchObj.manager.Start()

	return batchObj
}




