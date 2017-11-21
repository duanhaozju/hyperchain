//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package csmgr

import (
	"fmt"

	"github.com/hyperchain/hyperchain/common"
	cs "github.com/hyperchain/hyperchain/consensus"
	"github.com/hyperchain/hyperchain/consensus/helper"
	"github.com/hyperchain/hyperchain/consensus/rbft"
	"github.com/hyperchain/hyperchain/manager/event"
)

// newConsenter initializes the consentor instance, and now only rbft works.
func newConsenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	algo := conf.GetString(cs.CONSENSUS_ALGO)
	switch algo {
	case cs.RBFT:
		h := helper.NewHelper(eventMux, filterMux)
		return rbft.New(namespace, conf, h, n)
	case cs.NBFT:
		panic(fmt.Errorf("Not support yet %s", algo))
	default:
		panic(fmt.Errorf("Invalid consensus alorithm defined: %s", algo))
	}
}

// Consenter returns a Consenter instance in which eventMux and filterMux are the connections with outer services.
func Consenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	return newConsenter(namespace, conf, eventMux, filterMux, n)
}
