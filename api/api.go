package api

import (
	"fmt"
	"github.com/op/go-logging"
	"context"
	"hyperchain/common"
	"hyperchain/rpc"
)

type Test struct {
	namespace string
	height    uint64
	log       *logging.Logger
}


func NewTestAPI(namespace string) *Test {
	return &Test{
		namespace: namespace,
	}
}

// ===================== test ================
func (blk *Test) TestSub(ctx context.Context) (*common.Subscription, error) {
	notifier, supported := jsonrpc.NotifierFromContext(ctx)
	if !supported {
		return &common.Subscription{}, jsonrpc.ErrNotificationsUnsupported
	}

	rpcSub := notifier.CreateSubscription()
	fmt.Println(rpcSub.ID)

	//go func() {
	//	statuses := make(chan interface{})
	//	sub := api.SubscribeSyncStatus(statuses)
	//
	//	for {
	//		select {
	//		case status := <-statuses:
	//			notifier.Notify(rpcSub.ID, status)
	//		case <-rpcSub.Err():
	//			sub.Unsubscribe()
	//			return
	//		case <-notifier.Closed():
	//			sub.Unsubscribe()
	//			return
	//		}
	//	}
	//}()

	return rpcSub, nil
}

