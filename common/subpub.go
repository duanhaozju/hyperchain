package common

import (
	"sync"
	"context"
)

// ID defines a pseudo random number that is used to identify RPC subscriptions.
type ID string

// a Subscription is created by a notifier and tight to that notifier. The client can use
// this subscription to wait for an unsubscribe request for the client, see Err().
type Subscription struct {
	ID        ID
	Service string
	Method  string
	Namespace string
	Error       chan error // closed on unsubscribe
}

// Err returns a channel that is closed when the client send an unsubscribe request.
func (s *Subscription) Err() <-chan error {
	return s.Error
}

type NotifyPayload struct {
	SubID ID
	Data  interface{}
}

type Subchan struct {
	Ctx 		  context.Context
	//Mux               sync.Mutex
	SubscriptionChan  chan *Subscription
	NotifyDataChan    chan NotifyPayload
	Err		  chan error		// context error
	closed	          chan interface{}	// connection close
						 //Unsubscribe       chan ID	// event unsubscribe
}

var (
	SubCtxChan map[context.Context]*Subchan
	CtxChan chan context.Context
	mux sync.Mutex

)

func init() {
	SubCtxChan = make(map[context.Context]*Subchan)
	CtxChan = make(chan context.Context)
}

func GetSubChan(ctx context.Context) (*Subchan) {
	mux.Lock()
	defer mux.Unlock()

	if subchan, ok := SubCtxChan[ctx]; ok {
		return subchan
	} else {
		subchan := &Subchan{
			Ctx:		  ctx,
			SubscriptionChan: make(chan *Subscription),
			NotifyDataChan:   make(chan NotifyPayload),
			Err:              make(chan error),
			closed:           make(chan interface{}),
		}

		SubCtxChan[ctx] = subchan

		return subchan
	}
}

func (subchan *Subchan) Close() {
	// todo 是否需要加锁？
	close(subchan.closed)
	close(subchan.Err)
	delete(SubCtxChan, subchan.Ctx)
}


func (subchan *Subchan) Closed() <- chan interface{} {
	return subchan.closed
}
