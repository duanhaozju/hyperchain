package controller

import (
	"hyperchain/event"
	"hyperchain/consensus"
	"hyperchain/consensus/pbft"
	"hyperchain/consensus/helper"

	"github.com/op/go-logging"
)

var logger *logging.Logger // package-level logger

func init() {
	logger = logging.MustGetLogger("consensus/controller")
}

// NewConsenter constructs a Consenter object if not already present
func NewConsenter(id uint64, msgQ *event.TypeMux) consensus.Consenter {

	plugin := "pbft"
	logger.Infof("Creating consensus plugin %s", plugin)
	h := helper.NewHelper(msgQ)
	return pbft.GetPlugin(id, h)
}
