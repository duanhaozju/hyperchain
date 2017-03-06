//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.

package csmgr

import (
	"hyperchain/common"
	"fmt"

	"hyperchain/consensus/pbft"
	"hyperchain/event"
	"sync"
	cs "hyperchain/consensus"
	"hyperchain/consensus/helper"
)

var cr cs.Consenter
var once sync.Once

//new init the consentor instance.
func newConsenter(conf *common.Config, msgQ *event.TypeMux) cs.Consenter {
	h := helper.NewHelper(msgQ)
	algo := conf.GetString(cs.CONSENSUS_ALGO)
	switch algo {
	case cs.PBFT: return pbft.New(conf, h)
	case cs.NBFT: panic(fmt.Errorf("Not implemented yet %s", algo))
	default:
		panic(fmt.Errorf("Invalid consensus alorithm defined: %s", algo))
	}
	return nil
}

// Consenter return a Consenter instance
// msgQ is the connection with outer services.
func Consenter(conf *common.Config, msgQ *event.TypeMux) cs.Consenter {
	once.Do(func() {
		cr = newConsenter(conf, msgQ)
	})
	return cr
}
