package api

import (
	"fmt"
	"github.com/op/go-logging"
	"hyperchain/common"
	edb "hyperchain/core/ledger/chain"
	"hyperchain/core/types"
	"hyperchain/manager/event"
	flt "hyperchain/manager/filter"
	"sync"
	"time"
)

type PublicFilterAPI struct {
	namespace 	string
	config    	*common.Config
	log       	*logging.Logger
	events   	*flt.EventSystem
	filtersMu 	sync.Mutex
	filters   	map[string]*flt.Filter
}

func NewFilterAPI(namespace string,eventSystem  *flt.EventSystem, config *common.Config) *PublicFilterAPI {
	log := common.GetLogger(namespace, "api")
	api := &PublicFilterAPI{
		namespace: namespace,
		config:    config,
		log:       log,
		events:    eventSystem,
		filters:   make(map[string]*flt.Filter),
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

// NewBlockSubscription creates a filter that fetches blocks that are imported into the chain.
// It is part of the filter package since polling goes with getFilterChanges.
func (api *PublicFilterAPI) NewBlockSubscription(isVerbose bool) string {
	var (
		blockCh  = make(chan interface{})
		blockSub = api.events.NewCommonSubscription(blockCh, isVerbose, flt.BlocksSubscription, flt.FilterCriteria{})
	)
	api.filtersMu.Lock()
	api.filters[blockSub.ID] = flt.NewFilter(flt.BlocksSubscription, blockSub, flt.FilterCriteria{})
	api.filtersMu.Unlock()

	go func() {
		for {
			select {
			case b := <-blockCh:
				api.filtersMu.Lock()
				if f, found := api.filters[blockSub.ID]; found {
					f.AddData(b)
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

// NewEventSubscription subscribes a vm event. A filter will be created to fetch vm event,
// subscription ID will be returned.
func (api *PublicFilterAPI) NewEventSubscription(crit flt.FilterCriteria) string {
	var (
		logCh  = make(chan interface{})
		logSub = api.events.NewCommonSubscription(logCh, false, flt.LogsSubscription, crit)
	)
	api.filtersMu.Lock()
	api.filters[logSub.ID] = flt.NewFilter(flt.LogsSubscription, logSub, crit)
	api.filtersMu.Unlock()

	go func() {
		for {
			select {
			case b := <-logCh:
				api.filtersMu.Lock()
				if f, found := api.filters[logSub.ID]; found {
					f.AddData(b)
				}
				api.filtersMu.Unlock()
			case <-logSub.Err():
				api.filtersMu.Lock()
				delete(api.filters, logSub.ID)
				api.filtersMu.Unlock()
				return
			}
		}
	}()

	return logSub.ID

}

// NewArchiveSubscription subscribes a archive event. A filter will be created to fetch archive information,
// Subscription ID will be returned.
func (api *PublicFilterAPI) NewArchiveSubscription() string {
	var (
		ch  = make(chan interface{})
		sub = api.events.NewCommonSubscription(ch, false, flt.ArchiveSubscription, flt.FilterCriteria{})
	)
	api.filtersMu.Lock()
	api.filters[sub.ID] = flt.NewFilter(flt.ArchiveSubscription, sub, flt.FilterCriteria{})
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

// NewSystemStatusSubscription subscribes a system status event. Subscription ID will be returned.
func (api *PublicFilterAPI) NewSystemStatusSubscription(crit flt.FilterCriteria) string {
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

// GetFilterChanges returns the logs for the filter with the given id since
// last time was called. This can be used for polling.
//
// For pending transaction and block filters the result is []common.Hash.
// (pending)Log filters return []Log.
//
func (api *PublicFilterAPI) GetSubscriptionChanges(id string) (interface{}, error) {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	if f, found := api.filters[id]; found {
		if !f.GetDeadLine().Stop() {
			// timer expired but filter is not yet removed in timeout loop
			// receive timer value and reset timer
			<-f.GetDeadLine().C
		}
		f.ResetDeadline()

		switch f.GetType() {
		case flt.BlocksSubscription:
			if f.GetVerbose() {
				var ret []*BlockResult
				hashes := f.GetData()
				defer f.ClearData()
				for _, tmp := range hashes {
					hash, ok := tmp.(common.Hash)
					if !ok {
						continue
					}
					block, err := edb.GetBlock(api.namespace, hash.Bytes())
					if err != nil {
						api.log.Warningf("missing block data (#%s)", hash.Hex())
					} else {
						if wrappedBlock, err := outputBlockResult(api.namespace, block, true); err == nil {
							ret = append(ret, wrappedBlock)
						} else {
							api.log.Warningf("wrapper block data (#%s) failed", hash.Hex())
						}
					}
				}
				return ret, nil
			} else {
				hashes := f.GetData()
				defer f.ClearData()
				return returnHashes(hashes), nil
			}
		case flt.LogsSubscription:
			logs := f.GetData()
			defer f.ClearData()
			return returnLogs(logs), nil
		case flt.ArchiveSubscription:
			datas := f.GetData()
			defer f.ClearData()
			return datas, nil
		case flt.SystemStatusSubscription:
			data := f.GetData()
			defer f.ClearData()
			return returnSystemStatus(data), nil
		}
	}

	return []interface{}{}, &common.SubNotExistError{Message: "required subscription does not existed or has expired"}
}

// UnSubscription unsubscribes a given event.
func (api *PublicFilterAPI) UnSubscription(id string) error {
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()
	if f, found := api.filters[id]; found {
		f.GetSubsctiption().Unsubscribe()
		delete(api.filters, id)
		return nil
	}
	return &common.SubNotExistError{Message: "required subscription does not existed or has expired"}
}

// GetLogs returns eligible vm event logs.
func (api *PublicFilterAPI) GetLogs(crit flt.FilterCriteria) (interface{}, error) {
	genesis, err := edb.GetGenesisTag(api.namespace)
	if err != nil {
		return nil, &common.CallbackError{Message: "obtain genesis tag failed"}
	}
	head := edb.GetHeightOfChain(api.namespace)

	var beginNo, endNo uint64
	if crit.FromBlock == nil || crit.FromBlock.Uint64() < genesis {
		beginNo = genesis
	} else {
		beginNo = crit.FromBlock.Uint64()
	}

	if crit.ToBlock == nil || crit.ToBlock.Uint64() > head {
		endNo = head
	} else {
		endNo = crit.ToBlock.Uint64()
	}

	if beginNo >= endNo {
		return nil, &common.InvalidParamsError{Message: fmt.Sprintf("invalid params. current genesis %d, current head %d", genesis, head)}
	}

	searcher := flt.NewLogSearcher(beginNo, endNo, crit.Addresses, crit.Topics, api.namespace)
	return types.Logs(searcher.Search()).ToLogsTrans(), nil

}

// returnHashes is a helper that will return an empty hash array case the given hash array is nil,
// otherwise the given hashes array is returned.
func returnHashes(hashes []interface{}) []common.Hash {
	if hashes == nil {
		// Short circuit if hashes is empty
		return []common.Hash{}
	}
	var ret []common.Hash
	for _, tmp := range hashes {
		if hash, ok := tmp.(common.Hash); ok {
			ret = append(ret, hash)
		}
	}
	return ret
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []interface{}) []types.LogTrans {
	if logs == nil {
		// Short circuit if logs is empty
		return []types.LogTrans{}
	}
	var ret types.Logs
	for _, tmp := range logs {
		if log, ok := tmp.([]*types.Log); ok {
			ret = append(ret, log...)
		}
	}
	return ret.ToLogsTrans()
}

// returnSystemStatus is a helper that will return an empty system status array in case the given
// system status array is nil, otherwise the given logs array is returned.
func returnSystemStatus(data []interface{}) []event.FilterSystemStatusEvent {
	if len(data) == 0 {
		return []event.FilterSystemStatusEvent{}
	}
	var ret []event.FilterSystemStatusEvent
	for _, d := range data {
		if val, ok := d.(event.FilterSystemStatusEvent); ok {
			ret = append(ret, val)
		}
	}
	return ret
}
