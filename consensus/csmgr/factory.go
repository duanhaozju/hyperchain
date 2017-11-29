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
	"github.com/hyperchain/hyperchain/core/oplog"
	"github.com/hyperchain/hyperchain/core/fiber"
)

// newConsenter initializes the consentor instance, and now only rbft works.
func newConsenter(n int, namespace string, conf *common.Config, opLog oplog.OpLog, eventMux *event.TypeMux, filterMux *event.TypeMux, fiber fiber.Fiber) (cs.Consenter, error) {
	algo := conf.GetString(cs.CONSENSUS_ALGO)
	switch algo {
	case cs.RBFT:
		h := helper.NewHelper(eventMux, filterMux, opLog, fiber)
		return rbft.New(namespace, conf, h, n)
	case cs.NBFT:
		panic(fmt.Errorf("Not support yet %s", algo))
	default:
		panic(fmt.Errorf("Invalid consensus alorithm defined: %s", algo))
	}
}

// Consenter returns a Consenter instance in which eventMux and filterMux are the connections with outer services.
func Consenter(n int, namespace string, conf *common.Config, opLog oplog.OpLog, eventMux *event.TypeMux, filterMux *event.TypeMux, fiber fiber.Fiber) (cs.Consenter, error) {
	return newConsenter(n, namespace, conf, opLog, eventMux, filterMux, fiber)
}
