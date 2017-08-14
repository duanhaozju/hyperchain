//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package csmgr

import (
	"hyperchain/common"
	"fmt"

	"hyperchain/consensus/pbft"
	"hyperchain/manager/event"
	"sync"
	cs "hyperchain/consensus"
	"hyperchain/consensus/helper"
)

var cr cs.Consenter
var once sync.Once
var err error

//new init the consentor instance.
func newConsenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	h := helper.NewHelper(eventMux, filterMux)
	algo := conf.GetString(cs.CONSENSUS_ALGO)
	switch algo {
	case cs.PBFT: return pbft.New(namespace, conf, h, n)
	case cs.NBFT: panic(fmt.Errorf("Not implemented yet %s", algo))
	default:
		panic(fmt.Errorf("Invalid consensus alorithm defined: %s", algo))
	}
}

// Consenter return a Consenter instance
// msgQ is the connection with outer services.
func Consenter(namespace string, conf *common.Config, eventMux *event.TypeMux, filterMux *event.TypeMux, n int) (cs.Consenter, error) {
	//once.Do(func() {
		cr, err = newConsenter(namespace, conf, eventMux, filterMux, n)
	//})
	return cr, err
}
