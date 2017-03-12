//Hyperchain License
//Copyright (C) 2016 The Hyperchain Authors.
package controller

import (
	"hyperchain/consensus"
	"hyperchain/consensus/helper"
	"hyperchain/consensus/pbft"
	"hyperchain/event"

	"github.com/op/go-logging"
	"hyperchain/common"
)

var logger *logging.Logger // package-level logger
func init() {
	logger = logging.MustGetLogger("consensus")
}

// NewConsenter constructs a Consenter object if not already present
func NewConsenter(msgQ *event.TypeMux, conf *common.Config) consensus.Consenter {

	pbftConfigPath := conf.GetString(common.PBFT_CONFIG_PATH)
	id := uint64(conf.GetInt(common.C_NODE_ID))
	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	h := helper.NewHelper(msgQ)

	return pbft.GetPlugin(id, h, pbftConfigPath)
}
