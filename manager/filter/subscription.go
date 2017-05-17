package filter

import (
	"time"
	"sync"
	"hyperchain/core/vm"
	"hyperchain/common"
)

type subscription struct {
	id        string
	verbose   bool
	typ       Type
	created   time.Time
	logsCrit  FilterCriteria
	logs      chan []*vm.Log
	hashes    chan common.Hash
	installed chan struct{} // closed when the filter is installed
	err       chan error    // closed when the filter is uninstalled
}

// FilterCriteria represents a request to create a new filter.
type FilterCriteria struct {
	FromBlock uint64
	ToBlock   uint64
	Addresses []common.Address
	Topics    [][]common.Hash
}

// Subscription is created when the client registers itself for a particular event.
type Subscription struct {
	ID        string
	f         *subscription
	es        *EventSystem
	unsubOnce sync.Once
}
// Err returns a channel that is closed when unsubscribed.
func (sub *Subscription) Err() <-chan error {
	return sub.f.err
}

// Unsubscribe uninstalls the subscription from the event broadcast loop.
func (sub *Subscription) Unsubscribe() {
	sub.unsubOnce.Do(func() {
		uninstallLoop:
		for {
			// write uninstall request and consume logs/hashes. This prevents
			// the eventLoop broadcast method to deadlock when writing to the
			// filter event channel while the subscription loop is waiting for
			// this method to return (and thus not reading these events).
			select {
			case sub.es.uninstallC <- sub.f:
				break uninstallLoop
			case <-sub.f.logs:
			case <-sub.f.hashes:
			}
		}

		// wait for filter to be uninstalled in work loop before returning
		// this ensures that the manager won't use the event channel which
		// will probably be closed by the client asap after this method returns.
		<-sub.Err()
	})
}
