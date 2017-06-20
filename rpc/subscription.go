package jsonrpc

import (
	"errors"
	"sync"
	"context"
	"encoding/binary"
	"bufio"
	"time"
	"encoding/hex"
	"strings"
	crand "crypto/rand"
	"math/rand"
	"hyperchain/common"
)

const (
	SubscribeMethodSuffix    = "_subscribe"
	NotificationMethodSuffix = "_subscription"
	UnsubscribeMethodSuffix  = "_unsubscribe"
)

var (
	subscriptionIDGenMu sync.Mutex
	subscriptionIDGen   = idGenerator()
)

var (
	// ErrNotificationsUnsupported is returned when the connection doesn't support notifications
	ErrNotificationsUnsupported = errors.New("notifications not supported")
	// ErrNotificationNotFound is returned when the notification for the given id is not found
	ErrSubscriptionNotFound = errors.New("subscription not found")
)

//// ID defines a pseudo random number that is used to identify RPC subscriptions.
//type ID string
//
//// a Subscription is created by a notifier and tight to that notifier. The client can use
//// this subscription to wait for an unsubscribe request for the client, see Err().
//type Subscription struct {
//	ID        ID
//	Service string
//	Namespace string
//	Err       chan error // closed on unsubscribe
//}
//
//// Err returns a channel that is closed when the client send an unsubscribe request.
//func (s *Subscription) Err() <-chan error {
//	return s.Err
//}

// notifierKey is used to store a notifier within the connection context.
type NotifierKey struct{}

// Notifier is tight to a RPC connection that supports subscriptions.
// Server callbacks use the notifier to send notifications.
type Notifier struct {
	codec    ServerCodec
	subMu    sync.RWMutex // guards active and inactive maps
	stopped  bool
	active   map[common.ID]*common.Subscription
	inactive map[common.ID]*common.Subscription
}

// newNotifier creates a new notifier that can be used to send subscription
// notifications to the client.
func NewNotifier(codec ServerCodec) *Notifier {
//func NewNotifier() *Notifier {
	notifier := &Notifier{
			codec:    codec,
			active:   make(map[common.ID]*common.Subscription),
			inactive: make(map[common.ID]*common.Subscription),
		    }
	go eventloop()
	return notifier
}

func eventloop() {

	for {
		select {
			case ctx := <- common.CtxChan:
				log.Debug("receive context")
				subchan := common.GetSubChan(ctx)
				log.Debugf("current system SubCtxChan length = %v\n", len(common.SubCtxChan))
				notifier, supported := NotifierFromContext(ctx)
				if !supported {
					subchan.Err <- ErrNotificationsUnsupported
					continue
				}

				rpcSub := notifier.CreateSubscription()
				log.Debugf("create subscription %v\n", rpcSub.ID)
				subchan.SubscriptionChan <- rpcSub

				go dataListener(subchan, notifier)
			//case nd := <- common.GetSubChan().NotifyDataChan:
			//	//notifyMux.Lock()
			//	fmt.Printf("ready to send feedback: %#v\n", nd)
			//	id := nd.SubID
			//	data := nd.Data
			//	fmt.Printf("%v\n",notifier)
			//	fmt.Printf("%v\n",len(notifier.active))
			//	sub, active := notifier.active[id]
			//	fmt.Printf("len(n.active) = %v, subID: %v,  active = %v\n", len(notifier.active), id, active)
			//	for  k, v := range notifier.active {
			//		fmt.Printf("k=%v, v=%v\n", k, v)
			//	}
			//	if active {
			//		notification := notifier.codec.CreateNotification(id, sub.Service, sub.Method, sub.Namespace, data)
			//		if err := notifier.codec.WriteNotify(notification); err != nil {
			//			fmt.Errorf("%v",err)
			//			notifier.codec.Close()
			//			//return err
			//		}
			//	}
			//	//notifyMux.Unlock()

		}
	}
}

//func dataListener(ctx context.Context) {
func dataListener(subchan *common.Subchan, notifier *Notifier) {

	for {
		select {
			case nd := <- subchan.NotifyDataChan:
			//notifyMux.Lock()
				log.Debugf("ready to send feedback: %#v\n", nd.SubID)
				id := nd.SubID
				data := nd.Data
				sub, active := notifier.active[id]
				log.Debugf("len(n.active) = %v, subID: %v,  active = %v\n", len(notifier.active), id, active)

				if active {
					notification := notifier.codec.CreateNotification(id, sub.Service, sub.Method, sub.Namespace, data)
					if err := notifier.codec.WriteNotify(notification); err != nil {
						log.Errorf("%v",err)
						notifier.codec.Close()
						//return err
					}
				}
			case id := <-subchan.Unsubscribe:
				log.Debugf("notifier unsubscribe %v \n", id)
				notifier.Unsubscribe(id)
				break
				//common.DelSubChan(ctx) // todo 如果是退订某个事件，不应该删除上下文
			case <-subchan.Err:
				break
		}

	}

}

// NotifierFromContext returns the Notifier value stored in ctx, if any.
func NotifierFromContext(ctx context.Context) (*Notifier, bool) {
	n, ok := ctx.Value(NotifierKey{}).(*Notifier)
	return n, ok
}

// CreateSubscription returns a new subscription that is coupled to the
// RPC connection. By default subscriptions are inactive and notifications
// are dropped until the subscription is marked as active. This is done
// by the RPC server after the subscription ID is send to the client.
func (n *Notifier) CreateSubscription() *common.Subscription {
	s := &common.Subscription{ID: NewID(), Error: make(chan error)}
	n.subMu.Lock()
	n.inactive[s.ID] = s
	n.subMu.Unlock()
	return s
}

// Notify sends a notification to the client with the given data as payload.
// If an error occurs the RPC connection is closed and the error is returned.
//func (n *Notifier) Notify(id common.ID, data interface{}) error {
//func (n *Notifier) Notify() error {
//	n.subMu.RLock()
//	defer n.subMu.RUnlock()
//
//	nd := <- common.GetSubChan().NotifyDataChan
//	id := nd.SubID
//	data := nd.Data
//
//	sub, active := n.active[id]
//	if active {
//		notification := n.codec.CreateNotification(id, sub.Service, sub.Method, sub.Namespace, data)
//		if err := n.codec.Write(notification); err != nil {
//			n.codec.Close()
//			return err
//		}
//	}
//	return nil
//}

// Closed returns a channel that is closed when the RPC connection is closed.
//func (n *Notifier) Closed() <-chan interface{} {
//	return n.codec.Closed()
//}

// unsubscribe a subscription.
// If the subscription could not be found ErrSubscriptionNotFound is returned.
func (n *Notifier) Unsubscribe(id common.ID) error {
	n.subMu.Lock()
	defer n.subMu.Unlock()
	if s, found := n.active[id]; found {
		close(s.Error)
		delete(n.active, id)
		return nil
	}
	return ErrSubscriptionNotFound
}

// activate enables a subscription. Until a subscription is enabled all
// notifications are dropped. This method is called by the RPC server after
// the subscription ID was sent to client. This prevents notifications being
// send to the client before the subscription ID is send to the client.
func (n *Notifier) Activate(id common.ID, service, method, namespace string) {
	n.subMu.Lock()
	defer n.subMu.Unlock()
	if sub, found := n.inactive[id]; found {
		sub.Service = service
		sub.Method = method
		sub.Namespace = namespace
		n.active[id] = sub
		delete(n.inactive, id)
	}
}

// idGenerator helper utility that generates a (pseudo) random sequence of
// bytes that are used to generate identifiers.
func idGenerator() *rand.Rand {
	if seed, err := binary.ReadVarint(bufio.NewReader(crand.Reader)); err == nil {
		return rand.New(rand.NewSource(seed))
	}
	return rand.New(rand.NewSource(int64(time.Now().Nanosecond())))
}

// NewID generates a identifier that can be used as an identifier in the RPC interface.
// e.g. filter and subscription identifier.
func NewID() common.ID {
	subscriptionIDGenMu.Lock()
	defer subscriptionIDGenMu.Unlock()

	id := make([]byte, 16)
	for i := 0; i < len(id); i += 7 {
		val := subscriptionIDGen.Int63()
		for j := 0; i+j < len(id) && j < 7; j++ {
			id[i+j] = byte(val)
			val >>= 8
		}
	}

	rpcId := hex.EncodeToString(id)
	// rpc ID's are RPC quantities, no leading zero's and 0 is 0x0
	rpcId = strings.TrimLeft(rpcId, "0")
	if rpcId == "" {
		rpcId = "0"
	}

	return common.ID("0x" + rpcId)
}
