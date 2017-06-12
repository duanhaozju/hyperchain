package filter

import (
	"errors"
	"fmt"
	"hyperchain/common"
	"hyperchain/core/vm"
	"hyperchain/manager/event"
	"time"
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
	SnapshotSubscription
	DelSnapshotSubscription
	ArchiveSubscription
	// LastSubscription keeps track of the last index
	LastIndexSubscription
)

const (
	LatestBlock         int64 = -1
	EarliestBlockNumber int64 = 0
)

var (
	ErrInvalidSubscriptionID = errors.New("invalid id")
)

type filterIndex map[Type]map[string]*subscription

// EventSystem creates subscriptions, processes events and broadcasts them to the
// subscription which match the subscription criteria.
type EventSystem struct {
	mux        *event.TypeMux
	installC   chan *subscription
	uninstallC chan *subscription
}

// NewEventSystem creates a new manager that listens for event on the given mux,
// parses and filters them. It uses the all map to retrieve filter changes. The
// work loop holds its own index that is used to forward events to filters.
//
// The returned manager has a loop that needs to be stopped with the Stop function
// or by stopping the given mux.
func NewEventSystem(mux *event.TypeMux) *EventSystem {
	m := &EventSystem{
		mux:        mux,
		installC:   make(chan *subscription),
		uninstallC: make(chan *subscription),
	}
	go m.eventLoop()
	return m
}

// eventLoop (un)installs filters and processes mux events.
func (es *EventSystem) eventLoop() {
	var (
		index = make(filterIndex)
		sub   = es.mux.Subscribe(event.FilterNewBlockEvent{}, event.FilterNewLogEvent{}, event.FilterArchive{},
			event.FilterSnapshotEvent{}, event.FilterDeleteSnapshotEvent{})
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
func (es *EventSystem) broadcast(filters filterIndex, obj *event.Event) {
	if obj == nil {
		return
	}
	// dispatch all events
	switch ev := obj.Data.(type) {
	case event.FilterNewBlockEvent:
		for _, f := range filters[BlocksSubscription] {
			if obj.Time.After(f.created) {
				f.hashes <- common.BytesToHash(ev.Block.BlockHash)
			}
		}
	case event.FilterNewLogEvent:
		for _, f := range filters[LogsSubscription] {
			if obj.Time.After(f.created) {
				// filter logs
				ret := filterLogs(ev.Logs, &f.logsCrit)
				if len(ret) != 0 {
					f.logs <- ret
				}
			}
		}
	case event.FilterSnapshotEvent:
		for _, f := range filters[SnapshotSubscription] {
			if obj.Time.After(f.created) {
				f.extra <- ev
			}
		}
	case event.FilterDeleteSnapshotEvent:
		for _, f := range filters[DelSnapshotSubscription] {
			if obj.Time.After(f.created) {
				f.extra <- ev
			}
		}
	case event.FilterArchive:
		for _, f := range filters[ArchiveSubscription] {
			if obj.Time.After(f.created) {
				f.extra <- ev
			}
		}
	}
}

// subscribe installs the subscription in the event broadcast loop.
func (es *EventSystem) subscribe(sub *subscription) *Subscription {
	es.installC <- sub
	<-sub.installed
	return &Subscription{ID: sub.id, f: sub, es: es}
}

func (es *EventSystem) NewBlockSubscription(blockC chan common.Hash, isVerbose bool) *Subscription {
	sub := &subscription{
		id:        NewFilterID(),
		verbose:   isVerbose,
		typ:       BlocksSubscription,
		created:   time.Now(),
		logs:      make(chan []*vm.Log),
		hashes:    blockC,
		extra:     make(chan interface{}),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}

func (es *EventSystem) NewLogSubscription(logsCrit FilterCriteria, logC chan []*vm.Log) *Subscription {
	fmt.Println("############## [Criterias] ##############")
	fmt.Println("from block:", logsCrit.FromBlock)
	fmt.Println("to block:", logsCrit.ToBlock)
	sub := &subscription{
		id:        NewFilterID(),
		verbose:   true,
		typ:       LogsSubscription,
		created:   time.Now(),
		logsCrit:  logsCrit,
		logs:      logC,
		hashes:    make(chan common.Hash),
		extra:     make(chan interface{}),
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}

func (es *EventSystem) NewCommonSubscription(ch chan interface{}, verbose bool) *Subscription {
	sub := &subscription{
		id:        NewFilterID(),
		verbose:   verbose,
		typ:       SnapshotSubscription,
		created:   time.Now(),
		logsCrit:  FilterCriteria{},
		logs:      make(chan []*vm.Log),
		hashes:    make(chan common.Hash),
		extra:     ch,
		installed: make(chan struct{}),
		err:       make(chan error),
	}
	return es.subscribe(sub)
}
