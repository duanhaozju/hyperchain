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
	"fmt"
)

type PublicFilterAPI struct {
	namespace   string
	eh          *manager.EventHub
	config      *common.Config
	log         *logging.Logger
	events      *flt.EventSystem
	filtersMu   sync.Mutex
	filters     map[string]*flt.Filter
	subchan     *common.Subchan
}

func NewFilterAPI(namespace string, eh *manager.EventHub, config *common.Config, subchan *common.Subchan) *PublicFilterAPI {
	log := common.GetLogger(namespace, "api")
	api := &PublicFilterAPI{
		namespace:   namespace,
		eh:          eh,
		config:      config,
		log:         log,
		events:      eh.GetFilterSystem(),
		filters:     make(map[string]*flt.Filter),
		subchan:     subchan,
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
func (api *PublicFilterAPI) NewTxTest(ctx context.Context) (common.ID, error) {

	// todo 1. 对于订阅请求，往common包的subchan管道传入ctx，一定要将管道锁住，直到值被读出来以后才解锁（缺点：无法并发处理）
	// todo 这样的话得用三个管道，一个ctx，一个subid，一个告知通知的管道。是否可以用事件系统来替代？

	api.subchan.Mux.Lock()
	fmt.Println("ready to deal with 'context'")
	api.subchan.CtxChan <- ctx

	select {
	case err := <- api.subchan.Err:
		api.subchan.Mux.Unlock()
		return common.ID(""), err
	case rpcSub := <- api.subchan.SubscriptionChan:
		api.subchan.Mux.Unlock()
		fmt.Println("created subscription")

		go func() {
			//txHashes := make(chan common.Hash)
			//pendingTxSub := api.events.SubscribePendingTxEvents(txHashes)

			blockC   := make(chan common.Hash)
			blockSub := api.events.NewBlockSubscription(blockC, false)

			for {
				select {
				case h := <-blockC:
					fmt.Println("receive block")
				        payload := common.NotifyPayload{
						SubID: rpcSub.ID,
						Data:  h,
					}
				        api.subchan.NotifyDataChan <- payload
					//notifier.Notify(rpcSub.ID, h)
				case <-rpcSub.Err():
					blockSub.Unsubscribe()
					return
				//case <-notifier.Closed():
				//	pendingTxSub.Unsubscribe()
				//	return
				}
			}
		}()

		return rpcSub.ID, nil
	}

}

