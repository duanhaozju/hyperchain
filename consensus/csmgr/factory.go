//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package csmgr

import (
	"fmt"

	"hyperchain/common"
	cs "hyperchain/consensus"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/pbft"
	"hyperchain/manager/event"
)

//new init the consentor instance.
func newConsenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	h := helper.NewHelper(eventMux, filterMux)
	algo := conf.GetString(cs.CONSENSUS_ALGO)
	switch algo {
	case cs.PBFT:
		return pbft.New(namespace, conf, h, n)
	case cs.NBFT:
		panic(fmt.Errorf("Not implemented yet %s", algo))
	default:
		panic(fmt.Errorf("Invalid consensus alorithm defined: %s", algo))
	}
}

// Consenter return a Consenter instance
// msgQ is the connection with outer services.
func Consenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {

	return newConsenter(namespace, conf, eventMux, filterMux, n)

}
