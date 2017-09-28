//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package common

import (
	"context"
	"sync"
)

const (
	SubscribeMethodSuffix    = "_subscribe"
	UnsubscribeMethodSuffix  = "_unsubscribe"
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

// NotifyPayload created by server will be pushed to client.
type NotifyPayload struct {
	SubID ID
	Data  interface{}
}

// SubChs is the type that wraps several channel used to communication between jsonrpc module and api module, which
// prevents package circular import. When a websocket connection is established, we will create a new SubChs instance.
type SubChs struct {

	// context is unique identifier of a SubChs instance, it corresponds to a websocket connection
	Ctx context.Context

	SubscriptionCh chan *Subscription
	NotifyDataCh   chan NotifyPayload
	Err            chan error       	// error happened in the context
	closed         chan interface{} 	// connection close
	closer		   sync.Once		    // close once
}

var (
	// CtxCh channel will input a context when api service receive a request to subscribe, while notifier will
	// output it in a for loop and then create a subscription, this subscription will be input SubChs.SubscriptionCh.
	CtxCh     chan context.Context

	// subChsMap records current websocket connection resource, releases resource when error happened or connection closed.
	subChsMap map[context.Context]*SubChs
	mux       sync.Mutex
)

func init() {
	subChsMap = make(map[context.Context]*SubChs)
	CtxCh = make(chan context.Context)
}

// GetSubChs will create a new SubChs instance for the given context or return it if existed.
func GetSubChs(ctx context.Context) *SubChs {
	mux.Lock()
	defer mux.Unlock()

	if subChs, ok := subChsMap[ctx]; ok {
		return subChs
	} else {
		subChs := &SubChs{
			Ctx:            ctx,
			SubscriptionCh: make(chan *Subscription),
			NotifyDataCh:   make(chan NotifyPayload),
			Err:            make(chan error),
			closed:         make(chan interface{}),
		}

		subChsMap[ctx] = subChs

		return subChs
	}
}

// Close releases subChs resource when connection closed.
func (subChs *SubChs) Close() {
	subChs.closer.Do(func() {
		close(subChs.closed)
		close(subChs.Err)
		delete(subChsMap, subChs.Ctx)
	})
}

func (subChs *SubChs) Closed() <-chan interface{} {
	return subChs.closed
}
