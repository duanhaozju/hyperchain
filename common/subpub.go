package common

import (
	"context"
	"sync"
)

// ID defines a pseudo random number that is used to identify RPC subscriptions.
type ID string

// a Subscription is created by a notifier and tight to that notifier. The client can use
// this subscription to wait for an unsubscribe request for the client, see Err().
type Subscription struct {
	ID        ID
	Service   string
	Method    string
	Namespace string
	Error     chan error // closed on unsubscribe
}

// Err returns a channel that is closed when the client send an unsubscribe request.
func (s *Subscription) Err() <-chan error {
	return s.Error
}

type NotifyPayload struct {
	SubID ID
	Data  interface{}
}

type SubChs struct {
	Ctx context.Context
	//Mux           sync.Mutex
	SubscriptionCh chan *Subscription
	NotifyDataCh   chan NotifyPayload
	Err            chan error       // context error
	closed         chan interface{} // connection close
	//Unsubscribe       chan ID	// event unsubscribe
}

var (
	SubChsMap map[context.Context]*SubChs
	CtxCh     chan context.Context
	mux       sync.Mutex
)

func init() {
	SubChsMap = make(map[context.Context]*SubChs)
	CtxCh = make(chan context.Context)
}

func GetSubChs(ctx context.Context) *SubChs {
	mux.Lock()
	defer mux.Unlock()

	if subChs, ok := SubChsMap[ctx]; ok {
		return subChs
	} else {
		subChs := &SubChs{
			Ctx:            ctx,
			SubscriptionCh: make(chan *Subscription),
			NotifyDataCh:   make(chan NotifyPayload),
			Err:            make(chan error),
			closed:         make(chan interface{}),
		}

		SubChsMap[ctx] = subChs

		return subChs
	}
}

func (subChs *SubChs) Close() {
	// todo 是否需要加锁？
	close(subChs.closed)
	close(subChs.Err)
	delete(SubChsMap, subChs.Ctx)
}

func (subChs *SubChs) Closed() <-chan interface{} {
	return subChs.closed
}
