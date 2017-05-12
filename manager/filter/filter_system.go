package filter

import (
	"hyperchain/manager/event"
	"errors"
	"time"
	"hyperchain/core/vm"
	"hyperchain/common"
)

// Type determines the kind of filter and is used to put the filter in to
// the correct bucket when added.
type Type byte

const (
	// UnknownSubscription indicates an unknown subscription type
	UnknownSubscription Type = iota
	// LogsSubscription queries for new logs
	LogsSubscription
	// TransactionsSubscription queries for new txs
	TransactionsSubscription
	// BlocksSubscription queries hashes for blocks that are imported
	BlocksSubscription
	// LastSubscription keeps track of the last index
	LastIndexSubscription
)

var (
	ErrInvalidSubscriptionID = errors.New("invalid id")
)


type filterIndex map[Type]map[string]*subscription
// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria.
type EventSystem struct {
	mux           *event.TypeMux
	installC      chan  *subscription
	uninstallC    chan  *subscription
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(mux *event.TypeMux) *EventSystem {
	m := &EventSystem{
		mux:       mux,
		installC:   make(chan *subscription),
		uninstallC: make(chan *subscription),
	}

	return m
}


// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	var (
		index = make(filterIndex)
		sub   = es.mux.Subscribe()
		// sub   = es.mux.Subscribe(core.PendingLogsEvent{}, core.RemovedLogsEvent{}, []*types.Log{}, core.TxPreEvent{}, core.ChainEvent{})
	)

	for i := UnknownSubscription; i < LastIndexSubscription; i++ {
		index[i] = make(map[string]*subscription)
	}

	for {
		select {
		case ev, active := <-sub.Chan():
			if !active { // system stopped
				return
			}
			es.broadcast(index, ev)
		case f := <-es.installC:
			index[f.typ][f.id] = f
			close(f.installed)
		case f := <-es.uninstallC:
			delete(index[f.typ], f.id)
			close(f.err)
		}
	}
}

// broadcast event to filters that match criteria.
func (es *EventSystem) broadcast(filters filterIndex, ev *event.Event) {
	if ev == nil {
		return
	}
	// dispatch all events
}

// subscribe installs the subscription in the event broadcast loop.
func (es *EventSystem) subscribe(sub *subscription) *Subscription {
	es.installC <- sub
	<-sub.installed
	return &Subscription{ID: sub.id, f: sub, es: es}
}


func (es *EventSystem) NewBlockSubscription() *Subscription {
	sub := &subscription{
		id:        NewFilterID(),
		typ:       BlocksSubscription,
		created:   time.Now(),
		logs:      make(chan []*vm.Log),
		hashes:    make(chan common.Hash),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}
