package jsonrpc

import (
	"bufio"
	"context"
	crand "crypto/rand"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"github.com/hyperchain/hyperchain/common"
	"math/rand"
	"strings"
	"sync"
	"time"
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
	closed   chan interface{}
	subChs   *common.SubChs
}

// newNotifier creates a new notifier that can be used to send subscription
// notifications to the client.
func NewNotifier() *Notifier {
	notifier := &Notifier{
		active:   make(map[common.ID]*common.Subscription),
		inactive: make(map[common.ID]*common.Subscription),
		closed:   make(chan interface{}),
	}

	go notifier.waittingReq()
	return notifier
}

func (notifier *Notifier) waittingReq() {

	for {
		select {
		case ctx := <-common.CtxCh: // receive a request to subscribe
			log.Debug("receive context request")
			notifier, supported := NotifierFromContext(ctx)
			if !supported {
				notifier.subChs.Err <- ErrNotificationsUnsupported
				continue
			}

			rpcSub := notifier.CreateSubscription()
			log.Debugf("create subscription %v", rpcSub.ID)

			notifier.subChs.SubscriptionCh <- rpcSub

			go notifier.waittingSubData(rpcSub)
		case <-notifier.Closed(): // connection closed
			log.Debug("quit request listener of this connection")
			return

		}
	}
}

func (notifier *Notifier) waittingSubData(rpcSub *common.Subscription) {

	for {
		select {
		case nd := <-notifier.subChs.NotifyDataCh:
			// sends a notification to the client with the given data as payload.
			// If an error occurs the RPC connection is closed.
			log.Debugf("ready to send feedback: %#v", nd.SubID)
			id := nd.SubID
			data := nd.Data
			sub, active := notifier.active[id]

			if active {
				notification := notifier.codec.CreateNotification(id, sub.Service, sub.Method, sub.Namespace, data)
				if err := notifier.codec.WriteNotify(notification); err != nil {
					log.Errorf("%v", err)
					notifier.codec.Close()
					//return err
				}
			}
		case <-notifier.subChs.Err:
			// occurs error or connection close
			log.Debugf("quit data listener of subscription %v", rpcSub.ID)
			return
		case <-rpcSub.Err():
			// If the subscription is unsubscribed, quit its data listener.
			log.Debugf("quit data listener of subscription %v", rpcSub.ID)
			return
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

// Unsubscribe unsubscribes a subscription. If the subscription could not be found ErrSubscriptionNotFound is returned.
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

// Close releases notifier resources when connection closed.
func (n *Notifier) Close() {
	close(n.closed)
	n.codec.Close()
	n.subChs.Close()
}

// Closed returns a chnnel that is closed when the websocket connection is closed.
func (n *Notifier) Closed() <-chan interface{} {
	return n.closed
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
