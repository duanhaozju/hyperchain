//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package controller

import (
	"hyperchain/consensus"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/pbft"
	"hyperchain/event"

	"github.com/op/go-logging"
)

var logger *logging.Logger // package-level logger
func init() {
	logger = logging.MustGetLogger("consensus")
}

// NewConsenter constructs a Consenter object if not already present
func NewConsenter(namespace string, id uint64, msgQ *event.TypeMux, pbftConfigPath string) consensus.Consenter {

	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	h := helper.NewHelper(msgQ)

	return pbft.GetPlugin(namespace, id, h, pbftConfigPath)
}
