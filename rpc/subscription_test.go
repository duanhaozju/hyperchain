package jsonrpc

import (
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/hyperchain/hyperchain/common"
	"context"
)

func TestNewNotifier(t *testing.T) {
	log = common.GetLogger(common.DEFAULT_NAMESPACE, "jsonrpc")
	ctx := context.Background()

	notifier := NewNotifier()
	ctx = context.WithValue(ctx, NotifierKey{}, notifier)
	notifier.subChs = common.GetSubChs(ctx)
	notifier.codec = initCodec(`{"jsonrpc": "2.0", "method": "test_subscribe", "params": ["block"], "id": 1, "namespace":"global"}`)

	// create a new subscription
	sub := notifier.CreateSubscription()
	assert.Equal(t, 1, len(notifier.inactive))

	// active the subscription
	notifier.Activate(sub.ID, "test", "block", "global")
	assert.Equal(t, 0, len(notifier.inactive))
	assert.Equal(t, 1, len(notifier.active))

	// unsubscribe the subscription
	err := notifier.Unsubscribe(sub.ID)
	assert.Nil(t, err)
	assert.Equal(t, 0, len(notifier.active))

	// get the notifier from context
	n, isExist := NotifierFromContext(ctx)
	assert.True(t, isExist)
	assert.NotNil(t, n)

	notifier.Close()
}