package manager

import (
	"github.com/hyperchain/hyperchain/common"
	"github.com/hyperchain/hyperchain/manager/event"
	"testing"
)

func TestEventHubStopRepeatly(t *testing.T) {
	var (
		mux  = new(event.TypeMux)
		sub  = new(event.TypeMux)
		conf = common.NewRawConfig()
	)
	common.InitHyperLogger(common.DEFAULT_NAMESPACE, conf)
	hub := New(common.DEFAULT_NAMESPACE, mux, sub, nil, nil, nil, nil)
	hub.Start()
	hub.Stop()
	hub.Stop()
	hub.Start()
}
