package api

import (
	"hyperchain/manager"
	"hyperchain/common"
	"github.com/op/go-logging"
	"sync"
	flt "hyperchain/manager/filter"
	"time"
	"hyperchain/core/vm"
	edb "hyperchain/core/db_utils"
	"context"
)

type PublicFilterAPI struct {
	namespace   string
	eh          *manager.EventHub
	config      *common.Config
	log         *logging.Logger
	events      *flt.EventSystem
	filtersMu   sync.Mutex
	filters     map[string]*flt.Filter
	subscriptions      map[context.Context][]*flt.Subscription
}

//func NewFilterAPI(namespace string, eh *manager.EventHub, config *common.Config, subchan *common.Subchan) *PublicFilterAPI {
func NewFilterAPI(namespace string, eh *manager.EventHub, config *common.Config) *PublicFilterAPI {
	log := common.GetLogger(namespace, "api")
	api := &PublicFilterAPI{
		namespace:   	namespace,
		eh:          	eh,
		config:      	config,
		log:         	log,
		events:     	eh.GetFilterSystem(),
		filters:     	make(map[string]*flt.Filter),
		subscriptions:	make(map[context.Context][]*flt.Subscription, 0),
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
		blockC   = make(chan common.Hash)
		blockSub = api.events.NewBlockSubscription(blockC, isVerbose)
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
					f.AddHash(b)
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
		logC     = make(chan []*vm.Log)
		logSub   = api.events.NewLogSubscription(crit, logC)
	)
	api.filtersMu.Lock()
	api.filters[logSub.ID] = flt.NewFilter(flt.LogsSubscription, logSub, crit)
	api.filtersMu.Unlock()

	go func() {
		for {
			select {
			case b := <-logC:
				api.filtersMu.Lock()
				if f, found := api.filters[logSub.ID]; found {
					f.AddLog(b)
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
				hashes := f.GetHashes()
				defer f.ClearHash()
				for _, hash := range hashes {
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
				hashes := f.GetHashes()
				defer f.ClearHash()
				return returnHashes(hashes), nil
			}
		case flt.LogsSubscription:
			logs := f.GetLogs()
			defer f.Clearlog()
			return returnLogs(logs), nil
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
func returnHashes(hashes []common.Hash) []common.Hash {
	if hashes == nil {
		return []common.Hash{}
	}
	return hashes
}

// returnLogs is a helper that will return an empty log array in case the given logs array is nil,
// otherwise the given logs array is returned.
func returnLogs(logs []*vm.Log) []vm.LogTrans {
	if logs == nil {
		return []vm.LogTrans{}
	}
	_logs := vm.Logs(logs)
	return _logs.ToLogsTrans()
}

// ===================== test ================
func (api *PublicFilterAPI) NewBlock(ctx context.Context) (common.ID, error) {

	//api.subchan.Mux.Lock()
	api.filtersMu.Lock()
	defer api.filtersMu.Unlock()

	api.log.Debug("ready to deal with newBlock event request")
	common.CtxChan <- ctx

	select {
		case err := <- common.GetSubChan(ctx).Err:
			//api.subchan.Mux.Unlock()
			common.DelSubChan(ctx)
			return common.ID(""), err
		case rpcSub := <- common.GetSubChan(ctx).SubscriptionChan:
			//api.subchan.Mux.Unlock()
			api.log.Debugf("receive subscription %v\n", rpcSub.ID)

			go func() {

				blockC   := make(chan common.Hash)
				blockSub := api.events.NewBlockSubscription(blockC, false)
				api.subscriptions[ctx] = append(api.subscriptions[ctx], blockSub)

				for {
					select {
					case h := <-blockC:
						api.log.Debugf("receive block %v\n", h.Hex())
						payload := common.NotifyPayload{
							SubID: rpcSub.ID,
							Data:  h,
						}

						common.GetSubChan(ctx).NotifyDataChan <- payload
						//notifier.Notify(rpcSub.ID, h)
					case <-rpcSub.Err(): // todo 暂时没用到，但是这里也要改,以及subchan的err的处理; 管道重构，不直接引用管道，而是方法返回
						blockSub.Unsubscribe()
						//common.GetSubChan(ctx).Unsubscribe <- rpcSub.ID
						return
					case <-common.GetSubChan(ctx).Closed: // connection close, unsubscribe all the subscription in this context
						api.log.Debug("the websocket connection closed, release resource")
						api.unsubscribe(ctx)
						common.DelSubChan(ctx) // delete the communication channel of this context
						return
					}
				}
			}()

			return rpcSub.ID, nil
	}

}

func (api *PublicFilterAPI) unsubscribe(ctx context.Context) {
	for _,sub := range api.subscriptions[ctx] {
		sub.Unsubscribe()
	}
	delete(api.subscriptions, ctx)
}
