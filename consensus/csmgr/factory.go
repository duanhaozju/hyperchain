//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package csmgr

import (
	"fmt"

	"hyperchain/common"
	cs "hyperchain/consensus"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/rbft"
	"hyperchain/manager/event"
)

// newConsenter initializes the consentor instance, and now only rbft works.
func newConsenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	h := helper.NewHelper(eventMux, filterMux)
	algo := conf.GetString(cs.CONSENSUS_ALGO)
	switch algo {
	case cs.RBFT:
		return rbft.New(namespace, conf, h, n)
	case cs.NBFT:
		panic(fmt.Errorf("Not implemented yet %s", algo))
	default:
		panic(fmt.Errorf("Invalid consensus alorithm defined: %s", algo))
	}
}

// Consenter returns a Consenter instance in which eventMux and filterMux are the connections with outer services.
func Consenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	return newConsenter(namespace, conf, eventMux, filterMux, n)
}
