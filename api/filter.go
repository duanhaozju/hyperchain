package api

import (
	"github.com/op/go-logging"
	"hyperchain/common"
	"sync"
	flt "hyperchain/manager/filter"
	"time"
	edb "hyperchain/core/db_utils"
	"hyperchain/manager"
	"hyperchain/manager/event"
	"hyperchain/core/types"
	"context"
)

type PublicFilterAPI struct {
	namespace string
	eh        *manager.EventHub
	config    *common.Config
	log       *logging.Logger
	events    *flt.EventSystem
	filtersMu sync.Mutex
	filters   map[string]*flt.Filter
}

func NewFilterAPI(namespace string, eh *manager.EventHub, config *common.Config) *PublicFilterAPI {
	log := common.GetLogger(namespace, "api")
	api := &PublicFilterAPI{
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

func (api *PublicFilterAPI) NewEventSubscription(crit flt.FilterCriteria) string {
	var (
		logCh    = make(chan interface{})
		logSub   = api.events.NewCommonSubscription(logCh, false, flt.LogsSubscription, crit)
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

func (api *PublicFilterAPI) NewSnapshotSubscription() string {
	var (
		ch = make(chan interface{})
		sub = api.events.NewCommonSubscription(ch, false, flt.SnapshotSubscription, flt.FilterCriteria{})
	)
	api.filtersMu.Lock()
	api.filters[sub.ID] = flt.NewFilter(flt.SnapshotSubscription, sub, flt.FilterCriteria{})
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

func (api *PublicFilterAPI) NewExceptionSubscription(crit flt.FilterCriteria) string {
	var (
		ch     = make(chan interface{})
		sub   = api.events.NewCommonSubscription(ch, false, flt.ExceptionSubscription, crit)
	)
	api.filtersMu.Lock()
	api.filters[sub.ID] = flt.NewFilter(flt.ExceptionSubscription, sub, crit)
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
// last time is was called. This can be used for polling.
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
					if !ok { continue }
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
		case flt.SnapshotSubscription:
			datas := f.GetData()
			var ret []event.FilterSnapshotEvent
			for _, d := range datas {
				if val, ok := d.(event.FilterSnapshotEvent); ok {
					ret = append(ret, val)
				}
			}
			return ret, nil
		case flt.ExceptionSubscription:
			data := f.GetData()
			defer f.ClearData()
			return returnException(data), nil
		}
	}

	return []interface{}{}, &common.SubNotExistError{Message: "required subscription does not existed or has expired"}
}

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
	return ret.ToLogsTrans(types.Receipt_EVM)
}

func returnException(data []interface{}) []event.FilterExceptionEvent {
	if len(data) == 0 {
		return []event.FilterExceptionEvent{}
	}
	var ret []event.FilterExceptionEvent
	for _, d := range data {
		if val, ok := d.(event.FilterExceptionEvent); ok {
			ret = append(ret, val)
		}
	}
	return ret
}

/*******************************************************************************
**									      **
**			Webscoket Event	Subscription			      **
**									      **
********************************************************************************/

// Block creates a subscription that send a notification each time when a new block is appended to the chain.
func (api *PublicFilterAPI) Block(ctx context.Context) (common.ID, error) {

	api.log.Debug("ready to deal with newBlock event request")
	return api.handleWSSubscribe(ctx, flt.BlocksSubscription, flt.FilterCriteria{})
}

// Exception creates a subscription that send a notification each time when exception is threw.
func (api *PublicFilterAPI) Exception(ctx context.Context, crit flt.FilterCriteria) (common.ID, error) {

	api.log.Debug("ready to deal with newException event request")
	return api.handleWSSubscribe(ctx, flt.ExceptionSubscription, crit)
}

func (api *PublicFilterAPI) handleWSSubscribe(ctx context.Context, typ flt.Type, crit flt.FilterCriteria) (common.ID, error) {

	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	common.CtxCh <- ctx
	subChs := common.GetSubChs(ctx)

	select {
	case err := <- subChs.Err:
		return common.ID(""), err
	case rpcSub := <- subChs.SubscriptionCh:
		api.log.Debugf("receive subscription %v", rpcSub.ID)

		go func() {

			ch     := make(chan interface{})
			sub    := api.events.NewCommonSubscription(ch, false, typ, crit)

			for {
				select {
				case d := <-ch:
					api.log.Debugf("receive data")
					payload := common.NotifyPayload{
						SubID: rpcSub.ID,
						Data:  d,
					}

					subChs.NotifyDataCh <- payload
				case <-rpcSub.Err():	 // unsubscribe
					sub.Unsubscribe()
					return
				case <-subChs.Closed(): // connection close
					api.log.Debug("the websocket connection closed, release resource")
					sub.Unsubscribe()
					return
				}
			}
		}()

		return rpcSub.ID, nil
	}
}