package api

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"hyperchain/manager"
	flt "hyperchain/manager/filter"
	"sync"
	"time"
)

type HpcPublicFilterAPI struct {
	namespace string
	eh        *manager.EventHub
	config    *common.Config
	log       *logging.Logger
	events    *flt.EventSystem
	filtersMu sync.Mutex
	filters   map[string]*flt.Filter
}

func NewHpcFilterAPI(namespace string, eh *manager.EventHub, config *common.Config) *HpcPublicFilterAPI {
	log := common.GetLogger(namespace, "api")
	api := &HpcPublicFilterAPI{
		namespace: namespace,
		eh:        eh,
		config:    config,
		log:       log,
		events:    eh.GetFilterSystem(),
		filters:   make(map[string]*flt.Filter),
	}
	go api.timeoutLoop()
	return api

}

func (api *HpcPublicFilterAPI) timeoutLoop() {
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

// NewSystemStatusSubscription subscribes a system status event. Subscription ID will be returned.
func (api *HpcPublicFilterAPI) NewSystemStatusSubscription(crit flt.FilterCriteria) string {
	var (
		ch  = make(chan interface{})
		sub = api.events.NewCommonSubscription(ch, false, flt.SystemStatusSubscription, crit)
	)
	api.filtersMu.Lock()
	api.filters[sub.ID] = flt.NewFilter(flt.SystemStatusSubscription, sub, crit)
	api.filtersMu.Unlock()

	go func() {
		for {
			select {
			case d := <-ch:
				api.filtersMu.Lock()
				if f, found := api.filters[sub.ID]; found {
					f.AddData(d)
				}
				api.filtersMu.Unlock()
			case <-sub.Err():
				api.filtersMu.Lock()
				delete(api.filters, sub.ID)
				api.filtersMu.Unlock()
				return
			}
		}
	}()
	return sub.ID
}

func (api *HpcPublicFilterAPI) GetSubscriptionChanges(id string) (interface{}, error) {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	if f, found := api.filters[id]; found {
		if !f.GetDeadLine().Stop() {
			// timer expired but filter is not yet removed in timeout loop
			// receive timer value and reset timer
			<-f.GetDeadLine().C
		}
		f.ResetDeadline()

		data := f.GetData()
		defer f.ClearData()
		return returnSystemStatus(data), nil
	}

	return []interface{}{}, &common.SubNotExistError{Message: "required subscription does not existed or has expired"}
}
