package api

import (
	"hyperchain/manager"
	"hyperchain/common"
	"github.com/op/go-logging"
	"sync"
	flt "hyperchain/manager/filter"
	"time"
	"hyperchain/core/types"
)

type PublicFilterAPI struct {
	namespace   string
	eh          *manager.EventHub
	config      *common.Config
	log         *logging.Logger
	events      *flt.EventSystem
	filtersMu   sync.Mutex
	filters     map[string]*flt.Filter
}

func NewFilterAPI(namespace string, eh *manager.EventHub, config *common.Config) *PublicFilterAPI {
	log := common.GetLogger(namespace, "api")
	api := &PublicFilterAPI{
		namespace:   namespace,
		eh:          eh,
		config:      config,
		log:         log,
		events:      flt.NewEventSystem(eh.GetEventObject()),
		filters:     make(map[string]*flt.Filter),
	}
	go api.timeoutLoop()
	return api

}

// timeoutLoop runs every 5 minutes and deletes filters that have not been recently used.
// Tt is started when the api is created.
func (api *PublicFilterAPI) timeoutLoop() {
	ticker := time.NewTicker(5 * time.Minute)
	for {
		<-ticker.C
		api.filtersMu.Lock()
		for id, f := range api.filters {
			select {
			case <-f.GetDeadLine().C:
				f.GetSubsctiption().Unsubscribe()
				delete(api.filters, id)
			default:
				continue
			}
		}
		api.filtersMu.Unlock()
	}
}

// NewBlockFilter creates a filter that fetches blocks that are imported into the chain.
// It is part of the filter package since polling goes with getFilterChanges.
func (api *PublicFilterAPI) NewBlockFilter() string {
	var (
		blockC   = make(chan *types.Block)
		blockSub = api.events.NewBlockSubscription()
	)
	api.filtersMu.Lock()
	api.filters[blockSub.ID] = flt.NewFilter(flt.BlocksSubscription, blockSub, flt.FilterCriteria{})
	api.filtersMu.Unlock()

	go func() {
		for {
			select {
			case b := <-blockC:
				api.filtersMu.Lock()
				if f, found := api.filters[blockSub.ID]; found {
					f.AddHash(common.BytesToHash(b.BlockHash))
				}
				api.filtersMu.Unlock()
			case <-blockSub.Err():
				api.filtersMu.Lock()
				delete(api.filters, blockSub.ID)
				api.filtersMu.Unlock()
				return
			}
		}
	}()

	return blockSub.ID
}


