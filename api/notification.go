package api

import (
	"context"
	"hyperchain/common"
	edb "hyperchain/core/ledger/db_utils"
	flt "hyperchain/manager/filter"
	"sync"
)

// This file defines some public API handler for websocket message.
// If a kind of subscription is created successfully, server will
// push specific data to client.

var (
	subscribedEvents map[context.Context][]Event
	eventsMux        sync.Mutex
)

type Event struct {
	SubId common.ID `json:"subId"`
	Event string    `json:"event"`
}

func init() {
	subscribedEvents = make(map[context.Context][]Event)
}

// Block creates a subscription that sends a notification each time when a new block is appended to the chain.
func (api *PublicFilterAPI) Block(ctx context.Context, isVerbose bool) (common.ID, error) {
	api.log.Debug("ready to deal with newBlock event request")
	return api.handleWSSubscribe(ctx, isVerbose, flt.BlocksSubscription, flt.FilterCriteria{})
}

// SystemStatus creates a subscription that sends a notification each time when system status is changed.
func (api *PublicFilterAPI) SystemStatus(ctx context.Context, crit flt.FilterCriteria) (common.ID, error) {
	api.log.Debug("ready to deal with newException event request")
	return api.handleWSSubscribe(ctx, false, flt.SystemStatusSubscription, crit)
}

// Logs creates a subscription that sends a notification each time when contract event is triggered.
func (api *PublicFilterAPI) Logs(ctx context.Context, crit flt.FilterCriteria) (common.ID, error) {
	api.log.Debug("ready to deal with newLogs event request")
	return api.handleWSSubscribe(ctx, false, flt.LogsSubscription, crit)
}

// GetAllSubscription returns all the event has been subscribed and its subscription id for the connection.
func (api *PublicFilterAPI) GetAllSubscription(ctx context.Context) ([]Event, error) {
	events := subscribedEvents[ctx]
	if len(events) == 0 {
		return []Event{}, nil
	}
	return events, nil
}

func (api *PublicFilterAPI) handleWSSubscribe(ctx context.Context, isVerbose bool, typ flt.Type, crit flt.FilterCriteria) (common.ID, error) {

	eventsMux.Lock()
	defer eventsMux.Unlock()

	common.CtxCh <- ctx
	subChs := common.GetSubChs(ctx)

	select {
	case err := <-subChs.Err:
		return common.ID(""), err
	case rpcSub := <-subChs.SubscriptionCh: // a JSON-RPC subscription is created
		api.log.Debugf("receive subscription %v", rpcSub.ID)

		go func() {

			ch := make(chan interface{})
			sub := api.events.NewCommonSubscription(ch, isVerbose, typ, crit)

			for {
				select {
				case d := <-ch:
					api.log.Debugf("receive data")
					switch typ {
					case flt.BlocksSubscription:
						if isVerbose {
							hash, ok := d.(common.Hash)
							if !ok {
								continue
							}
							block, err := edb.GetBlock(api.namespace, hash.Bytes())
							if err != nil {
								api.log.Errorf("missing block data (#%s)", hash.Hex())
								continue
							} else if wrappedBlock, err := outputBlockResult(api.namespace, block, true); err != nil {
								api.log.Errorf("wrapper block data (#%s) failed", hash.Hex())
								continue
							} else {
								d = wrappedBlock
							}
						}
					case flt.LogsSubscription:
						logs := make([]interface{}, 1)
						logs = append(logs, d)
						d = returnLogs(logs)
					}

					payload := common.NotifyPayload{
						SubID: rpcSub.ID,
						Data:  d,
					}

					subChs.NotifyDataCh <- payload
				case <-rpcSub.Err(): // unsubscribe
					sub.Unsubscribe()
					subscribedEvents[ctx] = deleteEvent(subscribedEvents[ctx], rpcSub.ID)
					return
				case <-subChs.Closed(): // connection close
					api.log.Debug("the websocket connection closed, release resource")
					sub.Unsubscribe()
					delete(subscribedEvents, ctx)
					return
				}
			}
		}()

		// recode subscribed event
		switch typ {
		case flt.BlocksSubscription:
			subscribedEvents[ctx] = append(subscribedEvents[ctx], Event{
				SubId: rpcSub.ID,
				Event: "block",
			})
		case flt.LogsSubscription:
			subscribedEvents[ctx] = append(subscribedEvents[ctx], Event{
				SubId: rpcSub.ID,
				Event: "logs",
			})
		case flt.SystemStatusSubscription:
			subscribedEvents[ctx] = append(subscribedEvents[ctx], Event{
				SubId: rpcSub.ID,
				Event: "systemStatus",
			})
		}

		return rpcSub.ID, nil
	}
}

func deleteEvent(events []Event, id common.ID) []Event {
	eventsMux.Lock()
	defer eventsMux.Unlock()

	es := events
	for i, e := range es {
		if e.SubId == id {
			es = append(es[:i], es[i+1:]...)
		}
	}
	return es
}
